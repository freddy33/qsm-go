package client

import (
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/freddy33/qsm-go/m3util"
)

func GetApiPointPackData(env m3util.QsmEnvironment) *m3point.LoadedPointPackData {
	if env.GetData(m3util.PointIdx) == nil {
		ppd := new(m3point.LoadedPointPackData)
		ppd.EnvId = env.GetId()
		env.SetData(m3util.PointIdx, ppd)
		// do not return ppd but always the pointer in Env data array
	}
	return env.GetData(m3util.PointIdx).(*m3point.LoadedPointPackData)
}
