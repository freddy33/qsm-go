package m3db

import (
	"database/sql"
	"fmt"
	"github.com/freddy33/qsm-go/m3util"
)

type TableDefinition struct {
	Name           string
	DdlColumns     string
	DdlColumnsRefs []string
	Insert         string
	SelectAll      string
	Queries        []string
	QueryTableRefs map[int][]string
	ExpectedCount  int

	ErrorFilter func(err error) bool
}

type TableExec struct {
	tableName       string
	checked         bool
	created         bool
	queriesPrepared bool

	env         *QsmDbEnvironment
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

/***************************************************************/
// Global QsmDbEnvironment functions
/***************************************************************/

func (env *QsmDbEnvironment) GetOrCreateTableExec(tableName string) (*TableExec, error) {
	env.createTableMutex.Lock()
	defer env.createTableMutex.Unlock()

	tableExec, ok := env.tableExecs[tableName]
	if ok {
		if Log.IsTrace() {
			Log.Tracef("Table execution for environment %d and tableName '%s' already in map", env.GetId(), tableName)
		}
		if tableExec.checked {
			// Now the table exists
			tableExec.SetFilled()
			return tableExec, nil
		}
		if Log.IsDebug() {
			Log.Debugf("Table execution for environment %d tableName '%s' was not checked! Redoing checks.", env.GetId(), tableName)
		}
	} else {
		if Log.IsDebug() {
			Log.Debugf("Creating table execution for environment %d tableName=%s", env.GetId(), tableName)
		}
		tableExec = new(TableExec)
		tableExec.env = env
		env.tableExecs[tableName] = tableExec
	}

	err := env.CheckSchema()
	if err != nil {
		Log.Error(err)
		return nil, err
	}
	err = tableExec.initForTable(tableName)
	if err != nil {
		Log.Error(err)
		return nil, err
	}
	err = tableExec.fillStmt()
	if err != nil {
		Log.Fatal(err)
		return nil, err
	}

	tableExec.queriesPrepared = false

	return tableExec, nil
}

/***************************************************************/
// QsmWrongCount functions
/***************************************************************/

type QsmWrongCount struct {
	tableName        string
	actual, expected int
}

func (err *QsmWrongCount) Actual() int {
	return err.actual
}

func (err *QsmWrongCount) Error() string {
	return fmt.Sprintf("number of rows in %s is %d and should be %d", err.tableName, err.actual, err.expected)
}

/***************************************************************/
// TableExec functions
/***************************************************************/

func (te *TableExec) GetFullTableName() string {
	return te.env.GetSchemaName() + "." + te.tableName
}

func (te *TableExec) WasCreated() bool {
	return te.created
}

func (te *TableExec) SetFilled() {
	te.created = false
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

func (te *TableExec) IsFiltered(err error) bool {
	return err != nil && te.TableDef.ErrorFilter != nil && te.TableDef.ErrorFilter(err)
}

/*
Return the number of rows in the table, than a bool if the table should be filled and an error if something is wrong
*/
func (te *TableExec) GetForSaveAll() (int, bool, error) {
	if te.WasCreated() {
		return 0, true, nil
	} else {
		Log.Debugf("%s table was already created. Checking number of rows.", te.GetFullTableName())
		var nbRows int
		count, err := te.env.GetConnection().Query(fmt.Sprintf("select count(*) from %s", te.GetFullTableName()))
		if err != nil {
			Log.Error(err)
			return 0, false, err
		}
		if !count.Next() {
			err = m3util.MakeQsmErrorf("counting rows of table %s returned no results", te.GetFullTableName())
			Log.Error(err)
			return 0, false, err
		}
		err = count.Scan(&nbRows)
		if err != nil {
			Log.Error(err)
			return 0, false, err
		}
		if te.TableDef.ExpectedCount > 0 && nbRows != te.TableDef.ExpectedCount {
			if nbRows != 0 {
				// TODO: Delete all before refill. For now error
				return nbRows, false, &QsmWrongCount{tableName: te.GetFullTableName(), actual: nbRows, expected: te.TableDef.ExpectedCount}
			}
			return 0, true, nil
		}
		return nbRows, false, nil
	}
}

func (te *TableExec) SelectAllForLoad() (*sql.Rows, error) {
	if te.TableDef.ExpectedCount > 0 && te.WasCreated() {
		return nil, m3util.MakeQsmErrorf("could not load since table %s was just created", te.GetFullTableName())
	}
	rows, err := te.GetConnection().Query(fmt.Sprintf(te.TableDef.SelectAll, te.GetFullTableName()))
	if err != nil {
		return nil, m3util.MakeWrapQsmErrorf(err, "could not load all rows from %s due to: %v", te.GetFullTableName(), err)
	}
	return rows, nil
}

func (te *TableExec) Insert(args ...interface{}) error {
	res, err := te.InsertStmt.Exec(args...)
	if err != nil {
		return m3util.MakeWrapQsmErrorf(err, "executing insert for table %s with args %v got error %v", te.tableName, args, err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		if te.IsFiltered(err) {
			return err
		}
		return m3util.MakeWrapQsmErrorf(err, "after insert on table %s with args %v extracting rows received error %v", te.tableName, args, err)
	}
	if Log.IsTrace() {
		Log.Tracef("table %s inserted %v got %d response", te.tableName, args, rows)
	}
	if rows != int64(1) {
		err = m3util.MakeQsmErrorf("insert query on table %s should have receive one result, and got %d", te.tableName, rows)
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
		if te.IsFiltered(err) {
			return -1, err
		}
		return -1, m3util.MakeWrapQsmErrorf(err, "inserting on table %s using query row with args %v got error %v", te.tableName, args, err)
	}
	if Log.IsTrace() {
		Log.Tracef("table %s inserted %v got id %d", te.tableName, args, id)
	}
	return id, nil
}

func (te *TableExec) checkQueriesPrepared() {
	if !te.queriesPrepared {
		err := m3util.MakeQsmErrorf("Table exec %q did not prepare queries! Please call PrepareQueries()", te.GetFullTableName())
		Log.Fatal(err)
	}
}

func (te *TableExec) Update(queryId int, args ...interface{}) (int, error) {
	te.checkQueriesPrepared()
	res, err := te.QueriesStmt[queryId].Exec(args...)
	if err != nil {
		return 0, m3util.MakeWrapQsmErrorf(err, "executing update for table %s for query %d with args %v got error '%s'", te.tableName, queryId, args, err.Error())
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return 0, m3util.MakeWrapQsmErrorf(err, "after update on table %s for query %d with args %v extracting rows received error %v", te.tableName, queryId, args, err)
	}
	if Log.IsTrace() {
		Log.Tracef("updated table %s with query %d and args %v got %d response", te.tableName, queryId, args, rows)
	}
	return int(rows), nil
}

func (te *TableExec) Query(queryId int, args ...interface{}) (*sql.Rows, error) {
	te.checkQueriesPrepared()
	rows, err := te.QueriesStmt[queryId].Query(args...)
	if err != nil {
		return nil, m3util.MakeWrapQsmErrorf(err, "executing query %d for table %s with args %v got error %v", queryId, te.tableName, args, err)
	}
	if Log.IsTrace() {
		Log.Tracef("query %d on table %s with args %v got response", queryId, te.tableName, args)
	}
	return rows, nil
}

func (te *TableExec) QueryRow(queryId int, args ...interface{}) *sql.Row {
	te.checkQueriesPrepared()
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

func (te *TableExec) fillStmt() error {
	db := te.env.GetConnection()
	query := fmt.Sprintf("insert into %s "+te.TableDef.Insert, te.GetFullTableName())
	stmt, err := db.Prepare(query)
	if err != nil {
		Log.Fatalf("for table %s preparing insert query with '%s' got error %v", te.tableName, query, err)
		return err
	}
	te.InsertStmt = stmt
	return nil
}

func (te *TableExec) initForTable(tableName string) error {
	te.tableName = tableName
	te.checked = false

	var ok bool
	te.TableDef, ok = tableDefinitions[tableName]
	if !ok {
		return m3util.MakeQsmErrorf("Table definition for %s does not exists", tableName)
	}

	db := te.env.GetConnection()
	if db == nil {
		return m3util.MakeQsmErrorf("Got a nil connection for %d", te.env.GetId())
	}

	schemaName := te.env.GetSchemaName()
	resCheck := db.QueryRow("select 1 from information_schema.tables where table_schema=$1 and table_name=$2", schemaName, tableName)
	var one int
	err := resCheck.Scan(&one)

	fullTableName := te.GetFullTableName()
	var toCreate bool
	if err == nil {
		if one != 1 {
			return m3util.MakeQsmErrorf("checking for table existence of %s in %s returned %d instead of 1", fullTableName, te.env.dbDetails.DbName, one)
		} else {
			Log.Debugf("Table %s exists in %s", fullTableName, te.env.dbDetails.DbName)
		}
		toCreate = false
	} else {
		if err == sql.ErrNoRows {
			toCreate = true
		} else {
			return m3util.MakeWrapQsmErrorf(err, "could not check if table %s exists due to error %v", fullTableName, err)
		}
	}

	if !toCreate {
		if Log.IsDebug() {
			Log.Debugf("Table %s already exists", fullTableName)
		}
		te.created = false
		te.checked = true
		return nil
	}

	if Log.IsDebug() {
		Log.Debugf("Creating table %s", fullTableName)
	}
	params := te.convertTableNames(te.TableDef.DdlColumnsRefs)
	createQuery := fmt.Sprintf("create table %s "+te.TableDef.DdlColumns, params...)
	_, err = db.Exec(createQuery)
	if err != nil {
		return m3util.MakeWrapQsmErrorf(err, "could not create table %s using '%s' due to error %v", fullTableName, createQuery, err)
	}
	if Log.IsDebug() {
		Log.Debugf("Table %s created", fullTableName)
	}
	te.created = true
	te.checked = true
	return nil
}

func (te *TableExec) PrepareQueries() error {
	env := te.env

	env.createTableMutex.Lock()
	defer env.createTableMutex.Unlock()

	teInMap, ok := env.tableExecs[te.tableName]
	if !ok {
		return m3util.MakeQsmErrorf("cannot populate queries of not created table exec %s", te.tableName)
	}
	if teInMap != te {
		return m3util.MakeQsmErrorf("The table exec %s in map %v != %v", te.tableName, teInMap, te)
	}
	if te.queriesPrepared {
		Log.Debugf("table %q already prepared queries", te.GetFullTableName())
		return nil
	}

	nbQueries := len(te.TableDef.Queries)
	if nbQueries > 0 {
		db := te.env.GetConnection()
		te.QueriesStmt = make([]*sql.Stmt, nbQueries)
		for i, queryFormatSql := range te.TableDef.Queries {
			var extraTableNames []string
			if len(te.TableDef.QueryTableRefs) > 0 {
				extraTableNames, ok = te.TableDef.QueryTableRefs[i]
				if !ok {
					extraTableNames = nil
				}
			}
			params := te.convertTableNames(extraTableNames)
			querySql := fmt.Sprintf(queryFormatSql, params...)
			stmt, err := db.Prepare(querySql)
			if err != nil {
				return m3util.MakeWrapQsmErrorf(err, "for table %s preparing query %d with '%s' got error: %s", te.GetFullTableName(), i, querySql, err.Error())
			}
			te.QueriesStmt[i] = stmt
		}
	}

	te.queriesPrepared = true
	return nil
}

func (te *TableExec) convertTableNames(tableNames []string) []interface{} {
	fullTableName := te.GetFullTableName()
	params := make([]interface{}, len(tableNames)+1)
	params[0] = fullTableName
	if tableNames != nil {
		for i, r := range tableNames {
			params[i+1] = te.env.GetSchemaName() + "." + r
		}
	}
	return params
}
