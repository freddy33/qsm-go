package pathdb

import (
	"github.com/freddy33/qsm-go/backend/m3db"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3path"
	"github.com/freddy33/qsm-go/model/m3point"
)

type ServerPathPackData struct {
	m3path.BasePathPackData
	env *m3db.QsmDbEnvironment

	// All PathContexts centered at origin with growth type + offset
	AllCenterContexts       map[m3point.GrowthType][]m3path.PathContext
	AllCenterContextsLoaded bool
}

func makeServerPathPackData(env m3util.QsmEnvironment) *ServerPathPackData {
	res := new(ServerPathPackData)
	res.EnvId = env.GetId()
	res.env = env.(*m3db.QsmDbEnvironment)
	res.PathCtxMap = make(map[int]m3path.PathContext, 2^8)
	res.AllCenterContexts = make(map[m3point.GrowthType][]m3path.PathContext)
	res.AllCenterContextsLoaded = false
	return res
}

func GetServerPathPackData(env m3util.QsmEnvironment) m3path.PathPackDataIfc {
	if env.GetData(m3util.PathIdx) == nil {
		env.SetData(m3util.PathIdx, makeServerPathPackData(env))
	}
	return env.GetData(m3util.PathIdx).(m3path.PathPackDataIfc)
}

func (ppd *ServerPathPackData) addCenterPathContext(pathCtx m3path.PathContext) {
	if len(ppd.AllCenterContexts[pathCtx.GetGrowthType()]) == 0 {
		nbIndexes := pathCtx.GetGrowthType().GetNbIndexes()
		ppd.AllCenterContexts[pathCtx.GetGrowthType()] = make([]m3path.PathContext, nbIndexes)
		for i := 0; i < nbIndexes; i++ {
			ppd.AllCenterContexts[pathCtx.GetGrowthType()][i] = nil
		}
	}
	if ppd.AllCenterContexts[pathCtx.GetGrowthType()][pathCtx.GetGrowthOffset()] == nil {
		ppd.AllCenterContexts[pathCtx.GetGrowthType()][pathCtx.GetGrowthOffset()] = pathCtx
	}
}
