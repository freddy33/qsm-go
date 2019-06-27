package m3point

import (
	"database/sql"
	"fmt"
	"github.com/freddy33/qsm-go/m3db"
	"sync"
)

type QsmError string

func (qsmError QsmError) Error() string {
	return string(qsmError)
}

var createTableMutex sync.Mutex
var createdTable map[string]bool

func init() {
	createdTable = make(map[string]bool)
}

func saveAllConnectionDetails(envNumber m3db.QsmEnvironment) (int, error) {
	db := m3db.GetConnection(envNumber)
	defer m3db.CloseDb(db)

	tableName, err := checkConnDetailsTable(db)
	if err != nil {
		return 0, err
	}

	stmt, err := db.Prepare(fmt.Sprintf("insert into %s(id,x,y,z,ds) values($1,$2,$3,$4,$5)", tableName))
	if err != nil {
		Log.Error(err)
		return 0, err
	}

	inserted := 0
	for _, cd := range allConnections {
		res, err := stmt.Exec(cd.Id, cd.Vector.X(), cd.Vector.Y(), cd.Vector.Z(), cd.ConnDS)
		if err != nil {
			Log.Error(err)
			return inserted, err
		}
		rows, err := res.RowsAffected()
		if err != nil {
			Log.Error(err)
			return inserted, err
		}
		if Log.IsTrace() {
			Log.Tracef("inserted %v got %d response", *cd, rows)
		}
		if rows != int64(1) {
			err = QsmError(fmt.Sprintf("should have receive one result, and got %d", rows))
			Log.Error(err)
			return inserted, err
		}
		inserted++
	}
	return inserted, nil
}

func checkConnDetailsTable(db *sql.DB) (string, error) {
	createTableMutex.Lock()
	defer createTableMutex.Unlock()

	tableName := "connection_details"
	done, ok := createdTable[tableName]
	if ok && done {
		return tableName, nil
	}
	r, err := db.Exec(fmt.Sprintf("create table if not exists %s (id smallint, x integer, y integer, z integer, ds bigint, constraint %s_pkey primary key (id))", tableName, tableName))
	if err != nil {
		Log.Error(err)
		return "", err
	}
	fmt.Println(r)
	createdTable[tableName] = true
	return tableName, nil
}
