package m3point

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3db"
)

const (
	TrioContextsTable = "trio_contexts"
	ContextCubesTable = "context_cubes"
)

func init() {
	m3db.AddTableDef(createTrioContextTableDef())
	m3db.AddTableDef(createContextCubesTableDef())
}

func createTrioContextTableDef() *m3db.TableDefinition {
	res := m3db.TableDefinition{}
	res.Name = TrioContextsTable
	res.DdlColumns = "(id smallint PRIMARY KEY," +
		" ctx_type smallint," +
		" ctx_index smallint, UNIQUE (ctx_type, ctx_index) )"
	res.InsertStmt = "(id, ctx_type, ctx_index) values ($1,$2,$3)"
	res.SelectAll = "select id, ctx_type, ctx_index from trio_contexts"
	res.ExpectedCount = 52
	return &res
}

func createContextCubesTableDef() *m3db.TableDefinition {
	res := m3db.TableDefinition{}
	res.Name = ContextCubesTable
	res.DdlColumns = fmt.Sprintf("(id serial PRIMARY KEY," +
		" ctx_id smallint REFERENCES %s (id)," +
		" center smallint REFERENCES %s (id)," +
		" center_faces_PX smallint REFERENCES %s (id), center_faces_MX smallint REFERENCES %s (id)," +
		" center_faces_PY smallint REFERENCES %s (id), center_faces_MY smallint REFERENCES %s (id)," +
		" center_faces_PZ smallint REFERENCES %s (id), center_faces_MZ smallint REFERENCES %s (id)," +
		" middle_edges_PXPY smallint REFERENCES %s (id), middle_edges_PXMY smallint REFERENCES %s (id), middle_edges_PXPZ smallint REFERENCES %s (id), middle_edges_PXMZ smallint REFERENCES %s (id)," +
		" middle_edges_MXPY smallint REFERENCES %s (id), middle_edges_MXMY smallint REFERENCES %s (id), middle_edges_MXPZ smallint REFERENCES %s (id), middle_edges_MXMZ smallint REFERENCES %s (id)," +
		" middle_edges_PYPZ smallint REFERENCES %s (id), middle_edges_PYMZ smallint REFERENCES %s (id), middle_edges_MYPZ smallint REFERENCES %s (id), middle_edges_MYMZ smallint REFERENCES %s (id))",
		TrioContextsTable,
		TrioDetailsTable,
		TrioDetailsTable, TrioDetailsTable,
		TrioDetailsTable, TrioDetailsTable,
		TrioDetailsTable, TrioDetailsTable,
		TrioDetailsTable, TrioDetailsTable, TrioDetailsTable, TrioDetailsTable,
		TrioDetailsTable, TrioDetailsTable, TrioDetailsTable, TrioDetailsTable,
		TrioDetailsTable, TrioDetailsTable, TrioDetailsTable, TrioDetailsTable)
	res.InsertStmt = "(ctx_id, center," +
		" center_faces_PX, center_faces_MX, center_faces_PY, center_faces_MY, center_faces_PZ, center_faces_MZ, " +
		" middle_edges_PXPY, middle_edges_PXMY, middle_edges_PXPZ, middle_edges_PXMZ, " +
		" middle_edges_MXPY, middle_edges_MXMY, middle_edges_MXPZ, middle_edges_MXMZ, " +
		" middle_edges_PYPZ, middle_edges_PYMZ, middle_edges_MYPZ, middle_edges_MYMZ)" +
		" values ($1,$2," +
		" $3,$4,$5,$6,$7,$8," +
		" $9,$10,$11,$12," +
		" $13,$14,$15,$16," +
		" $17,$18,$19,$20)"
	res.SelectAll = fmt.Sprintf("select ctx_id, center," +
		" center_faces_PX, center_faces_MX, center_faces_PY, center_faces_MY, center_faces_PZ, center_faces_MZ, " +
		" middle_edges_PXPY, middle_edges_PXMY, middle_edges_PXPZ, middle_edges_PXMZ, " +
		" middle_edges_MXPY, middle_edges_MXMY, middle_edges_MXPZ, middle_edges_MXMZ, " +
		" middle_edges_PYPZ, middle_edges_PYMZ, middle_edges_MYPZ, middle_edges_MYMZ" +
		" from %s", ContextCubesTable)
	res.ExpectedCount = 5192
	return &res
}

/***************************************************************/
// Trio Contexts Load and Save
/***************************************************************/

func loadTrioContexts() []*TrioContext {
	te, rows := GetPointEnv().SelectAllForLoad(TrioContextsTable)
	res := make([]*TrioContext, 0, te.TableDef.ExpectedCount)

	for rows.Next() {
		trCtx := TrioContext{}
		err := rows.Scan(&trCtx.id, &trCtx.ctxType, &trCtx.ctxIndex)
		if err != nil {
			Log.Errorf("failed to load trio context line %d", len(res))
		} else {
			res = append(res, &trCtx)
		}
	}
	return res
}

func saveAllTrioContexts() (int, error) {
	te, inserted, err := GetPointEnv().GetForSaveAll(TrioContextsTable)
	if err != nil {
		return 0, err
	}
	if te.WasCreated() {
		trCtxs := calculateAllTrioContexts()
		if Log.IsDebug() {
			Log.Debugf("Populating table %s with %d elements", te.TableDef.Name, len(trCtxs))
		}
		for _, trCtx := range trCtxs {
			err := te.Insert(trCtx.id, trCtx.ctxType, trCtx.ctxIndex)
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
// Cubes Load and Save
/***************************************************************/

func loadContextCubes() []*CubeListPerContext {
	_, rows := GetPointEnv().SelectAllForLoad(ContextCubesTable)
	res := make([]*CubeListPerContext, GetTotalNbTrioContexts())

	loaded := 0
	for rows.Next() {
		var trCtxId int
		cubeKey := CubeKey{}
		err := rows.Scan(&trCtxId, &cubeKey.center,
			&cubeKey.centerFaces[0], &cubeKey.centerFaces[1], &cubeKey.centerFaces[2], &cubeKey.centerFaces[3], &cubeKey.centerFaces[4], &cubeKey.centerFaces[5],
			&cubeKey.middleEdges[0], &cubeKey.middleEdges[1], &cubeKey.middleEdges[2], &cubeKey.middleEdges[3],
			&cubeKey.middleEdges[4], &cubeKey.middleEdges[5], &cubeKey.middleEdges[6], &cubeKey.middleEdges[7],
			&cubeKey.middleEdges[8], &cubeKey.middleEdges[9], &cubeKey.middleEdges[10], &cubeKey.middleEdges[11])
		if err != nil {
			Log.Errorf("failed to load trio context line %d due to %v", loaded, err)
		} else {
			if res[trCtxId] == nil {
				cl := new(CubeListPerContext)
				cl.trCtx = GetTrioContextById(trCtxId)
				cl.allCubes = make([]CubeKey, 1, 15)
				cl.allCubes[0] = cubeKey
				res[trCtxId] = cl
			} else {
				cl := res[trCtxId]
				cl.allCubes = append(cl.allCubes, cubeKey)
			}
		}
		loaded++
	}
	return res
}

func saveAllContextCubes() (int, error) {
	te, inserted, err := GetPointEnv().GetForSaveAll(ContextCubesTable)
	if err != nil {
		return 0, err
	}
	if te.WasCreated() {
		contextCubes := calculateAllContextCubes()
		if Log.IsDebug() {
			Log.Debugf("Populating table %s with %d elements", te.TableDef.Name, len(contextCubes))
		}
		for _, ctxCube := range contextCubes {
			for _, cubeKey := range ctxCube.allCubes {
				err := te.Insert(ctxCube.trCtx.id, cubeKey.center,
					cubeKey.centerFaces[0], cubeKey.centerFaces[1], cubeKey.centerFaces[2], cubeKey.centerFaces[3], cubeKey.centerFaces[4], cubeKey.centerFaces[5],
					cubeKey.middleEdges[0], cubeKey.middleEdges[1], cubeKey.middleEdges[2], cubeKey.middleEdges[3],
					cubeKey.middleEdges[4], cubeKey.middleEdges[5], cubeKey.middleEdges[6], cubeKey.middleEdges[7],
					cubeKey.middleEdges[8], cubeKey.middleEdges[9], cubeKey.middleEdges[10], cubeKey.middleEdges[11])
				if err != nil {
					Log.Error(err)
				} else {
					inserted++
				}
			}
		}
	}
	return inserted, nil
}
