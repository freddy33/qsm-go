package m3path

import (
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/freddy33/qsm-go/utils/m3db"
	"github.com/freddy33/qsm-go/utils/m3util"
)

type PathPackData struct {
	env           *m3db.QsmDbEnvironment
	pathCtxMap    map[int]*PathContextDb
	pathNodeCache map[int64]*PathNodeDb

	// All PathContexts centered at origin with growth type + offset
	allCenterContexts       map[m3point.GrowthType][]PathContext
	allCenterContextsLoaded bool
}

func makePathPackData(env *m3db.QsmDbEnvironment) *PathPackData {
	res := new(PathPackData)
	res.env = env
	res.pathCtxMap = make(map[int]*PathContextDb, 2^8)
	res.pathNodeCache = make(map[int64]*PathNodeDb, 2^16)
	res.allCenterContexts = make(map[m3point.GrowthType][]PathContext)
	return res
}

func GetPathPackData(env *m3db.QsmDbEnvironment) *PathPackData {
	if env.GetData(m3util.PathIdx) == nil {
		env.SetData(m3util.PathIdx, makePathPackData(env))
	}
	return env.GetData(m3util.PathIdx).(*PathPackData)
}

func (ppd *PathPackData) GetEnvId() m3util.QsmEnvID {
	if ppd == nil {
		return m3util.NoEnv
	}
	if ppd.env == nil {
		return m3util.NoEnv
	}
	return ppd.env.GetId()
}

func (ppd *PathPackData) GetPathCtx(id int) PathContext {
	pathCtx, ok := ppd.pathCtxMap[id]
	if ok {
		return pathCtx
	}
	// TODO: Load from DB
	return nil
}

func (ppd *PathPackData) addPathCtx(pathCtx *PathContextDb) {
	ppd.pathCtxMap[pathCtx.id] = pathCtx
}
