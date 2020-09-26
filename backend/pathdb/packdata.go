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

	pointsTe    *m3db.TableExec
	pathCtxTe   *m3db.TableExec
	pathNodesTe *m3db.TableExec

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

func GetServerPathPackData(env m3util.QsmEnvironment) *ServerPathPackData {
	if env.GetData(m3util.PathIdx) == nil {
		env.SetData(m3util.PathIdx, makeServerPathPackData(env))
	}
	return env.GetData(m3util.PathIdx).(*ServerPathPackData)
}

func (pathData *ServerPathPackData) addCenterPathContext(pathCtx m3path.PathContext) {
	if len(pathData.AllCenterContexts[pathCtx.GetGrowthType()]) == 0 {
		nbIndexes := pathCtx.GetGrowthType().GetNbIndexes()
		pathData.AllCenterContexts[pathCtx.GetGrowthType()] = make([]m3path.PathContext, nbIndexes)
		for i := 0; i < nbIndexes; i++ {
			pathData.AllCenterContexts[pathCtx.GetGrowthType()][i] = nil
		}
	}
	if pathData.AllCenterContexts[pathCtx.GetGrowthType()][pathCtx.GetGrowthOffset()] == nil {
		pathData.AllCenterContexts[pathCtx.GetGrowthType()][pathCtx.GetGrowthOffset()] = pathCtx
	}
}
