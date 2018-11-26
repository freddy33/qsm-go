package m3space

import "fmt"

type Node struct {
	Pos          *Point
	outgrowths   []*EventOutgrowth
	connections  []*Connection
}

type Connection struct {
	N1, N2 *Node
}

/***************************************************************/
// Node Functions
/***************************************************************/

func (node *Node) HasFreeConnections(space *Space) bool {
	return node.connections == nil || len(node.connections) < space.MaxConnections
}

func (node *Node) AddConnection(conn *Connection, space *Space) int {
	if !node.HasFreeConnections(space) {
		return -1
	}
	if node.connections == nil {
		node.connections = make ([]*Connection, 0, 3)
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
	for _, eo := range node.outgrowths {
		if eo.IsRoot() {
			return true
		}
	}
	return false
}

func (node *Node) IsActive(threshold Distance) bool {
	for _, eo := range node.outgrowths {
		if eo.IsActive(threshold) {
			return true
		}
	}
	return false
}

func (node *Node) HowManyColors(threshold Distance) uint8 {
	return countOnes(node.GetColorMask(threshold))
}

func countOnes(m uint8) uint8 {
	return ((m>>7)&1) + ((m>>6)&1) + ((m>>5)&1) + ((m>>4)&1) + ((m>>3)&1) + ((m>>2)&1) + ((m>>1)&1) + (m&1)
}

func (node *Node) GetColorMask(threshold Distance) uint8 {
	m := uint8(0)
	for _, eo := range node.outgrowths {
		if eo.IsActive(threshold) {
			m |= uint8(eo.event.color)
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
	if node.outgrowths == nil || len(node.outgrowths) == 0 {
		return true
	}
	for _, eo := range node.outgrowths {
		if eo.event.id == id {
			return false
		}
	}
	return true
}

func (node *Node) AddOutgrowth(eo *EventOutgrowth) {
	if node.outgrowths == nil {
		node.outgrowths = make([]*EventOutgrowth,1,3)
		node.outgrowths[0] = eo
	} else {
		node.outgrowths = append(node.outgrowths, eo)
	}
}


func (node *Node) String() string {
	return fmt.Sprintf("%v:%d:%d", *(node.Pos), len(node.connections), len(node.outgrowths))
}

func (node *Node) GetStateString() string {
	evtIds := make([]EventID, len(node.outgrowths))
	for i, eo := range node.outgrowths {
		evtIds[i] = eo.event.id
	}
	connIds := make([]string, len(node.connections))
	for i, conn := range node.connections {
		var connVect Point
		if conn.N1 == node {
			connVect = conn.N2.Pos.Sub(*node.Pos)
		} else if conn.N2 == node {
			connVect = conn.N1.Pos.Sub(*node.Pos)
		} else {
			fmt.Println("ERROR: Connection",conn,"in list of node",node,"but not part of it?")
		}
		connIds[i] = AllConnectionsPossible[connVect].GetName()
	}
	return fmt.Sprintf("%v: %v, %v", *(node.Pos), evtIds, connIds)
}

/***************************************************************/
// Connection Functions
/***************************************************************/

func (conn *Connection) IsConnectedTo(node *Node) bool {
	return conn.N1 == node || conn.N2 == node
}

func (conn *Connection) GetColorMask(threshold Distance) uint8 {
	// Connection color mask of all event outgrowth that match
	m := uint8(0)
	if conn.N1 != nil && conn.N2 != nil {
		for _, eo1 := range conn.N1.outgrowths {
			if eo1.CameFrom(conn.N2) && eo1.IsActive(threshold) {
				m |= uint8(eo1.event.color)
			}
		}
		for _, eo2 := range conn.N2.outgrowths {
			if eo2.CameFrom(conn.N1) && eo2.IsActive(threshold) {
				m |= uint8(eo2.event.color)
			}
		}
	}
	return m
}

func (conn *Connection) IsActive(threshold Distance) bool {
	// 0 threshold cannot have active connections
	if threshold == 0 {
		return false
	}
	// Connection is active if event outgrowth latest match
	if conn.N1 != nil && conn.N2 != nil {
		for _, eo1 := range conn.N1.outgrowths {
			if eo1.CameFrom(conn.N2) && eo1.IsActive(threshold) {
				return true
			}
		}
		for _, eo2 := range conn.N2.outgrowths {
			if eo2.CameFrom(conn.N1) && eo2.IsActive(threshold) {
				return true
			}
		}
	}
	return false
}

func (conn *Connection) HowManyColors(threshold Distance) uint8 {
	return countOnes(conn.GetColorMask(threshold))
}

