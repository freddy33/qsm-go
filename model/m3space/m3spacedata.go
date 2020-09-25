package m3space

import (
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3point"
)

type SpaceIfc interface {
	GetId() int
	GetName() string
	CreateEvent(growthType m3point.GrowthType, growthIndex int, growthOffset int,
		creationTime DistAndTime, center m3point.Point, color EventColor) (EventIfc, error)
}

type SpacePackDataIfc interface {
	m3util.QsmDataPack
	GetAllSpaces() []SpaceIfc
	GetSpace(id int) SpaceIfc
}

type BaseSpacePackData struct {
	EnvId m3util.QsmEnvID
}

func (ppd *BaseSpacePackData) GetEnvId() m3util.QsmEnvID {
	if ppd == nil {
		return m3util.NoEnv
	}
	return ppd.EnvId
}
