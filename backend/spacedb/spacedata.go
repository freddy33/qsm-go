package spacedb

import (
	"github.com/freddy33/qsm-go/backend/m3db"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3space"
)

type ServerSpacePackData struct {
	m3space.BaseSpacePackData
	env *m3db.QsmDbEnvironment

	allSpaces map[int]*SpaceDb
}

func (spd *ServerSpacePackData) GetAllSpaces() []m3space.SpaceIfc {
	res := make([]m3space.SpaceIfc, len(spd.allSpaces))
	i := 0
	for _, s := range spd.allSpaces {
		res[i] = s
		i++
	}
	return res
}

func (spd *ServerSpacePackData) GetSpace(id int) m3space.SpaceIfc {
	return spd.allSpaces[id]
}

func makeServerSpacePackData(env m3util.QsmEnvironment) *ServerSpacePackData {
	res := new(ServerSpacePackData)
	res.EnvId = env.GetId()
	res.env = env.(*m3db.QsmDbEnvironment)
	res.allSpaces = make(map[int]*SpaceDb, 3)
	return res
}

func GetServerSpacePackData(env m3util.QsmEnvironment) *ServerSpacePackData {
	if env.GetData(m3util.SpaceIdx) == nil {
		env.SetData(m3util.SpaceIdx, makeServerSpacePackData(env))
	}
	return env.GetData(m3util.SpaceIdx).(*ServerSpacePackData)
}
