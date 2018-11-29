package m3space

import (
	"bytes"
	"fmt"
	"github.com/freddy33/qsm-go/m3util"
	"sort"
	"time"
)

var LogStat = m3util.NewLogger("m3stat ", m3util.INFO)

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
	OutgrowthCollectorStatSingle
	totalOriginalPossible       int
	nbEventsOriginalHistogram   []int
	totalOccupiedPossible       int
	nbEventsOccupiedHistogram   []int
	totalNoMoreConnPossible     int
	nbEventsNoMoreConnHistogram []int
	newPoint                    bool
	perEvent                    map[EventID]*OutgrowthCollectorStatSameEvent
	moreThan3EventsPerPoint     map[Point][]EventID
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

func (space *Space) ForwardTime() *FullOutgrowthCollector {
	nbLatest := 0
	for _, evt := range space.events {
		nbLatest += len(evt.latestOutgrowths)
	}
	Log.Infof("Stepping up to %d: %d events, %d actNodes, %d actConn, %d latestEO, %d oldNodes, %d oldConn",
		space.currentTime+1, len(space.events), len(space.activeNodesMap), len(space.activeConnections), nbLatest,
		len(space.oldNodesMap), len(space.oldConnections))
	LogStat.Infof("%d: %d, %d, %d, %d, %d, %d",
		space.currentTime, len(space.events), len(space.activeNodesMap), len(space.activeConnections), nbLatest,
		len(space.oldNodesMap), len(space.oldConnections))
	c := make(chan *NewPossibleOutgrowth, 100)
	for _, evt := range space.events {
		go evt.createNewPossibleOutgrowths(c)
	}
	collector := space.processNewOutgrowth(c, nbLatest)

	space.currentTime++

	space.realizeAllOutgrowth(collector)
	// Switch latest to old, and new to latest
	for _, evt := range space.events {
		evt.moveNewOutgrowthsToLatest()
	}
	space.moveOldToOldMaps()
	return collector
}

func (evt *Event) createNewPossibleOutgrowths(c chan *NewPossibleOutgrowth) {
	for _, eg := range evt.latestOutgrowths {
		if eg.state != EventOutgrowthLatest {
			Log.Fatalf("Wrong state of event! found non latest outgrowth %v at %v in latest list.", eg, *(eg.pos))
		}

		nextPoints := eg.pos.getNextPoints(&(evt.growthContext))
		for _, nextPoint := range nextPoints {
			if !eg.CameFromPoint(nextPoint) {
				sendOutgrowth := true
				nodeThere := evt.space.GetNode(nextPoint)
				if nodeThere != nil {
					sendOutgrowth = nodeThere.CanReceiveEvent(evt.id)
					Log.Trace("New EO on existing node", nodeThere.GetStateString(), "can receive=", sendOutgrowth)
				}
				if sendOutgrowth {
					Log.Trace("Creating new possible event outgrowth for", evt.id, "at", nextPoint)
					c <- &NewPossibleOutgrowth{nextPoint, evt, eg, eg.distance + 1, EventOutgrowthNew}
				}
			}
		}
	}
	Log.Debug("Finished with event outgrowth for", evt.id, "sending End state possible outgrowth")
	c <- &NewPossibleOutgrowth{*(evt.node.Pos), evt, nil, Distance(0), EventOutgrowthEnd}
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
				collector.addNewEo(newEo)
			default:
				Log.Error("Receive an event on channel with wrong state", newEo.state, "full=", *(newEo))
			}

			if nbEventsDone == nbEvents {
				stop = true
				break
			}
		case <-time.After(time.Duration(timeout) * time.Millisecond):
			stop = true
			Log.Error("Did not manage to process", nbLatest, "latest event outgrowth from", nbEvents, "events in", nbLatest*5, "msecs")
			break
		}
		if stop {
			break
		}
	}

	return collector
}

func (space *Space) realizeAllOutgrowth(collector *FullOutgrowthCollector) {
	collector.beginRealize()
	collector.displayTrace()

	// No problem just realize all single ones that fit
	for _, newPosEo := range collector.single.data {
		collector.single.realizeSingle(newPosEo)
	}
	collector.single.displayStat()

	// Realize only one of conflicting same event
	for _, newPosEoList := range collector.sameEvent.data {
		collector.sameEvent.setNewPoint()
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
		collector.multiEvents.setNewPoint()

		idsAlreadyDone := make(map[EventID]*EventOutgrowth, len(*evtMapList))
		for id, newPosEoList := range *evtMapList {
			for _, newPosEo := range *newPosEoList {
				doneEo, done := idsAlreadyDone[id]
				if !done {
					newEo := collector.multiEvents.realizeMultiEvent(newPosEo, len(*evtMapList), len(*newPosEoList))
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
			SortEventIDs(&ids)
			collector.multiEvents.moreThan3EventsPerPoint[pos] = ids
		}
	}
	collector.multiEvents.displayStat()
}

func SortEventIDs(ids *[]EventID) {
	sort.Slice(*ids, func(i, j int) bool {
		return (*ids)[i] < (*ids)[j]
	})
}

func MakeOutgrowthCollector(nbLatest int) *FullOutgrowthCollector {
	if nbLatest < 5 {
		nbLatest = 5
	}
	res := FullOutgrowthCollector{}
	res.single.data = make(map[Point]*NewPossibleOutgrowth, 2*nbLatest)
	res.sameEvent.data = make(map[Point]*[]*NewPossibleOutgrowth, nbLatest/3)
	res.multiEvents.data = make(map[Point]*map[EventID]*[]*NewPossibleOutgrowth, 10)
	res.single.name = "Single"
	res.sameEvent.name = "Same Event"
	res.multiEvents.name = "Multi Events"
	return &res
}

const (
	StartWithOneHistoDelta = 1
	StartWithTwoHistoDelta = 2
)

func (collector *FullOutgrowthCollector) beginRealize() {
	collector.single.beginRealize()
	collector.sameEvent.beginRealize()
	collector.multiEvents.beginRealize()
}

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

func (colStat *OutgrowthCollectorStatSameEvent) beginRealizeLine(l *[]*NewPossibleOutgrowth, histoDelta int) {
	size := len(*l)
	colStat.originalPossible += size
	currentSize := len(colStat.originalHistogram)
	currentPos := size - histoDelta
	if currentPos >= currentSize {
		newHistogram := make([]int, currentPos+1)
		for i, v := range colStat.originalHistogram {
			newHistogram[i] = v
		}
		colStat.originalHistogram = newHistogram
	}
	colStat.originalHistogram[currentPos]++
}

func (col *OutgrowthCollectorSameEvent) beginRealize() {
	col.originalPoints = len(col.data)
	if col.originalPoints == 0 {
		return
	}
	col.originalHistogram = make([]int, 1)
	col.originalPossible = 0
	for _, l := range col.data {
		col.beginRealizeLine(l, StartWithTwoHistoDelta)
	}
	col.occupiedHistogram = make([]int, len(col.originalHistogram))
	col.noMoreConnHistogram = make([]int, len(col.originalHistogram))
}

func (col *OutgrowthCollectorMultiEvent) beginRealize() {
	col.originalPoints = len(col.data)
	if col.originalPoints == 0 {
		return
	}
	col.totalOriginalPossible = 0
	col.moreThan3EventsPerPoint = make(map[Point][]EventID, 1+col.originalPoints/10)
	col.nbEventsOriginalHistogram = make([]int, 2)
	col.perEvent = make(map[EventID]*OutgrowthCollectorStatSameEvent, 3)
	for _, evtMapList := range col.data {
		size := len(*evtMapList)
		currentSize := len(col.nbEventsOriginalHistogram)
		currentPos := size - StartWithTwoHistoDelta
		if currentPos >= currentSize {
			newHistogram := make([]int, currentPos+1)
			for i, v := range col.nbEventsOriginalHistogram {
				newHistogram[i] = v
			}
			col.nbEventsOriginalHistogram = newHistogram
		}
		col.nbEventsOriginalHistogram[currentPos]++
		for id, l := range *evtMapList {
			stat, ok := col.perEvent[id]
			if !ok {
				stat = &OutgrowthCollectorStatSameEvent{}
				stat.name = fmt.Sprintf("Evt %d", id)
				stat.originalHistogram = make([]int, 1)
				stat.originalPossible = 0
				col.perEvent[id] = stat
			}
			stat.beginRealizeLine(l, StartWithOneHistoDelta)
			col.totalOriginalPossible += len(*l)
		}
	}
	col.nbEventsOccupiedHistogram = make([]int, len(col.nbEventsOriginalHistogram))
	col.nbEventsNoMoreConnHistogram = make([]int, len(col.nbEventsOriginalHistogram))
	for _, stat := range col.perEvent {
		stat.occupiedHistogram = make([]int, len(stat.originalHistogram))
		stat.noMoreConnHistogram = make([]int, len(stat.originalHistogram))
	}
}

func (colStat *OutgrowthCollectorStatSingle) processRealizeError(err error) {
	switch err.(type) {
	case *EventAlreadyGrewThereError:
		colStat.occupiedPoints++
	case *NoMoreConnectionsError:
		colStat.noMoreConnPoints++
	}
}

func (colStat *OutgrowthCollectorStatSingle) realizeSingle(newPosEo *NewPossibleOutgrowth) *EventOutgrowth {
	newEo, err := newPosEo.realize()
	if err != nil {
		colStat.processRealizeError(err)
		return nil
	}
	return newEo
}

func (colStat *OutgrowthCollectorStatSameEvent) setNewPoint() {
	colStat.newPoint = true
}

func (colStat *OutgrowthCollectorStatSameEvent) processRealizeError(err error, size int, histoDelta int) {
	switch err.(type) {
	case *EventAlreadyGrewThereError:
		colStat.occupiedPossible++
		if size > 1 {
			colStat.occupiedHistogram[size-histoDelta]++
		}
		if colStat.newPoint {
			colStat.occupiedPoints++
			colStat.newPoint = false
		}
	case *NoMoreConnectionsError:
		colStat.noMoreConnPossible++
		if size > 1 {
			colStat.noMoreConnHistogram[size-histoDelta]++
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
		colStat.processRealizeError(err, size, StartWithTwoHistoDelta)
		return nil, err
	}
	return newEo, nil
}

func (colStat *OutgrowthCollectorStatMultiEvent) setNewPoint() {
	colStat.newPoint = true
	for _, stat := range colStat.perEvent {
		stat.setNewPoint()
	}
}

func (colStat *OutgrowthCollectorStatMultiEvent) processRealizeError(err error, eventId EventID, nbEvents, size int) {
	stat := colStat.perEvent[eventId]
	stat.processRealizeError(err, size, StartWithOneHistoDelta)
	switch err.(type) {
	case *EventAlreadyGrewThereError:
		colStat.totalOccupiedPossible++
		if nbEvents > 1 {
			colStat.nbEventsOccupiedHistogram[nbEvents-StartWithTwoHistoDelta]++
		}
		if colStat.newPoint {
			colStat.occupiedPoints++
			colStat.newPoint = false
		}
	case *NoMoreConnectionsError:
		colStat.totalNoMoreConnPossible++
		if nbEvents > 1 {
			colStat.nbEventsNoMoreConnHistogram[nbEvents-StartWithTwoHistoDelta]++
		}
		if colStat.newPoint {
			colStat.noMoreConnPoints++
			colStat.newPoint = false
		}
	}
}

func (colStat *OutgrowthCollectorStatMultiEvent) realizeMultiEvent(newPosEo *NewPossibleOutgrowth, nbEvents, size int) *EventOutgrowth {
	newEo, err := newPosEo.realize()
	if err != nil {
		colStat.processRealizeError(err, newPosEo.event.id, nbEvents, size)
		return nil
	}
	return newEo
}

func (colStat *OutgrowthCollectorStatSingle) displayStat() {
	if colStat.originalPoints == 0 {
		// nothing to show skip
		return
	}
	// Only debug
	if LogStat.Level > m3util.DEBUG {
		return
	}
	LogStat.Debugf("%12s  : %6d / %6d / %6d", colStat.name,
		colStat.originalPoints, colStat.occupiedPoints, colStat.noMoreConnPoints)
}

func (colStat *OutgrowthCollectorStatSameEvent) displayStat() {
	if colStat.originalPoints == 0 {
		// nothing to show skip
		return
	}
	// Only debug
	if LogStat.Level > m3util.DEBUG {
		return
	}
	buf := bytes.NewBufferString("")
	fmt.Fprintf(buf, "%12s  : %6d / %6d / %6d | %6d / %6d / %6d", colStat.name,
		colStat.originalPoints, colStat.occupiedPoints, colStat.noMoreConnPoints,
		colStat.originalPossible, colStat.occupiedPossible, colStat.noMoreConnPossible)
	if len(colStat.originalHistogram) > 1 {
		for i, j := range colStat.originalHistogram {
			fmt.Fprintf(buf, "\n%12s %d: %6d / %6d / %6d", colStat.name, i+StartWithTwoHistoDelta, j,
				colStat.occupiedHistogram[i], colStat.noMoreConnHistogram[i])
		}
	}
	LogStat.Debug(buf.String())
}

func (colStat *OutgrowthCollectorStatMultiEvent) displayStat() {
	if colStat.originalPoints == 0 {
		// nothing to show skip
		return
	}

	// Only debug
	if LogStat.Level <= m3util.DEBUG {
		buf := bytes.NewBufferString("")
		fmt.Fprintf(buf, "%12s  : %6d / %6d / %6d | %6d / %6d / %6d", colStat.name,
			colStat.originalPoints, colStat.occupiedPoints, colStat.noMoreConnPoints,
			colStat.totalOriginalPossible, colStat.totalOccupiedPossible, colStat.totalNoMoreConnPossible)
		if len(colStat.nbEventsOriginalHistogram) > 1 {
			for i, j := range colStat.nbEventsOriginalHistogram {
				fmt.Fprintf(buf, "\n%12s %d: %6d / %6d / %6d", colStat.name, i+StartWithTwoHistoDelta, j,
					colStat.nbEventsOccupiedHistogram[i], colStat.nbEventsNoMoreConnHistogram[i])
			}
		}
		LogStat.Debug(buf.String())
		for _, se := range colStat.perEvent {
			se.displayStat()
		}
	}

	nbEventsPerPoints := len(colStat.moreThan3EventsPerPoint)
	if nbEventsPerPoints == 0 {
		// Nothing left
		return
	}
	buf := bytes.NewBufferString("")
	fmt.Fprintf(buf, "Points with more than 3 events %d \n", nbEventsPerPoints)
	for pos, ePerPos := range colStat.moreThan3EventsPerPoint {
		fmt.Fprintln(buf, "Triple sync at:", pos, "for", ePerPos)
	}
	Log.Info(buf.String())
	LogStat.Info(buf.String())
}

func (col *OutgrowthCollectorSameEvent) displayTrace() {
	dataLength := len(col.data)
	if dataLength == 0 {
		// nothing to show skip
		return
	}
	buf := bytes.NewBufferString("")
	bufExistNodes := bytes.NewBufferString("")
	fmt.Fprintf(buf, "%s %d:", col.name, dataLength)
	i := 0
	for p, l := range col.data {
		if i%6 == 0 {
			fmt.Fprint(buf, "\n")
		}
		fmt.Fprintf(buf, "%v=%d  ", p, len(*l))
		node := (*l)[0].event.space.GetNode(p)
		if node != nil {
			fmt.Fprintf(bufExistNodes, "%v:%s\n", p, node.GetStateString())
		}
		i++
	}
	Log.Trace(buf.String())
	if bufExistNodes.Len() > 0 {
		Log.Trace(bufExistNodes.String())
	}
}

func (col *OutgrowthCollectorMultiEvent) displayTrace() {
	dataLength := len(col.data)
	if dataLength == 0 {
		// nothing to show skip
		return
	}
	buf := bytes.NewBufferString("")
	bufExistNodes := bytes.NewBufferString("")
	fmt.Fprintf(buf, "%s %d\n", col.name, dataLength)
	for pos, idToList := range col.data {
		fmt.Fprintf(buf, "%v=%d", pos, len(*idToList))
		i := 0
		for id, l := range *idToList {
			if i%6 == 0 {
				fmt.Fprint(buf, "\n")
			}
			fmt.Fprintf(buf, "%d=%d  ", id, len(*l))
			node := (*l)[0].event.space.GetNode(pos)
			if node != nil {
				fmt.Fprintf(bufExistNodes, "%v:%s\n", pos, node.GetStateString())
			}
			i++
		}
		Log.Trace(buf.String())
		if bufExistNodes.Len() > 0 {
			Log.Trace(bufExistNodes.String())
		}
	}
}

func (collector *FullOutgrowthCollector) displayTrace() {
	if Log.Level <= m3util.TRACE {
		collector.sameEvent.displayTrace()
		collector.multiEvents.displayTrace()
	}
}

func (col *OutgrowthCollectorSameEvent) addNewEoFromSingle(newEo *NewPossibleOutgrowth, fromSingle *NewPossibleOutgrowth) {
	_, okSameEvent := col.data[newEo.pos]
	if okSameEvent {
		Log.Error("An event outgrowth in single map with state", fromSingle.state, "full=", *(fromSingle), "is state new but has an entry in the multi same event Map!!")
	}

	newSameEventList := make([]*NewPossibleOutgrowth, 2, 3)
	newSameEventList[0] = fromSingle
	newSameEventList[1] = newEo
	fromSingle.state = EventOutgrowthManySameEvent
	newEo.state = EventOutgrowthManySameEvent
	col.data[newEo.pos] = &newSameEventList
}

func (col *OutgrowthCollectorMultiEvent) addNewEoFromSingle(newEo *NewPossibleOutgrowth, fromSingle *NewPossibleOutgrowth) {
	_, okMultiEvent := col.data[newEo.pos]
	if okMultiEvent {
		Log.Error("An event outgrowth in single map with state", fromSingle.state, "full=", *(fromSingle), "is state new but has an entry in the multi events Map!!")
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
	col.data[newEo.pos] = &newMultiEventMap
}

func (col *OutgrowthCollectorMultiEvent) addNewEoAlreadyMultiEvent(newEo *NewPossibleOutgrowth, fromSingle *NewPossibleOutgrowth) {
	multiEventsMap, okMultiEventsMap := col.data[newEo.pos]
	if !okMultiEventsMap {
		Log.Error("An event outgrowth in single map with state", fromSingle.state, "full=", *(fromSingle), "does not have an entry in the multi events Map!!")
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

func (collector *FullOutgrowthCollector) addNewEo(newEo *NewPossibleOutgrowth) {
	fromSingle, ok := collector.single.data[newEo.pos]
	if !ok {
		collector.single.data[newEo.pos] = newEo
	} else {
		switch fromSingle.state {
		case EventOutgrowthNew:
			// First multiple entry, check if same event or not and move event outgrowth there
			if fromSingle.event.id == newEo.event.id {
				collector.sameEvent.addNewEoFromSingle(newEo, fromSingle)
			} else {
				collector.multiEvents.addNewEoFromSingle(newEo, fromSingle)
			}
		case EventOutgrowthManySameEvent:
			fromSameEvent, okSameEvent := collector.sameEvent.data[newEo.pos]
			if !okSameEvent {
				Log.Error("An event outgrowth in single map with state", fromSingle.state, "full=", *(fromSingle), "does not have an entry in the multi same event Map!!")
			} else {
				if fromSingle.event.id == newEo.event.id {
					newEo.state = EventOutgrowthManySameEvent
					*fromSameEvent = append(*fromSameEvent, newEo)
				} else {
					// Move all from1 same event to multi event
					_, okMultiEvent := collector.multiEvents.data[newEo.pos]
					if okMultiEvent {
						Log.Error("An event outgrowth in multi same event map with state", fromSingle.state, "full=", *(fromSingle), "is state same event but has an entry in the multi events Map!!")
					}

					newEo.state = EventOutgrowthMultipleEvents
					for _, eo := range *fromSameEvent {
						eo.state = EventOutgrowthMultipleEvents
					}
					// Just verify
					if newEo.state != EventOutgrowthMultipleEvents || fromSingle.state != EventOutgrowthMultipleEvents {
						Log.Error("Event outgrowth state change failed for", *fromSingle, "and", *newEo)
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
			collector.multiEvents.addNewEoAlreadyMultiEvent(newEo, fromSingle)
		}
	}
}

func (newPosEo *NewPossibleOutgrowth) realize() (*EventOutgrowth, error) {
	evt := newPosEo.event
	space := evt.space
	newNode := space.getOrCreateNode(newPosEo.pos)
	if !newNode.CanReceiveOutgrowth(newPosEo) {
		// Should have been filtered at new outgrowth creation
		Log.Warn("Event", newPosEo.event.id, "already occupy node", newNode.GetStateString())
		return nil, &EventAlreadyGrewThereError{newPosEo.event.id, newPosEo.pos,}
	}
	fromPoint := newPosEo.from.pos
	fromNode := space.GetNode(*fromPoint)
	if !fromNode.IsAlreadyConnected(newNode) {
		Log.Trace("Need to connect the two nodes", fromNode.GetStateString(), newNode.GetStateString())
		if space.makeConnection(fromNode, newNode) == nil {
			// No more connections
			Log.Debug("Two nodes", fromNode.GetStateString(), newNode.GetStateString(), "cannot be connected without exceeding", newPosEo.event.space.MaxConnections, "connections")
			return nil, &NoMoreConnectionsError{*(newNode.Pos), *(fromPoint)}
		}
	}
	newEo := &EventOutgrowth{newNode.Pos, []*EventOutgrowth{newPosEo.from,}, newPosEo.distance, EventOutgrowthNew}
	evt.latestOutgrowths = append(evt.latestOutgrowths, newEo)
	newNode.AddOutgrowth(evt.id, space.currentTime)
	Log.Trace("Created new outgrowth", newEo.String())
	return newEo, nil
}

func (evt *Event) moveNewOutgrowthsToLatest() {
	finalLatest := evt.latestOutgrowths[:0]
	for _, eg := range evt.latestOutgrowths {
		switch eg.state {
		case EventOutgrowthLatest:
			eg.state = EventOutgrowthCurrent
			evt.currentOutgrowths = append(evt.currentOutgrowths, eg)
		case EventOutgrowthNew:
			eg.state = EventOutgrowthLatest
			finalLatest = append(finalLatest, eg)
		}
	}
	evt.latestOutgrowths = finalLatest

	finalCurrent := evt.currentOutgrowths[:0]
	for _, eg := range evt.currentOutgrowths {
		if eg.state == EventOutgrowthCurrent && eg.IsOld(evt) {
			eg.state = EventOutgrowthOld
			evt.oldOutgrowths = append(evt.oldOutgrowths, eg)
		} else {
			finalCurrent = append(finalCurrent, eg)
		}
	}
	evt.currentOutgrowths = finalCurrent
}

func (space *Space) moveOldToOldMaps() {
	for p, node := range space.activeNodesMap {
		if node.IsOld(space) {
			delete(space.activeNodesMap, p)
			space.oldNodesMap[p] = node
		}
	}
	finalActive := space.activeConnections[:0]
	for _, conn := range space.activeConnections {
		if conn.IsOld(space) {
			space.oldConnections = append(space.oldConnections, conn)
		} else {
			finalActive = append(finalActive, conn)
		}
	}
	space.activeConnections = finalActive
}
