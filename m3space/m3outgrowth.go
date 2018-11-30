package m3space

import "fmt"

type NewPossibleOutgrowth struct {
	pos      Point
	event    *Event
	from     *EventOutgrowth
	distance Distance
	state    EventOutgrowthState
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
	return Distance(space.currentTime - evt.created) - eo.distance
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
	return Distance(space.currentTime - evt.created) - seo.distance
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
