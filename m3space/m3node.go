package m3space

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3db"
	"github.com/freddy33/qsm-go/m3path"
	"github.com/freddy33/qsm-go/m3point"
)

type Node interface {
	fmt.Stringer
	GetNbEvents() int
	GetNbLatestEvents() int
	GetLatestEventIds() []EventID
	GetNbActiveEvents(space *Space) int
	GetActiveEventIds(space *Space) []EventID
	GetActiveLinks(space *Space) NodeLinkList
	GetPoint() *m3point.Point
	IsEmpty() bool
	IsEventAlreadyPresent(id EventID) bool
	GetPathNode(id EventID) m3path.PathNode

	GetAccessed(evt *Event) DistAndTime

	GetLastAccessed(space *Space) DistAndTime
	GetLatestAccessed(space *Space) m3path.PathNode

	GetEventDistFromCurrent(evt *Event) DistAndTime
	HasRoot() bool
	GetEventForPathNode(pathNode m3path.PathNode, space *Space) *Event
	IsPathNodeActive(pathNode m3path.PathNode, space *Space) bool
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

type NodeLink interface {
	GetConnId() m3point.ConnectionId
	GetSrc() m3point.Point
}

type NodeList []Node
type NodeLinkList []NodeLink

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
/*
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
	// TODO: put the point in the node and test by point
 */
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
	nbEvts := bn.GetNbEvents()
	if nbEvts == 0 {
		return "EMPTY NODE"
	}
	p := m3point.Origin
	for _, pn := range bn.pathNodes {
		if pn != nil {
			p = pn.P()
			break
		}
	}
	return fmt.Sprintf("Node-%v-%d", p, nbEvts)
}

func (bn *BaseNode) GetNbEvents() int {
	res := 0
	for _, pn := range bn.pathNodes {
		if pn != nil {
			res++
		}
	}
	return res
}

func (bn *BaseNode) GetPointPackData() *m3point.PointPackData {
	return m3point.GetPointPackData(bn.GetEnv())
}

func (bn *BaseNode) GetEnv() *m3db.QsmEnvironment {
	for _, pn := range bn.pathNodes {
		if pn != nil {
			return pn.GetPathContext().GetGrowthCtx().GetEnv()
		}
	}
	return nil
}

func (bn *BaseNode) GetNbLatestEvents() int {
	res := 0
	for _, pn := range bn.pathNodes {
		if pn != nil && pn.IsLatest() {
			res++
		}
	}
	return res
}

func (bn *BaseNode) GetLatestEventIds() []EventID {
	res := make([]EventID, bn.GetNbLatestEvents())
	idx := 0
	for id, pn := range bn.pathNodes {
		if pn != nil && pn.IsLatest() {
			res[idx] = EventID(id)
			idx++
		}
	}
	return res
}

func (bn *BaseNode) GetNbActiveEvents(space *Space) int {
	res := 0
	for id, pn := range bn.pathNodes {
		if pn != nil {
			evt := space.GetEvent(EventID(id))
			if bn.IsEventActive(evt) {
				res++
			}
		}
	}
	return res
}

func (bn *BaseNode) GetActiveEventIds(space *Space) []EventID {
	res := make([]EventID, 0, 3)
	for id, pn := range bn.pathNodes {
		if pn != nil {
			evt := space.GetEvent(EventID(id))
			if bn.IsEventActive(evt) {
				res = append(res, evt.id)
			}
		}
	}
	return res
}

func (bn *BaseNode) GetActiveLinks(space *Space) NodeLinkList {
	if space.EventOutgrowthThreshold <= DistAndTime(0) {
		// No chance of active links with activity at 0
		return NodeLinkList(nil)
	}
	res := NodeLinkList(make([]NodeLink, 0, 3))
	for id, pn := range bn.pathNodes {
		if pn != nil && !pn.IsRoot() {
			evt := space.GetEvent(EventID(id))
			if bn.IsEventActive(evt) {
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
	}
	return res
}

func (bn *BaseNode) GetPoint() *m3point.Point {
	nbEvts := bn.GetNbEvents()
	if nbEvts == 0 {
		return nil
	}
	for _, pn := range bn.pathNodes {
		if pn != nil {
			p := pn.P()
			return &p
		}
	}
	return nil
}

func (bn *BaseNode) IsEmpty() bool {
	return bn.GetNbEvents() == 0
}

func (bn *BaseNode) IsEventAlreadyPresent(id EventID) bool {
	return bn.pathNodes[id] != nil
}

func (bn *BaseNode) GetPathNode(id EventID) m3path.PathNode {
	return bn.pathNodes[id]
}

func (bn *BaseNode) GetAccessed(evt *Event) DistAndTime {
	return DistAndTime(bn.pathNodes[evt.id].D()) + evt.created
}

func (bn *BaseNode) GetLastAccessed(space *Space) DistAndTime {
	maxAccess := DistAndTime(0)
	for id, n := range bn.pathNodes {
		if n != nil {
			a := DistAndTime(n.D()) + space.GetEvent(EventID(id)).created
			//a := bn.GetAccessed(space.GetEvent(EventID(id)))
			if a > maxAccess {
				maxAccess = a
			}
		}
	}
	return maxAccess
}

func (bn *BaseNode) GetLatestAccessed(space *Space) m3path.PathNode {
	maxAccess := bn.GetLastAccessed(space)
	for id, n := range bn.pathNodes {
		if n != nil {
			if maxAccess == bn.GetAccessed(space.GetEvent(EventID(id))) {
				return n
			}
		}
	}
	Log.Errorf("trying to find latest for node %s but did not find max access time %d", bn.String(), maxAccess)
	return nil
}

func (bn *BaseNode) GetEventDistFromCurrent(evt *Event) DistAndTime {
	return evt.space.currentTime - bn.GetAccessed(evt)
}

func (bn *BaseNode) HasRoot() bool {
	for _, pn := range bn.pathNodes {
		if pn != nil && pn.IsRoot() {
			return true
		}
	}
	return false
}

func (bn *BaseNode) GetEventForPathNode(pathNode m3path.PathNode, space *Space) *Event {
	for _, evt := range space.events {
		if evt != nil && evt.pathContext != nil && evt.pathContext.GetId() == pathNode.GetPathContext().GetId() {
			return evt
		}
	}
	return nil
}

func (bn *BaseNode) IsPathNodeActive(pathNode m3path.PathNode, space *Space) bool {
	evt := bn.GetEventForPathNode(pathNode, space)
	if evt != nil {
		return bn.IsEventActive(evt)
	}
	return false
}

func (bn *BaseNode) IsEventActive(evt *Event) bool {
	if evt == nil {
		return false
	}
	pn := bn.GetPathNode(evt.id)
	if pn == nil {
		return false
	}
	if pn.IsRoot() {
		return true
	}
	return bn.GetEventDistFromCurrent(evt) <= evt.space.EventOutgrowthThreshold
}

func (bn *BaseNode) IsEventOld(evt *Event) bool {
	n := bn.GetPathNode(evt.id)
	if n == nil {
		return false
	}
	if n.IsRoot() {
		return false
	}
	return bn.GetEventDistFromCurrent(evt) >= evt.space.EventOutgrowthOldThreshold
}

func (bn *BaseNode) IsEventDead(evt *Event) bool {
	n := bn.GetPathNode(evt.id)
	if n == nil {
		return true
	}
	if n.IsRoot() {
		return false
	}
	return bn.GetEventDistFromCurrent(evt) >= evt.space.EventOutgrowthDeadThreshold
}

func (bn *BaseNode) IsActive(space *Space) bool {
	if bn.HasRoot() {
		return true
	}
	return space.currentTime-bn.GetLastAccessed(space) <= space.EventOutgrowthThreshold
}

func (bn *BaseNode) HowManyColors(space *Space) uint8 {
	return countOnes(bn.GetColorMask(space))
}

func (bn *BaseNode) GetColorMask(space *Space) uint8 {
	m := uint8(0)
	if bn.IsEmpty() {
		return m
	}
	for id, n := range bn.pathNodes {
		if n != nil {
			evt := space.GetEvent(EventID(id))
			if n.IsRoot() {
				return uint8(evt.color)
			}
			if bn.IsEventActive(evt) {
				m |= uint8(evt.color)
			}
		}
	}
	return m
}

func (bn *BaseNode) IsOld(space *Space) bool {
	if bn.IsEmpty() {
		return false
	}
	for id, n := range bn.pathNodes {
		if n != nil {
			if n.IsRoot() {
				return false
			}
			evt := space.GetEvent(EventID(id))
			if !(bn.IsEventOld(evt) || bn.IsEventDead(evt)) {
				return false
			}
		}
	}
	return true
}

func (bn *BaseNode) IsDead(space *Space) bool {
	if bn.IsEmpty() {
		return false
	}
	for id, n := range bn.pathNodes {
		if n != nil {
			if n.IsRoot() {
				return false
			}
			evt := space.GetEvent(EventID(id))
			if !bn.IsEventDead(evt) {
				return false
			}
		}
	}
	return true
}

func (bn *BaseNode) GetStateString(space *Space) string {
	nbEvts := bn.GetNbEvents()
	evtIds := make([]EventID, nbEvts)
	idx := 0
	for id, n := range bn.pathNodes {
		if n != nil {
			evtIds[idx] = EventID(id)
			idx++
		}
	}
	latest := bn.GetLatestAccessed(space)
	if bn.HasRoot() {
		return fmt.Sprintf("root node %v, %s = %v", latest.P(), latest.GetTrioIndex(), evtIds)
	}
	return fmt.Sprintf("node %v, %s = %v", latest.P(), latest.GetTrioIndex(), evtIds)
}

func (bn *BaseNode) addPathNode(id EventID, n m3path.PathNode) {
	if bn.IsEventAlreadyPresent(id) {
		Log.Errorf("trying to add path node %s for node %s ")
	}
	bn.pathNodes[id] = n
}

func (bn *BaseNode) GetConnections() *UniqueConnectionsList {
	usedConns := UniqueConnectionsList{}
	for _, pn := range bn.pathNodes {
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
	}
	return &usedConns
}

func (bn *BaseNode) HasFreeConnections(space *Space) bool {
	usedConns := bn.GetConnections()
	return usedConns.size() < space.MaxConnections
}

func (bn *BaseNode) IsAlreadyConnected(opn *BaseNode) bool {
	if bn.IsEmpty() || opn.IsEmpty() {
		return false
	}

	pnp := *bn.GetPoint()
	opnp := *opn.GetPoint()
	cd := bn.GetPointPackData().GetConnDetailsByPoints(pnp, opnp)
	if cd == nil || !cd.IsValid() {
		Log.Errorf("finding if 2 nodes already connected but not separated by possible connection (%v, %v)", pnp, opnp)
		return false
	}
	for id, n := range bn.pathNodes {
		if n != nil {
			td := n.GetTrioDetails()
			isNext := n.IsNext(td.GetConnectionIdx(cd.GetId()))
			on := opn.GetPathNode(EventID(id))
			if on != nil && isNext {
				return true
			}
		}
	}
	return false
}
