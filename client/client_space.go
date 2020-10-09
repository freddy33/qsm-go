package client

import (
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3path"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/freddy33/qsm-go/model/m3space"
)

type ClientSpacePackData struct {
	Env       *QsmApiEnvironment
	allSpaces map[int]*SpaceCl
}

type SpaceCl struct {
	SpaceData *ClientSpacePackData

	id   int
	name string

	maxCoord m3point.CInt
	maxTime  m3space.DistAndTime

	activePathNodeThreshold m3space.DistAndTime
	maxTriosPerPoint        int
	maxPathNodesPerPoint    int
}

type EventCl struct {
	space        *SpaceCl
	id           m3space.EventId
	pathCtx      *PathContextCl
	CreationTime m3space.DistAndTime
	color        m3space.EventColor

	CenterNode *EventNodeCl
	// End time set equal to creation time when alive
	endTime m3space.DistAndTime
	// The biggest creation time of event node db
	MaxNodeTime m3space.DistAndTime
}

type EventNodeCl struct {
	Event *EventCl
	id    int64

	pointId int64
	point   m3point.Point

	creationTime m3space.DistAndTime
	d            m3space.DistAndTime
}

/***************************************************************/
// ClientSpacePackData Functions
/***************************************************************/

func (spd *ClientSpacePackData) GetEnvId() m3util.QsmEnvID {
	return spd.Env.GetId()
}

func (spd *ClientSpacePackData) GetAllSpaces() []m3space.SpaceIfc {

	res := make([]m3space.SpaceIfc, len(spd.allSpaces))
	i := 0
	for _, s := range spd.allSpaces {
		res[i] = s
		i++
	}
	return res
}

func (spd *ClientSpacePackData) GetSpace(id int) m3space.SpaceIfc {
	panic("implement me")
}

func (spd *ClientSpacePackData) CreateSpace(name string, activePathNodeThreshold m3space.DistAndTime, maxTriosPerPoint int, maxPathNodesPerPoint int) (m3space.SpaceIfc, error) {
	panic("implement me")
}

func (spd *ClientSpacePackData) DeleteSpace(id int, name string) (int, error) {
	panic("implement me")
}

/***************************************************************/
// SpaceCl Functions
/***************************************************************/

func (s *SpaceCl) String() string {
	panic("implement me")
}

func (s *SpaceCl) GetId() int {
	panic("implement me")
}

func (s *SpaceCl) GetName() string {
	panic("implement me")
}

func (s *SpaceCl) GetMaxTriosPerPoint() int {
	panic("implement me")
}

func (s *SpaceCl) GetActiveThreshold() m3space.DistAndTime {
	panic("implement me")
}

func (s *SpaceCl) GetMaxNodesPerPoint() int {
	panic("implement me")
}

func (s *SpaceCl) GetMaxTime() m3space.DistAndTime {
	panic("implement me")
}

func (s *SpaceCl) GetMaxCoord() m3point.CInt {
	panic("implement me")
}

func (s *SpaceCl) GetEvent(id m3space.EventId) m3space.EventIfc {
	panic("implement me")
}

func (s *SpaceCl) GetActiveEventsAt(time m3space.DistAndTime) []m3space.EventIfc {
	panic("implement me")
}

func (s *SpaceCl) GetSpaceTimeAt(time m3space.DistAndTime) m3space.SpaceTimeIfc {
	panic("implement me")
}

func (s *SpaceCl) CreateEvent(growthType m3point.GrowthType, growthIndex int, growthOffset int, creationTime m3space.DistAndTime, center m3point.Point, color m3space.EventColor) (m3space.EventIfc, error) {
	panic("implement me")
}

/***************************************************************/
// EventCl Functions
/***************************************************************/

func (evt *EventCl) String() string {
	panic("implement me")
}

func (evt *EventCl) GetId() m3space.EventId {
	panic("implement me")
}

func (evt *EventCl) GetSpace() m3space.SpaceIfc {
	panic("implement me")
}

func (evt *EventCl) GetPathContext() m3path.PathContext {
	panic("implement me")
}

func (evt *EventCl) GetCreationTime() m3space.DistAndTime {
	panic("implement me")
}

func (evt *EventCl) GetColor() m3space.EventColor {
	panic("implement me")
}

func (evt *EventCl) GetCenterNode() m3space.EventNodeIfc {
	panic("implement me")
}

func (evt *EventCl) GetActiveNodesAt(currentTime m3space.DistAndTime) ([]m3space.EventNodeIfc, error) {
	panic("implement me")
}

/***************************************************************/
// EventNodeCl Functions
/***************************************************************/

func (en *EventNodeCl) String() string {
	panic("implement me")
}

func (en *EventNodeCl) GetId() int64 {
	panic("implement me")
}

func (en *EventNodeCl) GetEventId() m3space.EventId {
	panic("implement me")
}

func (en *EventNodeCl) GetPointId() int64 {
	panic("implement me")
}

func (en *EventNodeCl) GetPathNodeId() int64 {
	panic("implement me")
}

func (en *EventNodeCl) GetCreationTime() m3space.DistAndTime {
	panic("implement me")
}

func (en *EventNodeCl) GetD() m3space.DistAndTime {
	panic("implement me")
}

func (en *EventNodeCl) GetColor() m3space.EventColor {
	panic("implement me")
}

func (en *EventNodeCl) GetPoint() (*m3point.Point, error) {
	panic("implement me")
}

func (en *EventNodeCl) GetPathNode() (m3path.PathNode, error) {
	panic("implement me")
}

func (en *EventNodeCl) GetTrioIndex() m3point.TrioIndex {
	panic("implement me")
}

func (en *EventNodeCl) GetTrioDetails(pointData m3point.PointPackDataIfc) *m3point.TrioDetails {
	panic("implement me")
}

func (en *EventNodeCl) HasOpenConnections() bool {
	panic("implement me")
}

func (en *EventNodeCl) IsFrom(connIdx int) bool {
	panic("implement me")
}

func (en *EventNodeCl) IsNext(connIdx int) bool {
	panic("implement me")
}

func (en *EventNodeCl) IsDeadEnd(connIdx int) bool {
	panic("implement me")
}

