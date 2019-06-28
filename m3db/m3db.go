package m3db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/freddy33/qsm-go/m3util"
	_ "github.com/lib/pq"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

var Log = m3util.NewLogger("m3db", m3util.INFO)

type QsmEnvironment int

const(
	NoEnv QsmEnvironment = iota
	MainEnv
	TempEnv
	TestEnv
	ShellEnv
	ConfEnv = QsmEnvironment(1234)
)

type DbConnDetails struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DbName   string `json:"dbName"`
}

type QsmError string

func (qsmError QsmError) Error() string {
	return string(qsmError)
}

type TableDefinition struct {
	Name       string
	Checked    bool
	Created    bool
	DdlColumns string
	InsertStmt string
	InsertFunc func(stmt *sql.Stmt) (sql.Result, error)
}

type TableExec struct {
	TableDef   *TableDefinition
	Db         *sql.DB
	InsertStmt *sql.Stmt
}

var createTableMutex sync.Mutex
var tableDefinitions map[string]*TableDefinition

func init() {
	tableDefinitions = make(map[string]*TableDefinition)
}

func AddTableDef(tDef *TableDefinition) {
	tableDefinitions[tDef.Name] = tDef
}

func readDbConf(envNumber QsmEnvironment) DbConnDetails {
	confData, err := ioutil.ReadFile(fmt.Sprintf("%s/dbconn%d.json", m3util.GetConfDir(), envNumber))
	if err != nil {
		log.Fatal(err)
	}
	var res DbConnDetails
	err = json.Unmarshal([]byte(confData), &res)
	if err != nil {
		log.Fatal(err)
	}
	return res
}

func CheckOrCreateEnv(envNumber QsmEnvironment) {
	rootDir := m3util.GetGitRootDir()
	m3util.ExitOnError(os.Setenv("QSM_ENV_NUMBER", fmt.Sprintf("%d", envNumber)))
	cmd := exec.Command("bash", filepath.Join(rootDir, "qsm"), "db", "check")
	out, err := cmd.CombinedOutput()
	fmt.Println(string(out))
	m3util.ExitOnError(err)
}

func DropEnv(envNumber QsmEnvironment) {
	rootDir := m3util.GetGitRootDir()
	m3util.ExitOnError(os.Setenv("QSM_ENV_NUMBER", fmt.Sprintf("%d", envNumber)))
	cmd := exec.Command("bash", filepath.Join(rootDir, "qsm"), "db", "drop")
	out, err := cmd.CombinedOutput()
	fmt.Println(string(out))
	m3util.ExitOnError(err)
}

func GetConnection(envNumber QsmEnvironment) *sql.DB {
	connDetails := readDbConf(envNumber)
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		connDetails.Host, connDetails.Port, connDetails.User, connDetails.Password, connDetails.DbName)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func CloseDb(db *sql.DB) {
	m3util.ExitOnError(db.Close())
}

func CloseTableExec(te *TableExec) {
	if te == nil || te.Db == nil {
		return
	}
	err := te.Db.Close()
	if err != nil {
		Log.Error(err)
	}
}

func GetTableExec(envNumber QsmEnvironment, tableName string) (*TableExec, error) {
	tableExec := TableExec{}
	tableExec.Db = GetConnection(envNumber)
	err := tableExec.getOrCreateTable(tableName)
	if err != nil {
		Log.Error(err)
		return nil, err
	}
	err = tableExec.fillStmt()
	if err != nil {
		Log.Error(err)
		return nil, err
	}
	return &tableExec, nil
}


func (te *TableExec) fillStmt() error {
	stmt, err := te.Db.Prepare(fmt.Sprintf("insert into %s "+te.TableDef.InsertStmt, te.TableDef.Name))
	if err != nil {
		Log.Error(err)
		return err
	}
	te.InsertStmt = stmt
	return nil
}

func (te *TableExec) getOrCreateTable(tableName string) error {
	createTableMutex.Lock()
	defer createTableMutex.Unlock()

	var ok bool
	te.TableDef, ok = tableDefinitions[tableName]
	if !ok {
		return QsmError(fmt.Sprintf("Table definition for %s does not exists", tableName))
	}
	if te.TableDef.Checked {
		if Log.IsTrace() {
			Log.Tracef("Table %s already checked", tableName)
		}
		return nil
	}
	checkQuery := fmt.Sprintf("select 1 from information_schema.tables where table_schema='public' and table_name='%s'", tableName)
	resCheck, err := te.Db.Query(checkQuery)
	if err != nil {
		Log.Errorf("could not check if table %s exists using '%s' due to error %v", tableName, checkQuery, err)
		return err
	}
	toCreate := !resCheck.Next()

	if !toCreate {
		if Log.IsDebug() {
			Log.Debugf("Table %s already exists", tableName)
		}
		te.TableDef.Created = false
		te.TableDef.Checked = true
		return nil
	}

	if Log.IsDebug() {
		Log.Debugf("Creating table %s", tableName)
	}
	createQuery := fmt.Sprintf("create table if not exists %s "+te.TableDef.DdlColumns, tableName)
	_, err = te.Db.Exec(createQuery)
	if err != nil {
		Log.Errorf("could not create table %s using '%s' due to error %v", tableName, createQuery, err)
		return err
	}
	if Log.IsDebug() {
		Log.Debugf("Table %s created", tableName)
	}
	te.TableDef.Created = true
	te.TableDef.Checked = true
	return nil
}

