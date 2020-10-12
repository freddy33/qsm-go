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
	Env             *QsmApiEnvironment
	allSpaces       map[int]*SpaceCl
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

	point m3point.Point

	creationTime m3space.DistAndTime
	d            m3space.DistAndTime

	pathNodeId     int64
	trioDetails    *m3point.TrioDetails
	connectionMask uint16
	linkNodes      [m3path.NbConnections]int64
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
		SpaceName:        name,
		ActiveThreshold:  int32(activePathNodeThreshold),
		MaxTriosPerPoint: int32(maxTriosPerPoint),
		MaxNodesPerPoint: int32(maxPathNodesPerPoint),
	}
	resMsg := &m3api.SpaceMsg{}
	_, err := spaceData.Env.clConn.ExecReq("POST", uri, reqMsg, resMsg, false)
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
		SpaceId:   int32(id),
		SpaceName: name,
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
	uri := "event"
	reqMsg := &m3api.FindEventsMsg{
		EventId:              int32(id),
		SpaceId:              int32(space.id),
		AtTime:               0,
	}
	resMsg := new(m3api.EventListMsg)
	_, err := space.SpaceData.Env.clConn.ExecReq("GET", uri, reqMsg, resMsg, true)
	if err != nil {
		Log.Error(err)
		return nil
	}
	if len(resMsg.Events) != 1 {
		Log.Errorf("Did not find 1 single event id %d at %s but %d array", id, space.String(), len(resMsg.Events))
		return nil
	}
	pathData := GetClientPathPackData(space.SpaceData.Env)
	pointData := GetClientPointPackData(space.SpaceData.Env)
	evtMsg := resMsg.Events[0]
	event, err := space.createEventFromMsg(pathData, pointData, evtMsg)
	if err != nil {
		Log.Error(err)
		return nil
	}
	return event
}

func (space *SpaceCl) GetActiveEventsAt(atTime m3space.DistAndTime) []m3space.EventIfc {
	uri := "event"
	reqMsg := &m3api.FindEventsMsg{
		EventId:              int32(-1),
		SpaceId:              int32(space.id),
		AtTime:               int32(atTime),
	}
	resMsg := new(m3api.EventListMsg)
	_, err := space.SpaceData.Env.clConn.ExecReq("GET", uri, reqMsg, resMsg, true)
	if err != nil {
		Log.Error(err)
		return nil
	}
	if len(resMsg.Events) > 0 {
		Log.Infof("Did not find a single event at time %d for %s", atTime, space.String())
		return nil
	}

	res := make([]m3space.EventIfc, len(resMsg.Events))
	pathData := GetClientPathPackData(space.SpaceData.Env)
	pointData := GetClientPointPackData(space.SpaceData.Env)
	for i, evtMsg := range resMsg.Events {
		res[i], err = space.createEventFromMsg(pathData, pointData, evtMsg)
		if err != nil {
			Log.Error(err)
			return nil
		}
	}
	return res
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
	_, err := space.SpaceData.Env.clConn.ExecReq("POST", uri, reqMsg, resMsg, false)
	if err != nil {
		return nil, err
	}

	pathData := GetClientPathPackData(space.SpaceData.Env)
	pointData := GetClientPointPackData(space.SpaceData.Env)
	evt, err := space.createEventFromMsg(pathData, pointData, resMsg)
	if err != nil {
		return nil, err
	}
	space.eventIds = append(space.eventIds, evt.id)

	return evt, nil
}

func (space *SpaceCl) setMaxCoordAndTime(evtNode *EventNodeCl) {
	if evtNode.creationTime > space.maxTime {
		space.maxTime = evtNode.creationTime
	}
	for _, c := range evtNode.point {
		if c > space.maxCoord {
			space.maxCoord = c
		}
		if -c > space.maxCoord {
			space.maxCoord = -c
		}
	}
}

func (space *SpaceCl) createEventFromMsg(pathData *ClientPathPackData, pointData *ClientPointPackData, resMsg *m3api.EventMsg) (*EventCl, error) {
	pathCtx := pathData.GetPathCtx(int(resMsg.PathCtxId))
	if pathCtx == nil {
		return nil, m3util.MakeQsmErrorf("Received path context id %d which does exists", resMsg.PathCtxId)
	}
	event := &EventCl{
		space:        space,
		id:           m3space.EventId(resMsg.EventId),
		pathCtx:      pathCtx.(*PathContextCl),
		CreationTime: m3space.DistAndTime(resMsg.CreationTime),
		color:        m3space.EventColor(resMsg.Color),
		endTime:      0,
		MaxNodeTime:  m3space.DistAndTime(resMsg.MaxNodeTime),
	}
	var err error
	event.CenterNode, err = event.createNodeFromMsg(pointData, resMsg.RootNode)
	if err != nil {
		return nil, err
	}
	return event, nil
}

/***************************************************************/
// EventCl Functions
/***************************************************************/

func (evt *EventCl) createNodeFromMsg(pointData *ClientPointPackData, neMsg *m3api.NodeEventMsg) (*EventNodeCl, error) {
	td := pointData.GetTrioDetails(m3point.TrioIndex(neMsg.TrioId))
	if td == nil {
		return nil, m3util.MakeQsmErrorf("Cannot create node %d since trio index %d does not exists", neMsg.EventNodeId, neMsg.TrioId)
	}
	ne := &EventNodeCl{
		Event:          evt,
		id:             neMsg.EventNodeId,
		point:          m3api.PointMsgToPoint(neMsg.Point),
		creationTime:   m3space.DistAndTime(neMsg.CreationTime),
		d:              m3space.DistAndTime(neMsg.D),
		pathNodeId:     neMsg.PathNodeId,
		trioDetails:    td,
		connectionMask: uint16(neMsg.ConnectionMask),
	}
	for i, nodeId := range neMsg.LinkedNodeIds {
		ne.linkNodes[i] = nodeId
	}
	evt.space.setMaxCoordAndTime(ne)

	return ne, nil
}

func (evt *EventCl) String() string {
	return fmt.Sprintf("Evt%02d:Sp%02d:CT=%d:%d", evt.id, evt.space.id, evt.CreationTime, evt.color)
}

func (evt *EventCl) GetId() m3space.EventId {
	return evt.id
}

func (evt *EventCl) GetSpace() m3space.SpaceIfc {
	return evt.space
}

func (evt *EventCl) GetPathContext() m3path.PathContext {
	return evt.pathCtx
}

func (evt *EventCl) GetCreationTime() m3space.DistAndTime {
	return evt.CreationTime
}

func (evt *EventCl) GetColor() m3space.EventColor {
	return evt.color
}

func (evt *EventCl) GetCenterNode() m3space.NodeEventIfc {
	return evt.CenterNode
}

func (evt *EventCl) GetActiveNodesAt(currentTime m3space.DistAndTime) ([]m3space.NodeEventIfc, error) {
	uri := "event-nodes"
	reqMsg := &m3api.FindNodeEventsMsg{
		EventId:              int32(evt.id),
		SpaceId:              int32(evt.space.id),
		AtTime:               int32(currentTime),
	}
	resMsg := new(m3api.NodeEventListMsg)
	env := evt.space.SpaceData.Env
	_, err := env.clConn.ExecReq("GET", uri, reqMsg, resMsg, true)
	if err != nil {
		return nil, err
	}
	if len(resMsg.Nodes) > 0 {
		return nil, m3util.MakeQsmErrorf("Did not find a single node event at time %d for %s", currentTime, evt.String())
	}

	res := make([]m3space.NodeEventIfc, len(resMsg.Nodes))
	pointData := GetClientPointPackData(env)
	for i, nodeMsg := range resMsg.Nodes {
		res[i], err = evt.createNodeFromMsg(pointData, nodeMsg)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

/***************************************************************/
// EventNodeCl Functions
/***************************************************************/

func (en *EventNodeCl) String() string {
	return fmt.Sprintf("EvtNode%02d:Evt%02d:P=%v:T=%d:%d", en.id, en.Event.id,
		en.point, en.creationTime, en.d)
}

func (en *EventNodeCl) GetId() int64 {
	return en.id
}

func (en *EventNodeCl) GetEventId() m3space.EventId {
	return en.Event.GetId()
}

func (en *EventNodeCl) GetPointId() int64 {
	panic("There is no point id on client side")
}

func (en *EventNodeCl) GetPathNodeId() int64 {
	return en.pathNodeId
}

func (en *EventNodeCl) GetCreationTime() m3space.DistAndTime {
	return en.creationTime
}

func (en *EventNodeCl) GetD() m3space.DistAndTime {
	return en.d
}

func (en *EventNodeCl) GetColor() m3space.EventColor {
	return en.Event.color
}

func (en *EventNodeCl) GetPoint() (*m3point.Point, error) {
	return &en.point, nil
}

func (en *EventNodeCl) GetPathNode() (m3path.PathNode, error) {
	return en.Event.pathCtx.pathNodeMap.GetPathNodeById(m3point.Int64Id(en.pathNodeId)), nil
}

func (en *EventNodeCl) GetTrioIndex() m3point.TrioIndex {
	return en.trioDetails.Id
}

func (en *EventNodeCl) GetTrioDetails(pointData m3point.PointPackDataIfc) *m3point.TrioDetails {
	return en.trioDetails
}

func (en *EventNodeCl) GetConnectionState(connIdx int) m3path.ConnectionState {
	return m3path.GetConnectionState(en.connectionMask, connIdx)
}

func (en *EventNodeCl) HasOpenConnections() bool {
	for i := 0; i < m3path.NbConnections; i++ {
		if en.GetConnectionState(i) == m3path.ConnectionNotSet {
			return true
		}
	}
	return false
}

func (en *EventNodeCl) IsFrom(connIdx int) bool {
	return en.GetConnectionState(connIdx) == m3path.ConnectionFrom
}

func (en *EventNodeCl) IsNext(connIdx int) bool {
	return en.GetConnectionState(connIdx) == m3path.ConnectionNext
}

func (en *EventNodeCl) IsDeadEnd(connIdx int) bool {
	return en.GetConnectionState(connIdx) == m3path.ConnectionBlocked
}
