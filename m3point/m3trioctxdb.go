package m3point

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3db"
)

const (
	TrioContextsTable    = "trio_contexts"
	CubesTable           = "cubes"
	CubesPerContextTable = "cubes_per_context"
)

func init() {
	m3db.AddTableDef(createTrioContextTableDef())
	m3db.AddTableDef(createCubesTableDef())
	m3db.AddTableDef(createCubesPerContextTableDef())
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

func createCubesTableDef() *m3db.TableDefinition {
	res := m3db.TableDefinition{}
	res.Name = CubesTable
	res.DdlColumns = "(id smallint PRIMARY KEY," +
		" center smallint," +
		" center_faces smallint []," +
		" middle_edges smallint [])"
	res.InsertStmt = "(id, center, center_faces, middle_edges) values ($1,$2,$3,$4)"
	res.ExpectedCount = 52
	return &res
}

func createCubesPerContextTableDef() *m3db.TableDefinition {
	res := m3db.TableDefinition{}
	res.Name = CubesPerContextTable
	res.DdlColumns = fmt.Sprintf("(ctx_id smallint REFERENCES %s (id),"+
		" cube_id smallint  REFERENCES %s (id),"+
		" UNIQUE (ctx_id, cube_id) )", TrioContextsTable, CubesTable)
	res.InsertStmt = "(ctx_id, cube_id) values ($1,$2)"
	res.ExpectedCount = 52
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
		}
		res = append(res, &trCtx)
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
