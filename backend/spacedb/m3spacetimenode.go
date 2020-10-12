package spacedb

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/freddy33/qsm-go/model/m3space"
)

type SpaceTimeNode struct {
	spaceTime *SpaceTime
	pointId   int64
	head      *NodeEventList
}

type NodeEventList struct {
	cur  *EventNodeDb
	next *NodeEventList
}

func (nel *NodeEventList) Add(en *EventNodeDb) {
	if nel.cur == nil {
		nel.cur = en
	} else if nel.next == nil {
		nel.next = &NodeEventList{cur: en}
	} else {
		nel.next.Add(en)
	}
}

func (nel *NodeEventList) Size() int {
	if nel.cur == nil {
		return 0
	}
	if nel.next == nil {
		return 1
	}
	return 1 + nel.next.Size()
}

func countOnes(m uint8) uint8 {
	return ((m >> 7) & 1) + ((m >> 6) & 1) + ((m >> 5) & 1) + ((m >> 4) & 1) + ((m >> 3) & 1) + ((m >> 2) & 1) + ((m >> 1) & 1) + (m & 1)
}

/***************************************************************/
// SpaceTimeNode Functions
/***************************************************************/

func (stn *SpaceTimeNode) GetSpaceTime() m3space.SpaceTimeIfc {
	return stn.spaceTime
}

func (stn *SpaceTimeNode) GetPointId() int64 {
	return stn.pointId
}

func (stn *SpaceTimeNode) GetNbEventNodes() int {
	if stn.head == nil {
		return 0
	}
	return stn.head.Size()
}

func (stn *SpaceTimeNode) GetEventNodes() []m3space.NodeEventIfc {
	res := make([]m3space.NodeEventIfc, 0, stn.GetNbEventNodes())
	nel := stn.head
	for nel != nil {
		res = append(res, nel.cur)
		nel = nel.next
	}
	return res
}

func (stn *SpaceTimeNode) String() string {
	return fmt.Sprintf("Node-%d-%d", stn.pointId, stn.GetNbEventNodes())
}

func (stn *SpaceTimeNode) GetEventIds() []m3space.EventId {
	res := make([]m3space.EventId, 0, 3)
	nel := stn.head
	for nel != nil {
		if nel.cur != nil {
			res = append(res, nel.cur.GetEventId())
		}
		nel = nel.next
	}
	return res
}

func (stn *SpaceTimeNode) VisitConnections(visitConn func(evtNode *EventNodeDb, connId m3point.ConnectionId, linkId int64)) {
	pointData := stn.spaceTime.space.pointData
	nel := stn.head
	for nel != nil {
		// Need to be active on the next round also to have from link activated
		if nel.cur != nil {
			td := nel.cur.GetTrioDetails(pointData)
			for connIdx, linkId := range nel.cur.LinkIds {
				connId := td.Conns[connIdx].Id
				visitConn(nel.cur, connId, linkId)
			}
		}
		nel = nel.next
	}
}

func (stn *SpaceTimeNode) GetPoint() (*m3point.Point, error) {
	if stn.IsEmpty() {
		return nil, m3util.MakeQsmErrorf("cannot get point id %d since not event node set here at time=%d",
			stn.pointId, stn.spaceTime.GetCurrentTime())
	}
	return stn.head.cur.GetPoint()
}

func (stn *SpaceTimeNode) IsEmpty() bool {
	return stn.head == nil || stn.head.cur == nil
}

func (stn *SpaceTimeNode) IsEventAlreadyPresent(id m3space.EventId) bool {
	return stn.GetNodeEvent(id) != nil
}

func (stn *SpaceTimeNode) GetNodeEvent(id m3space.EventId) m3space.NodeEventIfc {
	nel := stn.head
	for nel != nil {
		if nel.cur.GetEventId() == id {
			return nel.cur
		}
		nel = nel.next
	}
	return nil
}

func (stn *SpaceTimeNode) GetLastAccessed() m3space.DistAndTime {
	maxAccess := m3space.ZeroDistAndTime
	nel := stn.head
	for nel != nil {
		a := nel.cur.GetCreationTime()
		if a > maxAccess {
			maxAccess = a
		}
		nel = nel.next
	}
	return maxAccess
}

func (stn *SpaceTimeNode) HasRoot() bool {
	nel := stn.head
	for nel != nil {
		if nel.cur.IsRoot() {
			return true
		}
		nel = nel.next
	}
	return false
}

func (stn *SpaceTimeNode) HowManyColors() uint8 {
	return countOnes(stn.GetColorMask())
}

func (stn *SpaceTimeNode) GetColorMask() uint8 {
	m := uint8(0)
	if stn.IsEmpty() {
		return m
	}
	nel := stn.head
	for nel != nil {
		ne := nel.cur
		if ne.IsRoot() {
			return uint8(ne.GetColor())
		}
		m |= uint8(ne.GetColor())
		nel = nel.next
	}
	return m
}

func (stn *SpaceTimeNode) GetStateString() string {
	evtIds := make([]m3space.EventId, 0, 3)
	nel := stn.head
	for nel != nil {
		evtIds = append(evtIds, nel.cur.GetEventId())
		nel = nel.next
	}
	name := "node"
	if stn.HasRoot() {
		name = "root node"
	}
	p, err := stn.GetPoint()
	if err != nil {
		Log.Error(err)
		return fmt.Sprintf("%s %d:FAIL:%v", name, stn.pointId, evtIds)
	}
	return fmt.Sprintf("%s %d:%v:%v", name, stn.pointId, *p, evtIds)
}

func (stn *SpaceTimeNode) GetConnections() []m3point.ConnectionId {
	res := make([]m3point.ConnectionId, 0, 5)
	stn.VisitConnections(func(evtNode *EventNodeDb, connId m3point.ConnectionId, linkId int64) {
		for _, cId := range res {
			if cId == connId {
				return
			}
		}
		res = append(res, connId)
	})
	return res
}

func (stn *SpaceTimeNode) HasFreeConnections() bool {
	// TODO: Use GetMaxNodesPerPoint() and GetMaxTriosPerPoint()
	return true
}
