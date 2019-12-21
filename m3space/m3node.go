package m3space

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3path"
	"github.com/freddy33/qsm-go/m3point"
	"sync/atomic"
	"unsafe"
)

type UniqueConnectionsList struct {
	conns []m3point.ConnectionId
}

type BaseNodeLink struct {
	connId m3point.ConnectionId
	point  m3point.Point
}

func (bnl *BaseNodeLink) GetConnId() m3point.ConnectionId {
	return bnl.connId
}

func (bnl *BaseNodeLink) GetSrc() m3point.Point {
	return bnl.point
}

type BaseNode struct {
	p    m3point.Point
	head *NodeEventList
}

func countOnes(m uint8) uint8 {
	return ((m >> 7) & 1) + ((m >> 6) & 1) + ((m >> 5) & 1) + ((m >> 4) & 1) + ((m >> 3) & 1) + ((m >> 2) & 1) + ((m >> 1) & 1) + (m & 1)
}

/***************************************************************/
// UniqueConnectionsList Functions
/***************************************************************/

func (cl *UniqueConnectionsList) size() int {
	return len(cl.conns)
}

func (cl *UniqueConnectionsList) exist(connId m3point.ConnectionId) bool {
	for _, cId := range cl.conns {
		if cId == connId {
			return true
		}
	}
	return false
}

func (cl *UniqueConnectionsList) addLink(nl NodeLink) {
	if nl != nil {
		cl.add(nl.GetConnId())
	}
}

func (cl *UniqueConnectionsList) addFromLink(nl NodeLink) {
	if nl != nil {
		cl.add(nl.GetConnId().GetNegId())
	}
}

func (cl *UniqueConnectionsList) add(connId m3point.ConnectionId) {
	if !cl.exist(connId) {
		if cl.conns == nil {
			cl.conns = make([]m3point.ConnectionId, 1, 3)
			cl.conns[0] = connId
		} else {
			cl.conns = append(cl.conns, connId)
		}
	}
}

/***************************************************************/
// NodeLinkList Functions
/***************************************************************/

func (nl *NodeList) addNode(newNode Node) {
	if newNode == nil {
		return
	}
	p := newNode.GetPoint()
	if p == nil {
		return
	}
	for _, n := range *nl {
		if n == newNode {
			return
		}
	}

	*nl = append(*nl, newNode)
}

/***************************************************************/
// NodeLinkList Functions
/***************************************************************/

func (pll *NodeLinkList) addAll(links NodeLinkList) {
	for _, pl := range links {
		*pll = append(*pll, pl)
	}
}

/***************************************************************/
// BaseNode Functions
/***************************************************************/

func (bn *BaseNode) String() string {
	return fmt.Sprintf("Node-%v-%d", bn.p, bn.GetNbEvents())
}

func (bn *BaseNode) GetNbEvents() int {
	res := 0
	nel := bn.head
	for nel != nil {
		res++
		nel = nel.next
	}
	return res
}

func (bn *BaseNode) GetNbLatestEvents() int {
	res := 0
	nel := bn.head
	for nel != nil {
		if nel.cur.IsLatest() {
			res++
		}
		nel = nel.next
	}
	return res
}

func (bn *BaseNode) GetLatestEventIds() []EventID {
	res := make([]EventID, 0, 3)
	nel := bn.head
	for nel != nil {
		if nel.cur.IsLatest() {
			res = append(res, nel.cur.GetEventId())
		}
		nel = nel.next
	}
	return res
}

func (bn *BaseNode) GetNbActiveEvents(space *Space) int {
	res := 0
	nel := bn.head
	for nel != nil {
		if nel.cur.IsActive(space) {
			res++
		}
		nel = nel.next
	}
	return res
}

func (bn *BaseNode) GetActiveEventIds(space *Space) []EventID {
	res := make([]EventID, 0, 3)
	nel := bn.head
	for nel != nil {
		if nel.cur.IsActive(space) {
			res = append(res, nel.cur.GetEventId())
		}
		nel = nel.next
	}
	return res
}

func (bn *BaseNode) GetActiveLinks(space *Space) NodeLinkList {
	if space.EventOutgrowthThreshold <= DistAndTime(0) {
		// No chance of active links with activity at 0
		return NodeLinkList(nil)
	}
	res := NodeLinkList(make([]NodeLink, 0, 3))
	nel := bn.head
	for nel != nil {
		// Need to be active on the next round also to have from link activated
		if nel.cur.IsActiveNext(space) {
			pn := nel.cur.GetPathNode()
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
		nel = nel.next
	}
	return res
}

func (bn *BaseNode) GetPoint() *m3point.Point {
	return &bn.p
}

func (bn *BaseNode) IsEmpty() bool {
	return bn.head == nil
}

func (bn *BaseNode) IsEventAlreadyPresent(id EventID) bool {
	return bn.GetNodeEvent(id) != nil
}

func (bn *BaseNode) GetNodeEvent(id EventID) NodeEvent {
	nel := bn.head
	for nel != nil {
		if nel.cur.evtId == id {
			return nel.cur
		}
		nel = nel.next
	}
	return nil
}

// Deprecated
func (bn *BaseNode) GetPathNode(id EventID) m3path.PathNode {
	res := bn.GetNodeEvent(id)
	if res != nil {
		return res.GetPathNode()
	}
	return nil
}

// Deprecated
func (bn *BaseNode) GetAccessed(evt *Event) DistAndTime {
	res := bn.GetNodeEvent(evt.id)
	if res != nil {
		return res.GetAccessedTime()
	}
	Log.Errorf("Trying to retrieve access time for event %d and base node %s but not accessed yet", evt.id, bn.String())
	return DistAndTime(0)
}

func (bn *BaseNode) GetLastAccessed(space *Space) DistAndTime {
	maxAccess := DistAndTime(0)
	nel := bn.head
	for nel != nil {
		a := nel.cur.GetAccessedTime()
		if a > maxAccess {
			maxAccess = a
		}
		nel = nel.next
	}
	return maxAccess
}

// Deprecated
func (bn *BaseNode) GetEventDistFromCurrent(evt *Event) DistAndTime {
	return evt.space.currentTime - bn.GetAccessed(evt)
}

func (bn *BaseNode) HasRoot(space *Space) bool {
	nel := bn.head
	for nel != nil {
		evt := space.GetEvent(nel.cur.evtId)
		if nel.cur.IsRoot(evt) {
			return true
		}
		nel = nel.next
	}
	return false
}

func (bn *BaseNode) GetEventForPathNode(pathNode m3path.PathNode, space *Space) *Event {
	nel := bn.head
	for nel != nil {
		if nel.cur.pathNodeId == pathNode.GetId() {
			return space.GetEvent(nel.cur.evtId)
		}
		nel = nel.next
	}
	return nil
}

func (bn *BaseNode) IsPathNodeActive(pathNode m3path.PathNode, space *Space) bool {
	pnId := pathNode.GetId()
	nel := bn.head
	for nel != nil {
		if nel.cur.pathNodeId == pnId {
			if nel.cur.IsActive(space) {
				return true
			}
		}
		nel = nel.next
	}
	return false
}

func (bn *BaseNode) HowManyColors(space *Space) uint8 {
	return countOnes(bn.GetColorMask(space))
}

func (bn *BaseNode) GetColorMask(space *Space) uint8 {
	m := uint8(0)
	if bn.IsEmpty() {
		return m
	}
	nel := bn.head
	for nel != nil {
		ne := nel.cur
		evt := space.GetEvent(ne.evtId)
		if ne.IsRoot(evt) {
			return uint8(evt.color)
		}
		if ne.IsActive(space) {
			m |= uint8(evt.color)
		}
		nel = nel.next
	}
	return m
}

// Node is active if any node events it has is active
func (bn *BaseNode) IsActive(space *Space) bool {
	nel := bn.head
	for nel != nil {
		if nel.cur.IsActive(space) {
			return true
		}
		nel = nel.next
	}
	return false
}

// Node is old if all node events it has are old. Empty node are dead and so also old
func (bn *BaseNode) IsOld(space *Space) bool {
	if bn.IsEmpty() {
		return true
	}
	nel := bn.head
	for nel != nil {
		if !nel.cur.IsOld(space) {
			return false
		}
		nel = nel.next
	}
	return true
}

// Node is dead if all node events it has are dead. Empty node are dead
func (bn *BaseNode) IsDead(space *Space) bool {
	if bn.IsEmpty() {
		return true
	}
	nel := bn.head
	for nel != nil {
		if !nel.cur.IsDead(space) {
			return false
		}
		nel = nel.next
	}
	return true
}

func (bn *BaseNode) GetStateString(space *Space) string {
	evtIds := make([]EventID, 0, 3)
	nel := bn.head
	for nel != nil {
		evtIds = append(evtIds, nel.cur.evtId)
		nel = nel.next
	}
	if bn.HasRoot(space) {
		return fmt.Sprintf("root node %v:%v", bn.p, evtIds)
	}
	return fmt.Sprintf("node %v: %v", bn.p, evtIds)
}

func (bn *BaseNode) addPathNode(id EventID, pn m3path.PathNode, space *Space) {
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
	if bn.IsEventAlreadyPresent(id) {
		Log.Errorf("trying to add path node %d for event %d but already one %d", pnId, id, bn.GetNodeEvent(id).GetPathNodeId())
		return
	}
	// Insert at end of linked list without lock
	newNE := BaseNodeEvent{id, pnId, DistAndTime(pn.D()) + evt.created, pn}
	newEl := NodeEventList{&newNE, nil}
	if bn.head == nil {
		success := atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&bn.head)), unsafe.Pointer(nil), unsafe.Pointer(&newEl))
		if !success {
			bn.addPathNode(id, pn, space)
		}
	} else {
		prev := bn.head
		tail := prev.next
		for tail != nil {
			prev = tail
			tail = tail.next
		}
		success := atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&prev.next)), unsafe.Pointer(nil), unsafe.Pointer(&newEl))
		if !success {
			bn.addPathNode(id, pn, space)
		}
	}
}

func (bn *BaseNode) GetConnections() *UniqueConnectionsList {
	usedConns := UniqueConnectionsList{}
	nel := bn.head
	for nel != nil {
		pn := nel.cur.GetPathNode()
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

func (bn *BaseNode) HasFreeConnections(space *Space) bool {
	usedConns := bn.GetConnections()
	return usedConns.size() < space.MaxConnections
}

func (bn *BaseNode) IsAlreadyConnected(opn *BaseNode, space *Space) bool {
	if bn.IsEmpty() || opn.IsEmpty() {
		return false
	}

	p1 := *bn.GetPoint()
	p2 := *opn.GetPoint()

	cd := space.GetPointPackData().GetConnDetailsByPoints(p1, p2)
	if cd == nil || !cd.IsValid() {
		Log.Errorf("finding if 2 nodes already connected but not separated by possible connection (%v, %v)", p1, p2)
		return false
	}
	nel := bn.head
	for nel != nil {
		pn1 := nel.cur.GetPathNode()
		pn2 := opn.GetPathNode(nel.cur.evtId)
		if pn1 != nil && pn2 != nil {
			td := pn1.GetTrioDetails()
			return pn1.IsNext(td.GetConnectionIdx(cd.GetId()))
		}
		nel = nel.next
	}
	return false
}
