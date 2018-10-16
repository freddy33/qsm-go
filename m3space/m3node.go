package m3space

import "fmt"

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
