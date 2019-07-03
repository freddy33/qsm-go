package m3point

import (
	"github.com/freddy33/qsm-go/m3db"
)

const (
	TrioContextsTable = "trio_contexts"
)

func init() {
	m3db.AddTableDef(createTrioContextTableDef())
}

func createTrioContextTableDef() *m3db.TableDefinition {
	res := m3db.TableDefinition{}
	res.Name = TrioContextsTable
	res.DdlColumns = "(id smallint PRIMARY KEY," +
		" ctx_type smallint," +
		" ctx_index smallint, UNIQUE (ctx_type, ctx_index) )"
	res.Insert = "(id, ctx_type, ctx_index) values ($1,$2,$3)"
	res.SelectAll = "select id, ctx_type, ctx_index from trio_contexts"
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
