package spacedb

import (
	"github.com/freddy33/qsm-go/backend/m3db"
	"github.com/freddy33/qsm-go/backend/pathdb"
	"github.com/freddy33/qsm-go/backend/pointdb"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/freddy33/qsm-go/model/m3space"
)

type SpaceDb struct {
	spaceData *ServerSpacePackData
	pathData  *pathdb.ServerPathPackData

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
	events      []*EventDb
}

func CreateSpace(env *m3db.QsmDbEnvironment,
	name string, activePathNodeThreshold m3space.DistAndTime,
	maxTriosPerPoint int, maxPathNodesPerPoint int) (*SpaceDb, error) {
	space := new(SpaceDb)
	space.spaceData = GetServerSpacePackData(env)
	space.pathData = pathdb.GetServerPathPackData(env)
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
	space.events = make([]*EventDb, 0, 8)
	space.spaceData.allSpaces[space.id] = space
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
		res[i] = int32(evt.GetId())
	}
	return res
}

func (space *SpaceDb) insertInDb() error {
	te := space.spaceData.spacesTe
	id64, err := te.InsertReturnId(space.name, space.activePathNodeThreshold, space.maxTriosPerPoint, space.maxPathNodesPerPoint, space.maxCoord)
	if err != nil {
		return m3util.MakeWrapQsmErrorf(err, "could not insert space %q in %q due to: %s", space.GetName(), te.GetFullTableName(), err.Error())
	}
	space.id = int(id64)
	return nil
}

func (space *SpaceDb) CreateEvent(growthType m3point.GrowthType, growthIndex int, growthOffset int,
	creationTime m3space.DistAndTime, center m3point.Point, color m3space.EventColor) (m3space.EventIfc, error) {
	env := space.spaceData.env
	pointData := pointdb.GetPointPackData(env)
	growthCtx := pointData.GetGrowthContextByTypeAndIndex(growthType, growthIndex)
	if growthCtx == nil {
		return nil, m3util.MakeQsmErrorf("Growth context with type=%d and index=%d does not exists", growthType, growthIndex)
	}
	centerPoint := center
	pointId := space.pathData.GetOrCreatePoint(centerPoint)
	pathCtx := space.pathData.CreatePathCtxFromAttributes(growthCtx, growthOffset, m3point.Origin)
	rootPathNode := pathCtx.GetRootPathNode().(*pathdb.PathNodeDb)
	evt := &EventDb{
		space:        space,
		pathCtx:      pathCtx.(*pathdb.PathContextDb),
		creationTime: creationTime,
		color:        color,
		endTime:      creationTime,
	}
	evt.centerNode = &EventNodeDb{
		event:        evt,
		pointId:      pointId,
		pathNodeId:   rootPathNode.GetId(),
		creationTime: creationTime,
		d:            0,
		point:        &centerPoint,
		pathNode:     rootPathNode,
	}
	evt.centerNode.SetLinksToNil()
	evt.centerNode.SetFullConnectionMask(rootPathNode.GetConnectionMask())

	space.events = append(space.events, evt)

	err := evt.insertInDb()
	if err != nil {
		return nil, err
	}

	return evt, nil
}

func (spd *ServerSpacePackData) LoadAllSpaces() error {
	rows, err := spd.spacesTe.SelectAllForLoad()
	if err != nil {
		return err
	}
	for rows.Next() {
		space := SpaceDb{spaceData: spd}
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
