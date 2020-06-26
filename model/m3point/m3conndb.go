package m3point

import (
	"fmt"
	"github.com/freddy33/qsm-go/utils/m3db"
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
	res.SelectAll = "select id,x,y,z,ds from connection_details"
	res.ExpectedCount = 50
	return &res
}

func createTrioDetailsTableDef() *m3db.TableDefinition {
	res := m3db.TableDefinition{}
	res.Name = TrioDetailsTable
	res.DdlColumns = fmt.Sprintf("(id smallint PRIMARY KEY,"+
		" conn1 smallint REFERENCES %s (id),"+
		" conn2 smallint REFERENCES %s (id),"+
		" conn3 smallint REFERENCES %s (id))", ConnectionDetailsTable, ConnectionDetailsTable, ConnectionDetailsTable)
	res.Insert = "(id,conn1,conn2,conn3) values ($1,$2,$3,$4)"
	res.SelectAll = "select id, conn1, conn2, conn3 from trio_details"
	res.ExpectedCount = 200
	return &res
}

/***************************************************************/
// Connection Details Load and Save
/***************************************************************/

func loadConnectionDetails(env *m3db.QsmEnvironment) ([]*ConnectionDetails, map[Point]*ConnectionDetails) {
	te, rows := env.SelectAllForLoad(ConnectionDetailsTable)

	res := make([]*ConnectionDetails, 0, te.TableDef.ExpectedCount)
	connMap := make(map[Point]*ConnectionDetails, te.TableDef.ExpectedCount)

	for rows.Next() {
		cd := ConnectionDetails{}
		err := rows.Scan(&cd.Id, &cd.Vector[0], &cd.Vector[1], &cd.Vector[2], &cd.ConnDS)
		if err != nil {
			Log.Errorf("failed to load connection details line %d", len(res))
		} else {
			res = append(res, &cd)
			connMap[cd.Vector] = &cd
		}
	}
	return res, connMap
}

func (ppd *PointPackData) saveAllConnectionDetails() (int, error) {
	te, inserted, toFill, err := ppd.env.GetForSaveAll(ConnectionDetailsTable)
	if err != nil {
		return 0, err
	}
	if toFill {
		connections, _ := ppd.calculateConnectionDetails()
		if Log.IsDebug() {
			Log.Debugf("Populating table %s with %d elements", te.TableDef.Name, len(connections))
		}
		for _, cd := range connections {
			err := te.Insert(cd.Id, cd.Vector.X(), cd.Vector.Y(), cd.Vector.Z(), cd.ConnDS)
			if err != nil {
				Log.Error(err)
			} else {
				inserted++
			}
		}
	}
	return inserted, nil
}

/***************************************************************/
// trio Details Load and Save
/***************************************************************/

func (ppd *PointPackData) loadTrioDetails() TrioDetailList {
	te, rows := ppd.env.SelectAllForLoad(TrioDetailsTable)

	res := TrioDetailList(make([]*TrioDetails, 0, te.TableDef.ExpectedCount))

	for rows.Next() {
		td := TrioDetails{}
		connIds := [3]ConnectionId{}
		err := rows.Scan(&td.id, &connIds[0], &connIds[1], &connIds[2])
		if err != nil {
			Log.Errorf("failed to load trio details line %d", len(res))
		} else {
			for i, cId := range connIds {
				td.conns[i] = ppd.GetConnDetailsById(cId)
			}
			res = append(res, &td)
		}
	}
	return res
}

func (ppd *PointPackData) saveAllTrioDetails() (int, error) {
	te, inserted, toFill, err := ppd.env.GetForSaveAll(TrioDetailsTable)
	if te == nil {
		return 0, err
	}

	if toFill {
		trios := ppd.calculateAllTrioDetails()
		if Log.IsDebug() {
			Log.Debugf("Populating table %s with %d elements", te.TableDef.Name, len(trios))
		}
		for _, td := range trios {
			err := te.Insert(td.id, td.conns[0].Id, td.conns[1].Id, td.conns[2].Id)
			if err != nil {
				Log.Error(err)
			} else {
				inserted++
			}
		}
	}
	return inserted, nil
}
