package spacedb

import (
	"github.com/freddy33/qsm-go/backend/m3db"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/freddy33/qsm-go/model/m3space"
)

type SpaceDb struct {
	spd *ServerSpacePackData

	// Unique keys
	id   int
	name string

	// global counters to quickly retrieve space metrics
	maxCoord      m3point.CInt
	maxTime       m3space.DistAndTime
	nbActiveNodes int

	// Space behavior configuration parameters
	activePathNodeThreshold m3space.DistAndTime
	maxTriosPerPoint        int
	maxPathNodesPerPoint    int

	// Current state of this space
	currentTime m3space.DistAndTime
	events      []m3space.Event
}

func CreateSpace(env *m3db.QsmDbEnvironment,
	name string, activePathNodeThreshold m3space.DistAndTime,
	maxTriosPerPoint int, maxPathNodesPerPoint int) (*SpaceDb, error) {
	space := new(SpaceDb)
	space.spd = GetServerSpacePackData(env)
	space.name = name
	space.activePathNodeThreshold = activePathNodeThreshold
	space.maxTriosPerPoint = maxTriosPerPoint
	space.maxPathNodesPerPoint = maxPathNodesPerPoint

	err := space.insertInDb()
	if err != nil {
		return nil, err
	}

	// 2*9 is the minimum ;-)
	space.maxCoord = 2 * 9
	space.maxTime = 0

	space.finalInit()

	return space, nil
}

func (space *SpaceDb) finalInit() {
	space.nbActiveNodes = 0
	space.currentTime = 0
	space.events = make([]m3space.Event, 0, 8)
	space.spd.allSpaces[space.id] = space
}

func (space *SpaceDb) GetId() int {
	return space.id
}

func (space *SpaceDb) GetName() string {
	return space.name
}

func (space *SpaceDb) GetMaxCoord() m3point.CInt {
	return space.maxCoord
}

func (space *SpaceDb) GetMaxTime() m3space.DistAndTime {
	return space.maxTime
}

func (space *SpaceDb) GetNbActiveNodes() int {
	return space.nbActiveNodes
}

func (space *SpaceDb) GetActivePathNodeThreshold() m3space.DistAndTime {
	return space.activePathNodeThreshold
}

func (space *SpaceDb) GetMaxTriosPerPoint() int {
	return space.maxTriosPerPoint
}

func (space *SpaceDb) GetMaxPathNodesPerPoint() int {
	return space.maxPathNodesPerPoint
}

func (space *SpaceDb) GetCurrentTime() m3space.DistAndTime {
	return space.currentTime
}

func (space *SpaceDb) GetEventIdsForMsg() []int32 {
	res := make([]int32, len(space.events))
	for i, evt := range space.events {
		res[i] = int32(evt.Id)
	}
	return res
}

func (space *SpaceDb) insertInDb() error {
	te, err := space.spd.env.GetOrCreateTableExec(SpaceTable)
	if err != nil {
		return m3util.MakeWrapQsmErrorf(err, "could not get table details %s out of %s space due to '%s'", SpaceTable, space.GetName(), err.Error())
	}
	id64, err := te.InsertReturnId(space.name, space.activePathNodeThreshold, space.maxTriosPerPoint, space.maxPathNodesPerPoint, space.maxCoord)
	if err != nil {
		return m3util.MakeWrapQsmErrorf(err, "could not insert space %s due to '%s'", space.GetName(), err.Error())
	}
	space.id = int(id64)
	return nil
}

func (spd *ServerSpacePackData) LoadAllSpaces() error {
	_, rows := spd.env.SelectAllForLoad(SpaceTable)
	for rows.Next() {
		space := SpaceDb{spd: spd}
		err := rows.Scan(&space.id, &space.name, &space.activePathNodeThreshold,
			&space.maxTriosPerPoint, &space.maxPathNodesPerPoint, &space.maxCoord, &space.maxTime)
		if err != nil {
			return err
		}
		existingSpace, ok := spd.allSpaces[space.id]
		if ok {
			// Make sure same data
			if existingSpace.name != space.name {
				return m3util.MakeQsmErrorf("got different spaces in memory %v and DB %v", existingSpace, space)
			}
		} else {
			space.finalInit()
		}
	}
	return nil
}
