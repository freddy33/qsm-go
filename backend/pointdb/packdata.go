package pointdb

import (
	"github.com/freddy33/qsm-go/backend/m3db"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3point"
)

type ServerPointPackData struct {
	m3point.BasePointPackData
	Env *m3db.QsmDbEnvironment
}

func GetPointPackData(env m3util.QsmEnvironment) m3point.PointPackDataIfc {
	ppd, _ := GetServerPointPackData(env)
	return ppd
}

func GetServerPointPackData(env m3util.QsmEnvironment) (*ServerPointPackData, bool) {
	newData := env.GetData(m3util.PointIdx) == nil
	if newData {
		ppd := new(ServerPointPackData)
		ppd.EnvId = env.GetId()
		ppd.Env = env.(*m3db.QsmDbEnvironment)
		env.SetData(m3util.PointIdx, ppd)
		// do not return ppd but always the pointer in env data array
	}
	return env.GetData(m3util.PointIdx).(*ServerPointPackData), newData
}
