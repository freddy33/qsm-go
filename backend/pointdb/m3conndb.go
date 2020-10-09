package pointdb

import (
	"github.com/freddy33/qsm-go/backend/m3db"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3point"
)

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
	res.Insert = "(id,x,y,z,ds) values ($1,$2,$3,$4,$5)"
	res.SelectAll = "select id,x,y,z,ds from %s"
	res.ExpectedCount = 50
	return &res
}

func createTrioDetailsTableDef() *m3db.TableDefinition {
	res := m3db.TableDefinition{}
	res.Name = TrioDetailsTable
	res.DdlColumns = "(id smallint PRIMARY KEY," +
		" conn1 smallint REFERENCES %s (id)," +
		" conn2 smallint REFERENCES %s (id)," +
		" conn3 smallint REFERENCES %s (id))"
	res.DdlColumnsRefs = []string{ConnectionDetailsTable, ConnectionDetailsTable, ConnectionDetailsTable}
	res.Insert = "(id,conn1,conn2,conn3) values ($1,$2,$3,$4)"
	res.SelectAll = "select id, conn1, conn2, conn3 from %s"
	res.ExpectedCount = 200
	return &res
}

/***************************************************************/
// Connection Details Load and Save
/***************************************************************/

func (pointData *ServerPointPackData) loadConnectionDetails() error {
	te := pointData.connDetailsTe
	rows, err := te.SelectAllForLoad()
	if err != nil {
		return err
	}

	res := make([]*m3point.ConnectionDetails, 0, te.TableDef.ExpectedCount)
	connMap := make(map[m3point.Point]*m3point.ConnectionDetails, te.TableDef.ExpectedCount)

	for rows.Next() {
		cd := m3point.ConnectionDetails{}
		err := rows.Scan(&cd.Id, &cd.Vector[0], &cd.Vector[1], &cd.Vector[2], &cd.ConnDS)
		if err != nil {
			return m3util.MakeWrapQsmErrorf(err, "failed to load connection details line %d", len(res))
		} else {
			res = append(res, &cd)
			connMap[cd.Vector] = &cd
		}
	}

	pointData.AllConnections = res
	pointData.AllConnectionsByVector = connMap
	pointData.ConnectionsLoaded = true
	return nil
}

func (pointData *ServerPointPackData) saveAllConnectionDetails() (int, error) {
	te := pointData.connDetailsTe
	inserted, toFill, err := te.GetForSaveAll()
	if err != nil {
		return 0, err
	}
	if toFill {
		connections, _ := pointData.calculateConnectionDetails()
		if Log.IsDebug() {
			Log.Debugf("Populating table %s with %d elements", te.GetFullTableName(), len(connections))
		}
		for _, cd := range connections {
			err := te.Insert(cd.Id, cd.Vector.X(), cd.Vector.Y(), cd.Vector.Z(), cd.ConnDS)
			if err != nil {
				return inserted, err
			} else {
				inserted++
			}
		}
		te.SetFilled()
	}
	return inserted, nil
}
