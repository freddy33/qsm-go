package m3space

import "fmt"

type EventOutgrowthState uint8

const (
	EventOutgrowthLatest EventOutgrowthState = iota
	EventOutgrowthNew
	EventOutgrowthCurrent
	EventOutgrowthOld
	EventOutgrowthEnd
	EventOutgrowthManySameEvent
	EventOutgrowthMultipleEvents
)

type NewPossibleOutgrowth struct {
	pos      Point
	event    *Event
	from     *EventOutgrowth
	distance Distance
	state    EventOutgrowthState
}

type EventOutgrowth struct {
	pos      *Point
	fromList []Outgrowth
	distance Distance
	state    EventOutgrowthState
	rootPath PathElement
}

type SavedEventOutgrowth struct {
	pos             Point
	fromConnections []int8
	distance        Distance
	rootPath        PathElement
}

type Outgrowth interface {
	GetPoint() *Point
	GetDistance() Distance
	GetState() EventOutgrowthState
	AddFrom(from Outgrowth)
	GetFromConnIds() []int8
	BuildPath(path PathElement) PathElement
	CameFromPoint(point Point) bool
	FromLength() int
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

/***************************************************************/
// Realize Errors
/***************************************************************/

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

/***************************************************************/
// NewPossibleOutgrowth Functions
/***************************************************************/

func (newPosEo *NewPossibleOutgrowth) String() string {
	return fmt.Sprintf("NP %v %d: %s, %d", newPosEo.pos, newPosEo.event.id, newPosEo.state.String(), newPosEo.distance)
}

/***************************************************************/
// EventOutgrowth Functions
/***************************************************************/

func (eo *EventOutgrowth) String() string {
	return fmt.Sprintf("%v: %s, %d, %d", *(eo.pos), eo.state.String(), eo.distance, len(eo.fromList))
}

func (eo *EventOutgrowth) GetPoint() *Point {
	return eo.pos
}

func (eo *EventOutgrowth) GetDistance() Distance {
	return eo.distance
}

func (eo *EventOutgrowth) GetState() EventOutgrowthState {
	return eo.state
}

func (eo *EventOutgrowth) AddFrom(from Outgrowth) {
	if eo.fromList == nil {
		eo.fromList = []Outgrowth{from,}
	} else {
		eo.fromList = append(eo.fromList, from)
	}
}

func (eo *EventOutgrowth) HasFrom() bool {
	return eo.FromLength() > 0
}

func (eo *EventOutgrowth) FromLength() int {
	return len(eo.fromList)
}

func (eo *EventOutgrowth) GetFromConnIds() []int8 {
	res := make([]int8, len(eo.fromList))
	if len(res) == 0 {
		return res
	}
	for i, from := range eo.fromList {
		bv := MakeVector(*eo.pos, *from.GetPoint())
		res[i] = AllConnectionsPossible[bv].Id
	}
	return res
}

func (eo *EventOutgrowth) CameFromPoint(point Point) bool {
	if !eo.HasFrom() {
		return false
	}
	for _, from := range eo.fromList {
		if *(from.GetPoint()) == point {
			return true
		}
	}
	return false
}

func (eo *EventOutgrowth) IsRoot() bool {
	return !eo.HasFrom()
}

func (eo *EventOutgrowth) DistanceFromLatest(evt *Event) Distance {
	space := evt.space
	return Distance(space.currentTime-evt.created) - eo.distance
}

func (eo *EventOutgrowth) IsOld(evt *Event) bool {
	if eo.IsRoot() {
		// Root event always active
		return false
	}
	return eo.DistanceFromLatest(evt) >= evt.space.EventOutgrowthOldThreshold
}

func (eo *EventOutgrowth) IsActive(evt *Event) bool {
	if eo.IsRoot() {
		// Root event always active
		return true
	}
	return eo.DistanceFromLatest(evt) <= evt.space.EventOutgrowthThreshold
}

/***************************************************************/
// EventOutgrowth Functions
/***************************************************************/

func (seo *SavedEventOutgrowth) String() string {
	return fmt.Sprintf("%v: %s, %d, %d", seo.pos, EventOutgrowthOld.String(), seo.distance, len(seo.fromConnections))
}

func (seo *SavedEventOutgrowth) GetPoint() *Point {
	return &seo.pos
}

func (seo *SavedEventOutgrowth) GetDistance() Distance {
	return seo.distance
}

func (seo *SavedEventOutgrowth) GetState() EventOutgrowthState {
	return EventOutgrowthOld
}

func (seo *SavedEventOutgrowth) AddFrom(from Outgrowth) {
	Log.Errorf("Cannot add to from list on saved outgrowth %v <- %v", seo, from)
}

func (seo *SavedEventOutgrowth) HasFrom() bool {
	return seo.FromLength() > 0
}

func (seo *SavedEventOutgrowth) FromLength() int {
	return len(seo.fromConnections)
}

func (seo *SavedEventOutgrowth) GetFromConnIds(point Point) []int8 {
	return seo.fromConnections
}

func (seo *SavedEventOutgrowth) CameFromPoint(point Point) bool {
	if !seo.HasFrom() {
		return false
	}
	for _, fromConnId := range seo.fromConnections {
		cd := AllConnectionsIds[fromConnId]
		if seo.pos.Add(cd.Vector) == point {
			return true
		}
	}
	return false
}

func (seo *SavedEventOutgrowth) IsRoot() bool {
	return !seo.HasFrom()
}

func (seo *SavedEventOutgrowth) DistanceFromLatest(evt *Event) Distance {
	space := evt.space
	return Distance(space.currentTime-evt.created) - seo.distance
}

func (seo *SavedEventOutgrowth) IsOld(evt *Event) bool {
	if seo.IsRoot() {
		// Root event always active
		return false
	}
	return seo.DistanceFromLatest(evt) >= evt.space.EventOutgrowthOldThreshold
}

func (seo *SavedEventOutgrowth) IsActive(evt *Event) bool {
	if seo.IsRoot() {
		// Root event always active
		return true
	}
	return seo.DistanceFromLatest(evt) <= evt.space.EventOutgrowthThreshold
}

/***************************************************************/
// Path building Functions
/***************************************************************/

// An element in the path from event base node to latest outgrowth
// Forward is from event to outgrowth
// Backwards is from latest outgrowth to event
type PathElement interface {
	NbForwardElements() int
	GetForwardConnId(idx int) int8
	GetForwardElement(idx int) PathElement
	Copy() PathElement
	SetLastNext(path PathElement)
}

// The int8 here is the forward connection Id
type SimplePathElement struct {
	forwardConnId int8
	next          PathElement
}

func (spe *SimplePathElement) NbForwardElements() int {
	return 1
}

func (spe *SimplePathElement) GetForwardConnId(idx int) int8 {
	return spe.forwardConnId
}

func (spe *SimplePathElement) GetForwardElement(idx int) PathElement {
	return spe.next
}

func (spe *SimplePathElement) Copy() PathElement {
	if spe.next == nil {
		return &SimplePathElement{spe.forwardConnId, nil}
	}
	return &SimplePathElement{spe.forwardConnId, spe.next.Copy()}
}

func (spe *SimplePathElement) SetLastNext(path PathElement) {
	if spe.next == nil {
		spe.next = path
	} else {
		spe.next.SetLastNext(path)
	}
}

// We count only forward fork
type ForkPathElement []PathElement

func (fpe *ForkPathElement) NbForwardElements() int {
	return len(*fpe)
}

func (fpe *ForkPathElement) GetForwardConnId(idx int) int8 {
	return (*fpe)[idx].GetForwardConnId(0)
}

func (fpe *ForkPathElement) GetForwardElement(idx int) PathElement {
	return (*fpe)[idx]
}

func (fpe *ForkPathElement) Copy() PathElement {
	res := ForkPathElement(make([]PathElement, len(*fpe)))
	for i, spe := range *fpe {
		res[i] = spe.Copy()
	}
	return &res
}

func (fpe *ForkPathElement) SetLastNext(path PathElement) {
	for _, spe := range *fpe {
		spe.SetLastNext(path)
	}
}

type ConnectionListBuilder struct {
	length      int
	currentPath PathElement
}

func (eo *EventOutgrowth) GetRootPathElement() PathElement {
	if eo.rootPath == nil {
		eo.rootPath = eo.BuildPath(nil)
	}
	return eo.rootPath
}

func (seo *SavedEventOutgrowth) GetRootPathElement() PathElement {
	return seo.rootPath
}

func (eo *EventOutgrowth) BuildPath(path PathElement) PathElement {
	if eo.IsRoot() {
		return path
	}
	fromConnIds := eo.GetFromConnIds()
	firstPath := eo.fromList[0].BuildPath(&SimplePathElement{-fromConnIds[0], path,})
	if len(eo.fromList) == 1 {
		return firstPath
	}
	for i := 1; i < len(eo.fromList); i++ {
		newPath := eo.fromList[i].BuildPath(&SimplePathElement{-fromConnIds[i], path,})
		firstPath = MergePath(firstPath, newPath)
	}
	return firstPath
}

func MergePath(path1, path2 PathElement) PathElement {
	if path1 == nil && path2 == nil {
		return nil
	}
	if path1 != nil && path2 == nil {
		return path1
	}
	if path1 == nil && path2 != nil {
		return path2
	}
	nb1 := path1.NbForwardElements()
	nb2 := path2.NbForwardElements()
	if nb1 == 1 && nb2 == 1 {
		if path1.GetForwardConnId(0) == path2.GetForwardConnId(0) {
			return &SimplePathElement{path1.GetForwardConnId(0), MergePath(path1.GetForwardElement(0), path2.GetForwardElement(0))}
		}
		fpe := ForkPathElement(make([]PathElement, 2))
		fpe[0] = &SimplePathElement{path1.GetForwardConnId(0), path1.GetForwardElement(0)}
		fpe[2] = &SimplePathElement{path2.GetForwardConnId(0), path2.GetForwardElement(0)}
		return &fpe
	}
	pathsPerConnId := make(map[int8][]PathElement)
	for i := 0; i < nb1; i++ {
		connId := path1.GetForwardConnId(i)
		paths, ok := pathsPerConnId[connId]
		newPath := &SimplePathElement{connId, path1.GetForwardElement(i)}
		if !ok {
			paths = make([]PathElement, 1)
			paths[0] = newPath
		} else {
			paths = append(paths, newPath)
		}
		pathsPerConnId[connId] = paths
	}
	for i := 0; i < nb2; i++ {
		connId := path2.GetForwardConnId(i)
		paths, ok := pathsPerConnId[connId]
		newPath := &SimplePathElement{connId, path2.GetForwardElement(i)}
		if !ok {
			paths = make([]PathElement, 1)
			paths[0] = newPath
		} else {
			paths = append(paths, newPath)
		}
		pathsPerConnId[connId] = paths
	}
	i := 0
	res := ForkPathElement(make([]PathElement, len(pathsPerConnId)))
	for connId, paths := range pathsPerConnId {
		if len(paths) == 1 {
			res[i] = paths[0]
			i++
		} else if len(paths) == 2 {
			res[i] = MergePath(paths[0], paths[1])
			i++
		} else {
			Log.Errorf("Cannot have paths in merge for same connection ids not 1 or 2 for %d %d", connId, len(paths))
		}
	}
	return &res
}

func (seo *SavedEventOutgrowth) BuildPath(path PathElement) PathElement {
	newPath := seo.rootPath.Copy()
	newPath.SetLastNext(path)
	return newPath
}
