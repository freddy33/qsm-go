package m3space

type Node struct {
	P *Point
	E []*EventOutgrowth
	C [3]*Connection
}

type Connection struct {
	N1, N2 *Node
}

func (n *Node) HasFreeConnections() bool {
	for _, c := range n.C {
		if c == nil {
			return true
		}
	}
	return false
}

