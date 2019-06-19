package m3space

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3path"
	"github.com/freddy33/qsm-go/m3point"
)

type Node interface {
	fmt.Stringer
	GetNbEvents() int
	GetNbActiveEvents() int
	GetActiveEventIds() []EventID
	GetPoint() *m3point.Point
	IsEmpty() bool
	IsEventAlreadyPresent(id EventID) bool
	GetPathNode(id EventID) m3path.PathNode

	GetAccessed(evt *Event) DistAndTime

	GetLastAccessed(space *Space) DistAndTime
	GetLatest(space *Space) m3path.PathNode

	GetEventDistFromCurrent(evt *Event) DistAndTime
	HasRoot() bool
	IsEventActive(evt *Event) bool
	IsEventOld(evt *Event) bool
	IsEventDead(evt *Event) bool

	HowManyColors(space *Space) uint8
	GetColorMask(space *Space) uint8

	IsActive(space *Space) bool
	IsOld(space *Space) bool
	IsDead(space *Space) bool

	GetStateString(space *Space) string

	addPathNode(id EventID, pn m3path.PathNode)
}

type UniqueConnectionsList struct {
	conns []m3point.ConnectionId
}

type PointNode struct {
	pathNodes []m3path.PathNode
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

func (cl *UniqueConnectionsList) addLink(pl m3path.PathLink) {
	if pl != nil && pl.HasDestination() {
		cl.add(pl.GetConnId())
	}
}

func (cl *UniqueConnectionsList) addFromLink(pl m3path.PathLink) {
	if pl != nil {
		cl.add(pl.GetConnId().GetNegId())
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
// PointNode Functions
/***************************************************************/

func (pn *PointNode) String() string {
	nbEvts := pn.GetNbEvents()
	if nbEvts == 0 {
		return "EMPTY NODE"
	}
	p := m3point.Origin
	for _, n := range pn.pathNodes {
		if n != nil && !n.IsEnd() {
			p = n.P()
			break
		}
	}
	return fmt.Sprintf("Node-%v-%d", p, nbEvts)
}

func (pn *PointNode) GetNbEvents() int {
	res := 0
	for _, n := range pn.pathNodes {
		if n != nil && !n.IsEnd() {
			res++
		}
	}
	return res
}

func (pn *PointNode) GetNbActiveEvents() int {
	res := 0
	for _, n := range pn.pathNodes {
		if n != nil && n.IsActive() {
			res++
		}
	}
	return res
}

func (pn *PointNode) GetActiveEventIds() []EventID {
	res := make([]EventID, pn.GetNbActiveEvents())
	idx := 0
	for id, n := range pn.pathNodes {
		if n != nil && n.IsActive() {
			res[idx] = EventID(id)
			idx++
		}
	}
	return res
}

func (pn *PointNode) GetPoint() *m3point.Point {
	nbEvts := pn.GetNbEvents()
	if nbEvts == 0 {
		return nil
	}
	for _, n := range pn.pathNodes {
		if n != nil && !n.IsEnd() {
			p := n.P()
			return &p
		}
	}
	return nil
}

func (pn *PointNode) IsEmpty() bool {
	return pn.GetNbEvents() == 0
}

func (pn *PointNode) IsEventAlreadyPresent(id EventID) bool {
	return pn.pathNodes[id] != nil && !pn.pathNodes[id].IsEnd()
}

func (pn *PointNode) GetPathNode(id EventID) m3path.PathNode {
	return pn.pathNodes[id]
}

func (pn *PointNode) GetAccessed(evt *Event) DistAndTime {
	return DistAndTime(pn.pathNodes[evt.id].D()) + evt.created
}

func (pn *PointNode) GetLastAccessed(space *Space) DistAndTime {
	maxAccess := DistAndTime(0)
	for id, n := range pn.pathNodes {
		if n != nil {
			a := DistAndTime(n.D()) + space.GetEvent(EventID(id)).created
			//a := pn.GetAccessed(space.GetEvent(EventID(id)))
			if a > maxAccess {
				maxAccess = a
			}
		}
	}
	return maxAccess
}

func (pn *PointNode) GetLatest(space *Space) m3path.PathNode {
	maxAccess := pn.GetLastAccessed(space)
	for id, n := range pn.pathNodes {
		if n != nil {
			if maxAccess == pn.GetAccessed(space.GetEvent(EventID(id))) {
				return n
			}
		}
	}
	Log.Errorf("trying to find latest for node %s but did not find max access time %d", pn.String(), maxAccess)
	return nil
}

func (pn *PointNode) GetEventDistFromCurrent(evt *Event) DistAndTime {
	return evt.space.currentTime - pn.GetAccessed(evt)
}

func (pn *PointNode) HasRoot() bool {
	for _, n := range pn.pathNodes {
		if n != nil && n.IsRoot() {
			return true
		}
	}
	return false
}

func (pn *PointNode) IsEventActive(evt *Event) bool {
	n := pn.GetPathNode(evt.id)
	if n == nil {
		return false
	}
	if n.IsRoot() {
		return true
	}
	return pn.GetEventDistFromCurrent(evt) >= evt.space.EventOutgrowthThreshold
}

func (pn *PointNode) IsEventOld(evt *Event) bool {
	n := pn.GetPathNode(evt.id)
	if n == nil {
		return false
	}
	if n.IsRoot() {
		return false
	}
	return pn.GetEventDistFromCurrent(evt) >= evt.space.EventOutgrowthOldThreshold
}

func (pn *PointNode) IsEventDead(evt *Event) bool {
	n := pn.GetPathNode(evt.id)
	if n == nil {
		return true
	}
	if n.IsRoot() {
		return false
	}
	return pn.GetEventDistFromCurrent(evt) >= evt.space.EventOutgrowthDeadThreshold
}

func (pn *PointNode) IsActive(space *Space) bool {
	if pn.HasRoot() {
		return true
	}
	return space.currentTime-pn.GetLastAccessed(space) >= space.EventOutgrowthThreshold
}

func (pn *PointNode) HowManyColors(space *Space) uint8 {
	return countOnes(pn.GetColorMask(space))
}

func (pn *PointNode) GetColorMask(space *Space) uint8 {
	m := uint8(0)
	if pn.IsEmpty() {
		return m
	}
	for id, n := range pn.pathNodes {
		if n != nil {
			evt := space.GetEvent(EventID(id))
			if n.IsRoot() {
				return uint8(evt.color)
			}
			if pn.IsEventActive(evt) {
				m |= uint8(evt.color)
			}
		}
	}
	return m
}

func (pn *PointNode) IsOld(space *Space) bool {
	if pn.IsEmpty() {
		return false
	}
	for id, n := range pn.pathNodes {
		if n != nil {
			if n.IsRoot() {
				return false
			}
			evt := space.GetEvent(EventID(id))
			if !(pn.IsEventOld(evt) || pn.IsEventDead(evt)) {
				return false
			}
		}
	}
	return true
}

func (pn *PointNode) IsDead(space *Space) bool {
	if pn.IsEmpty() {
		return false
	}
	for id, n := range pn.pathNodes {
		if n != nil && !n.IsEnd() {
			if n.IsRoot() {
				return false
			}
			evt := space.GetEvent(EventID(id))
			if !pn.IsEventDead(evt) {
				return false
			}
		}
	}
	return true
}

func (pn *PointNode) GetStateString(space *Space) string {
	nbEvts := pn.GetNbEvents()
	evtIds := make([]EventID, nbEvts)
	idx := 0
	for id, n := range pn.pathNodes {
		if n != nil && !n.IsEnd() {
			evtIds[idx] = EventID(id)
			idx++
		}
	}
	latest := pn.GetLatest(space)
	if pn.HasRoot() {
		return fmt.Sprintf("root node %v, %s = %v", latest.P(), latest.GetTrioIndex(), evtIds)
	}
	return fmt.Sprintf("node %v, %s = %v", latest.P(), latest.GetTrioIndex(), evtIds)
}

func (pn *PointNode) addPathNode(id EventID, n m3path.PathNode) {
	if pn.IsEventAlreadyPresent(id) {
		Log.Errorf("trying to add path node %s for node %s ")
	}
	pn.pathNodes[id] = n
}

func (pn *PointNode) GetConnections() *UniqueConnectionsList {
	usedConns := UniqueConnectionsList{}
	for _, n := range pn.pathNodes {
		if n != nil && !n.IsEnd() {
			max := 2
			if n.IsRoot() {
				max = 3
			}
			for j := 0; j < max; j++ {
				usedConns.addLink(n.GetNext(j))
			}
			if !n.IsRoot() {
				usedConns.addFromLink(n.GetFrom())
				usedConns.addFromLink(n.GetOtherFrom())
			}
		}
	}
	return &usedConns
}

func (pn *PointNode) HasFreeConnections(space *Space) bool {
	usedConns := pn.GetConnections()
	return usedConns.size() < space.MaxConnections
}

func (pn *PointNode) IsAlreadyConnected(opn *PointNode) bool {
	if pn.IsEmpty() || opn.IsEmpty() {
		return false
	}
	pnp := *pn.GetPoint()
	opnp := *opn.GetPoint()
	cd := m3point.GetConnDetailsByPoints(pnp, opnp)
	if cd == nil || !cd.IsValid() {
		Log.Errorf("finding if 2 nodes already connected but not separated by possible connection (%v, %v)", pnp, opnp)
		return false
	}
	for id, n := range pn.pathNodes {
		if n != nil && !n.IsEnd() {
			pl := n.GetNextConnection(cd.GetId())
			on := opn.GetPathNode(EventID(id))
			if on != nil && !on.IsEnd() && pl != nil && pl.GetSrc() == on {
				return true
			}
		}
	}
	return false
}
