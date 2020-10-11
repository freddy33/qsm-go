package client

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3api"
	"github.com/freddy33/qsm-go/model/m3path"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/freddy33/qsm-go/model/m3space"
)

type ClientSpacePackData struct {
	Env       *QsmApiEnvironment
	allSpaces map[int]*SpaceCl
	allSpacesLoaded bool
}

type SpaceCl struct {
	SpaceData *ClientSpacePackData

	id   int
	name string

	maxCoord m3point.CInt
	maxTime  m3space.DistAndTime
	eventIds []m3space.EventId

	activeThreshold      m3space.DistAndTime
	maxTriosPerPoint     int
	maxPathNodesPerPoint int
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

func (spaceData *ClientSpacePackData) GetEnvId() m3util.QsmEnvID {
	return spaceData.Env.GetId()
}

func (spaceData *ClientSpacePackData) LoadAllSpaces() error {
	if spaceData.allSpacesLoaded {
		return nil
	}
	uri := "space"
	pMsg := &m3api.SpaceListMsg{}
	_, err := spaceData.Env.clConn.ExecReq("GET", uri, nil, pMsg, false)
	if err != nil {
		return err
	}
	res := make(map[int]*SpaceCl, len(pMsg.Spaces))
	for i, sp := range pMsg.Spaces {
		res[i] = createSpaceClFromMsg(spaceData, sp)
	}
	spaceData.allSpaces = res
	spaceData.allSpacesLoaded = true
	return nil
}

func createSpaceClFromMsg(spd *ClientSpacePackData, sp *m3api.SpaceMsg) *SpaceCl {
	space := &SpaceCl{
		SpaceData:            spd,
		id:                   int(sp.SpaceId),
		name:                 sp.SpaceName,
		maxCoord:             m3point.CInt(sp.GetMaxCoord()),
		maxTime:              m3space.DistAndTime(sp.GetMaxTime()),
		activeThreshold:      m3space.DistAndTime(sp.GetActiveThreshold()),
		maxTriosPerPoint:     int(sp.GetMaxTriosPerPoint()),
		maxPathNodesPerPoint: int(sp.GetMaxNodesPerPoint()),
	}
	if len(sp.EventIds) > 0 {
		space.eventIds = make([]m3space.EventId, len(sp.EventIds))
		for i, evtId := range sp.EventIds {
			space.eventIds[i] = m3space.EventId(evtId)
		}
	} else {
		space.eventIds = nil
	}
	return space
}

func (spaceData *ClientSpacePackData) GetAllSpaces() []m3space.SpaceIfc {
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

func (spaceData *ClientSpacePackData) GetSpace(id int) m3space.SpaceIfc {
	err := spaceData.LoadAllSpaces()
	if err != nil {
		Log.Error(err)
		return nil
	}
	return spaceData.allSpaces[id]
}

func (spaceData *ClientSpacePackData) CreateSpace(name string, activePathNodeThreshold m3space.DistAndTime, maxTriosPerPoint int, maxPathNodesPerPoint int) (m3space.SpaceIfc, error) {
	uri := "space"
	reqMsg := &m3api.SpaceMsg{
		SpaceName:            name,
		ActiveThreshold:      int32(activePathNodeThreshold),
		MaxTriosPerPoint:     int32(maxTriosPerPoint),
		MaxNodesPerPoint:     int32(maxPathNodesPerPoint),
	}
	resMsg := &m3api.SpaceMsg{}
	_, err := spaceData.Env.clConn.ExecReq("PUT", uri, reqMsg, resMsg, false)
	if err != nil {
		return nil, err
	}
	space := createSpaceClFromMsg(spaceData, resMsg)
	spaceData.allSpaces[space.id] = space
	return space, nil
}

func (spaceData *ClientSpacePackData) DeleteSpace(id int, name string) (int, error) {
	uri := "space"
	reqMsg := &m3api.SpaceMsg{
		SpaceId:            int32(id),
		SpaceName:            name,
	}
	_, err := spaceData.Env.clConn.ExecReq("DELETE", uri, reqMsg, nil, true)
	if err != nil {
		return 0, err
	}
	delete(spaceData.allSpaces, id)
	return 1, nil
}

/***************************************************************/
// SpaceCl Functions
/***************************************************************/

func (space *SpaceCl) String() string {
	return fmt.Sprintf("SpaceCl:%d:%s-%d-%d", space.id, space.name, space.maxTime, space.maxCoord)
}

func (space *SpaceCl) GetId() int {
	return space.id
}

func (space *SpaceCl) GetName() string {
	return space.name
}

func (space *SpaceCl) GetMaxTriosPerPoint() int {
	return space.maxTriosPerPoint
}

func (space *SpaceCl) GetActiveThreshold() m3space.DistAndTime {
	return space.activeThreshold
}

func (space *SpaceCl) GetMaxNodesPerPoint() int {
	return space.maxPathNodesPerPoint
}

func (space *SpaceCl) GetMaxTime() m3space.DistAndTime {
	return space.maxTime
}

func (space *SpaceCl) GetMaxCoord() m3point.CInt {
	return space.maxCoord
}

func (space *SpaceCl) GetEvent(id m3space.EventId) m3space.EventIfc {

	return
}

func (space *SpaceCl) GetActiveEventsAt(time m3space.DistAndTime) []m3space.EventIfc {
	panic("implement me")
}

func (space *SpaceCl) GetSpaceTimeAt(time m3space.DistAndTime) m3space.SpaceTimeIfc {
	panic("implement me")
}

func (space *SpaceCl) CreateEvent(growthType m3point.GrowthType, growthIndex int, growthOffset int,
	creationTime m3space.DistAndTime, center m3point.Point, color m3space.EventColor) (m3space.EventIfc, error) {

	uri := "event"

	reqMsg := &m3api.CreateEventRequestMsg{
		SpaceId:      int32(space.id),
		GrowthType:   int32(growthType),
		GrowthIndex:  int32(growthIndex),
		GrowthOffset: int32(growthOffset),
		CreationTime: int32(creationTime),
		Center:       m3api.PointToPointMsg(center),
		Color:        uint32(color),
	}
	resMsg := new(m3api.EventMsg)
	_, err := space.SpaceData.Env.clConn.ExecReq("PUT", uri, reqMsg, resMsg, false)

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

