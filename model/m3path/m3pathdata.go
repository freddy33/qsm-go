package m3path

import (
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3point"
)

type PathPackDataIfc interface {
	m3util.QsmDataPack
	AddPathCtx(pathCtx PathContext)
	GetPathCtx(id int) PathContext
	CreatePathCtxFromAttributes(growthCtx m3point.GrowthContext, offset int, center m3point.Point) PathContext
}

type BasePathPackData struct {
	EnvId m3util.QsmEnvID

	PathCtxMap map[int]PathContext
}

func (ppd *BasePathPackData) GetEnvId() m3util.QsmEnvID {
	if ppd == nil {
		return m3util.NoEnv
	}
	return ppd.EnvId
}

func (ppd *BasePathPackData) GetPathCtx(id int) PathContext {
	pathCtx, ok := ppd.PathCtxMap[id]
	if ok {
		return pathCtx
	}
	// TODO: Load from DB
	return nil
}

func (ppd *BasePathPackData) AddPathCtx(pathCtx PathContext) {
	ppd.PathCtxMap[pathCtx.GetId()] = pathCtx
}


