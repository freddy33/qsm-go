package m3space

import "fmt"

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
			if m&1<<eo.event.color == uint8(0) {
				r++
			}
			m |= 1<<eo.event.color
		}
	}
	return r
}

func (eo *EventOutgrowth) IsRoot() bool {
	if eo.distance == Distance(0) {
		if eo.from != nil {
			fmt.Println("An event outgrowth of",eo.event.id,"has distance 0 and from not nil!")
		}
	}
	return eo.from == nil
}

func (e *Event) LatestDistance() Distance {
	// Usually the latest outgrowth are the last in the list
	lastEO := len(e.outgrowths) - 1
	if lastEO < 0 {
		return Distance(0)
	}
	eo := e.outgrowths[lastEO]
	if eo.state == EventOutgrowthLatest {
		return eo.distance
	}
	for _, eo = range e.outgrowths {
		if eo.state == EventOutgrowthLatest {
			return eo.distance
		}
	}
	fmt.Println("Did not find any latest in the list of Outgrowth! Impossible!")
	return Distance(0)
}

func (eo *EventOutgrowth) LatestDistance() Distance {
	if eo.state == EventOutgrowthLatest {
		return eo.distance
	}
	return eo.event.LatestDistance()
}

func (eo *EventOutgrowth) DistanceFromLatest() Distance {
	latestDistance := eo.LatestDistance()
	return latestDistance - eo.distance
}

func (eo *EventOutgrowth) IsDrawn(filter SpaceDrawingFilter) bool {
	return eo.IsActive(filter.EventOutgrowthThreshold) &&
		filter.EventColorMask&1<<eo.event.color != uint8(0) &&
		eo.node.HowManyColors(filter.EventOutgrowthThreshold) >= filter.EventOutgrowthManyColorsThreshold
}

func (eo *EventOutgrowth) IsActive(threshold Distance) bool {
	if eo.IsRoot() {
		// Root event always active
		return true
	}
	return eo.DistanceFromLatest() <= threshold
}