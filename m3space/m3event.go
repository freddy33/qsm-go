package m3space

import "fmt"

type EventID uint64

type Distance uint64

type EventColor uint8

type EventOutgrowthState uint8

const (
	RedEvent    EventColor = 1 << iota
	GreenEvent
	BlueEvent
	YellowEvent
)

const (
	EventOutgrowthLatest            EventOutgrowthState = iota
	EventOutgrowthNew
	EventOutgrowthOld
	EventOutgrowthEnd
	EventOutgrowthMultipleSameEvent
	EventOutgrowthMultipleEvents
)

var AllColors = [4]EventColor{RedEvent, GreenEvent, BlueEvent, YellowEvent}

type Event struct {
	space            *Space
	id               EventID
	node             *Node
	created          TickTime
	color            EventColor
	growthContext    GrowthContext
	oldOutgrowths    []*EventOutgrowth
	latestOutgrowths []*EventOutgrowth
}

type EventOutgrowth struct {
	node     *Node
	event    *Event
	from1    *EventOutgrowth
	from2    *EventOutgrowth
	distance Distance
	state    EventOutgrowthState
}

type NewPossibleOutgrowth struct {
	pos      Point
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
		make([]*EventOutgrowth, 0, 100), make([]*EventOutgrowth, 1, 100),}
	e.latestOutgrowths[0] = &EventOutgrowth{n, &e, nil, nil, Distance(0), EventOutgrowthLatest}
	n.outgrowths = make([]*EventOutgrowth, 1)
	n.outgrowths[0] = e.latestOutgrowths[0]
	space.events[id] = &e
	ctx.center = n.Pos
	return &e
}

func (evt *Event) createNewPossibleOutgrowths(c chan *NewPossibleOutgrowth) {
	for _, eg := range evt.latestOutgrowths {
		if eg.state != EventOutgrowthLatest {
			panic(fmt.Sprintf("wrong state of event! found non latest outgrowth %v at %v in latest list.", eg, *(eg.node.Pos)))
		}

		nextPoints := eg.node.Pos.getNextPoints(&(evt.growthContext))
		for _, nextPoint := range nextPoints {
			if !eg.CameFromPoint(nextPoint) {
				if DEBUG {
					fmt.Println("Creating new possible event outgrowth for", evt.id, "at", nextPoint)
				}
				c <- &NewPossibleOutgrowth{nextPoint, evt, eg, eg.distance + 1, EventOutgrowthNew}
			}
		}
	}
	c <- &NewPossibleOutgrowth{*(evt.node.Pos), evt, nil, Distance(0), EventOutgrowthEnd}
}

func (evt *Event) moveNewOutgrowthsToLatest() {
	finalLatest := evt.latestOutgrowths[:0]
	for _, eg := range evt.latestOutgrowths {
		switch eg.state {
		case EventOutgrowthLatest:
			eg.state = EventOutgrowthOld
			evt.oldOutgrowths = append(evt.oldOutgrowths, eg)
		case EventOutgrowthNew:
			eg.state = EventOutgrowthLatest
			finalLatest = append(finalLatest, eg)
		}
	}
	evt.latestOutgrowths = finalLatest
}

func (evt *Event) LatestDistance() Distance {
	// Distance and time are the same...
	return Distance(evt.space.currentTime - evt.created)
}

func (newPosEo *NewPossibleOutgrowth) realize() *EventOutgrowth {
	evt := newPosEo.event
	space := evt.space
	newNode := space.getOrCreateNode(newPosEo.pos)
	if !newNode.CanReceiveOutgrowth(newPosEo) {
		return nil
	}
	fromNode := newPosEo.from.node
	if !fromNode.IsAlreadyConnected(newNode) {
		space.makeConnection(fromNode, newNode)
	}
	newEo := &EventOutgrowth{newNode, evt, newPosEo.from, nil, newPosEo.distance, EventOutgrowthNew}
	evt.latestOutgrowths = append(evt.latestOutgrowths, newEo)
	newNode.AddOutgrowth(newEo)
	return newEo
}

func (eo *EventOutgrowth) HasFrom() bool {
	return eo.from1 != nil
}

func (eo *EventOutgrowth) HasFrom2() bool {
	return eo.from2 != nil
}

func (eo *EventOutgrowth) CameFrom(node *Node) bool {
	return (eo.HasFrom() && eo.from1.node == node) || (eo.HasFrom2() && eo.from2.node == node)
}

func (eo *EventOutgrowth) CameFromPoint(point Point) bool {
	return (eo.HasFrom() && *(eo.from1.node.Pos) == point) || (eo.HasFrom2() && *(eo.from2.node.Pos) == point)
}

func (eo *EventOutgrowth) IsRoot() bool {
	return !eo.HasFrom()
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

func (eo *EventOutgrowth) IsActive(threshold Distance) bool {
	if eo.IsRoot() {
		// Root event always active
		return true
	}
	return eo.DistanceFromLatest() <= threshold
}
