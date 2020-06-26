package m3db

import (
	"database/sql"
	"fmt"
)

type TableDefinition struct {
	Name          string
	DdlColumns    string
	Insert        string
	SelectAll     string
	Queries       []string
	ExpectedCount int

	ErrorFilter func(err error) bool
}

type TableExec struct {
	envId     QsmEnvID
	tableName string
	checked   bool
	created   bool

	env         *QsmEnvironment
	TableDef    *TableDefinition
	InsertStmt  *sql.Stmt
	QueriesStmt []*sql.Stmt
}

var tableDefinitions map[string]*TableDefinition

func init() {
	tableDefinitions = make(map[string]*TableDefinition)
}

func AddTableDef(tDef *TableDefinition) {
	tableDefinitions[tDef.Name] = tDef
}

func (env *QsmEnvironment) SelectAllForLoad(tableName string) (*TableExec, *sql.Rows) {
	te, err := env.GetOrCreateTableExec(tableName)
	if err != nil {
		Log.Fatalf("could not load due to error while getting table exec %v", err)
		return nil, nil
	}
	if te.WasCreated() {
		Log.Fatalf("could not load since table %s was just created", te.GetTableName())
		return nil, nil
	}
	rows, err := te.GetConnection().Query(te.TableDef.SelectAll)
	if err != nil {
		Log.Fatalf("could not load due to error while select all %v", err)
		return nil, nil
	}
	return te, rows
}

func (env *QsmEnvironment) GetForSaveAll(tableName string) (*TableExec, int, bool, error) {
	te, err := env.GetOrCreateTableExec(tableName)
	if err != nil {
		return te, 0, false, err
	}
	if te.WasCreated() {
		return te, 0, true, nil
	} else {
		Log.Debugf("%s table was already created. Checking number of rows.", tableName)
		var nbRows int
		count, err := env.GetConnection().Query(fmt.Sprintf("select count(*) from %s", te.TableDef.Name))
		if err != nil {
			Log.Error(err)
			return te, 0, false, err
		}
		if !count.Next() {
			err = MakeQsmErrorf("counting rows of table %s returned no results", te.TableDef.Name)
			Log.Error(err)
			return te, 0, false, err
		}
		err = count.Scan(&nbRows)
		if err != nil {
			Log.Error(err)
			return te, 0, false, err
		}
		if te.TableDef.ExpectedCount > 0 && nbRows != te.TableDef.ExpectedCount {
			if nbRows != 0 {
				// TODO: Delete all before refill. For now error
				return te, nbRows, false, &QsmWrongCount{tableName, nbRows, te.TableDef.ExpectedCount}
			}
			return te, 0, true, nil
		}
		return te, nbRows, false, nil
	}
}

type QsmWrongCount struct {
	tableName string
	actual, expected int
}

func (err *QsmWrongCount) Actual() int {
	return err.actual
}

func (err *QsmWrongCount) Error() string {
	return fmt.Sprintf("number of rows in %s is %d and should be %d", err.tableName, err.actual, err.expected)
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

	nbQueries := len(tableExec.TableDef.Queries)
	if nbQueries > 0 {
		tableExec.QueriesStmt = make([]*sql.Stmt, nbQueries)
		for i, query := range tableExec.TableDef.Queries {
			err = tableExec.fillQuery(i, query)
			if err != nil {
				Log.Error(err)
				return nil, err
			}
		}
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

func (te *TableExec) GetConnection() *sql.DB {
	return te.env.db
}

func (te *TableExec) Close() error {
	if te.InsertStmt != nil {
		return te.InsertStmt.Close()
	}
	return nil
}

func (te *TableExec) fillStmt() error {
	db := te.env.GetConnection()
	query := fmt.Sprintf("insert into %s "+te.TableDef.Insert, te.TableDef.Name)
	stmt, err := db.Prepare(query)
	if err != nil {
		Log.Errorf("for table %s preparing insert query with '%s' got error %v", te.tableName, query, err)
		return err
	}
	te.InsertStmt = stmt
	return nil
}

func (te *TableExec) fillQuery(i int, query string) error {
	db := te.env.GetConnection()
	stmt, err := db.Prepare(query)
	if err != nil {
		Log.Errorf("for table %s preparing query %d with '%s' got error %v", te.tableName, i, query, err)
		return err
	}
	te.QueriesStmt[i] = stmt
	return nil
}

func (te *TableExec) IsFiltered(err error) bool {
	return err != nil && te.TableDef.ErrorFilter != nil && te.TableDef.ErrorFilter(err)
}

func (te *TableExec) Insert(args ...interface{}) error {
	res, err := te.InsertStmt.Exec(args...)
	if err != nil {
		Log.Errorf("executing insert for table %s with args %v got error %v", te.tableName, args, err)
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		if !te.IsFiltered(err) {
			Log.Errorf("after insert on table %s with args %v extracting rows received error '%s'", te.tableName, args, err.Error())
		}
		return err
	}
	if Log.IsTrace() {
		Log.Tracef("table %s inserted %v got %d response", te.tableName, args, rows)
	}
	if rows != int64(1) {
		err = MakeQsmErrorf("insert query on table %s should have receive one result, and got %d", te.tableName, rows)
		if !te.IsFiltered(err) {
			Log.Error(err)
		}
		return err
	}
	return nil
}

func (te *TableExec) InsertReturnId(args ...interface{}) (int64, error) {
	row := te.InsertStmt.QueryRow(args...)
	var id int64
	err := row.Scan(&id)
	if err != nil {
		if !te.IsFiltered(err) {
			Log.Errorf("inserting on table %s using query row with args %v got error '%s'", te.tableName, args, err.Error())
		}
		return -1, err
	}
	if Log.IsTrace() {
		Log.Tracef("table %s inserted %v got id %d", te.tableName, args, id)
	}
	return id, nil
}

func (te *TableExec) Update(queryId int, args ...interface{}) (int, error) {
	res, err := te.QueriesStmt[queryId].Exec(args...)
	if err != nil {
		Log.Errorf("executing update for table %s for query %d with args %v got error '%s'", te.tableName, queryId, args, err.Error())
		return 0, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		Log.Errorf("after update on table %s for query %d with args %v extracting rows received error %v", te.tableName, queryId, args, err)
		return 0, err
	}
	if Log.IsTrace() {
		Log.Tracef("updated table %s with query %d and args %v got %d response", te.tableName, queryId, args, rows)
	}
	return int(rows), nil
}

func (te *TableExec) Query(queryId int, args ...interface{}) (*sql.Rows, error) {
	rows, err := te.QueriesStmt[queryId].Query(args...)
	if err != nil {
		Log.Errorf("executing query %d for table %s with args %v got error %v", queryId, te.tableName, args, err)
		return nil, err
	}
	if Log.IsTrace() {
		Log.Tracef("query %d on table %s with args %v got response", queryId, te.tableName, args)
	}
	return rows, nil
}

func (te *TableExec) QueryRow(queryId int, args ...interface{}) *sql.Row {
	row := te.QueriesStmt[queryId].QueryRow(args...)
	if Log.IsTrace() {
		Log.Tracef("query row %d on table %s with args %v", queryId, te.tableName, args)
	}
	return row
}

func (te *TableExec) CloseRows(rows *sql.Rows) {
	err := rows.Close()
	if err != nil {
		Log.Errorf("error closing %s result set %v", te.tableName, err)
	}
}

func (te *TableExec) initForTable(tableName string) error {
	te.tableName = tableName
	te.checked = false

	var ok bool
	te.TableDef, ok = tableDefinitions[tableName]
	if !ok {
		return MakeQsmErrorf("Table definition for %s does not exists", tableName)
	}

	db := te.env.GetConnection()
	if db == nil {
		return MakeQsmErrorf("Got a nil connection for %d", te.env.id)
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
