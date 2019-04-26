package m3space

import (
	"github.com/freddy33/qsm-go/m3path"
	"github.com/freddy33/qsm-go/m3point"
)

type EventID uint64

const (
	NilEvent = EventID(0)
)

type Distance uint64

type EventColor uint8

const (
	RedEvent EventColor = 1 << iota
	GreenEvent
	BlueEvent
	YellowEvent
)

var AllColors = [4]EventColor{RedEvent, GreenEvent, BlueEvent, YellowEvent}

type Event struct {
	space             *Space
	id                EventID
	node              *ActiveNode
	created           TickTime
	color             EventColor
	growthContext     *m3path.GrowthContext
	currentOutgrowths []*EventOutgrowth
	latestOutgrowths  []*EventOutgrowth
}

type SavedEvent struct {
	id                   EventID
	node                 SavedNode
	created              TickTime
	color                EventColor
	growthContext        m3path.GrowthContext
	savedLatestOutgrowth []SavedEventOutgrowth
}

func (space *Space) CreateEvent(p m3point.Point, k EventColor) *Event {
	ctx := m3path.CreateGrowthContext(m3point.Origin, 8, 0, 0)
	switch k {
	case RedEvent:
		// No change
	case GreenEvent:
		ctx.SetIndexOffset(4, 0)
	case BlueEvent:
		ctx.SetIndexOffset(8, 0)
	case YellowEvent:
		ctx.SetIndexOffset(10, 4)
	}
	return space.CreateEventWithGrowthContext(p, k, ctx)
}

func (space *Space) CreateEventWithGrowthContext(p m3point.Point, k EventColor, ctx *m3path.GrowthContext) *Event {
	n := space.getOrCreateNode(p)
	id := space.currentId
	n.SetRoot(id, space.currentTime)
	if Log.IsInfo() {
		Log.Info("Creating new event at node", n.GetStateString())
	}
	space.currentId++
	e := Event{space, id, n, space.currentTime, k,
		ctx,
		make([]*EventOutgrowth, 0, 100), make([]*EventOutgrowth, 1, 100)}
	e.latestOutgrowths[0] = MakeActiveOutgrowth(n.Pos, Distance(0), EventOutgrowthLatest)
	space.events[id] = &e
	ctx.SetCenter(n.Pos)
	return &e
}

func (evt *Event) LatestDistance() Distance {
	// Distance and time are the same...
	return Distance(evt.space.currentTime - evt.created)
}
