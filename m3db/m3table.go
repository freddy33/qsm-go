package m3db

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/freddy33/qsm-go/m3util"
	"sync"
)

type TableDefinition struct {
	Name       string
	Checked    bool
	Created    bool
	DdlColumns string
	InsertStmt string
	InsertFunc func(stmt *sql.Stmt) (sql.Result, error)
}

type TableExec struct {
	env        *QsmEnvironment
	TableDef   *TableDefinition
	InsertStmt *sql.Stmt
}

var createTableMutex sync.Mutex
var tableDefinitions map[string]*TableDefinition

var ctx context.Context

func init() {
	tableDefinitions = make(map[string]*TableDefinition)
	ctx = context.TODO()
}

func AddTableDef(tDef *TableDefinition) {
	tableDefinitions[tDef.Name] = tDef
}

func (env *QsmEnvironment) CreateTableExec(tableName string) (*TableExec, error) {
	if Log.IsDebug() {
		Log.Debugf("Creating table execution for environment %d tableName=%s", env.id, tableName)
	}
	tableExec := TableExec{}
	tableExec.env = env
	err := tableExec.initForTable(tableName)
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

func CloseTableExec(te *TableExec) {
	if te == nil {
		Log.Warn("Closing nil Table Exec")
		return
	}
	m3util.ExitOnError(te.Close())
}

func (te *TableExec) Close() error {
	if te.InsertStmt != nil {
		return te.InsertStmt.Close()
	}
	return nil
}

func (te *TableExec) fillStmt() error {
	db := te.env.GetConnection()
	stmt, err := db.Prepare(fmt.Sprintf("insert into %s "+te.TableDef.InsertStmt, te.TableDef.Name))
	if err != nil {
		Log.Error(err)
		return err
	}
	te.InsertStmt = stmt
	return nil
}

func closeTxQuietly(tx *sql.Tx) {
	if tx != nil {
		err := tx.Rollback()
		if err != nil {
			Log.Errorf("Rollback threw %v", err)
		}
	}
}

func (te *TableExec) initForTable(tableName string) error {
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

	db := te.env.GetConnection()
	if db == nil {
		return QsmError(fmt.Sprintf("Got a nil connection for %d", te.env.id))
	}

	resCheck := db.QueryRowContext(ctx, "select 1 from information_schema.tables where table_schema='public' and table_name=$1", tableName)
	var one int
	err := resCheck.Scan(&one)

	var toCreate bool
	if err == nil {
		if one != 1 {
			Log.Errorf("checking for table existence of %s in %s returned %d instead of 1", tableName, te.env.dbDetails.DbName, one)
		} else {
			Log.Debugf("Table %s exists in %s", tableName, te.env.dbDetails.DbName)
		}
		toCreate = false
	} else {
		if err == sql.ErrNoRows {
			toCreate = true
		} else {
			Log.Errorf("could not check if table %s exists due to error %v", tableName, err)
			return err
		}
	}

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
	_, err = db.Exec(createQuery)
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
