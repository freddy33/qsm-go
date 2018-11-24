package m3space

import (
	"fmt"
	"time"
)

const (
	// Where the number matters and appear. Remember that 3 is the number!
	THREE = 3
)

var DEBUG = false

type TickTime uint64

type SpaceVisitor interface {
	VisitNode(space *Space, node *Node)
	VisitConnection(space *Space, conn *Connection)
}

type Space struct {
	events      map[EventID]*Event
	nodesMap    map[Point]*Node
	connections []*Connection
	currentId   EventID
	currentTime TickTime
	// Max size of all CurrentSpace. TODO: Make it variable using the furthest node from1 origin
	Max int64
	// Max number of connections per node
	MaxConnections int
	// Distance from1 latest to consider event outgrowth active
	EventOutgrowthThreshold Distance
}

func MakeSpace(max int64) Space {
	space := Space{}
	space.events = make(map[EventID]*Event)
	space.nodesMap = make(map[Point]*Node)
	space.connections = make([]*Connection, 0, 500)
	space.currentId = 0
	space.currentTime = 0
	space.Max = max
	space.MaxConnections = 6
	space.EventOutgrowthThreshold = Distance(1)
	return space
}

func (space *Space) GetCurrentTime() TickTime {
	return space.currentTime
}

func (space *Space) GetNbNodes() int {
	return len(space.nodesMap)
}

func (space *Space) GetNbConnections() int {
	return len(space.connections)
}

func (space *Space) GetNbEvents() int {
	return len(space.events)
}

func (space *Space) VisitAll(visitor SpaceVisitor) {
	for _, node := range space.nodesMap {
		visitor.VisitNode(space, node)
	}
	for _, conn := range space.connections {
		visitor.VisitConnection(space, conn)
	}
}

func (space *Space) CreateSingleEventCenter() *Event {
	return space.CreateEvent(Origin, RedEvent)
}

func (space *Space) CreatePyramid(pyramidSize int64) {
	space.CreateEvent(Point{3, 0, 3}.Mul(pyramidSize), RedEvent)
	space.CreateEvent(Point{-3, 3, 3}.Mul(pyramidSize), GreenEvent)
	space.CreateEvent(Point{-3, -3, 3}.Mul(pyramidSize), BlueEvent)
	space.CreateEvent(Point{0, 0, -3}.Mul(pyramidSize), YellowEvent)
}

func (space *Space) ForwardTime() {
	fmt.Println("\n**********\nStepping up time from", space.currentTime, "=>", space.currentTime+1, "for", len(space.events), "events")
	nbLatest := 0
	c := make(chan *NewPossibleOutgrowth, 100)
	for _, evt := range space.events {
		go evt.createNewPossibleOutgrowths(c)
		nbLatest += len(evt.latestOutgrowths)
	}
	collector := space.processNewOutgrowth(c, nbLatest)
	space.realizeAllOutgrowth(collector)

	// Switch latest to old, and new to latest
	for _, evt := range space.events {
		evt.moveNewOutgrowthsToLatest()
	}
	space.currentTime++
}

func (space *Space) realizeAllOutgrowth(collector *OutgrowthCollector) {
	fmt.Println("Found", len(collector.single), "single new outgrowth,",
		len(collector.sameEvent), "overlap outgrowth on same event,",
		len(collector.multiEvents), "overlap on multi events.")

	occupied := 0
	noMoreConn := 0
	// No problem just realize all single ones that fit
	for _, newPosEo := range collector.single {
		_, err := newPosEo.realize()
		if err != nil {
			switch err.(type) {
			case *EventAlreadyGrewThereError:
				occupied++
			case *NoMoreConnectionsError:
				noMoreConn++
			}
		}
	}
	fmt.Printf("Single: %6d / %6d / %6d", len(collector.single), occupied, noMoreConn)

	notRealized = 0
	// Realize only one of conflicting same event
	for _, newPosEoList := range collector.sameEvent {
		var newEo *EventOutgrowth
		for _, newPosEo := range *newPosEoList {
			if newEo == nil {
				newEo = newPosEo.realize()
				if newEo == nil {
					notRealized++
				}
			} else {
				newEo.AddFrom(newPosEo.from)
			}
		}
	}
	fmt.Println("Multi same event new outgrowth not realized=", notRealized)

	notRealized = 0
	// Realize only one per event of conflicting multi events
	// Collect all more than 3 event outgrowth
	moreThan3 := make(map[Point][]EventID)
	for pos, newPosEoList := range collector.multiEvents {
		idsAlreadyDone := make(map[EventID]*EventOutgrowth, 2)
		for _, newPosEo := range *newPosEoList {
			doneEo, done := idsAlreadyDone[newPosEo.event.id]
			if !done {
				newEo := newPosEo.realize()
				if newEo != nil {
					idsAlreadyDone[newPosEo.event.id] = newEo
				} else {
					notRealized++
				}
			} else {
				doneEo.AddFrom(newPosEo.from)
			}
		}
		if len(idsAlreadyDone) >= THREE {
			ids := make([]EventID, len(idsAlreadyDone))
			i := 0
			for id := range idsAlreadyDone {
				ids[i] = id
				i++
			}
			moreThan3[pos] = ids
		}
	}
	fmt.Println("Multi event new outgrowth not realized=", notRealized, "found", len(moreThan3), "more than 3 positions")
}

type OutgrowthCollectorStat struct {
	original, occupied, noMoreConn int
}

type OutgrowthCollector struct {
	single         map[Point]*NewPossibleOutgrowth
	sameEvent      map[Point]*[]*NewPossibleOutgrowth
	multiEvents    map[Point]*[]*NewPossibleOutgrowth
	moreThan3      map[Point][]EventID
	singleStat     OutgrowthCollectorStat
	sameEventStat  OutgrowthCollectorStat
	multiEventStat OutgrowthCollectorStat
}

func MakeOutgrowthCollector(nbLatest int) *OutgrowthCollector {
	if nbLatest < 5 {
		nbLatest = 5
	}
	res := OutgrowthCollector{}
	res.single = make(map[Point]*NewPossibleOutgrowth, 2*nbLatest)
	res.sameEvent = make(map[Point]*[]*NewPossibleOutgrowth, nbLatest/3)
	res.multiEvents = make(map[Point]*[]*NewPossibleOutgrowth, 100)
	res.moreThan3 = make(map[Point][]EventID, 3)
	return &res
}

func (space *Space) processNewOutgrowth(c chan *NewPossibleOutgrowth, nbLatest int) *OutgrowthCollector {
	// Protect full calculation timeout with 5 milliseconds per latest outgrowth
	timeout := int64(5 * nbLatest)
	if timeout < 1000 {
		timeout = 1000
	}
	nbEvents := len(space.events)
	nbEventsDone := 0
	collector := MakeOutgrowthCollector(nbLatest)
	for {
		stop := false
		select {
		case newEo := <-c:

			switch newEo.state {
			case EventOutgrowthEnd:
				nbEventsDone++
			case EventOutgrowthNew:

				fromSingle, ok := collector.single[newEo.pos]
				if !ok {
					collector.single[newEo.pos] = newEo
				} else {
					switch fromSingle.state {
					case EventOutgrowthNew:
						// First multiple entry, check if same event or not and move event outgrowth there
						if fromSingle.event.id == newEo.event.id {
							_, okSameEvent := collector.sameEvent[newEo.pos]
							if okSameEvent {
								fmt.Println("ERROR: An event outgrowth in single map with state", fromSingle.state, "full=", *(fromSingle), "is state new but has an entry in the multi same event Map!!")
							}

							newSameEventList := make([]*NewPossibleOutgrowth, 2, 3)
							newSameEventList[0] = fromSingle
							newSameEventList[1] = newEo
							fromSingle.state = EventOutgrowthMultipleSameEvent
							newEo.state = EventOutgrowthMultipleSameEvent
							collector.sameEvent[newEo.pos] = &newSameEventList
						} else {
							_, okMultiEvent := collector.multiEvents[newEo.pos]
							if okMultiEvent {
								fmt.Println("ERROR: An event outgrowth in single map with state", fromSingle.state, "full=", *(fromSingle), "is state new but has an entry in the multi events Map!!")
							}

							newMultiEventList := make([]*NewPossibleOutgrowth, 2, 3)
							newMultiEventList[0] = fromSingle
							newMultiEventList[1] = newEo
							fromSingle.state = EventOutgrowthMultipleEvents
							newEo.state = EventOutgrowthMultipleEvents
							collector.multiEvents[newEo.pos] = &newMultiEventList
						}
					case EventOutgrowthMultipleSameEvent:
						fromSameEvent, okSameEvent := collector.sameEvent[newEo.pos]
						if !okSameEvent {
							fmt.Println("ERROR: An event outgrowth in single map with state", fromSingle.state, "full=", *(fromSingle), "does not have an entry in the multi same event Map!!")
						} else {
							if fromSingle.event.id == newEo.event.id {
								newEo.state = EventOutgrowthMultipleSameEvent
								*fromSameEvent = append(*fromSameEvent, newEo)
							} else {
								// Move all from1 same event to multi event
								_, okMultiEvent := collector.multiEvents[newEo.pos]
								if okMultiEvent {
									fmt.Println("ERROR: An event outgrowth in multi same event map with state", fromSingle.state, "full=", *(fromSingle), "is state same event but has an entry in the multi events Map!!")
								}

								*fromSameEvent = append(*fromSameEvent, newEo)
								for _, eo := range *fromSameEvent {
									eo.state = EventOutgrowthMultipleEvents
								}
								// Just verify
								if newEo.state != EventOutgrowthMultipleEvents || fromSingle.state != EventOutgrowthMultipleEvents {
									fmt.Println("ERROR: Event outgrowth state change failed for", *fromSingle, "and", *newEo)
								}
								collector.multiEvents[newEo.pos] = fromSameEvent
								delete(collector.sameEvent, newEo.pos)
							}
						}
					case EventOutgrowthMultipleEvents:
						fromMultiEvent, okMultiEvent := collector.multiEvents[newEo.pos]
						if !okMultiEvent {
							fmt.Println("ERROR: An event outgrowth in single map with state", fromSingle.state, "full=", *(fromSingle), "does not have an entry in the multi events Map!!")
						} else {
							newEo.state = EventOutgrowthMultipleEvents
							*fromMultiEvent = append(*fromMultiEvent, newEo)
						}
					}
				}
			default:
				fmt.Println("ERROR: Receive an event on channel with wrong state", newEo.state, "full=", *(newEo))
			}

			if nbEventsDone == nbEvents {
				stop = true
				break
			}
		case <-time.After(time.Duration(timeout) * time.Millisecond):
			stop = true
			fmt.Println("ERROR: Did not manage to process", nbLatest, "latest event outgrowth from1", nbEvents, "events in", nbLatest*5, "msecs")
			break
		}
		if stop {
			break
		}
	}

	return collector
}

func (space *Space) GetNode(p Point) *Node {
	n, ok := space.nodesMap[p]
	if ok {
		return n
	}
	return nil
}

func (space *Space) getOrCreateNode(p Point) *Node {
	n := space.GetNode(p)
	if n != nil {
		return n
	}
	n = &Node{&p, nil, nil,}
	space.nodesMap[p] = n
	return n
}

func (space *Space) makeConnection(n1, n2 *Node) *Connection {
	if !n1.HasFreeConnections(space) {
		if DEBUG {
			fmt.Println("Node 1", n1, "does not have free connections")
		}
		return nil
	}
	if !n2.HasFreeConnections(space) {
		if DEBUG {
			fmt.Println("Node 2", n2, "does not have free connections")
		}
		return nil
	}
	if n1.IsAlreadyConnected(n2) {
		if DEBUG {
			fmt.Println("Connection between 2 points", *(n1.Pos), *(n2.Pos), "already connected!")
		}
		return nil
	}

	// Flipping if needed to make sure n1 is main
	if n2.Pos.IsMainPoint() {
		temp := n1
		n1 = n2
		n2 = temp
	}
	d := DS(n1.Pos, n2.Pos)
	if !(d == 1 || d == 2 || d == 3 || d == 5) {
		fmt.Println("ERROR: Connection between 2 points", *(n1.Pos), *(n2.Pos), "that are not 1, 2, 3 or 5 DS away!")
		return nil
	}
	// All good create connection
	c := &Connection{n1, n2}
	space.connections = append(space.connections, c)
	n1done := n1.AddConnection(c, space)
	n2done := n2.AddConnection(c, space)
	if n1done < 0 || n2done < 0 {
		fmt.Println("ERROR: Node1 connection association", n1done, "or Node2", n2done, "did not happen!!")
		return nil
	}
	return c
}

func (space *Space) DisplaySettings() {
	fmt.Println("========= Space Settings =========")
	fmt.Println("Current Time", space.currentTime)
	fmt.Println("Nb Nodes", len(space.nodesMap), ", Nb Connections", len(space.connections), ", Nb Events", len(space.events))
}
