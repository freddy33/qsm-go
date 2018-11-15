package m3space


type Node struct {
	Pos          *Point
	outgrowths   []*EventOutgrowth
	connections  []*Connection
}

type Connection struct {
	N1, N2 *Node
}

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

func (conn *Connection) IsConnectedTo(node *Node) bool {
	return conn.N1 == node || conn.N2 == node
}

func (conn *Connection) GetColorMask(threshold Distance) uint8 {
	// Connection color mask of all event outgrowth that match
	m := uint8(0)
	if conn.N1 != nil && conn.N2 != nil {
		for _, eo1 := range conn.N1.outgrowths {
			if eo1.from != nil && eo1.from.node == conn.N2 && eo1.IsActive(threshold) {
				m |= uint8(eo1.event.color)
			}
		}
		for _, eo2 := range conn.N2.outgrowths {
			if eo2.from != nil && eo2.from.node == conn.N1 && eo2.IsActive(threshold) {
				m |= uint8(eo2.event.color)
			}
		}
	}
	return m
}

func (conn *Connection) IsActive(threshold Distance) bool {
	// Connection is active if event outgrowth latest match
	if conn.N1 != nil && conn.N2 != nil {
		for _, eo1 := range conn.N1.outgrowths {
			if eo1.from != nil && eo1.from.node == conn.N2 && eo1.IsActive(threshold) {
				return true
			}
		}
		for _, eo2 := range conn.N2.outgrowths {
			if eo2.from != nil && eo2.from.node == conn.N1 && eo2.IsActive(threshold) {
				return true
			}
		}
	}
	return false
}

func (conn *Connection) HowManyColors(threshold Distance) uint8 {
	return countOnes(conn.GetColorMask(threshold))
}

