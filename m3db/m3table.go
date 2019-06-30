package m3db

import (
	"database/sql"
	"fmt"
)

type TableDefinition struct {
	Name       string
	DdlColumns string
	InsertStmt string
}

type TableExec struct {
	envId     QsmEnvID
	tableName string
	checked   bool
	created   bool

	env        *QsmEnvironment
	TableDef   *TableDefinition
	InsertStmt *sql.Stmt
}

var tableDefinitions map[string]*TableDefinition

func init() {
	tableDefinitions = make(map[string]*TableDefinition)
}

func AddTableDef(tDef *TableDefinition) {
	tableDefinitions[tDef.Name] = tDef
}

func (env *QsmEnvironment) GetOrCreateTableExec(tableName string) (*TableExec, error) {
	env.createTableMutex.Lock()
	defer env.createTableMutex.Unlock()

	tableExec, ok := env.tableExecs[tableName]
	if ok {
		if Log.IsTrace() {
			Log.Tracef("Table execution for environment %d and tableName '%s' already in map", env.id, tableName)
		}
		if tableExec.checked {
			// Now the table exists
			tableExec.created = false
			return tableExec, nil
		}
		if Log.IsDebug() {
			Log.Debugf("Table execution for environment %d tableName '%s' was not checked! Redoing checks.", env.id, tableName)
		}
	} else {
		if Log.IsDebug() {
			Log.Debugf("Creating table execution for environment %d tableName=%s", env.id, tableName)
		}
		tableExec = new(TableExec)
		tableExec.envId = env.id
		tableExec.env = env
	}

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
	return tableExec, nil
}

func (te *TableExec) GetTableName() string {
	return te.tableName
}

func (te *TableExec) WasCreated() bool {
	return te.created
}

func (te *TableExec) WasChecked() bool {
	return te.checked
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
	te.tableName = tableName
	te.checked = false

	var ok bool
	te.TableDef, ok = tableDefinitions[tableName]
	if !ok {
		return QsmError(fmt.Sprintf("Table definition for %s does not exists", tableName))
	}

	db := te.env.GetConnection()
	if db == nil {
		return QsmError(fmt.Sprintf("Got a nil connection for %d", te.env.id))
	}

	resCheck := db.QueryRow("select 1 from information_schema.tables where table_schema='public' and table_name=$1", tableName)
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
		te.created = false
		te.checked = true
		return nil
	}

	if Log.IsDebug() {
		Log.Debugf("Creating table %s", tableName)
	}
	createQuery := fmt.Sprintf("create table %s "+te.TableDef.DdlColumns, tableName)
	_, err = db.Exec(createQuery)
	if err != nil {
		Log.Errorf("could not create table %s using '%s' due to error %v", tableName, createQuery, err)
		return err
	}
	if Log.IsDebug() {
		Log.Debugf("Table %s created", tableName)
	}
	te.created = true
	te.checked = true
	return nil
}
