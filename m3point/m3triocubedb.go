package m3point

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3db"
)

const (
	TrioCubesTable = "trio_cubes"
)

func init() {
	m3db.AddTableDef(createContextCubesTableDef())
}

func createContextCubesTableDef() *m3db.TableDefinition {
	res := m3db.TableDefinition{}
	res.Name = TrioCubesTable
	res.DdlColumns = fmt.Sprintf("(id smallint PRIMARY KEY,"+
		" ctx_id smallint REFERENCES %s (id),"+
		" center smallint REFERENCES %s (id),"+
		" center_faces_PX smallint REFERENCES %s (id), center_faces_MX smallint REFERENCES %s (id),"+
		" center_faces_PY smallint REFERENCES %s (id), center_faces_MY smallint REFERENCES %s (id),"+
		" center_faces_PZ smallint REFERENCES %s (id), center_faces_MZ smallint REFERENCES %s (id),"+
		" middle_edges_PXPY smallint REFERENCES %s (id), middle_edges_PXMY smallint REFERENCES %s (id), middle_edges_PXPZ smallint REFERENCES %s (id), middle_edges_PXMZ smallint REFERENCES %s (id),"+
		" middle_edges_MXPY smallint REFERENCES %s (id), middle_edges_MXMY smallint REFERENCES %s (id), middle_edges_MXPZ smallint REFERENCES %s (id), middle_edges_MXMZ smallint REFERENCES %s (id),"+
		" middle_edges_PYPZ smallint REFERENCES %s (id), middle_edges_PYMZ smallint REFERENCES %s (id), middle_edges_MYPZ smallint REFERENCES %s (id), middle_edges_MYMZ smallint REFERENCES %s (id))",
		GrowthContextsTable,
		TrioDetailsTable,
		TrioDetailsTable, TrioDetailsTable,
		TrioDetailsTable, TrioDetailsTable,
		TrioDetailsTable, TrioDetailsTable,
		TrioDetailsTable, TrioDetailsTable, TrioDetailsTable, TrioDetailsTable,
		TrioDetailsTable, TrioDetailsTable, TrioDetailsTable, TrioDetailsTable,
		TrioDetailsTable, TrioDetailsTable, TrioDetailsTable, TrioDetailsTable)
	res.Insert = "(id, ctx_id, center," +
		" center_faces_PX, center_faces_MX, center_faces_PY, center_faces_MY, center_faces_PZ, center_faces_MZ, " +
		" middle_edges_PXPY, middle_edges_PXMY, middle_edges_PXPZ, middle_edges_PXMZ, " +
		" middle_edges_MXPY, middle_edges_MXMY, middle_edges_MXPZ, middle_edges_MXMZ, " +
		" middle_edges_PYPZ, middle_edges_PYMZ, middle_edges_MYPZ, middle_edges_MYMZ)" +
		" values ($1,$2,$3," +
		" $4,$5,$6,$7,$8,$9," +
		" $10,$11,$12,$13," +
		" $14,$15,$16,$17," +
		" $18,$19,$20,$21)"
	res.SelectAll = fmt.Sprintf("select id, ctx_id, center,"+
		" center_faces_PX, center_faces_MX, center_faces_PY, center_faces_MY, center_faces_PZ, center_faces_MZ, "+
		" middle_edges_PXPY, middle_edges_PXMY, middle_edges_PXPZ, middle_edges_PXMZ, "+
		" middle_edges_MXPY, middle_edges_MXMY, middle_edges_MXPZ, middle_edges_MXMZ, "+
		" middle_edges_PYPZ, middle_edges_PYMZ, middle_edges_MYPZ, middle_edges_MYMZ"+
		" from %s", TrioCubesTable)
	res.ExpectedCount = TotalNumberOfCubes
	return &res
}

/***************************************************************/
// Cubes Load and Save
/***************************************************************/

func loadContextCubes() map[CubeKeyId]int {
	te, rows := GetPointEnv().SelectAllForLoad(TrioCubesTable)
	res := make(map[CubeKeyId]int, te.TableDef.ExpectedCount)

	loaded := 0
	for rows.Next() {
		var cubeId int
		var trCtxId int
		cube := CubeOfTrioIndex{}
		err := rows.Scan(&cubeId, &trCtxId, &cube.center,
			&cube.centerFaces[0], &cube.centerFaces[1], &cube.centerFaces[2], &cube.centerFaces[3], &cube.centerFaces[4], &cube.centerFaces[5],
			&cube.middleEdges[0], &cube.middleEdges[1], &cube.middleEdges[2], &cube.middleEdges[3],
			&cube.middleEdges[4], &cube.middleEdges[5], &cube.middleEdges[6], &cube.middleEdges[7],
			&cube.middleEdges[8], &cube.middleEdges[9], &cube.middleEdges[10], &cube.middleEdges[11])
		if err != nil {
			Log.Errorf("failed to load trio context line %d due to %v", loaded, err)
		} else {
			key := CubeKeyId{trCtxId,cube}
			res[key] = cubeId
		}
		loaded++
	}
	return res
}

func saveAllContextCubes() (int, error) {
	te, inserted, err := GetPointEnv().GetForSaveAll(TrioCubesTable)
	if err != nil {
		return 0, err
	}
	if te.WasCreated() {
		cubeKeys := calculateAllContextCubes()
		if Log.IsDebug() {
			Log.Debugf("Populating table %s with %d elements", te.TableDef.Name, len(cubeKeys))
		}
		for cubeKey, cubeId := range cubeKeys {
			cube := cubeKey.cube
			err := te.Insert(cubeId, cubeKey.trCtxId, cube.center,
				cube.centerFaces[0], cube.centerFaces[1], cube.centerFaces[2], cube.centerFaces[3], cube.centerFaces[4], cube.centerFaces[5],
				cube.middleEdges[0], cube.middleEdges[1], cube.middleEdges[2], cube.middleEdges[3],
				cube.middleEdges[4], cube.middleEdges[5], cube.middleEdges[6], cube.middleEdges[7],
				cube.middleEdges[8], cube.middleEdges[9], cube.middleEdges[10], cube.middleEdges[11])
			if err != nil {
				Log.Error(err)
			} else {
				inserted++
			}
		}
	}
	return inserted, nil
}
