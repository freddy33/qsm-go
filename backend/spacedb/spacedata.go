package spacedb

import (
	"github.com/freddy33/qsm-go/backend/m3db"
	"github.com/freddy33/qsm-go/backend/pathdb"
	"github.com/freddy33/qsm-go/backend/pointdb"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3space"
)

type ServerSpacePackData struct {
	m3space.BaseSpacePackData
	env *m3db.QsmDbEnvironment

	spacesTe *m3db.TableExec
	eventsTe *m3db.TableExec
	nodesTe  *m3db.TableExec

	allSpaces       map[int]*SpaceDb
	allSpacesLoaded bool
}

func (spaceData *ServerSpacePackData) CreateSpace(name string, activePathNodeThreshold m3space.DistAndTime, maxTriosPerPoint int, maxPathNodesPerPoint int) (m3space.SpaceIfc, error) {
	return CreateSpace(spaceData.env, name, activePathNodeThreshold, maxTriosPerPoint, maxPathNodesPerPoint)
}

func (spaceData *ServerSpacePackData) GetAllSpaces() []m3space.SpaceIfc {
	err := spaceData.LoadAllSpaces()
	if err != nil {
		Log.Error(err)
		return nil
	}
	res := make([]m3space.SpaceIfc, len(spaceData.allSpaces))
	i := 0
	for _, s := range spaceData.allSpaces {
		res[i] = s
		i++
	}
	return res
}

func (spaceData *ServerSpacePackData) LoadAllSpaces() error {
	if spaceData.allSpacesLoaded {
		return nil
	}
	pathData := pathdb.GetServerPathPackData(spaceData.env)
	pointData := pointdb.GetServerPointPackData(spaceData.env)
	rows, err := spaceData.spacesTe.SelectAllForLoad()
	if err != nil {
		return err
	}
	for rows.Next() {
		space := SpaceDb{spaceData: spaceData, pathData: pathData, pointData: pointData}
		err := rows.Scan(&space.id, &space.name, &space.activePathNodeThreshold,
			&space.maxTriosPerPoint, &space.maxPathNodesPerPoint, &space.maxCoord, &space.maxTime)
		if err != nil {
			return err
		}
		existingSpace, ok := spaceData.allSpaces[space.id]
		if ok {
			// Make sure same data
			if existingSpace.name != space.name {
				return m3util.MakeQsmErrorf("got different spaces in memory %v and DB %v", existingSpace, space)
			}
		} else {
			err = space.finalInit()
			if err != nil {
				return err
			}
		}
	}
	spaceData.allSpacesLoaded = true
	return nil
}

func (spaceData *ServerSpacePackData) GetSpace(id int) m3space.SpaceIfc {
	err := spaceData.LoadAllSpaces()
	if err != nil {
		Log.Error(err)
		return nil
	}
	return spaceData.allSpaces[id]
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
