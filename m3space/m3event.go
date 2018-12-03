package m3space

type EventID uint64

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
	growthContext     *GrowthContext
	currentOutgrowths []*EventOutgrowth
	latestOutgrowths  []*EventOutgrowth
}

type SavedEvent struct {
	id                   EventID
	node                 SavedNode
	created              TickTime
	color                EventColor
	growthContext        GrowthContext
	savedLatestOutgrowth []SavedEventOutgrowth
}

func (space *Space) CreateEvent(p Point, k EventColor) *Event {
	ctx := GrowthContext{Origin, 8, 0, false, 0}
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
	return space.CreateEventWithGrowthContext(p, k, &ctx)
}

func (space *Space) CreateEventWithGrowthContext(p Point, k EventColor, ctx *GrowthContext) *Event {
	n := space.getOrCreateNode(p)
	id := space.currentId
	n.SetRoot(id, space.currentTime)
	Log.Info("Creating new event at node", n.GetStateString())
	space.currentId++
	e := Event{space, id, n, space.currentTime, k,
		ctx,
		make([]*EventOutgrowth, 0, 100), make([]*EventOutgrowth, 1, 100),}
	e.latestOutgrowths[0] = MakeActiveOutgrowth(n.Pos, Distance(0), EventOutgrowthLatest)
	space.events[id] = &e
	ctx.center = n.Pos
	return &e
}

func (evt *Event) LatestDistance() Distance {
	// Distance and time are the same...
	return Distance(evt.space.currentTime - evt.created)
}
