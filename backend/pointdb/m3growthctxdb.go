package pointdb

import (
	"github.com/freddy33/qsm-go/backend/m3db"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3point"
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
	res.SelectAll = "select id, ctx_type, ctx_index from %s"
	res.ExpectedCount = 52
	return &res
}

/***************************************************************/
// trio Contexts Load and Save
/***************************************************************/

func (ppd *ServerPointPackData) loadGrowthContexts() error {
	te := ppd.growthCtxTe

	rows, err := te.SelectAllForLoad()
	if err != nil {
		return err
	}
	res := make([]m3point.GrowthContext, 0, te.TableDef.ExpectedCount)

	for rows.Next() {
		growthCtx := m3point.BaseGrowthContext{}
		growthCtx.Env = ppd.env
		err := rows.Scan(&growthCtx.Id, &growthCtx.GrowthType, &growthCtx.GrowthIndex)
		if err != nil {
			return m3util.MakeWrapQsmErrorf(err, "failed to load trio context line %d", len(res))
		} else {
			res = append(res, &growthCtx)
		}
	}

	ppd.AllGrowthContexts = res
	ppd.GrowthContextsLoaded = true

	return nil
}

func (ppd *ServerPointPackData) saveAllGrowthContexts() (int, error) {
	te := ppd.growthCtxTe
	inserted, toFill, err := te.GetForSaveAll()
	if err != nil {
		return 0, err
	}
	if toFill {
		growthContexts := ppd.calculateAllGrowthContexts()
		if Log.IsDebug() {
			Log.Debugf("Populating table %s with %d elements", te.GetFullTableName(), len(growthContexts))
		}
		for _, growthCtx := range growthContexts {
			err := te.Insert(growthCtx.GetId(), growthCtx.GetGrowthType(), growthCtx.GetGrowthIndex())
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

func (ppd *ServerPointPackData) calculateAllGrowthContexts() []m3point.GrowthContext {
	res := make([]m3point.GrowthContext, m3point.TotalNbContexts)
	idx := 0
	for _, ctxType := range m3point.GetAllGrowthTypes() {
		nbIndexes := ctxType.GetNbIndexes()
		for pIdx := 0; pIdx < nbIndexes; pIdx++ {
			growthCtx := m3point.BaseGrowthContext{Env: ppd.env, Id: idx, GrowthType: ctxType, GrowthIndex: pIdx}
			res[idx] = &growthCtx
			idx++
		}
	}
	return res
}

