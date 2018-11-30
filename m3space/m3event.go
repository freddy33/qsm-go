package m3space

type EventID uint64

type Distance uint64

type EventColor uint8

type EventOutgrowthState uint8

const (
	RedEvent EventColor = 1 << iota
	GreenEvent
	BlueEvent
	YellowEvent
)

const (
	EventOutgrowthLatest EventOutgrowthState = iota
	EventOutgrowthNew
	EventOutgrowthCurrent
	EventOutgrowthOld
	EventOutgrowthEnd
	EventOutgrowthManySameEvent
	EventOutgrowthMultipleEvents
)

var AllColors = [4]EventColor{RedEvent, GreenEvent, BlueEvent, YellowEvent}

type Event struct {
	space             *Space
	id                EventID
	node              *ActiveNode
	created           TickTime
	color             EventColor
	growthContext     GrowthContext
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

type EventOutgrowth struct {
	pos      *Point
	fromList []Outgrowth
	distance Distance
	state    EventOutgrowthState
}

type SavedEventOutgrowth struct {
	pos             Point
	fromConnections []int8
	distance        Distance
}

type Outgrowth interface {
	GetPoint() *Point
	GetDistance() Distance
	GetState() EventOutgrowthState
	AddFrom(from Outgrowth)
	CameFromPoint(point Point) bool
	HasFrom() bool
	IsRoot() bool
	DistanceFromLatest(evt *Event) Distance
	IsActive(evt *Event) bool
	IsOld(evt *Event) bool
}

func (eos EventOutgrowthState) String() string {
	switch eos {
	case EventOutgrowthLatest:
		return "Latest"
	case EventOutgrowthNew:
		return "New"
	case EventOutgrowthCurrent:
		return "Current"
	case EventOutgrowthOld:
		return "Old"
	case EventOutgrowthEnd:
		return "End"
	case EventOutgrowthManySameEvent:
		return "SameEvent"
	case EventOutgrowthMultipleEvents:
		return "MultiEvents"
	default:
		Log.Error("Got an event outgrowth state unknown:", int(eos))
	}
	return "unknown"
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
	n.SetRoot(id, space.currentTime)
	Log.Info("Creating new event at node", n.GetStateString())
	space.currentId++
	e := Event{space, id, n, space.currentTime, k,
		ctx,
		make([]*EventOutgrowth, 0, 100), make([]*EventOutgrowth, 1, 100),}
	e.latestOutgrowths[0] = &EventOutgrowth{n.Pos, nil, Distance(0), EventOutgrowthLatest}
	space.events[id] = &e
	ctx.center = n.Pos
	return &e
}

func (evt *Event) LatestDistance() Distance {
	// Distance and time are the same...
	return Distance(evt.space.currentTime - evt.created)
}
