package client

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3api"
	"github.com/freddy33/qsm-go/model/m3path"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/freddy33/qsm-go/model/m3space"
)


type SpaceTimeCl struct {
	space       *SpaceCl
	currentTime m3space.DistAndTime

	activeEvents  map[m3space.EventId]*EventCl
	nbActiveNodes int
	nbActiveLinks int
	stNodes       []*SpaceTimeNodeCl
}

type SpaceTimeNodeCl struct {
	spaceTime *SpaceTimeCl
	pointId   int64
	point     m3point.Point

	hasRoot   bool
	colorMask uint8

	nodeEvents []*SpaceTimeNodeEventCl
}

type SpaceTimeNodeEventCl struct {
	eventId        m3space.EventId
	creationTime   m3space.DistAndTime
	d              m3space.DistAndTime
	trioDetails    *m3point.TrioDetails
	connectionMask uint16
}

/***************************************************************/
// SpaceTimeCl Functions
/***************************************************************/

func createSpaceTimeFromMsg(space *SpaceCl, resMsg *m3api.SpaceTimeResponseMsg) *SpaceTimeCl {
	res := &SpaceTimeCl{
		space:         space,
		currentTime:   m3space.DistAndTime(resMsg.CurrentTime),
		activeEvents:  make(map[m3space.EventId]*EventCl, len(resMsg.ActiveEvents)),
		nbActiveNodes: int(resMsg.NbActiveNodes),
		nbActiveLinks: -1,
		stNodes:       make([]*SpaceTimeNodeCl, len(resMsg.FilteredNodes)),
	}
	pathData := GetClientPathPackData(space.SpaceData.Env)
	pointData := GetClientPointPackData(space.SpaceData.Env)
	for _, evtMsg := range resMsg.ActiveEvents {
		evt, err := space.createEventFromMsg(pathData, pointData, evtMsg)
		if err != nil {
			Log.Error(err)
			return nil
		}
		res.activeEvents[evt.GetId()] = evt
	}
	for i, stn := range resMsg.FilteredNodes {
		res.stNodes[i] = res.createStNodeFromMsg(pointData, stn)
	}
	return res
}

func (st *SpaceTimeCl) String() string {
	return fmt.Sprintf("SpaceTimeCl-S:%d-T:%d", st.space.GetId(), st.currentTime)
}

func (st *SpaceTimeCl) GetSpace() m3space.SpaceIfc {
	return st.space
}

func (st *SpaceTimeCl) GetCurrentTime() m3space.DistAndTime {
	return st.currentTime
}

func (st *SpaceTimeCl) GetActiveEvents() []m3space.EventIfc {
	res := make([]m3space.EventIfc, len(st.activeEvents))
	i := 0
	for _, evt := range st.activeEvents {
		res[i] = evt
		i++
	}
	return res
}

func (st *SpaceTimeCl) Next() m3space.SpaceTimeIfc {
	// TODO: Be more efficient here
	return st.space.GetSpaceTimeAt(st.currentTime + 1)
}

func (st *SpaceTimeCl) GetNbActiveNodes() int {
	return st.nbActiveNodes
}

func (st *SpaceTimeCl) GetNbActiveLinks() int {
	if st.nbActiveLinks < 0 {
		st.nbActiveLinks = 0
		threshold := st.space.GetActiveThreshold()
		if threshold > 0 {
			for _, stn := range st.stNodes {
				connIdsAlreadyDone := make(map[m3point.ConnectionId]bool)
				for _, nodeEvt := range stn.nodeEvents {
					for connIdx, conn := range nodeEvt.trioDetails.GetConnections() {
						connId := conn.GetId()
						// Take only the from connections
						if m3path.GetConnectionState(nodeEvt.connectionMask, connIdx) == m3path.ConnectionFrom &&
							st.currentTime-nodeEvt.creationTime < threshold {
							alreadyDone, ok := connIdsAlreadyDone[connId]
							if !ok || !alreadyDone {
								st.nbActiveLinks++
							}
							connIdsAlreadyDone[connId] = true
						}
					}
				}
			}
		}
	}
	return st.nbActiveLinks
}

func (st *SpaceTimeCl) VisitNodes(visitor m3space.SpaceTimeNodeVisitor) {
	for _, node := range st.stNodes {
		visitor.VisitNode(node)
	}
}

func (st *SpaceTimeCl) VisitLinks(visitor m3space.SpaceTimeLinkVisitor) {
	threshold := st.space.GetActiveThreshold()
	if threshold == 0 {
		// Nothing to do
		return
	}

	for _, stn := range st.stNodes {
		point, _ := stn.GetPoint()
		connIdsAlreadyDone := make(map[m3point.ConnectionId]bool)
		for _, nodeEvt := range stn.nodeEvents {
			for connIdx, conn := range nodeEvt.trioDetails.GetConnections() {
				connId := conn.GetId()
				// Take only the from connections
				if m3path.GetConnectionState(nodeEvt.connectionMask, connIdx) == m3path.ConnectionFrom &&
					st.currentTime-nodeEvt.creationTime < threshold {
					alreadyDone, ok := connIdsAlreadyDone[connId]
					if !ok || !alreadyDone {
						visitor.VisitLink(stn, *point, connId)
					}
					connIdsAlreadyDone[connId] = true
				}
			}
		}
	}
}

func (st *SpaceTimeCl) GetDisplayState() string {
	return fmt.Sprintf("%s NbActive=%d Filtered=%d", st.String(), st.nbActiveNodes, len(st.stNodes))
}

func (st *SpaceTimeCl) createStNodeFromMsg(pointData *ClientPointPackData, nodeMsg *m3api.SpaceTimeNodeMsg) *SpaceTimeNodeCl {
	res := &SpaceTimeNodeCl{
		spaceTime:  st,
		pointId:    nodeMsg.PointId,
		point:      m3api.PointMsgToPoint(nodeMsg.Point),
		hasRoot:    nodeMsg.HasRoot,
		colorMask:  uint8(nodeMsg.ColorMask),
		nodeEvents: make([]*SpaceTimeNodeEventCl, len(nodeMsg.Nodes)),
	}
	for i, nodeEvt := range nodeMsg.Nodes {
		res.nodeEvents[i] = &SpaceTimeNodeEventCl{
			eventId:        m3space.EventId(nodeEvt.EventId),
			creationTime:   m3space.DistAndTime(nodeEvt.CreationTime),
			d:              m3space.DistAndTime(nodeEvt.D),
			trioDetails:    pointData.GetTrioDetails(m3point.TrioIndex(nodeEvt.TrioId)),
			connectionMask: uint16(nodeEvt.ConnectionMask),
		}
	}
	return res
}

/***************************************************************/
// SpaceTimeNodeCl Functions
/***************************************************************/

func (stn *SpaceTimeNodeCl) GetSpaceTime() m3space.SpaceTimeIfc {
	return stn.spaceTime
}

func (stn *SpaceTimeNodeCl) GetPointId() int64 {
	return stn.pointId
}

func (stn *SpaceTimeNodeCl) GetPoint() (*m3point.Point, error) {
	return &stn.point, nil
}

func (stn *SpaceTimeNodeCl) IsEmpty() bool {
	return len(stn.nodeEvents) == 0
}

func (stn *SpaceTimeNodeCl) GetEventIds() []m3space.EventId {
	res := make([]m3space.EventId, len(stn.nodeEvents))
	for i, nodeEvt := range stn.nodeEvents {
		res[i] = nodeEvt.eventId
	}
	return res
}

func (stn *SpaceTimeNodeCl) HasRoot() bool {
	return stn.hasRoot
}

func (stn *SpaceTimeNodeCl) GetLastAccessed() m3space.DistAndTime {
	res := m3space.ZeroDistAndTime
	for _, nodeEvt := range stn.nodeEvents {
		if res < nodeEvt.creationTime {
			res = nodeEvt.creationTime
		}
	}
	return res
}

func (stn *SpaceTimeNodeCl) HowManyColors() uint8 {
	return m3util.CountTheOnes(stn.colorMask)
}

func (stn *SpaceTimeNodeCl) GetColorMask() uint8 {
	return stn.colorMask
}

func (stn *SpaceTimeNodeCl) GetStateString() string {
	return fmt.Sprintf("STN of %s last=%d root=%v colors=%d events=%v",
		stn.spaceTime.String(), stn.GetLastAccessed(), stn.hasRoot, stn.HowManyColors(), stn.GetEventIds())
}

func (stn *SpaceTimeNodeCl) VisitConnections(visitConn func(evtNode *SpaceTimeNodeEventCl, connIdx int, connId m3point.ConnectionId)) {
	for _, nodeEvt := range stn.nodeEvents {
		for connIdx, conn := range nodeEvt.trioDetails.GetConnections() {
			visitConn(nodeEvt, connIdx, conn.GetId())
		}
	}
}
