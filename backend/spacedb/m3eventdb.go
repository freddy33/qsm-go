package spacedb

import (
	"fmt"
	"github.com/freddy33/qsm-go/backend/pathdb"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3path"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/freddy33/qsm-go/model/m3space"
)

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

func (en *EventNodeDb) insertInDb() error {
	evt := en.event
	linkForDb := en.GetLinkIdsForDb()
	var err error
	en.id, err = evt.space.spaceData.nodesTe.InsertReturnId(evt.id, en.pathNodeId, en.pointId, en.d, en.creationTime,
		en.GetConnectionMask(), linkForDb[0], linkForDb[1], linkForDb[2])
	if err != nil {
		return err
	}
	return nil
}

type EventDb struct {
	space        *SpaceDb
	id           m3space.EventId
	pathCtx      *pathdb.PathContextDb
	creationTime m3space.DistAndTime
	color        m3space.EventColor
	centerNode   *EventNodeDb
	// End time set equal to creation time when alive
	endTime m3space.DistAndTime
}

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
	id64, err := evt.space.spaceData.eventsTe.InsertReturnId(evt.space.GetId(), evt.GetPathContext().GetId(), evt.GetCreationTime(), evt.GetColor(), evt.endTime)
	if err != nil {
		return m3util.MakeWrapQsmErrorf(err, "could not insert event %s due to '%s'", evt.String(), err.Error())
	}
	evt.id = m3space.EventId(id64)
	eventNode := evt.centerNode
	err = eventNode.insertInDb()
	if err != nil {
		return err
	}
	return nil
}
