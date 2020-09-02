package m3path

import (
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3point"
)

type PathPackDataIfc interface {
	m3util.QsmDataPack
	GetPathCtx(id int) PathContext
}

type PathPackData struct {
	EnvId m3util.QsmEnvID

	pathCtxMap    map[int]PathContext

	// All PathContexts centered at origin with growth type + offset
	AllCenterContexts       map[m3point.GrowthType][]PathContext
	AllCenterContextsLoaded bool
}

func makePathPackData(env m3util.QsmEnvironment) *PathPackData {
	res := new(PathPackData)
	res.EnvId = env.GetId()
	res.pathCtxMap = make(map[int]PathContext, 2^8)
	res.AllCenterContexts = make(map[m3point.GrowthType][]PathContext)
	res.AllCenterContextsLoaded = false
	return res
}

func GetPathPackData(env m3util.QsmEnvironment) *PathPackData {
	if env.GetData(m3util.PathIdx) == nil {
		env.SetData(m3util.PathIdx, makePathPackData(env))
	}
	return env.GetData(m3util.PathIdx).(*PathPackData)
}

func (ppd *PathPackData) GetEnvId() m3util.QsmEnvID {
	if ppd == nil {
		return m3util.NoEnv
	}
	return ppd.EnvId
}

func (ppd *PathPackData) GetPathCtx(id int) PathContext {
	pathCtx, ok := ppd.pathCtxMap[id]
	if ok {
		return pathCtx
	}
	// TODO: Load from DB
	return nil
}

func (ppd *PathPackData) AddPathCtx(pathCtx PathContext) {
	ppd.pathCtxMap[pathCtx.GetId()] = pathCtx
}
