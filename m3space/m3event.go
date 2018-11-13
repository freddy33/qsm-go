package m3space

import "fmt"

type EventID uint64

type Distance uint64

type EventColor uint8

type EventOutgrowthState uint8

const (
	RedEvent    EventColor = 0x01
	GreenEvent  EventColor = 0x02
	BlueEvent   EventColor = 0x04
	YellowEvent EventColor = 0x08
)

const (
	EventOutgrowthLatest EventOutgrowthState = iota
	EventOutgrowthNew
	EventOutgrowthOld
	EventOutgrowthDead
)

var AllColors = [4]EventColor{RedEvent, GreenEvent, BlueEvent, YellowEvent}

type Event struct {
	space         *Space
	id            EventID
	node          *Node
	created       TickTime
	color         EventColor
	growthContext GrowthContext
	outgrowths    []*EventOutgrowth
	newOutgrowths []*EventOutgrowth
}

type EventOutgrowth struct {
	node     *Node
	event    *Event
	from     *EventOutgrowth
	distance Distance
	state    EventOutgrowthState
}

func (space *Space) CreateEvent(p Point, k EventColor) *Event {
	ctx := GrowthContext{&Origin, 8, 0, false, 0}
	switch k {
	case RedEvent:
		// No change
	case GreenEvent:
		ctx.permutationIndex = 4
		ctx.permutationOffset = 0
	case BlueEvent:
		ctx.permutationIndex = 8
		ctx.permutationOffset = 0
	case YellowEvent:
		ctx.permutationIndex = 10
		ctx.permutationOffset = 4
	}
	return space.CreateEventWithGrowthContext(p, k, ctx)
}

func (space *Space) CreateEventWithGrowthContext(p Point, k EventColor, ctx GrowthContext) *Event {
	n := space.getOrCreateNode(p)
	id := space.currentId
	space.currentId++
	e := Event{space, id, n, space.currentTime, k,
		ctx,
		make([]*EventOutgrowth, 1, 100), make([]*EventOutgrowth, 0, 10),}
	e.outgrowths[0] = &EventOutgrowth{n, &e, nil, Distance(0), EventOutgrowthLatest}
	n.outgrowths = make([]*EventOutgrowth, 1)
	n.outgrowths[0] = e.outgrowths[0]
	space.events[id] = &e
	ctx.center = n.point
	return &e
}

func (evt *Event) createNewOutgrowths() {
	evt.newOutgrowths = evt.newOutgrowths[:0]
	for _, eg := range evt.outgrowths {
		if eg.state == EventOutgrowthLatest && eg.node.HasFreeConnections() {
			nextPoints := eg.node.point.getNextPoints(&(evt.growthContext))
			var otherNodes []*Node
			if eg.from == nil {
				otherNodes = make([]*Node,3)
			} else {
				otherNodes = make([]*Node,2)
			}
			offset := 0
			for _, nextPoint := range nextPoints {
				otherNode := evt.space.getOrCreateNode(nextPoint)
				if eg.from == nil || otherNode != eg.from.node {
					otherNodes[offset] = otherNode
					offset++
				}
			}

			for _, otherNode := range otherNodes {
				// Roots cannot have outgrowth that not theirs (TODO: why?)
				hasAlreadyEvent := otherNode.IsRoot()
				for _, eo := range otherNode.outgrowths {
					if eo.event.id == evt.id {
						hasAlreadyEvent = true
					}
				}
				if !hasAlreadyEvent {
					if DEBUG {
						fmt.Println("Creating new event outgrowth for", evt.id, "at", otherNode.point)
					}
					if !otherNode.IsAlreadyConnected(eg.node) && otherNode.HasFreeConnections() && eg.node.HasFreeConnections() {
						evt.space.makeConnection(eg.node, otherNode)
					}
					if otherNode.IsAlreadyConnected(eg.node) {
						newEo := &EventOutgrowth{otherNode, evt, eg, eg.distance + 1, EventOutgrowthNew}
						otherNode.outgrowths = append(otherNode.outgrowths, newEo)
						evt.newOutgrowths = append(evt.newOutgrowths, newEo)
					} else {
						fmt.Println("Could NOT create new event outgrowth for", evt.id, "at", otherNode.point, "since no more free connections!")
					}
				}
			}
		}
	}
}

func (evt *Event) moveNewOutgrowthsToLatest() {
	for _, eg := range evt.outgrowths {
		switch eg.state {
		case EventOutgrowthLatest:
			eg.state = EventOutgrowthOld
		}
	}
	for _, eg := range evt.newOutgrowths {
		switch eg.state {
		case EventOutgrowthNew:
			eg.state = EventOutgrowthLatest
			evt.outgrowths = append(evt.outgrowths, eg)
		}
	}
}

func (eventOutgrowth *EventOutgrowth) IsRoot() bool {
	if eventOutgrowth.distance == Distance(0) {
		if eventOutgrowth.from != nil {
			fmt.Println("An event outgrowth of", eventOutgrowth.event.id, "has distance 0 and from not nil!")
		}
	}
	return eventOutgrowth.from == nil
}

func (evt *Event) LatestDistance() Distance {
	// Usually the latest outgrowth are the last in the list
	lastEO := len(evt.outgrowths) - 1
	if lastEO < 0 {
		return Distance(0)
	}
	eo := evt.outgrowths[lastEO]
	if eo.state == EventOutgrowthLatest {
		return eo.distance
	}
	for _, eo = range evt.outgrowths {
		if eo.state == EventOutgrowthLatest {
			return eo.distance
		}
	}
	fmt.Println("Did not find any latest in the list of Outgrowth! Impossible!")
	return Distance(0)
}

func (eventOutgrowth *EventOutgrowth) LatestDistance() Distance {
	if eventOutgrowth.state == EventOutgrowthLatest {
		return eventOutgrowth.distance
	}
	return eventOutgrowth.event.LatestDistance()
}

func (eventOutgrowth *EventOutgrowth) DistanceFromLatest() Distance {
	latestDistance := eventOutgrowth.LatestDistance()
	return latestDistance - eventOutgrowth.distance
}

func (eventOutgrowth *EventOutgrowth) IsDrawn(filter SpaceDrawingFilter) bool {
	return eventOutgrowth.IsActive(eventOutgrowth.event.space.EventOutgrowthThreshold) &&
		filter.EventColorMask&uint8(eventOutgrowth.event.color) != uint8(0) &&
		eventOutgrowth.node.HowManyColors(eventOutgrowth.event.space.EventOutgrowthThreshold) >= filter.EventOutgrowthManyColorsThreshold
}

func (eventOutgrowth *EventOutgrowth) IsActive(threshold Distance) bool {
	if eventOutgrowth.IsRoot() {
		// Root event always active
		return true
	}
	return eventOutgrowth.DistanceFromLatest() <= threshold
}
