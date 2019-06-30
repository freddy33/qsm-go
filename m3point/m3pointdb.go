package m3point

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3db"
)

var pointEnv *m3db.QsmEnvironment

const (
	ConnectionDetailsTable = "connection_details"
	TrioDetailsTable       = "trio_details"
)

func init() {
	m3db.AddTableDef(createConnectionDetailsTableDef())
	m3db.AddTableDef(createTrioDetailsTableDef())
}

func createConnectionDetailsTableDef() *m3db.TableDefinition {
	res := m3db.TableDefinition{}
	res.Name = ConnectionDetailsTable
	res.DdlColumns = "(id smallint PRIMARY KEY," +
		" x integer," +
		" y integer," +
		" z integer," +
		" ds bigint)"
	res.InsertStmt = "(id,x,y,z,ds) values ($1,$2,$3,$4,$5)"
	return &res
}

func createTrioDetailsTableDef() *m3db.TableDefinition {
	res := m3db.TableDefinition{}
	res.Name = TrioDetailsTable
	res.DdlColumns = fmt.Sprintf("(id smallint PRIMARY KEY,"+
		" conn1 smallint REFERENCES %s (id),"+
		" conn2 smallint REFERENCES %s (id),"+
		" conn3 smallint REFERENCES %s (id))", ConnectionDetailsTable, ConnectionDetailsTable, ConnectionDetailsTable)
	res.InsertStmt = "(id,conn1,conn2,conn3) values ($1,$2,$3,$4)"
	return &res
}

func GetPointEnv() *m3db.QsmEnvironment {
	if pointEnv == nil {
		pointEnv = m3db.GetDefaultEnvironment()
	}
	return pointEnv
}

/***************************************************************/
// Connection Details Load and Save
/***************************************************************/

func loadConnectionDetails() ([]*ConnectionDetails, map[Point]*ConnectionDetails) {
	res := make([]*ConnectionDetails, 0, 50)
	connMap := make(map[Point]*ConnectionDetails)

	env := GetPointEnv()
	te, err := env.GetOrCreateTableExec(ConnectionDetailsTable)

	if err != nil {
		Log.Fatalf("could not load connection details due to %v", err)
		return res, connMap
	}

	if !te.WasChecked() {
		Log.Fatalf("could not load connection details since table %s was not checked", te.GetTableName())
		return res, connMap
	}
	if te.WasCreated() {
		Log.Fatalf("could not load connection details since table %s was just created", te.GetTableName())
		return res, connMap
	}

	rows, err := env.GetConnection().Query("select id,x,y,z,ds from connection_details")
	if err != nil {
		Log.Fatalf("could not load connection details due to %v", err)
		return res, connMap
	}
	for rows.Next() {
		cd := ConnectionDetails{}
		err = rows.Scan(&cd.Id, &cd.Vector[0], &cd.Vector[1], &cd.Vector[2], &cd.ConnDS)
		if err != nil {
			Log.Errorf("failed to load connection details line %d", len(res))
		}
		res = append(res, &cd)
		connMap[cd.Vector] = &cd
	}
	return res, connMap
}

func saveAllConnectionDetails() (int, error) {
	env := GetPointEnv()
	te, err := env.GetOrCreateTableExec(ConnectionDetailsTable)

	if err != nil {
		return 0, err
	}

	inserted := 0
	if !te.WasChecked() {
		return inserted, m3db.QsmError(fmt.Sprintf("Table execution for %s in env %d was not checked", te.GetTableName(), env.GetId()))
	}

	if te.WasCreated() {
		connections, _ := calculateConnectionDetails()
		if Log.IsDebug() {
			Log.Debugf("Populating table %s with %d elements", te.TableDef.Name, len(connections))
		}
		for _, cd := range connections {
			res, err := te.InsertStmt.Exec(cd.Id, cd.Vector.X(), cd.Vector.Y(), cd.Vector.Z(), cd.ConnDS)

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
				err = m3db.QsmError(fmt.Sprintf("should have receive one result, and got %d", rows))
				Log.Error(err)
				return inserted, err
			}
			inserted++
		}
	} else {
		Log.Debugf("Connection details table already created")
		count, err := env.GetConnection().Query(fmt.Sprintf("select count(*) from %s", te.TableDef.Name))
		if err != nil {
			Log.Error(err)
			return inserted, err
		}
		if !count.Next() {
			err = m3db.QsmError(fmt.Sprintf("counting rows of table %s returned no results", te.TableDef.Name))
			Log.Error(err)
			return inserted, err
		}
		err = count.Scan(&inserted)
		if err != nil {
			Log.Error(err)
			return inserted, err
		}
	}
	return inserted, nil
}

/***************************************************************/
// Trio Details Load and Save
/***************************************************************/

func loadTrioDetails() TrioDetailList {
	res := TrioDetailList(make([]*TrioDetails, 0, 200))

	env := GetPointEnv()
	te, err := env.GetOrCreateTableExec(TrioDetailsTable)

	if err != nil {
		Log.Fatalf("could not load trio details due to %v", err)
		return res
	}

	if !te.WasChecked() {
		Log.Fatalf("could not load trio details since table %s was not checked", te.GetTableName())
		return res
	}
	if te.WasCreated() {
		Log.Fatalf("could not load trio details since table %s was just created", te.GetTableName())
		return res
	}

	rows, err := env.GetConnection().Query("select id,conn1,conn2,conn3 from trio_details")
	if err != nil {
		Log.Fatalf("could not load trio details due to %v", err)
		return res
	}
	for rows.Next() {
		td := TrioDetails{}
		connIds := [3]ConnectionId{}
		err = rows.Scan(&td.id, &connIds[0], &connIds[1], &connIds[2])
		if err != nil {
			Log.Errorf("failed to load connection details line %d", len(res))
		}
		for i, cId := range connIds {
			td.conns[i] = GetConnDetailsById(cId)
		}
		res = append(res, &td)
	}
	return res
}

func saveAllTrioDetails() (int, error) {
	env := GetPointEnv()
	te, err := env.GetOrCreateTableExec(TrioDetailsTable)

	if err != nil {
		return 0, err
	}

	inserted := 0
	if !te.WasChecked() {
		return inserted, m3db.QsmError(fmt.Sprintf("Table execution for %s in env %d was not checked", te.GetTableName(), env.GetId()))
	}

	if te.WasCreated() {
		trios := calculateAllTrioDetails()
		if Log.IsDebug() {
			Log.Debugf("Populating table %s with %d elements", te.TableDef.Name, len(trios))
		}
		for _, td := range trios {
			res, err := te.InsertStmt.Exec(td.id, td.conns[0].Id, td.conns[1].Id, td.conns[2].Id)

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
				Log.Tracef("inserted %v got %d response", *td, rows)
			}
			if rows != int64(1) {
				err = m3db.QsmError(fmt.Sprintf("should have receive one result, and got %d", rows))
				Log.Error(err)
				return inserted, err
			}
			inserted++
		}
	} else {
		Log.Debugf("trio details table already created")
		count, err := env.GetConnection().Query(fmt.Sprintf("select count(*) from %s", te.TableDef.Name))
		if err != nil {
			Log.Error(err)
			return inserted, err
		}
		if !count.Next() {
			err = m3db.QsmError(fmt.Sprintf("counting rows of table %s returned no results", te.TableDef.Name))
			Log.Error(err)
			return inserted, err
		}
		err = count.Scan(&inserted)
		if err != nil {
			Log.Error(err)
			return inserted, err
		}
	}
	return inserted, nil
}

func FillDb() {
	env := GetPointEnv()
	defer m3db.CloseEnv(env)

	n, err := saveAllConnectionDetails()
	if err != nil {
		Log.Fatalf("could not save all connections due to %v", err)
		return
	}
	if Log.IsInfo() {
		Log.Infof("Environment %d has %d connection details", env.GetId(), n)
	}

	// Init from DB
	allConnections, allConnectionsByVector = loadConnectionDetails()
	detailsInitialized = true

	n, err = saveAllTrioDetails()
	if err != nil {
		Log.Fatalf("could not save all trios due to %v", err)
		return
	}
	if Log.IsInfo() {
		Log.Infof("Environment %d has %d trio details", env.GetId(), n)
	}
	allTrioDetails = loadTrioDetails()
}
