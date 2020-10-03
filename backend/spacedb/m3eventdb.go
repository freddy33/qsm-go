package spacedb

import (
	"database/sql"
	"fmt"
	"github.com/freddy33/qsm-go/backend/pathdb"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3path"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/freddy33/qsm-go/model/m3space"
	"sync"
)

type EventDb struct {
	space        *SpaceDb
	id           m3space.EventId
	pathCtx      *pathdb.PathContextDb
	creationTime m3space.DistAndTime
	color        m3space.EventColor
	centerNode   *EventNodeDb
	// End time set equal to creation time when alive
	endTime m3space.DistAndTime
	// The biggest creation time of event node db
	maxNodeTime m3space.DistAndTime

	increaseNodeMutex sync.Mutex
}

type EventNodeDb struct {
	pathdb.ConnectionsStateDb

	event *EventDb
	id    int64

	pointId    int64
	pathNodeId int64

	creationTime m3space.DistAndTime
	d            m3space.DistAndTime

	// Loaded on demand
	point    *m3point.Point
	pathNode *pathdb.PathNodeDb
}

/***************************************************************/
// EventDb Functions
/***************************************************************/

func (evt *EventDb) String() string {
	return fmt.Sprintf("Evt%02d:Sp%02d:T=%d:%d", evt.id, evt.space.id, evt.creationTime, evt.color)
}

func (evt *EventDb) GetId() m3space.EventId {
	return evt.id
}

func (evt *EventDb) GetSpace() m3space.SpaceIfc {
	return evt.space
}

func (evt *EventDb) GetPathContext() m3path.PathContext {
	return evt.pathCtx
}

func (evt *EventDb) GetCreationTime() m3space.DistAndTime {
	return evt.creationTime
}

func (evt *EventDb) GetColor() m3space.EventColor {
	return evt.color
}

func (evt *EventDb) GetCenterNode() m3space.EventNodeIfc {
	return evt.centerNode
}

func (evt *EventDb) insertInDb() error {
	// max and end time set to creation time at first insert
	evt.endTime = evt.creationTime
	evt.maxNodeTime = evt.creationTime

	id64, err := evt.space.spaceData.eventsTe.InsertReturnId(
		evt.space.GetId(), evt.GetPathContext().GetId(), evt.GetCreationTime(), evt.GetColor(), evt.endTime, evt.maxNodeTime)
	if err != nil {
		return m3util.MakeWrapQsmErrorf(err, "could not insert event %s due to '%s'", evt.String(), err.Error())
	}
	evt.id = m3space.EventId(id64)
	evt.space.events[evt.id] = evt
	err = evt.centerNode.insertInDb()
	if err != nil {
		return err
	}
	return nil
}

func (evt *EventDb) increaseMaxNodeTime() error {
	evt.increaseNodeMutex.Lock()
	defer evt.increaseNodeMutex.Unlock()

	nextTime := evt.maxNodeTime + 1

	center, err := evt.GetCenterNode().GetPoint()
	if err != nil {
		return err
	}
	dTime := nextTime - evt.creationTime
	pathNodes, err := evt.pathCtx.GetPathNodesAt(int(dTime))
	if err != nil {
		return err
	}
	for _, pn := range pathNodes {
		point := (*center).Add(pn.P())
		pointId := evt.space.pathData.GetOrCreatePoint(point)
		pathNodeDb := pn.(*pathdb.PathNodeDb)
		evtNode := &EventNodeDb{
			ConnectionsStateDb: pathdb.ConnectionsStateDb{
				ConnectionMask: pathNodeDb.ConnectionMask,
				LinkIds:        [3]int64{pathNodeDb.LinkIds[0], pathNodeDb.LinkIds[1], pathNodeDb.LinkIds[2]},
				TrioId:         pathNodeDb.TrioId,
				TrioDetails:    nil,
			},
			event:        evt,
			pointId:      pointId,
			point:        &point,
			pathNodeId:   pn.GetId(),
			creationTime: nextTime,
			d:            dTime,
			pathNode:     pathNodeDb,
		}
		err = evtNode.insertInDb()
		if err != nil {
			return err
		}
		evt.space.setMaxCoordAndTime(evtNode)
	}

	evt.maxNodeTime = nextTime
	rowAffected, err := evt.space.spaceData.eventsTe.Update(UpdateMaxNodeTime, evt.id, evt.maxNodeTime)
	if err != nil {
		return m3util.MakeWrapQsmErrorf(err, "could not update event %s with new max node time %d due to %v", evt.String(), evt.maxNodeTime, err)
	}
	if rowAffected != 1 {
		return m3util.MakeQsmErrorf("updating event %s with new max node time %d returned wrong rows %d", evt.String(), evt.maxNodeTime, rowAffected)
	}
	// TODO: Needs to update space maxCoord and maxTime also
	return nil
}

func (evt *EventDb) GetActiveNodesAt(currentTime m3space.DistAndTime) ([]*EventNodeDb, error) {
	var err error
	for evt.maxNodeTime < currentTime {
		err = evt.increaseMaxNodeTime()
		if err != nil {
			return nil, err
		}
	}
	te := evt.space.spaceData.nodesTe
	from, to, useBetween := evt.getFromToTime(currentTime)
	var rows *sql.Rows
	var expectedNbNodes int
	if useBetween {
		expectedNbNodes = evt.pathCtx.GetNumberOfNodesBetween(int(from-evt.creationTime), int(to-evt.creationTime))
		rows, err = te.Query(SelectNodesBetween, evt.GetId(), from, to)
	} else {
		expectedNbNodes = evt.pathCtx.GetNumberOfNodesAt(int(currentTime - evt.creationTime))
		rows, err = te.Query(SelectNodesAt, evt.GetId(), currentTime)
	}
	if err != nil {
		return nil, err
	}
	res := make([]*EventNodeDb, 0, expectedNbNodes)
	for rows.Next() {
		evtNode, err := evt.CreateEventNodeFromDbRows(rows)
		if err != nil {
			return nil, err
		}
		res = append(res, evtNode)
	}
	return res, nil
}

/*
Return fromTime, toTime, and use between flag based on space threshold, current time passed and state of event
*/
func (evt *EventDb) getFromToTime(currentTime m3space.DistAndTime) (m3space.DistAndTime, m3space.DistAndTime, bool) {
	availableDelta := currentTime - evt.creationTime
	if availableDelta < 0 {
		Log.Errorf("asking from and to time for inactive event %s at time %d", evt.String(), currentTime)
		return -1, -1, false
	}
	if evt.endTime != evt.creationTime && evt.endTime < currentTime {
		Log.Errorf("asking from and to time for event %s already dead at time %d", evt.String(), currentTime)
		return -1, -1, false
	}

	threshold := evt.space.GetActivePathNodeThreshold()
	if threshold == 0 || availableDelta == 0 {
		return currentTime, currentTime, false
	}
	if availableDelta >= threshold {
		return currentTime - threshold, currentTime, true
	}
	return evt.creationTime, currentTime, true
}

func CreateEventFromDbRows(rows *sql.Rows, space *SpaceDb) error {
	evt := EventDb{space: space}
	rootNode := EventNodeDb{event: &evt}
	point := m3point.Point{}
	linkIds := [m3path.NbConnections]sql.NullInt64{}
	var pathCtxId int
	err := rows.Scan(&evt.id, &pathCtxId, &evt.creationTime,
		&evt.color, &evt.endTime, &evt.maxNodeTime,
		&rootNode.id, &rootNode.pathNodeId, &rootNode.TrioId, &rootNode.pointId,
		&rootNode.ConnectionMask, &linkIds[0], &linkIds[1], &linkIds[2],
		&point[0], &point[1], &point[2])
	if err != nil {
		return err
	}
	evt.pathCtx = space.pathData.GetPathCtxDb(pathCtxId)
	if evt.pathCtx == nil {
		return m3util.MakeQsmErrorf("got event %d from db with wrong path context id %d", evt.id, pathCtxId)
	}
	rootNode.SetLinkIdsFromDbData(linkIds)
	rootNode.d = m3space.DistAndTime(0)
	rootNode.creationTime = evt.creationTime
	rootNode.TrioDetails = nil
	rootNode.pathNode = nil
	rootNode.point = &point
	evt.centerNode = &rootNode
	space.events[evt.GetId()] = &evt
	return nil
}

func (evt *EventDb) CreateEventNodeFromDbRows(rows *sql.Rows) (*EventNodeDb, error) {
	evtNode := EventNodeDb{event: evt}
	point := m3point.Point{}
	linkIds := [m3path.NbConnections]sql.NullInt64{}
	var eventId m3space.EventId
	err := rows.Scan(&evtNode.id, &eventId, &evtNode.pathNodeId, &evtNode.TrioId, &evtNode.pointId, &evtNode.d, &evtNode.creationTime,
		&evtNode.ConnectionMask, &linkIds[0], &linkIds[1], &linkIds[2],
		&point[0], &point[1], &point[2])
	if err != nil {
		return nil, err
	}
	if eventId != evt.GetId() {
		return nil, m3util.MakeQsmErrorf("got event node %d from db with wrong event id %d instead of %d", evtNode.id, eventId, evt.GetId())
	}
	evtNode.SetLinkIdsFromDbData(linkIds)
	evtNode.pathNode = nil
	evtNode.point = &point

	return &evtNode, nil
}

/***************************************************************/
// EventNodeDb Functions
/***************************************************************/

func (en *EventNodeDb) String() string {
	return fmt.Sprintf("EvtNode%02d:Evt%02d:P=%04d,%v:T=%d:%d", en.id, en.event.id,
		en.pointId, en.point, en.creationTime, en.d)
}

func (en *EventNodeDb) GetId() int64 {
	return en.id
}

func (en *EventNodeDb) GetEventId() m3space.EventId {
	return en.event.GetId()
}

func (en *EventNodeDb) GetColor() m3space.EventColor {
	return en.event.GetColor()
}

func (en *EventNodeDb) GetPointId() int64 {
	return en.pointId
}

func (en *EventNodeDb) GetPoint() (*m3point.Point, error) {
	if en.pointId < 0 {
		return nil, m3util.MakeQsmErrorf("No point id in event %s", en.String())
	}
	if en.point != nil {
		return en.point, nil
	}
	var err error
	en.point, err = en.event.space.pathData.GetPoint(en.pointId)
	if err != nil {
		return nil, err
	}
	return en.point, nil
}

func (en *EventNodeDb) GetPathNodeId() int64 {
	return en.pathNodeId
}

func (en *EventNodeDb) GetPathNode() (m3path.PathNode, error) {
	if en.pathNodeId < 0 {
		return nil, m3util.MakeQsmErrorf("No path node id in event %s", en.String())
	}
	if en.pathNode != nil {
		return en.pathNode, nil
	}
	var err error
	en.pathNode, err = en.event.pathCtx.GetPathNodeDb(en.pathNodeId)
	if err != nil {
		return nil, err
	}
	return en.pathNode, nil
}

func (en *EventNodeDb) GetCreationTime() m3space.DistAndTime {
	return en.creationTime
}

func (en *EventNodeDb) GetD() m3space.DistAndTime {
	return en.d
}

func (en *EventNodeDb) IsRoot() bool {
	return en.d == m3space.DistAndTime(0)
}

func (en *EventNodeDb) insertInDb() error {
	evt := en.event
	linkForDb := en.GetLinkIdsForDb()
	var err error
	en.id, err = evt.space.spaceData.nodesTe.InsertReturnId(
		evt.id, en.pathNodeId, en.GetTrioIndex(), en.pointId, en.d, en.creationTime,
		en.GetConnectionMask(), linkForDb[0], linkForDb[1], linkForDb[2])
	if err != nil {
		return err
	}
	return nil
}
