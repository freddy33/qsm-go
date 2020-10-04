package m3space

import (
	"github.com/freddy33/qsm-go/m3util"
)

var Log = m3util.NewLogger("m3space", m3util.INFO)

type SpacePackDataIfc interface {
	m3util.QsmDataPack
	GetAllSpaces() []SpaceIfc
	GetSpace(id int) SpaceIfc
	CreateSpace(name string, activePathNodeThreshold DistAndTime,
		maxTriosPerPoint int, maxPathNodesPerPoint int) (SpaceIfc, error)
	DeleteSpace(id int, name string) error
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
