package spacedb

func (ne *BaseNodeEvent) GetDistFromCurrent(spaceTime SpaceTimeIfc) DistAndTime {
	return spaceTime.GetCurrentTime() - ne.accessedTime
}

// Return true if path node is currently active
func (ne *BaseNodeEvent) IsActive(spaceTime SpaceTimeIfc) bool {
	space := spaceTime.GetSpace()
	evt := space.GetEvent(ne.evtId)
	if ne.IsRoot(evt) {
		return true
	}
	return ne.GetDistFromCurrent(spaceTime) <= space.GetActivePathNodeThreshold()
}

// Return true if path node is currently and next step active
func (ne *BaseNodeEvent) IsActiveNext(spaceTime SpaceTimeIfc) bool {
	space := spaceTime.GetSpace()
	evt := space.GetEvent(ne.evtId)
	if ne.IsRoot(evt) {
		return false
	}
	return ne.GetDistFromCurrent(spaceTime) < space.GetActivePathNodeThreshold()
}
