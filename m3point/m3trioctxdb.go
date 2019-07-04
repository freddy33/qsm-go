package m3point

import (
	"github.com/freddy33/qsm-go/m3db"
)

const (
	GrowthContextsTable = "growth_contexts"
)

func init() {
	m3db.AddTableDef(createGrowthContextsTableDef())
}

func createGrowthContextsTableDef() *m3db.TableDefinition {
	res := m3db.TableDefinition{}
	res.Name = GrowthContextsTable
	res.DdlColumns = "(id smallint PRIMARY KEY," +
		" ctx_type smallint," +
		" ctx_index smallint, UNIQUE (ctx_type, ctx_index) )"
	res.Insert = "(id, ctx_type, ctx_index) values ($1,$2,$3)"
	res.SelectAll = "select id, ctx_type, ctx_index from growth_contexts"
	res.ExpectedCount = 52
	return &res
}

/***************************************************************/
// trio Contexts Load and Save
/***************************************************************/

func loadGrowthContexts() []GrowthContext {
	te, rows := GetPointEnv().SelectAllForLoad(GrowthContextsTable)
	res := make([]GrowthContext, 0, te.TableDef.ExpectedCount)

	for rows.Next() {
		growthCtx := BaseGrowthContext{}
		err := rows.Scan(&growthCtx.id, &growthCtx.growthType, &growthCtx.growthIndex)
		if err != nil {
			Log.Errorf("failed to load trio context line %d", len(res))
		} else {
			res = append(res, &growthCtx)
		}
	}
	return res
}

func saveAllGrowthContexts() (int, error) {
	te, inserted, err := GetPointEnv().GetForSaveAll(GrowthContextsTable)
	if err != nil {
		return 0, err
	}
	if te.WasCreated() {
		growthCtxs := calculateAllGrowthContexts()
		if Log.IsDebug() {
			Log.Debugf("Populating table %s with %d elements", te.TableDef.Name, len(growthCtxs))
		}
		for _, growthCtx := range growthCtxs {
			err := te.Insert(growthCtx.GetId(), growthCtx.GetGrowthType(), growthCtx.GetGrowthIndex())
			if err != nil {
				Log.Error(err)
			} else {
				inserted++
			}
		}
	}
	return inserted, nil
}
