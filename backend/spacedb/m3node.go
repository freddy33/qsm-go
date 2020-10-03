package spacedb

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3path"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/freddy33/qsm-go/model/m3space"
	"sync/atomic"
	"unsafe"
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

func (stn *SpaceTimeNode) GetEventNodes() []m3space.EventNodeIfc {
	res := make([]m3space.EventNodeIfc, 0, stn.GetNbEventNodes())
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

func (stn *SpaceTimeNode) GetActiveLinks() NodeLinkList {

	space := spaceTime.GetSpace()
	if space.GetActivePathNodeThreshold() <= DistAndTime(0) {
		// No chance of active links with activity at 0
		return NodeLinkList(nil)
	}
	res := NodeLinkList(make([]NodeLink, 0, 3))
	nel := stn.head
	for nel != nil {
		// Need to be active on the next round also to have from link activated
		if nel.cur.IsActiveNext(spaceTime) {
			pn, err := nel.cur.GetPathNode()
			if err != nil {
				Log.Error(err)
			} else {
				td := pn.GetTrioDetails()
				for i := 0; i < m3path.NbConnections; i++ {
					if pn.IsFrom(i) {
						conn := td.GetConnections()[i]
						fromP := pn.P().Add(conn.Vector)
						nl := BaseNodeLink{
							conn.GetNegId(),
							fromP,
						}
						res = append(res, &nl)
					}
				}
			}
		}
		nel = nel.next
	}
	return res
}

func (stn *SpaceTimeNode) GetPoint() (*m3point.Point, error) {
	if stn.head == nil || stn.head.cur == nil {
		return nil, m3util.MakeQsmErrorf("cannot get point id %d since not event node set here at time=%d",
			stn.pointId, stn.spaceTime.GetCurrentTime())
	}
	return stn.head.cur.GetPoint()
}

func (stn *SpaceTimeNode) IsEmpty() bool {
	return stn.head == nil
}

func (stn *SpaceTimeNode) IsEventAlreadyPresent(id m3space.EventId) bool {
	return stn.GetNodeEvent(id) != nil
}

func (stn *SpaceTimeNode) GetNodeEvent(id m3space.EventId) m3space.EventNodeIfc {
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
	maxAccess := m3space.DistAndTime(0)
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

func (stn *SpaceTimeNode) GetEventForPathNode(pathNode m3path.PathNode, space SpaceIfc) EventIfc {
	nel := stn.head
	for nel != nil {
		if nel.cur.pathNodeId == pathNode.GetId() {
			return space.GetEvent(nel.cur.evtId)
		}
		nel = nel.next
	}
	return nil
}

func (stn *SpaceTimeNode) IsPathNodeActive(pathNode m3path.PathNode, spaceTime SpaceTimeIfc) bool {
	pnId := pathNode.GetId()
	nel := stn.head
	for nel != nil {
		if nel.cur.pathNodeId == pnId {
			if nel.cur.IsActive(spaceTime) {
				return true
			}
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

// Node is active if any node events it has is active
func (stn *SpaceTimeNode) IsActive(spaceTime SpaceTimeIfc) bool {
	nel := stn.head
	for nel != nil {
		if nel.cur.IsActive(spaceTime) {
			return true
		}
		nel = nel.next
	}
	return false
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

func (stn *SpaceTimeNode) addPathNode(id EventId, pn m3path.PathNode, space SpaceIfc) {
	pnId := pn.GetId()
	if pnId < int64(0) {
		Log.Fatalf("trying to add non saved path node %d for event %d", pnId, id)
		return
	}
	evt := space.GetEvent(id)
	if evt == nil {
		Log.Fatalf("trying to add path node %d for non existing event %d", pnId, id)
		return
	}
	if stn.IsEventAlreadyPresent(id) {
		Log.Errorf("trying to add path node %d for event %d but already one %d", pnId, id, stn.GetNodeEvent(id).GetPathNodeId())
		return
	}
	// Insert at end of linked list without lock
	newNE := BaseNodeEvent{id, pnId, DistAndTime(pn.D()) + evt.GetCreationTime(), pn}
	newEl := NodeEventList{&newNE, nil}
	if stn.head == nil {
		success := atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&stn.head)), unsafe.Pointer(nil), unsafe.Pointer(&newEl))
		if !success {
			stn.addPathNode(id, pn, space)
		}
	} else {
		prev := stn.head
		tail := prev.next
		for tail != nil {
			prev = tail
			tail = tail.next
		}
		success := atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&prev.next)), unsafe.Pointer(nil), unsafe.Pointer(&newEl))
		if !success {
			stn.addPathNode(id, pn, space)
		}
	}
}

func (stn *SpaceTimeNode) GetConnections() *UniqueConnectionsList {
	// TODO: This is wrong. Should be base on event node at space time not path node
	usedConns := UniqueConnectionsList{}
	nel := stn.head
	for nel != nil {
		pn, err := nel.cur.GetPathNode()
		if err != nil {
			Log.Error(err)
		}
		if pn != nil {
			td := pn.GetTrioDetails()
			for i := 0; i < m3path.NbConnections; i++ {
				if pn.IsFrom(i) {
					usedConns.add(td.GetConnections()[i].GetNegId())
				}
				if pn.IsNext(i) {
					usedConns.add(td.GetConnections()[i].GetId())
				}
			}
		}
		nel = nel.next
	}
	return &usedConns
}

func (stn *SpaceTimeNode) HasFreeConnections(spaceTime SpaceTimeIfc) bool {
	// TODO: Use GetMaxPathNodesPerPoint() and GetMaxTriosPerPoint()
	usedConns := stn.GetConnections()
	return usedConns.size() < spaceTime.GetSpace().GetMaxTriosPerPoint()*3
}

func (stn *SpaceTimeNode) IsAlreadyConnected(opn *SpaceTimeNode, spaceTime SpaceTimeIfc) bool {
	if stn.IsEmpty() || opn.IsEmpty() {
		return false
	}

	p1, err := stn.GetPoint()
	if err != nil {
		Log.Error(err)
	}
	p2, err := opn.GetPoint()
	if err != nil {
		Log.Error(err)
	}

	cd := spaceTime.GetPointPackData().GetConnDetailsByPoints(*p1, *p2)
	if cd == nil || !cd.IsValid() {
		Log.Errorf("finding if 2 nodes already connected but not separated by possible connection (%v, %v)", p1, p2)
		return false
	}
	nel := stn.head
	for nel != nil {
		pn1, err := nel.cur.GetPathNode()
		if err != nil {
			Log.Error(err)
		}
		pn2, err := opn.GetPathNode(nel.cur.evtId)
		if err != nil {
			Log.Error(err)
		}
		if pn1 != nil && pn2 != nil {
			td := pn1.GetTrioDetails()
			return pn1.IsNext(td.GetConnectionIdx(cd.GetId()))
		}
		nel = nel.next
	}
	return false
}
