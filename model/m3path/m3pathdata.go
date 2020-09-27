package m3path

import (
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3point"
)

type PathPackDataIfc interface {
	m3util.QsmDataPack
	GetPathCtx(id int) PathContext
	CreatePathCtxFromAttributes(growthCtx m3point.GrowthContext, offset int) (PathContext, error)
}



