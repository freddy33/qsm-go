package m3space

import (
	"fmt"
	"time"
)

type OutgrowthCollectorStat struct {
	name                                 string
	originalPoints, originalPossible     int
	originalHistogram                    []int
	occupiedPoints, occupiedPossible     int
	occupiedHistogram                    []int
	noMoreConnPoints, noMoreConnPossible int
	noMoreConnHistogram                  []int
	newPoint                             bool
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

func (space *Space) ForwardTime() {
	fmt.Printf("\n**********\nStepping up time from %d => %d for %d events", space.currentTime, space.currentTime+1, len(space.events))
	nbLatest := 0
	c := make(chan *NewPossibleOutgrowth, 100)
	for _, evt := range space.events {
		go evt.createNewPossibleOutgrowths(c)
		nbLatest += len(evt.latestOutgrowths)
	}
	fmt.Printf(" and %d latest outgrowths\n", nbLatest)
	collector := space.processNewOutgrowth(c, nbLatest)
	space.realizeAllOutgrowth(collector)

	// Switch latest to old, and new to latest
	for _, evt := range space.events {
		evt.moveNewOutgrowthsToLatest()
	}
	space.currentTime++
}

func (evt *Event) createNewPossibleOutgrowths(c chan *NewPossibleOutgrowth) {
	for _, eg := range evt.latestOutgrowths {
		if eg.state != EventOutgrowthLatest {
			panic(fmt.Sprintf("wrong state of event! found non latest outgrowth %v at %v in latest list.", eg, *(eg.node.Pos)))
		}

		nextPoints := eg.node.Pos.getNextPoints(&(evt.growthContext))
		for _, nextPoint := range nextPoints {
			if !eg.CameFromPoint(nextPoint) {
				sendOutgrowth := true
				nodeThere := evt.space.GetNode(nextPoint)
				if nodeThere != nil {
					sendOutgrowth = nodeThere.CanReceiveEvent(evt.id)
					if DEBUG {
						fmt.Println("New EO on existing node", nodeThere.GetStateString(), "can receive=", sendOutgrowth)
					}
				}
				if sendOutgrowth {
					if DEBUG {
						fmt.Println("Creating new possible event outgrowth for", evt.id, "at", nextPoint)
					}
					c <- &NewPossibleOutgrowth{nextPoint, evt, eg, eg.distance + 1, EventOutgrowthNew}
				}
			}
		}
	}
	if DEBUG {
		fmt.Println("Finished with event outgrowth for", evt.id, "sending End state possible outgrowth")
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

func (colStat *OutgrowthCollectorStat) realizeAndStat(newPosEo *NewPossibleOutgrowth, size int) *EventOutgrowth {
	newEo, err := newPosEo.realize()
	if err != nil {
		switch err.(type) {
		case *EventAlreadyGrewThereError:
			colStat.occupiedPossible++
			if size > 1 {
				colStat.occupiedHistogram[size-HistogramDelta]++
			}
			if colStat.newPoint {
				colStat.occupiedPoints++
				colStat.newPoint = false
			}
		case *NoMoreConnectionsError:
			colStat.noMoreConnPossible++
			if size > 1 {
				colStat.noMoreConnHistogram[size-HistogramDelta]++
			}
			if colStat.newPoint {
				colStat.noMoreConnPoints++
				colStat.newPoint = false
			}
		}
		return nil
	}
	return newEo
}

func (colStat *OutgrowthCollectorStat) displayStat() {
	if colStat.originalPoints == 0 {
		// nothing to show skip
		return
	}
	fmt.Printf("%12s  : %6d / %6d / %6d | %6d / %6d / %6d\n", colStat.name,
		colStat.originalPoints, colStat.occupiedPoints, colStat.noMoreConnPoints,
		colStat.originalPossible, colStat.occupiedPossible, colStat.noMoreConnPossible)
	if len(colStat.originalHistogram) > 1 {
		for i, j := range colStat.originalHistogram {
			fmt.Printf("%12s %d: %6d / %6d / %6d\n", colStat.name, i+HistogramDelta, j,
				colStat.occupiedHistogram[i], colStat.noMoreConnHistogram[i])
		}
	}
}

func (colStat *OutgrowthCollectorStat) displayDebug(data map[Point]*[]*NewPossibleOutgrowth) {
	dataLength := len(data)
	if dataLength == 0 {
		// nothing to show skip
		return
	}
	fmt.Printf("%s %d:", colStat.name, dataLength)
	i := 0
	onExistingNodes := make([]string, 0)
	for p, l := range data {
		if i%6 == 0 {
			fmt.Println("")
		}
		fmt.Printf("%v=%d  ", p, len(*l))
		node := (*l)[0].event.space.GetNode(p)
		if node != nil {
			onExistingNodes = append(onExistingNodes, fmt.Sprintf("%v:%s", p, node.GetStateString()))
		}
		i++
	}
	fmt.Println("")
	for _, s := range onExistingNodes {
		fmt.Println(s)
	}
}

func (collector *OutgrowthCollector) displayDebug() {
	collector.sameEventStat.displayDebug(collector.sameEvent)
	collector.multiEventStat.displayDebug(collector.multiEvents)
}

const (
	HistogramDelta = 2
)

func (colStat *OutgrowthCollectorStat) beginRealize(data map[Point]*[]*NewPossibleOutgrowth) {
	colStat.originalPoints = len(data)
	colStat.originalHistogram = make([]int, 1)
	origPos := 0
	for _, l := range data {
		size := len(*l)
		origPos += size
		currentSize := len(colStat.originalHistogram)
		currentPos := size - HistogramDelta
		if currentPos >= currentSize {
			newHistogram := make([]int, currentPos+1)
			for i, v := range colStat.originalHistogram {
				newHistogram[i] = v
			}
			colStat.originalHistogram = newHistogram
		}
		colStat.originalHistogram[currentPos]++
	}
	colStat.occupiedHistogram = make([]int, len(colStat.originalHistogram))
	colStat.noMoreConnHistogram = make([]int, len(colStat.originalHistogram))
	colStat.originalPossible = origPos
}

func (collector *OutgrowthCollector) beginRealize() {
	// Remove all single that don't have new state (usually same event or multi event)
	newSingleMap := make(map[Point]*NewPossibleOutgrowth, len(collector.single)-len(collector.sameEvent)-len(collector.multiEvents))
	for p, e := range collector.single {
		if e.state == EventOutgrowthNew {
			newSingleMap[p] = e
		}
	}
	collector.single = newSingleMap

	// For single points and possible are the same
	collector.singleStat.originalPoints = len(collector.single)
	collector.singleStat.originalPossible = len(collector.single)

	collector.sameEventStat.beginRealize(collector.sameEvent)
	collector.multiEventStat.beginRealize(collector.multiEvents)
}

func (space *Space) realizeAllOutgrowth(collector *OutgrowthCollector) {
	collector.beginRealize()
	if DEBUG {
		collector.displayDebug()
	}

	// No problem just realize all single ones that fit
	for _, newPosEo := range collector.single {
		collector.singleStat.newPoint = true
		collector.singleStat.realizeAndStat(newPosEo, 1)
	}
	collector.singleStat.displayStat()

	// Realize only one of conflicting same event
	for _, newPosEoList := range collector.sameEvent {
		collector.sameEventStat.newPoint = true
		var newEo *EventOutgrowth
		for _, newPosEo := range *newPosEoList {
			if newEo == nil {
				newEo = collector.sameEventStat.realizeAndStat(newPosEo, len(*newPosEoList))
			} else {
				newEo.AddFrom(newPosEo.from)
			}
		}
	}
	collector.sameEventStat.displayStat()

	// Realize only one per event of conflicting multi events
	// Collect all more than 3 event outgrowth
	for pos, newPosEoList := range collector.multiEvents {
		collector.multiEventStat.newPoint = true
		idsAlreadyDone := make(map[EventID]*EventOutgrowth, 2)
		for _, newPosEo := range *newPosEoList {
			doneEo, done := idsAlreadyDone[newPosEo.event.id]
			if !done {
				newEo := collector.multiEventStat.realizeAndStat(newPosEo, len(*newPosEoList))
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
		fmt.Println("Finally got triple sync at:", collector.moreThan3)
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
			fmt.Println("ERROR: Did not manage to process", nbLatest, "latest event outgrowth from", nbEvents, "events in", nbLatest*5, "msecs")
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
		if DEBUG {
			fmt.Println("Already event", newNode.GetStateString())
		}
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
