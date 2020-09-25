package m3space

import "github.com/freddy33/qsm-go/m3util"

type SpaceIfc interface {
	GetId() int
	GetName() string
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
