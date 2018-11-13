package m3space


type Node struct {
	space       *Space
	point       *Point
	outgrowths  []*EventOutgrowth
	connections []*Connection
}

type Connection struct {
	N1, N2 *Node
}

func (node *Node) HasFreeConnections() bool {
	return node.connections == nil || len(node.connections) < node.space.MaxConnections
}

func (node *Node) AddConnection(conn *Connection) int {
	if !node.HasFreeConnections() {
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
	r := uint8(0)
	m := uint8(0)
	for _, eo := range node.outgrowths {
		if eo.IsActive(threshold) {
			if m&uint8(eo.event.color) == uint8(0) {
				r++
			}
			m |= uint8(eo.event.color)
		}
	}
	return r
}

func (conn *Connection) IsConnectedTo(node *Node) bool {
	return conn.N1 == node || conn.N2 == node
}
