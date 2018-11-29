package m3space

import "fmt"

type AccessedEventID struct {
	id     EventID
	access TickTime
}

type Node struct {
	Pos              *Point
	root             bool
	accessedEventIDS []AccessedEventID
	connections      []*Connection
}

type SavedNode struct {
	root             bool
	accessedEventIDS []AccessedEventID
	connections      []int8
}

type Connection struct {
	N1, N2 *Node
}

/***************************************************************/
// Node Functions
/***************************************************************/

func NewNode(p *Point) *Node {
	n := Node{}
	n.Pos = p
	return &n
}

func (node *Node) SetRoot(id EventID, time TickTime) {
	node.root = true
	node.accessedEventIDS = make([]AccessedEventID, 1)
	node.accessedEventIDS[0] = AccessedEventID{id, time,}
}

func (node *Node) HasFreeConnections(space *Space) bool {
	return node.connections == nil || len(node.connections) < space.MaxConnections
}

func (node *Node) AddConnection(conn *Connection, space *Space) int {
	if !node.HasFreeConnections(space) {
		return -1
	}
	if node.connections == nil {
		node.connections = make([]*Connection, 0, 3)
	}
	index := len(node.connections)
	node.connections = append(node.connections, conn)
	return index
}

func (node *Node) IsAlreadyConnected(otherNode *Node) bool {
	for _, conn := range node.connections {
		if conn.IsConnectedTo(otherNode) {
			return true
		}
	}
	return false
}

func (node *Node) IsRoot() bool {
	return node.root
}

func (node *Node) GetLastAccessed() TickTime {
	bestTime := TickTime(0)
	for _, ae := range node.accessedEventIDS {
		if ae.access > bestTime {
			bestTime = ae.access
		}
	}
	return bestTime
}

func (ae AccessedEventID) IsActive(space *Space) bool {
	return Distance(space.currentTime-ae.access) <= space.EventOutgrowthThreshold
}

func (node *Node) IsActive(space *Space) bool {
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

func (node *Node) HowManyColors(space *Space) uint8 {
	return countOnes(node.GetColorMask(space))
}

func countOnes(m uint8) uint8 {
	return ((m >> 7) & 1) + ((m >> 6) & 1) + ((m >> 5) & 1) + ((m >> 4) & 1) + ((m >> 3) & 1) + ((m >> 2) & 1) + ((m >> 1) & 1) + (m & 1)
}

func (node *Node) GetColorMask(space *Space) uint8 {
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

func (node *Node) CanReceiveOutgrowth(newPosEo *NewPossibleOutgrowth) bool {
	if !node.CanReceiveEvent(newPosEo.event.id) {
		return false
	}
	return true
}

func (node *Node) CanReceiveEvent(id EventID) bool {
	for _, ae := range node.accessedEventIDS {
		if ae.id == id {
			return false
		}
	}
	return true
}

func (node *Node) AddOutgrowth(id EventID, time TickTime) {
	node.accessedEventIDS = append(node.accessedEventIDS, AccessedEventID{id, time,})
}

func (node *Node) IsOld(space *Space) bool {
	if node.IsRoot() {
		return false
	}
	return Distance(space.currentTime-node.GetLastAccessed()) >= space.EventOutgrowthOldThreshold
}

func (node *Node) String() string {
	return fmt.Sprintf("%v:%d:%d", *(node.Pos), len(node.connections), len(node.accessedEventIDS))
}

func (node *Node) GetStateString() string {
	connIds := make([]string, len(node.connections))
	for i, conn := range node.connections {
		var connVect Point
		if conn.N1 == node {
			connVect = conn.N2.Pos.Sub(*node.Pos)
		} else if conn.N2 == node {
			connVect = conn.N1.Pos.Sub(*node.Pos)
		} else {
			Log.Error("Connection", conn, "in list of node", node, "but not part of it?")
		}
		connIds[i] = AllConnectionsPossible[connVect].GetName()
	}

	if node.root {
		return fmt.Sprintf("%v: root %v, %v", *(node.Pos), node.accessedEventIDS, connIds)
	}
	return fmt.Sprintf("%v: %v, %v", *(node.Pos), node.accessedEventIDS, connIds)
}

/***************************************************************/
// Connection Functions
/***************************************************************/

func (conn *Connection) IsConnectedTo(node *Node) bool {
	return conn.N1 == node || conn.N2 == node
}

func (conn *Connection) GetColorMask(space *Space) uint8 {
	// Connection color mask of all event outgrowth that match
	if conn.N1 != nil && conn.N2 != nil {
		return conn.N1.GetColorMask(space) & conn.N2.GetColorMask(space)
	}
	return uint8(0)
}

func (conn *Connection) HowManyColors(space *Space) uint8 {
	return countOnes(conn.GetColorMask(space))
}

func (conn *Connection) IsOld(space *Space) bool {
	return conn.N1.IsOld(space) && conn.N2.IsOld(space)
}
