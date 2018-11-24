package m3space

import (
	"fmt"
	"time"
)

type OutgrowthCollectorStat struct {
	name string
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
	res.singleStat.name = "Single"
	res.sameEventStat.name = "Same Event"
	res.multiEventStat.name = "Multi Events"
	return &res
}

func (colStat *OutgrowthCollectorStat) realizeAndStat(newPosEo *NewPossibleOutgrowth) *EventOutgrowth {
	newEo, err := newPosEo.realize()
	if err != nil {
		switch err.(type) {
		case *EventAlreadyGrewThereError:
			colStat.occupied++
		case *NoMoreConnectionsError:
			colStat.noMoreConn++
		}
		return nil
	}
	return newEo
}

func (colStat *OutgrowthCollectorStat) displayStat() {
	fmt.Printf("%12s: %6d / %6d / %6d\n", colStat.name, colStat.original, colStat.occupied, colStat.noMoreConn)
}

func (collector *OutgrowthCollector) beginRealize() {
	collector.singleStat.original = len(collector.single)
	collector.sameEventStat.original = len(collector.sameEvent)
	collector.multiEventStat.original = len(collector.multiEvents)
}

func (space *Space) realizeAllOutgrowth(collector *OutgrowthCollector) {
	collector.beginRealize()

	// No problem just realize all single ones that fit
	for _, newPosEo := range collector.single {
		collector.singleStat.realizeAndStat(newPosEo)
	}
	collector.singleStat.displayStat()

	// Realize only one of conflicting same event
	for _, newPosEoList := range collector.sameEvent {
		var newEo *EventOutgrowth
		for _, newPosEo := range *newPosEoList {
			if newEo == nil {
				newEo = collector.sameEventStat.realizeAndStat(newPosEo)
			} else {
				newEo.AddFrom(newPosEo.from)
			}
		}
	}
	collector.sameEventStat.displayStat()

	// Realize only one per event of conflicting multi events
	// Collect all more than 3 event outgrowth
	for pos, newPosEoList := range collector.multiEvents {
		idsAlreadyDone := make(map[EventID]*EventOutgrowth, 2)
		for _, newPosEo := range *newPosEoList {
			doneEo, done := idsAlreadyDone[newPosEo.event.id]
			if !done {
				newEo := collector.multiEventStat.realizeAndStat(newPosEo)
				if newEo != nil {
					idsAlreadyDone[newPosEo.event.id] = newEo
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
			collector.moreThan3[pos] = ids
		}
	}
	collector.multiEventStat.displayStat()
	if len(collector.moreThan3) > 0 {
		fmt.Println("###############################################")
		fmt.Println("Finally got triple sync at:",collector.moreThan3)
		fmt.Println("###############################################")
	}
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

type EventAlreadyGrewThereError struct {
	id  EventID
	pos Point
}

func (e *EventAlreadyGrewThereError) Error() string {
	return fmt.Sprintf("event with id %d already has outgrowth at %v", e.id, e.pos)
}

type NoMoreConnectionsError struct {
	pos      Point
	otherPos Point
}

func (e *NoMoreConnectionsError) Error() string {
	return fmt.Sprintf("node at %v already has full connections and cannot connect to %v", e.pos, e.otherPos)
}

func (newPosEo *NewPossibleOutgrowth) realize() (*EventOutgrowth, error) {
	evt := newPosEo.event
	space := evt.space
	newNode := space.getOrCreateNode(newPosEo.pos)
	if !newNode.CanReceiveOutgrowth(newPosEo) {
		return nil, &EventAlreadyGrewThereError{newPosEo.event.id, newPosEo.pos,}
	}
	fromNode := newPosEo.from.node
	if !fromNode.IsAlreadyConnected(newNode) {
		if space.makeConnection(fromNode, newNode) == nil {
			// No more connections
			return nil, &NoMoreConnectionsError{*(newNode.Pos), *(fromNode.Pos)}
		}
	}
	newEo := &EventOutgrowth{newNode, evt, []*EventOutgrowth{newPosEo.from,}, newPosEo.distance, EventOutgrowthNew}
	evt.latestOutgrowths = append(evt.latestOutgrowths, newEo)
	newNode.AddOutgrowth(newEo)
	return newEo, nil
}

func (eo *EventOutgrowth) AddFrom(from *EventOutgrowth) {
	if eo.fromList == nil {
		eo.fromList = []*EventOutgrowth{from,}
	} else {
		eo.fromList = append(eo.fromList, from)
	}
}

func (eo *EventOutgrowth) HasFrom() bool {
	return eo.fromList != nil && len(eo.fromList) > 0
}

func (eo *EventOutgrowth) CameFrom(node *Node) bool {
	if !eo.HasFrom() {
		return false
	}
	for _, from := range eo.fromList {
		if from.node == node {
			return true
		}
	}
	return false
}

func (eo *EventOutgrowth) CameFromPoint(point Point) bool {
	if !eo.HasFrom() {
		return false
	}
	for _, from := range eo.fromList {
		if *(from.node.Pos) == point {
			return true
		}
	}
	return false
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
