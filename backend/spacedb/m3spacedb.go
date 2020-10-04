package spacedb

import (
	"fmt"
	"github.com/freddy33/qsm-go/backend/m3db"
	"github.com/freddy33/qsm-go/backend/pathdb"
	"github.com/freddy33/qsm-go/backend/pointdb"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/freddy33/qsm-go/model/m3space"
	"sync"
)

type SpaceDb struct {
	spaceData *ServerSpacePackData
	pathData  *pathdb.ServerPathPackData
	pointData *pointdb.ServerPointPackData

	// Unique keys
	id   int
	name string

	// global counters to quickly retrieve space metrics
	maxCoord m3point.CInt
	maxTime  m3space.DistAndTime

	// SpaceTime behavior configuration parameters
	activeThreshold  m3space.DistAndTime
	maxTriosPerPoint int
	maxNodesPerPoint int

	events map[m3space.EventId]*EventDb
}

type SpaceTime struct {
	space       *SpaceDb
	currentTime m3space.DistAndTime

	populated      bool
	populatedMutex sync.Mutex
	populatedError error
	activeEvents   []*EventDb
	stNodes        map[int64]*SpaceTimeNode
}

/***************************************************************/
// SpaceDb Functions
/***************************************************************/

func CreateSpace(env *m3db.QsmDbEnvironment,
	name string, activePathNodeThreshold m3space.DistAndTime,
	maxTriosPerPoint int, maxPathNodesPerPoint int) (*SpaceDb, error) {
	space := new(SpaceDb)
	space.spaceData = GetServerSpacePackData(env)
	space.pathData = pathdb.GetServerPathPackData(env)
	space.pointData = pointdb.GetServerPointPackData(env)
	space.name = name
	space.activeThreshold = activePathNodeThreshold
	space.maxTriosPerPoint = maxTriosPerPoint
	space.maxNodesPerPoint = maxPathNodesPerPoint

	// 2*9 is the minimum ;-)
	space.maxCoord = m3space.MinMaxCoord
	space.maxTime = 0

	err := space.insertInDb()
	if err != nil {
		return nil, err
	}

	err = space.finalInit()
	if err != nil {
		return nil, err
	}

	return space, nil
}

func (space *SpaceDb) finalInit() error {
	rows, err := space.spaceData.eventsTe.Query(SelectEventsPerSpace, space.GetId())
	if err != nil {
		return err
	}
	space.events = make(map[m3space.EventId]*EventDb, 8)
	for rows.Next() {
		err = CreateEventFromDbRows(rows, space)
		if err != nil {
			return err
		}
	}
	space.spaceData.allSpaces[space.id] = space
	return nil
}

func (space *SpaceDb) String() string {
	return fmt.Sprintf("%d:%s-%d", space.id, space.name, len(space.events))
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

func (space *SpaceDb) GetActiveThreshold() m3space.DistAndTime {
	return space.activeThreshold
}

func (space *SpaceDb) GetMaxTriosPerPoint() int {
	return space.maxTriosPerPoint
}

func (space *SpaceDb) GetMaxNodesPerPoint() int {
	return space.maxNodesPerPoint
}

func (space *SpaceDb) GetEvent(id m3space.EventId) m3space.EventIfc {
	return space.events[id]
}

func (space *SpaceDb) GetEventIdsForMsg() []int32 {
	res := make([]int32, len(space.events))
	i := 0
	for _, evt := range space.events {
		res[i] = int32(evt.GetId())
		i++
	}
	return res
}

func (space *SpaceDb) GetNbNodesBetween(from, to m3space.DistAndTime) (int, error) {
	row := space.spaceData.nodesTe.QueryRow(CountNodesPerSpaceBetween, space.GetId(), from, to)
	var res int
	err := row.Scan(&res)
	return res, err
}

func (space *SpaceDb) insertInDb() error {
	te := space.spaceData.spacesTe
	id64, err := te.InsertReturnId(space.name, space.activeThreshold, space.maxTriosPerPoint, space.maxNodesPerPoint, space.maxCoord, space.maxTime)
	if err != nil {
		return m3util.MakeWrapQsmErrorf(err, "could not insert space %q in %q due to: %s", space.GetName(), te.GetFullTableName(), err.Error())
	}
	space.id = int(id64)
	return nil
}

func (space *SpaceDb) CreateEvent(growthType m3point.GrowthType, growthIndex int, growthOffset int,
	creationTime m3space.DistAndTime, center m3point.Point, color m3space.EventColor) (m3space.EventIfc, error) {
	centerPoint := center
	pointId := space.pathData.GetOrCreatePoint(centerPoint)
	pathCtx, err := space.pathData.GetPathCtxDbFromAttributes(growthType, growthIndex, growthOffset)
	if err != nil {
		return nil, err
	}
	rootPathNode := pathCtx.GetRootPathNode().(*pathdb.PathNodeDb)
	evt := &EventDb{
		space:        space,
		pathCtx:      pathCtx,
		creationTime: creationTime,
		color:        color,
		maxNodeTime:  m3space.ZeroDistAndTime,
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
	evt.centerNode.SetTrioDetails(rootPathNode.GetTrioDetails(space.pointData))
	evt.centerNode.ConnectionMask = rootPathNode.GetConnectionMask()
	evt.centerNode.LinkIds, err = evt.getNodeLinkIds(rootPathNode)
	if err != nil {
		return nil, err
	}

	err = evt.insertInDb()
	if err != nil {
		return nil, err
	}

	space.setMaxCoordAndTime(evt.centerNode)

	return evt, nil
}

func (space *SpaceDb) setMaxCoordAndTime(evtNode *EventNodeDb) {
	if evtNode.creationTime > space.maxTime {
		space.maxTime = evtNode.creationTime
	}
	if evtNode.point != nil {
		for _, c := range *evtNode.point {
			if c > space.maxCoord {
				space.maxCoord = c
			}
			if -c > space.maxCoord {
				space.maxCoord = -c
			}
		}
	}
}

func (space *SpaceDb) CreateSingleEventCenter() *EventDb {
	return space.CreateEventFromColor(m3point.Origin, m3space.RedEvent)
}

func (space *SpaceDb) CreatePyramid(pyramidSize m3point.CInt) {
	space.CreateEventFromColor(m3point.Point{3, 0, 3}.Mul(pyramidSize), m3space.RedEvent)
	space.CreateEventFromColor(m3point.Point{-3, 3, 3}.Mul(pyramidSize), m3space.GreenEvent)
	space.CreateEventFromColor(m3point.Point{-3, -3, 3}.Mul(pyramidSize), m3space.BlueEvent)
	space.CreateEventFromColor(m3point.Point{0, 0, -3}.Mul(pyramidSize), m3space.YellowEvent)
}

func (space *SpaceDb) CreateEventFromColor(p m3point.Point, k m3space.EventColor) *EventDb {
	idx, offset := getIndexAndOffsetForColor(k)
	evt, err := space.CreateEvent(m3point.GrowthType(8), idx, offset, m3space.ZeroDistAndTime, p, k)
	if err != nil {
		Log.Error(err)
		return nil
	}
	return evt.(*EventDb)
}

func (space *SpaceDb) GetSpaceTimeAt(time m3space.DistAndTime) m3space.SpaceTimeIfc {
	st := new(SpaceTime)
	st.space = space
	st.currentTime = time
	return st
}

func (space *SpaceDb) GetActiveEventsAt(time m3space.DistAndTime) []m3space.EventIfc {
	res := make([]m3space.EventIfc, 0, len(space.events))
	for _, evt := range space.events {
		if evt.creationTime <= time {
			res = append(res, evt)
		}
	}
	return res
}

func (space *SpaceDb) GetNbEventsAt(time m3space.DistAndTime) int {
	return len(space.GetActiveEventsAt(time))
}

func getIndexAndOffsetForColor(k m3space.EventColor) (int, int) {
	switch k {
	case m3space.RedEvent:
		return 0, 0
	case m3space.GreenEvent:
		return 4, 0
	case m3space.BlueEvent:
		return 8, 0
	case m3space.YellowEvent:
		return 10, 4
	}
	Log.Errorf("Event color unknown %v", k)
	return -1, -1
}

/***************************************************************/
// SpaceTime Functions
/***************************************************************/

func (st *SpaceTime) GetCurrentTime() m3space.DistAndTime {
	return st.currentTime
}

func (st *SpaceTime) GetRuleAnalyzer() *SpaceTimeRuleAnalyzer {
	res := MakeRuleAnalyzer(st)
	st.VisitAll(res)
	return res
}

/*
Return fromDist, toDist, and use between flag for this event in this space time
*/
func (st *SpaceTime) queryPathContext(evt *EventDb) (int, int, bool) {
	toDist := st.currentTime - evt.GetCreationTime()
	threshold := st.space.GetActiveThreshold()
	if threshold > 0 && toDist > 0 {
		fromDist := toDist - threshold
		if fromDist < 0 {
			fromDist = m3space.ZeroDistAndTime
		}
		return int(fromDist), int(toDist), true
	} else {
		return 0, int(toDist), false
	}
}


func (st *SpaceTime) populate() error {
	if st.populated {
		return st.populatedError
	}

	st.populatedMutex.Lock()
	defer st.populatedMutex.Unlock()

	if st.populated {
		return st.populatedError
	}

	events := st.GetActiveEvents()
	st.activeEvents = make([]*EventDb, len(events))
	nodesMap := make(map[m3space.EventId][]*EventNodeDb, len(events))
	nbPathNodes := 0
	for i, _ := range events {
		evt := events[i].(*EventDb)
		st.activeEvents[i] = evt
		nodeList, err := evt.GetActiveNodesAt(st.currentTime)
		if err != nil {
			st.populatedError = err
			st.populated = true
			return err
		}
		nbPathNodes += len(nodeList)
		nodesMap[evt.GetId()] = nodeList
	}
	st.stNodes = make(map[int64]*SpaceTimeNode, nbPathNodes)
	for _, nodeList := range nodesMap {
		for _, en := range nodeList {
			stn, ok := st.stNodes[en.GetPointId()]
			if ok {
				stn.head.Add(en)
			} else {
				st.stNodes[en.GetPointId()] = &SpaceTimeNode{
					spaceTime: st,
					pointId:   en.GetPointId(),
					head:      &NodeEventList{cur: en},
				}
			}
		}
	}
	st.populated = true
	return nil
}

func (st *SpaceTime) GetNbActiveNodes() int {
	err := st.populate()
	if err != nil {
		Log.Error(err)
		return -1
	}
	return len(st.stNodes)
}

func (st *SpaceTime) GetSpace() m3space.SpaceIfc {
	return st.space
}

func (st *SpaceTime) GetActiveEvents() []m3space.EventIfc {
	return st.space.GetActiveEventsAt(st.currentTime)
}

func (st *SpaceTime) Next() m3space.SpaceTimeIfc {
	// TODO: Place to optimize when threshold is > 0 and conflict at nodes level appears
	// TODO: For now just recreate all
	return st.space.GetSpaceTimeAt(st.currentTime + 1)
}

func (st *SpaceTime) GetNbActiveLinks() int {
	err := st.populate()
	if err != nil {
		Log.Error(err)
		return -1
	}
	threshold := st.space.GetActiveThreshold()
	if threshold == 0 {
		return 0
	}
	// Count all link ids positive for distance to latest ( currentTime - creationTime )
	// strictly smaller than threshold
	nbActiveLinks := 0
	for _, stn := range st.stNodes {
		connIdsAlreadyDone := make(map[m3point.ConnectionId]bool)
		stn.VisitConnections(func(evtNode *EventNodeDb, connId m3point.ConnectionId, linkId int64) {
			if linkId > 0 && st.currentTime-evtNode.creationTime < threshold {
				alreadyDone, ok := connIdsAlreadyDone[connId]
				if !ok || !alreadyDone {
					nbActiveLinks++
				}
				connIdsAlreadyDone[connId] = true
			}
		})
	}
	return nbActiveLinks
}

func (st *SpaceTime) VisitAll(visitor m3space.SpaceTimeVisitor) {
	err := st.populate()
	if err != nil {
		Log.Error(err)
		return
	}
	threshold := st.space.GetActiveThreshold()
	for _, stn := range st.stNodes {
		// Visit all the nodes
		visitor.VisitNode(stn)

		// Visit links only if threshold above 0
		if threshold != 0 {
			point, err := stn.GetPoint()
			if err != nil {
				Log.Error(err)
				return
			}
			connIdsAlreadyDone := make(map[m3point.ConnectionId]bool)
			stn.VisitConnections(func(evtNode *EventNodeDb, connId m3point.ConnectionId, linkId int64) {
				if linkId > 0 && st.currentTime-evtNode.creationTime < threshold {
					alreadyDone, ok := connIdsAlreadyDone[connId]
					if !ok || !alreadyDone {
						visitor.VisitLink(stn, *point, connId)
					}
					connIdsAlreadyDone[connId] = true
				}
			})
		}
	}

}

func (st *SpaceTime) GetDisplayState() string {
	return fmt.Sprintf("========= SpaceTime State =========\n"+
		"Current Time: %d, Nb Active Nodes: %d, Nb Events %d",
		st.currentTime, st.GetNbActiveNodes(), st.space.GetNbEventsAt(st.currentTime))
}
