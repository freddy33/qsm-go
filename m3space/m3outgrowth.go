package m3space

import (
	"fmt"
)

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
