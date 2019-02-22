package m3space

import "fmt"

type AccessedEventID struct {
	id     EventID
	access TickTime
}

type Node interface {
	IsRoot() bool
	GetLastAccessed() TickTime
	IsActive(space *Space) bool
	HowManyColors(space *Space) uint8
	GetColorMask(space *Space) uint8
	IsEventAlreadyPresent(id EventID) bool
	IsOld(space *Space) bool
	IsDead(space *Space) bool
	GetStateString() string
}

type SavedNode struct {
	root             bool
	accessedEventIDS []AccessedEventID
	connections      []int8
}

type ActiveNode struct {
	Pos Point
	SavedNode
}

type Connection struct {
	Id     int8
	P1, P2 Point
}

func countOnes(m uint8) uint8 {
	return ((m >> 7) & 1) + ((m >> 6) & 1) + ((m >> 5) & 1) + ((m >> 4) & 1) + ((m >> 3) & 1) + ((m >> 2) & 1) + ((m >> 1) & 1) + (m & 1)
}

func (ae AccessedEventID) IsActive(space *Space) bool {
	return Distance(space.currentTime-ae.access) <= space.EventOutgrowthThreshold
}

/***************************************************************/
// Node Functions
/***************************************************************/

func NewNode(p Point) *ActiveNode {
	n := ActiveNode{}
	n.Pos = p
	return &n
}

func (s *SavedNode) ConvertToActive(p Point) *ActiveNode {
	n := ActiveNode{}
	n.Pos = p
	n.root = s.root
	n.accessedEventIDS = s.accessedEventIDS
	n.connections = s.connections
	return &n
}

func (a *ActiveNode) ConvertToSaved() *SavedNode {
	s := SavedNode{}
	s.root = a.root
	s.accessedEventIDS = a.accessedEventIDS
	s.connections = a.connections
	return &s
}

func (node *ActiveNode) SetRoot(id EventID, time TickTime) {
	node.root = true
	node.accessedEventIDS = make([]AccessedEventID, 1)
	node.accessedEventIDS[0] = AccessedEventID{id, time}
}

func (node *ActiveNode) HasFreeConnections(space *Space) bool {
	return node.connections == nil || len(node.connections) < space.MaxConnections
}

func (node *ActiveNode) AddConnection(conn *Connection, space *Space) int {
	if !node.HasFreeConnections(space) {
		return -1
	}
	if node.connections == nil {
		node.connections = make([]int8, 0, 3)
	}
	index := len(node.connections)
	if node.Pos == conn.P1 {
		node.connections = append(node.connections, conn.Id)
	} else if node.Pos == conn.P2 {
		node.connections = append(node.connections, -conn.Id)
	} else {
		Log.Errorf("Trying to add connection %v that does connect to node %v", *conn, *node)
		return -1
	}
	return index
}

func (node *ActiveNode) IsAlreadyConnected(otherNode *ActiveNode) bool {
	bv := MakeVector(node.Pos, otherNode.Pos)
	cd, ok := AllConnectionsPossible[bv]
	if !ok {
		Log.Errorf("Cannot determine an already connected nodes P1=%v P2=%v that is not reachable by a base connection %v",
			node.Pos, otherNode.Pos, bv)
		return false
	}
	for _, conn := range node.connections {
		if conn == cd.Id {
			return true
		}
	}
	return false
}

func (node *ActiveNode) CanReceiveOutgrowth(newPosEo *NewPossibleOutgrowth) bool {
	if !node.IsEventAlreadyPresent(newPosEo.event.id) {
		return false
	}
	return true
}

func (node *ActiveNode) AddOutgrowth(id EventID, time TickTime) {
	node.accessedEventIDS = append(node.accessedEventIDS, AccessedEventID{id, time})
}

/***************************************************************/
// Connection Functions
/***************************************************************/

func (conn *Connection) GetConnId() int8 {
	return conn.Id
}

func (conn *Connection) GetConnectionDetails() ConnectionDetails {
	return AllConnectionsIds[conn.Id]
}

func (conn *Connection) IsConnectedTo(point Point) bool {
	return conn.P1 == point || conn.P2 == point
}

func (conn *Connection) GetColorMask(space *Space) uint8 {
	n1 := space.GetNode(conn.P1)
	n2 := space.GetNode(conn.P2)
	// Connection color mask of all event outgrowth that match
	if n1 != nil && n2 != nil {
		return n1.GetColorMask(space) & n2.GetColorMask(space)
	}
	return uint8(0)
}

func (conn *Connection) HowManyColors(space *Space) uint8 {
	return countOnes(conn.GetColorMask(space))
}

func (conn *Connection) IsOld(space *Space) bool {
	n1 := space.GetNode(conn.P1)
	n2 := space.GetNode(conn.P2)
	if n1 != nil && n2 != nil {
		return n1.IsOld(space) && n2.IsOld(space)
	}
	return false
}

func (space *Space) makeConnection(n1, n2 *ActiveNode) *Connection {
	if !n1.HasFreeConnections(space) {
		Log.Trace("Node 1", n1, "does not have free connections")
		return nil
	}
	if !n2.HasFreeConnections(space) {
		Log.Trace("Node 2", n2, "does not have free connections")
		return nil
	}
	if n1.IsAlreadyConnected(n2) {
		Log.Trace("Connection between 2 points", n1.Pos, n2.Pos, "already connected!")
		return nil
	}

	d := DS(n1.Pos, n2.Pos)
	if !(d == 1 || d == 2 || d == 3 || d == 5) {
		Log.Error("Connection between 2 points", n1.Pos, n2.Pos, "that are not 1, 2, 3 or 5 DS away!")
		return nil
	}
	// All good create connection
	bv := MakeVector(n1.Pos, n2.Pos)
	cd := AllConnectionsPossible[bv]
	c1 := &Connection{cd.GetIntId(), n1.Pos, n2.Pos}
	space.activeConnections = append(space.activeConnections, c1)
	n1done := n1.AddConnection(c1, space)
	c2 := &Connection{-cd.GetIntId(), n2.Pos, n1.Pos}
	n2done := n2.AddConnection(c2, space)
	if n1done < 0 || n2done < 0 {
		Log.Error("Node1 connection association", n1done, "or Node2", n2done, "did not happen!!")
		return nil
	}
	return c1
}

/***************************************************************/
// Saved Node Functions
/***************************************************************/

func (node *SavedNode) IsRoot() bool {
	return node.root
}

func (node *SavedNode) GetLastAccessed() TickTime {
	bestTime := TickTime(0)
	for _, ae := range node.accessedEventIDS {
		if ae.access > bestTime {
			bestTime = ae.access
		}
	}
	return bestTime
}

func (node *SavedNode) IsActive(space *Space) bool {
	if node.IsRoot() {
		return true
	}
	for _, ae := range node.accessedEventIDS {
		if ae.IsActive(space) {
			return true
		}
	}
	return false
}

func (node *SavedNode) HowManyColors(space *Space) uint8 {
	return countOnes(node.GetColorMask(space))
}

func (node *SavedNode) GetColorMask(space *Space) uint8 {
	if node.root {
		return uint8(space.events[node.accessedEventIDS[0].id].color)
	}
	m := uint8(0)
	for _, ae := range node.accessedEventIDS {
		if ae.IsActive(space) {
			m |= uint8(space.events[ae.id].color)
		}
	}
	return m
}

func (node *SavedNode) IsEventAlreadyPresent(id EventID) bool {
	for _, ae := range node.accessedEventIDS {
		if ae.id == id {
			return false
		}
	}
	return true
}

func (node *SavedNode) IsOld(space *Space) bool {
	if node.IsRoot() {
		return false
	}
	return Distance(space.currentTime-node.GetLastAccessed()) >= space.EventOutgrowthOldThreshold
}

func (node *SavedNode) IsDead(space *Space) bool {
	if node.IsRoot() {
		return false
	}
	return Distance(space.currentTime-node.GetLastAccessed()) >= space.EventOutgrowthDeadThreshold
}

func (node *SavedNode) String() string {
	return fmt.Sprintf("%s:%d:%d", "Saved", len(node.connections), len(node.accessedEventIDS))
}

func (node *SavedNode) GetStateString() string {
	connIds := make([]string, len(node.connections))
	for i, connId := range node.connections {
		connIds[i] = AllConnectionsIds[connId].GetName()
	}
	if node.root {
		return fmt.Sprintf("%s: root %v, %v", "Saved", node.accessedEventIDS, connIds)
	}
	return fmt.Sprintf("%s: %v, %v", "Saved", node.accessedEventIDS, connIds)
}
