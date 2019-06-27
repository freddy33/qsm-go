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

