package m3point

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3db"
)

var pointEnv *m3db.QsmEnvironment

const (
	ConnectionDetailsTable = "connection_details"
)

func init() {
	m3db.AddTableDef(createConnectionDetailsTableDef())
}

func createConnectionDetailsTableDef() *m3db.TableDefinition {
	res := m3db.TableDefinition{}
	res.Name = ConnectionDetailsTable
	res.DdlColumns = fmt.Sprintf("(id smallint, x integer, y integer, z integer, ds bigint, constraint %s_pkey primary key (id))", res.Name)
	res.InsertStmt = "(id,x,y,z,ds) values ($1,$2,$3,$4,$5)"
	return &res
}

func GetPointEnv() *m3db.QsmEnvironment {
	if pointEnv == nil {
		pointEnv = m3db.GetDefaultEnvironment()
	}
	return pointEnv
}

func saveAllConnectionDetails() (int, error) {
	env := GetPointEnv()
	te, err := env.GetOrCreateTableExec(ConnectionDetailsTable)
	defer m3db.CloseTableExec(te)
	if err != nil {
		return 0, err
	}

	inserted := 0
	if !te.WasChecked() {
		return inserted, m3db.QsmError(fmt.Sprintf("Table execution for %s in env %d was not checked", te.GetTableName(), env.GetId()))
	}

	if te.WasCreated() {
		if Log.IsDebug() {
			Log.Debugf("Populating table %s with %d elements", te.TableDef.Name, len(allConnections))
		}
		for _, cd := range allConnections {
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
}