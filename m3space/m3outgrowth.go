package m3space

import (
	"fmt"
	"sort"
	"time"
)

type OutgrowthCollectorStatSingle struct {
	name             string
	originalPoints   int
	occupiedPoints   int
	noMoreConnPoints int
}

type OutgrowthCollectorStatSameEvent struct {
	OutgrowthCollectorStatSingle
	originalPossible    int
	originalHistogram   []int
	occupiedPossible    int
	occupiedHistogram   []int
	noMoreConnPossible  int
	noMoreConnHistogram []int
	newPoint            bool
}

type OutgrowthCollectorStatMultiEvent struct {
	OutgrowthCollectorStatSameEvent
	perEvent       map[EventID]*OutgrowthCollectorStatSingle
	eventsPerPoint map[Point][]EventID
}

type OutgrowthCollectorSingle struct {
	OutgrowthCollectorStatSingle
	data map[Point]*NewPossibleOutgrowth
}

type OutgrowthCollectorSameEvent struct {
	OutgrowthCollectorStatSameEvent
	data map[Point]*[]*NewPossibleOutgrowth
}

type OutgrowthCollectorMultiEvent struct {
	OutgrowthCollectorStatMultiEvent
	data map[Point]*map[EventID]*[]*NewPossibleOutgrowth
}

type FullOutgrowthCollector struct {
	single      OutgrowthCollectorSingle
	sameEvent   OutgrowthCollectorSameEvent
	multiEvents OutgrowthCollectorMultiEvent
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

func MakeOutgrowthCollector(nbLatest int) *FullOutgrowthCollector {
	if nbLatest < 5 {
		nbLatest = 5
	}
	res := FullOutgrowthCollector{}
	res.single.data = make(map[Point]*NewPossibleOutgrowth, 2*nbLatest)
	res.sameEvent.data = make(map[Point]*[]*NewPossibleOutgrowth, nbLatest/3)
	res.multiEvents.data = make(map[Point]*map[EventID]*[]*NewPossibleOutgrowth, 10)
	res.multiEvents.eventsPerPoint = make(map[Point][]EventID, 3)
	res.single.name = "Single"
	res.sameEvent.name = "Same Event"
	res.multiEvents.name = "Multi Events"
	return &res
}

func (colStat *OutgrowthCollectorStatSingle) processRealizeError(err error) {
	switch err.(type) {
	case *EventAlreadyGrewThereError:
		colStat.occupiedPoints++
	case *NoMoreConnectionsError:
		colStat.noMoreConnPoints++
	}
}

func (colStat *OutgrowthCollectorStatSingle) realize(newPosEo *NewPossibleOutgrowth) *EventOutgrowth {
	newEo, err := newPosEo.realize()
	if err != nil {
		colStat.processRealizeError(err)
		return nil
	}
	return newEo
}

func (colStat *OutgrowthCollectorStatSameEvent) processRealizeError(err error, size int) {
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
}

func (colStat *OutgrowthCollectorStatSameEvent) realizeSameEvent(newPosEo *NewPossibleOutgrowth, size int) (*EventOutgrowth, error) {
	newEo, err := newPosEo.realize()
	if err != nil {
		colStat.processRealizeError(err, size)
		return nil, err
	}
	return newEo, nil
}

func (colStat *OutgrowthCollectorStatMultiEvent) realizeMultiEvent(newPosEo *NewPossibleOutgrowth, size int) *EventOutgrowth {
	newEo, err := (*colStat).OutgrowthCollectorStatSameEvent.realizeSameEvent(newPosEo, size)
	if err != nil {
		colStat.perEvent[newPosEo.event.id].processRealizeError(err)
		return nil
	}
	return newEo
}

func (colStat *OutgrowthCollectorStatSingle) displayStat() {
	if colStat.originalPoints == 0 {
		// nothing to show skip
		return
	}
	fmt.Printf("%12s  : %6d / %6d / %6d\n", colStat.name,
		colStat.originalPoints, colStat.occupiedPoints, colStat.noMoreConnPoints)
}

func (colStat *OutgrowthCollectorStatSameEvent) displayStat() {
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

func (colStat *OutgrowthCollectorStatMultiEvent) displayStat() {
	(*colStat).OutgrowthCollectorStatSameEvent.displayStat()
	for _, se := range colStat.perEvent {
		se.displayStat()
	}
	for pos, ePerPos := range colStat.eventsPerPoint {
		if len(ePerPos) >= THREE {
			fmt.Println("###############################################")
			fmt.Println("Finally got triple sync at:", pos, "for", ePerPos)
			fmt.Println("###############################################")
		}
	}
}

func (col *OutgrowthCollectorSameEvent) displayDebug() {
	dataLength := len(col.data)
	if dataLength == 0 {
		// nothing to show skip
		return
	}
	fmt.Printf("%s %d:", col.name, dataLength)
	i := 0
	onExistingNodes := make([]string, 0)
	for p, l := range col.data {
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

func (col *OutgrowthCollectorMultiEvent) displayDebug() {
	dataLength := len(col.data)
	if dataLength == 0 {
		// nothing to show skip
		return
	}
	fmt.Printf("%s %d\n", col.name, dataLength)
	for pos, idToList := range col.data {
		fmt.Printf("%v=%d", pos, len(*idToList))
		onExistingNodes := make([]string, 0)
		i := 0
		for id, l := range *idToList {
			if i%6 == 0 {
				fmt.Println("")
			}
			fmt.Printf("%d=%d  ", id, len(*l))
			node := (*l)[0].event.space.GetNode(pos)
			if node != nil {
				onExistingNodes = append(onExistingNodes, fmt.Sprintf("%v:%s", pos, node.GetStateString()))
			}
			i++
		}
		fmt.Println("")
		for _, s := range onExistingNodes {
			fmt.Println(s)
		}
	}
}

func (collector *FullOutgrowthCollector) displayDebug() {
	collector.sameEvent.displayDebug()
	collector.multiEvents.displayDebug()
}

const (
	HistogramDelta = 2
)

func (col *OutgrowthCollectorSingle) beginRealize() {
	// Remove all single that don't have new state (usually same event or multi event)
	for p, e := range col.data {
		if e.state != EventOutgrowthNew {
			delete(col.data, p)
		}
	}

	// For single points and possible are the same
	col.originalPoints = len(col.data)
}

func (col *OutgrowthCollectorSameEvent) beginRealize() {
	col.originalPoints = len(col.data)
	col.originalHistogram = make([]int, 1)
	origPos := 0
	for _, l := range col.data {
		size := len(*l)
		origPos += size
		currentSize := len(col.originalHistogram)
		currentPos := size - HistogramDelta
		if currentPos >= currentSize {
			newHistogram := make([]int, currentPos+1)
			for i, v := range col.originalHistogram {
				newHistogram[i] = v
			}
			col.originalHistogram = newHistogram
		}
		col.originalHistogram[currentPos]++
	}
	col.occupiedHistogram = make([]int, len(col.originalHistogram))
	col.noMoreConnHistogram = make([]int, len(col.originalHistogram))
	col.originalPossible = origPos
}

func (col *OutgrowthCollectorMultiEvent) beginRealize() {
	col.originalPoints = len(col.data)
	col.originalHistogram = make([]int, 1)
	col.perEvent = make(map[EventID]*OutgrowthCollectorStatSingle, 3)
	origPos := 0
	for _, evtMapList := range col.data {
		size := len(*evtMapList)
		origPos += size
		currentSize := len(col.originalHistogram)
		currentPos := size - HistogramDelta
		if currentPos >= currentSize {
			newHistogram := make([]int, currentPos+1)
			for i, v := range col.originalHistogram {
				newHistogram[i] = v
			}
			col.originalHistogram = newHistogram
		}
		col.originalHistogram[currentPos]++
		for id, l := range *evtMapList {
			stat, ok := col.perEvent[id]
			if !ok {
				stat = &OutgrowthCollectorStatSingle{}
				stat.name = fmt.Sprintf("Evt %d", id)
				col.perEvent[id] = stat
			}
			stat.originalPoints += len(*l)
			origPos += len(*l)
		}
	}
	col.occupiedHistogram = make([]int, len(col.originalHistogram))
	col.noMoreConnHistogram = make([]int, len(col.originalHistogram))
	col.originalPossible = origPos
}

func (collector *FullOutgrowthCollector) beginRealize() {
	collector.single.beginRealize()
	collector.sameEvent.beginRealize()
	collector.multiEvents.beginRealize()
}

func (space *Space) realizeAllOutgrowth(collector *FullOutgrowthCollector) {
	collector.beginRealize()
	if DEBUG {
		collector.displayDebug()
	}

	// No problem just realize all single ones that fit
	for _, newPosEo := range collector.single.data {
		collector.single.realize(newPosEo)
	}
	collector.single.displayStat()

	// Realize only one of conflicting same event
	for _, newPosEoList := range collector.sameEvent.data {
		collector.sameEvent.newPoint = true
		var newEo *EventOutgrowth
		for _, newPosEo := range *newPosEoList {
			if newEo == nil {
				newEo, _ = collector.sameEvent.realizeSameEvent(newPosEo, len(*newPosEoList))
			} else {
				newEo.AddFrom(newPosEo.from)
			}
		}
	}
	collector.sameEvent.displayStat()

	// Realize only one per event of conflicting multi events
	// Collect all more than 3 event outgrowth
	for pos, evtMapList := range collector.multiEvents.data {
		collector.multiEvents.newPoint = true

		idsAlreadyDone := make(map[EventID]*EventOutgrowth, len(*evtMapList))
		for id, newPosEoList := range *evtMapList {
			for _, newPosEo := range *newPosEoList {
				doneEo, done := idsAlreadyDone[id]
				if !done {
					newEo := collector.multiEvents.realizeMultiEvent(newPosEo, len(*newPosEoList))
					if newEo != nil {
						idsAlreadyDone[id] = newEo
					}
				} else {
					doneEo.AddFrom(newPosEo.from)
				}
			}
		}
		if len(idsAlreadyDone) >= THREE {
			ids := make([]EventID, len(idsAlreadyDone))
			i := 0
			for id := range idsAlreadyDone {
				ids[i] = id
				i++
			}
			sort.Slice(ids, func(i, j int) bool {
				return ids[i] < ids[j]
			})
			collector.multiEvents.eventsPerPoint[pos] = ids
		}
	}
	collector.multiEvents.displayStat()
}

func (space *Space) processNewOutgrowth(c chan *NewPossibleOutgrowth, nbLatest int) *FullOutgrowthCollector {
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
				fromSingle, ok := collector.single.data[newEo.pos]
				if !ok {
					collector.single.data[newEo.pos] = newEo
				} else {
					switch fromSingle.state {
					case EventOutgrowthNew:
						// First multiple entry, check if same event or not and move event outgrowth there
						if fromSingle.event.id == newEo.event.id {
							_, okSameEvent := collector.sameEvent.data[newEo.pos]
							if okSameEvent {
								fmt.Println("ERROR: An event outgrowth in single map with state", fromSingle.state, "full=", *(fromSingle), "is state new but has an entry in the multi same event Map!!")
							}

							newSameEventList := make([]*NewPossibleOutgrowth, 2, 3)
							newSameEventList[0] = fromSingle
							newSameEventList[1] = newEo
							fromSingle.state = EventOutgrowthMultipleSameEvent
							newEo.state = EventOutgrowthMultipleSameEvent
							collector.sameEvent.data[newEo.pos] = &newSameEventList
						} else {
							_, okMultiEvent := collector.multiEvents.data[newEo.pos]
							if okMultiEvent {
								fmt.Println("ERROR: An event outgrowth in single map with state", fromSingle.state, "full=", *(fromSingle), "is state new but has an entry in the multi events Map!!")
							}

							newMultiEventMap := make(map[EventID]*[]*NewPossibleOutgrowth, 2)
							newMultiEventList1 := make([]*NewPossibleOutgrowth, 1)
							newMultiEventList2 := make([]*NewPossibleOutgrowth, 1)
							newMultiEventList1[0] = fromSingle
							newMultiEventList2[0] = newEo
							fromSingle.state = EventOutgrowthMultipleEvents
							newEo.state = EventOutgrowthMultipleEvents
							newMultiEventMap[fromSingle.event.id] = &newMultiEventList1
							newMultiEventMap[newEo.event.id] = &newMultiEventList2
							collector.multiEvents.data[newEo.pos] = &newMultiEventMap
						}
					case EventOutgrowthMultipleSameEvent:
						fromSameEvent, okSameEvent := collector.sameEvent.data[newEo.pos]
						if !okSameEvent {
							fmt.Println("ERROR: An event outgrowth in single map with state", fromSingle.state, "full=", *(fromSingle), "does not have an entry in the multi same event Map!!")
						} else {
							if fromSingle.event.id == newEo.event.id {
								newEo.state = EventOutgrowthMultipleSameEvent
								*fromSameEvent = append(*fromSameEvent, newEo)
							} else {
								// Move all from1 same event to multi event
								_, okMultiEvent := collector.multiEvents.data[newEo.pos]
								if okMultiEvent {
									fmt.Println("ERROR: An event outgrowth in multi same event map with state", fromSingle.state, "full=", *(fromSingle), "is state same event but has an entry in the multi events Map!!")
								}

								newEo.state = EventOutgrowthMultipleEvents
								for _, eo := range *fromSameEvent {
									eo.state = EventOutgrowthMultipleEvents
								}
								// Just verify
								if newEo.state != EventOutgrowthMultipleEvents || fromSingle.state != EventOutgrowthMultipleEvents {
									fmt.Println("ERROR: Event outgrowth state change failed for", *fromSingle, "and", *newEo)
								}

								delete(collector.sameEvent.data, newEo.pos)

								newMultiEventMap := make(map[EventID]*[]*NewPossibleOutgrowth, 2)
								newMultiEventList2 := make([]*NewPossibleOutgrowth, 1)
								newMultiEventList2[0] = newEo
								newMultiEventMap[fromSingle.event.id] = fromSameEvent
								newMultiEventMap[newEo.event.id] = &newMultiEventList2
								collector.multiEvents.data[newEo.pos] = &newMultiEventMap
							}
						}
					case EventOutgrowthMultipleEvents:
						multiEventsMap, okMultiEventsMap := collector.multiEvents.data[newEo.pos]
						if !okMultiEventsMap {
							fmt.Println("ERROR: An event outgrowth in single map with state", fromSingle.state, "full=", *(fromSingle), "does not have an entry in the multi events Map!!")
						} else {
							newEo.state = EventOutgrowthMultipleEvents
							fromMultiEvents, okMultiEvent := (*multiEventsMap)[newEo.event.id]
							if !okMultiEvent {
								newMultiEventList2 := make([]*NewPossibleOutgrowth, 1)
								newMultiEventList2[0] = newEo
								(*multiEventsMap)[newEo.event.id] = &newMultiEventList2
							} else {
								*fromMultiEvents = append(*fromMultiEvents, newEo)
							}
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
