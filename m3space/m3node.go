package m3space

type Node struct {
	point       *Point
	outgrowths  []*EventOutgrowth
	connections [3]*Connection
}

type Connection struct {
	N1, N2 *Node
}

func (n *Node) HasFreeConnections() bool {
	for _, c := range n.connections {
		if c == nil {
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
