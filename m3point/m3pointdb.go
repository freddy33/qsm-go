package m3point

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3db"
)


const (
	CONNECTION_DETAILS_TABLE = "connection_details"
)

func init() {
	m3db.AddTableDef(createConnectionDetailsTableDef())
}

func createConnectionDetailsTableDef() *m3db.TableDefinition {
	res := m3db.TableDefinition{}
	res.Name = CONNECTION_DETAILS_TABLE
	res.Created = false
	res.DdlColumns = fmt.Sprintf("(id smallint, x integer, y integer, z integer, ds bigint, constraint %s_pkey primary key (id))", res.Name)
	res.InsertStmt = "(id,x,y,z,ds) values ($1,$2,$3,$4,$5)"
	return &res
}

func saveAllConnectionDetails(envNumber m3db.QsmEnvironment) (int, error) {
	te, err := m3db.GetTableExec(envNumber, CONNECTION_DETAILS_TABLE)
	defer m3db.CloseTableExec(te)
	if err != nil {
		return 0, err
	}

	inserted := 0
	if te.TableDef.Created {
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
	}
	return inserted, nil
}

