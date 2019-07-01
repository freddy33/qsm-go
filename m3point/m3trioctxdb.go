package m3point

import (
	"database/sql"
	"fmt"
	"github.com/freddy33/qsm-go/m3db"
	"github.com/lib/pq"
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
		" center smallint," +
		" center_faces smallint []," +
		" middle_edges smallint [])", TrioContextsTable)
	res.InsertStmt = "(ctx_id, center, center_faces, middle_edges) values ($1,$2,$3,$4)"
	res.SelectAll = fmt.Sprintf("select ctx_id, center, center_faces, middle_edges from %s", ContextCubesTable)
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
		var center int
		centerFaces := make([]sql.NullInt64, 0, 6)
		middleEdges := make([]sql.NullInt64, 0, 12)
		err := rows.Scan(&trCtxId, &center, pq.Array(&centerFaces), pq.Array(&middleEdges))
		if err != nil {
			Log.Errorf("failed to load trio context line %d due to %v", loaded, err)
		} else {
			cubeKey := CubeKey{}
			cubeKey.center = TrioIndex(uint8(center))
			for i := 0; i < 6; i++ {
				if !centerFaces[i].Valid {
					Log.Errorf("center face %d for trio context line %d invalid", i, loaded)
				}
				cubeKey.centerFaces[i] = TrioIndex(uint8(centerFaces[i].Int64))
			}
			for i := 0; i < 12; i++ {
				if !middleEdges[i].Valid {
					Log.Errorf("middle egde %d for trio context line %d invalid", i, loaded)
				}
				cubeKey.middleEdges[i] = TrioIndex(uint8(middleEdges[i].Int64))
			}
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
				err := te.Insert(ctxCube.trCtx.id, cubeKey.center, pq.Array(cubeKey.centerFaces), pq.Array(cubeKey.middleEdges))
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
