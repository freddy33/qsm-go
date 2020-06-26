package m3path

import (
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/freddy33/qsm-go/utils/m3db"
)

type PathPackData struct {
	env           *m3db.QsmEnvironment
	pathCtxMap    map[int]*PathContextDb
	pathNodeCache map[int64]*PathNodeDb

	// All PathContexts centered at origin with growth type + offset
	allCenterContexts       map[m3point.GrowthType][]PathContext
	allCenterContextsLoaded bool
}

func makePathPackData(env *m3db.QsmEnvironment) *PathPackData {
	res := new(PathPackData)
	res.env = env
	res.pathCtxMap = make(map[int]*PathContextDb, 2^8)
	res.pathNodeCache = make(map[int64]*PathNodeDb, 2^16)
	res.allCenterContexts = make(map[m3point.GrowthType][]PathContext)
	return res
}

func GetPathPackData(env *m3db.QsmEnvironment) *PathPackData {
	if env.GetData(m3db.PathIdx) == nil {
		env.SetData(m3db.PathIdx, makePathPackData(env))
	}
	return env.GetData(m3db.PathIdx).(*PathPackData)
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
