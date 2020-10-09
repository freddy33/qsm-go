package m3space

import (
	"fmt"
	"github.com/freddy33/qsm-go/model/m3path"
	"github.com/freddy33/qsm-go/model/m3point"
)

type EventId int

const (
	NilEvent = EventId(-1)
	MinMaxCoord = m3point.CInt(2 * 9)
)

type DistAndTime int

const ZeroDistAndTime = DistAndTime(0)

type EventColor uint8

const (
	RedEvent EventColor = 1 << iota
	GreenEvent
	BlueEvent
	YellowEvent
)

// TODO: This should be in the space data entry of the environment
var AllColors = [4]EventColor{RedEvent, GreenEvent, BlueEvent, YellowEvent}

type SpaceIfc interface {
	fmt.Stringer
	GetId() int
	GetName() string
	GetActiveThreshold() DistAndTime
	GetMaxTriosPerPoint() int
	GetMaxNodesPerPoint() int
	GetMaxTime() DistAndTime
	GetMaxCoord() m3point.CInt
	GetEvent(id EventId) EventIfc
	GetActiveEventsAt(time DistAndTime) []EventIfc
	GetSpaceTimeAt(time DistAndTime) SpaceTimeIfc
	CreateEvent(growthType m3point.GrowthType, growthIndex int, growthOffset int,
		creationTime DistAndTime, center m3point.Point, color EventColor) (EventIfc, error)
}

type SpaceTimeVisitor interface {
	VisitNode(node SpaceTimeNodeIfc)
	VisitLink(node SpaceTimeNodeIfc, srcPoint m3point.Point, connId m3point.ConnectionId)
}

type SpaceTimeIfc interface {
	GetSpace() SpaceIfc
	GetCurrentTime() DistAndTime
	GetActiveEvents() []EventIfc
	Next() SpaceTimeIfc
	GetNbActiveNodes() int
	GetNbActiveLinks() int
	VisitAll(visitor SpaceTimeVisitor)
	GetDisplayState() string
}

type SpaceTimeNodeIfc interface {
	GetSpaceTime() SpaceTimeIfc
	GetPointId() int64
	GetPoint() (*m3point.Point, error)
	IsEmpty() bool
	GetNbEventNodes() int
	GetEventNodes() []EventNodeIfc
	GetEventIds() []EventId

	HasRoot() bool
	GetLastAccessed() DistAndTime
	HowManyColors() uint8
	GetColorMask() uint8

	GetStateString() string
}

type EventIfc interface {
	fmt.Stringer
	GetId() EventId
	GetSpace() SpaceIfc
	GetPathContext() m3path.PathContext
	GetCreationTime() DistAndTime
	GetColor() EventColor
	GetCenterNode() EventNodeIfc
	GetActiveNodesAt(currentTime DistAndTime) ([]EventNodeIfc, error)
}

type EventNodeIfc interface {
	fmt.Stringer
	m3path.ConnectionStateIfc
	GetId() int64
	GetEventId() EventId
	GetPointId() int64
	GetPathNodeId() int64
	GetCreationTime() DistAndTime
	GetD() DistAndTime
	GetColor() EventColor
	GetPoint() (*m3point.Point, error)
	GetPathNode() (m3path.PathNode, error)
}


