package m3space

import "github.com/freddy33/qsm-go/model/m3path"

type NodeEventList struct {
	cur  *BaseNodeEvent
	next *NodeEventList
}

type BaseNodeEvent struct {
	evtId        EventID
	pathNodeId   int64
	accessedTime DistAndTime
	pathNode     m3path.PathNode
}

func (ne *BaseNodeEvent) GetEventId() EventID {
	return ne.evtId
}

func (ne *BaseNodeEvent) GetPathNodeId() int64 {
	return ne.pathNodeId
}

func (ne *BaseNodeEvent) GetPathNode() m3path.PathNode {
	pn := ne.pathNode
	// TODO: Should be a method with bool m3path.InPoolId {
	if pn != nil && pn.GetId() == int64(-2) {
		// nilify for now
		ne.pathNode = nil
		return nil
	}
	return pn
}

func (ne *BaseNodeEvent) IsRoot(evt *Event) bool {
	return ne.evtId == evt.id && ne.accessedTime == evt.created
}

func (ne *BaseNodeEvent) GetAccessedTime() DistAndTime {
	return ne.accessedTime
}

func (ne *BaseNodeEvent) GetDistFromCurrent(space *Space) DistAndTime {
	return space.CurrentTime - ne.accessedTime
}

func (ne *BaseNodeEvent) IsLatest() bool {
	pn := ne.GetPathNode()
	if pn != nil {
		return pn.IsLatest()
	}
	return false
}

// Return true if path node is currently active
func (ne *BaseNodeEvent) IsActive(space *Space) bool {
	evt := space.GetEvent(ne.evtId)
	if ne.IsRoot(evt) {
		return true
	}
	return ne.GetDistFromCurrent(space) <= space.EventOutgrowthThreshold
}

// Return true if path node is currently and next step active
func (ne *BaseNodeEvent) IsActiveNext(space *Space) bool {
	evt := space.GetEvent(ne.evtId)
	if ne.IsRoot(evt) {
		return false
	}
	return ne.GetDistFromCurrent(space) < space.EventOutgrowthThreshold
}

// Return true if path node is old. Dead node are also old
func (ne *BaseNodeEvent) IsOld(space *Space) bool {
	return ne.GetDistFromCurrent(space) >= space.EventOutgrowthOldThreshold
}

// Return true if path node is dead
func (ne *BaseNodeEvent) IsDead(space *Space) bool {
	return ne.GetDistFromCurrent(space) >= space.EventOutgrowthDeadThreshold
}
