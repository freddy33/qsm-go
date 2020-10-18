package m3server

import (
	"github.com/freddy33/qsm-go/backend/pathdb"
	"github.com/freddy33/qsm-go/backend/spacedb"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3api"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/freddy33/qsm-go/model/m3space"
	"math"
	"net/http"
)

func getSpaces(w http.ResponseWriter, r *http.Request) {
	Log.Infof("Receive getSpaces")
	env := GetEnvironment(r)
	spd := spacedb.GetServerSpacePackData(env)
	err := spd.LoadAllSpaces()
	if err != nil {
		SendResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	allSpaces := spd.GetAllSpaces()
	resMsg := &m3api.SpaceListMsg{
		Spaces: make([]*m3api.SpaceMsg, len(allSpaces)),
	}
	for i, space := range allSpaces {
		resMsg.Spaces[i] = spaceDbToMsg(space.(*spacedb.SpaceDb))
	}
	WriteResponseMsg(w, r, resMsg)
}

func createSpace(w http.ResponseWriter, r *http.Request) {
	Log.Infof("Receive createSpace")

	reqMsg := &m3api.SpaceMsg{}
	if !ReadRequestMsg(w, r, reqMsg) {
		return
	}

	env := GetEnvironment(r)

	space, err := spacedb.CreateSpace(env, reqMsg.SpaceName, m3space.DistAndTime(reqMsg.ActiveThreshold),
		int(reqMsg.MaxTriosPerPoint), int(reqMsg.MaxNodesPerPoint))
	if err != nil {
		SendResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteResponseMsg(w, r, spaceDbToMsg(space))
}

func deleteSpace(w http.ResponseWriter, r *http.Request) {
	Log.Infof("Receive deleteSpace")

	reqMsg := &m3api.SpaceMsg{}
	if !ReadRequestMsg(w, r, reqMsg) {
		return
	}

	env := GetEnvironment(r)
	spaceData := spacedb.GetServerSpacePackData(env)

	spaceId := int(reqMsg.SpaceId)
	spaceName := reqMsg.SpaceName
	nbDeleted, err := spaceData.DeleteSpace(spaceId, spaceName)
	if err != nil {
		SendResponse(w, http.StatusInternalServerError, "Trying to delete %d %q got: %s", spaceId, spaceName, err.Error())
		return
	}

	SendResponse(w, http.StatusOK, "Space ID %d %q deleted %d elements", spaceId, spaceName, nbDeleted)
}

func spaceDbToMsg(space *spacedb.SpaceDb) *m3api.SpaceMsg {
	return &m3api.SpaceMsg{
		SpaceId:          int32(space.GetId()),
		SpaceName:        space.GetName(),
		ActiveThreshold:  int32(space.GetActiveThreshold()),
		MaxTriosPerPoint: int32(space.GetMaxTriosPerPoint()),
		MaxNodesPerPoint: int32(space.GetMaxNodesPerPoint()),
		MaxTime:          int32(space.GetMaxTime()),
		MaxCoord:         int32(space.GetMaxCoord()),
		EventIds:         space.GetEventIdsForMsg(),
	}
}

func getEvents(w http.ResponseWriter, r *http.Request) {
	Log.Infof("Receive getEvents")

	reqMsg := &m3api.FindEventsMsg{}
	if !ReadRequestMsg(w, r, reqMsg) {
		return
	}

	env := GetEnvironment(r)
	spaceData := spacedb.GetServerSpacePackData(env)
	space := spaceData.GetSpace(int(reqMsg.SpaceId)).(*spacedb.SpaceDb)
	if space == nil {
		SendResponse(w, http.StatusNotFound, "Space id %d does not exists", reqMsg.SpaceId)
		return
	}

	var events []m3space.EventIfc
	if reqMsg.EventId > 0 {
		Log.Debugf("Using event id %d to find event in space %s", reqMsg.EventId, space.String())
		events = make([]m3space.EventIfc, 1)
		events[0] = space.GetEvent(m3space.EventId(reqMsg.EventId))
	} else {
		Log.Debugf("Using time %d to find events in space %s", reqMsg.AtTime, space.String())
		if reqMsg.AtTime < 0 {
			// Get all events
			events = space.GetAllEvents()
		} else {
			events = space.GetActiveEventsAt(m3space.DistAndTime(reqMsg.AtTime))
		}
	}

	var err error
	resMsg := new(m3api.EventListMsg)
	resMsg.Events = make([]*m3api.EventMsg, len(events))
	for i, evt := range events {
		resMsg.Events[i], err = createEventMsg(evt)
		if err != nil {
			SendResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	WriteResponseMsg(w, r, resMsg)
}

func createEvent(w http.ResponseWriter, r *http.Request) {
	Log.Infof("Receive createEvent")

	reqMsg := &m3api.CreateEventRequestMsg{}
	if !ReadRequestMsg(w, r, reqMsg) {
		return
	}

	env := GetEnvironment(r)
	spaceData := spacedb.GetServerSpacePackData(env)
	space := spaceData.GetSpace(int(reqMsg.SpaceId))
	if space == nil {
		SendResponse(w, http.StatusNotFound, "Space id %d does not exists", reqMsg.SpaceId)
		return
	}
	event, err := space.CreateEvent(m3point.GrowthType(reqMsg.GrowthType), int(reqMsg.GrowthIndex), int(reqMsg.GrowthOffset),
		m3space.DistAndTime(reqMsg.CreationTime), m3api.PointMsgToPoint(reqMsg.Center), m3space.EventColor(reqMsg.Color))
	if err != nil {
		SendResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	resMsg, err := createEventMsg(event)
	if err != nil {
		SendResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteResponseMsg(w, r, resMsg)
}

func createNodeEventMsg(node m3space.NodeEventIfc) (*m3api.NodeEventMsg, error) {
	point, err := node.GetPoint()
	if err != nil {
		return nil, err
	}

	pathNode, err := node.GetPathNode()
	if err != nil {
		return nil, err
	}

	pathNodeDb := pathNode.(*pathdb.PathNodeDb)

	return &m3api.NodeEventMsg{
		EventNodeId:    int64(node.GetId()),
		EventId:        int32(node.GetEventId()),
		Point:          m3api.PointToPointMsg(*point),
		CreationTime:   int32(node.GetCreationTime()),
		D:              int32(node.GetD()),
		TrioId:         int32(pathNode.GetTrioIndex()),
		ConnectionMask: uint32(pathNodeDb.GetConnectionMask()),
	}, nil
}

func createEventMsg(event m3space.EventIfc) (*m3api.EventMsg, error) {
	space := event.GetSpace()
	rootNode, err := createNodeEventMsg(event.GetCenterNode())
	if err != nil {
		return nil, err
	}
	pathContext := event.GetPathContext()
	resMsg := &m3api.EventMsg{
		EventId:      int32(event.GetId()),
		SpaceId:      int32(space.GetId()),
		GrowthType:   int32(pathContext.GetGrowthType()),
		GrowthIndex:  int32(pathContext.GetGrowthIndex()),
		GrowthOffset: int32(pathContext.GetGrowthOffset()),
		CreationTime: int32(event.GetCreationTime()),
		PathCtxId:    int32(pathContext.GetId()),
		Color:        uint32(event.GetColor()),
		RootNode:     rootNode,
		MaxNodeTime:  int32(event.(*spacedb.EventDb).GetMaxNodeTime()),
	}
	return resMsg, nil
}

func getNodeEvents(w http.ResponseWriter, r *http.Request) {
	Log.Infof("Receive getNodeEvents")

	reqMsg := &m3api.FindNodeEventsMsg{}
	if !ReadRequestMsg(w, r, reqMsg) {
		return
	}

	env := GetEnvironment(r)
	spaceData := spacedb.GetServerSpacePackData(env)
	space := spaceData.GetSpace(int(reqMsg.SpaceId)).(*spacedb.SpaceDb)
	if space == nil {
		SendResponse(w, http.StatusNotFound, "Space id %d does not exists", reqMsg.SpaceId)
		return
	}
	event := space.GetEvent(m3space.EventId(reqMsg.EventId))
	if event == nil {
		SendResponse(w, http.StatusNotFound, "Space id %d does not have events %d", reqMsg.SpaceId, reqMsg.EventId)
		return
	}

	nodes, err := event.GetActiveNodesAt(m3space.DistAndTime(reqMsg.AtTime))
	if err != nil {
		SendResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	resMsg := new(m3api.NodeEventListMsg)
	resMsg.Nodes = make([]*m3api.NodeEventMsg, len(nodes))
	for i, node := range nodes {
		resMsg.Nodes[i], err = createNodeEventMsg(node)
		if err != nil {
			SendResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	WriteResponseMsg(w, r, resMsg)
}

func getSpaceTime(w http.ResponseWriter, r *http.Request) {
	Log.Infof("Receive getSpaceTime")

	reqMsg := &m3api.SpaceTimeRequestMsg{}
	if !ReadRequestMsg(w, r, reqMsg) {
		return
	}

	env := GetEnvironment(r)
	spaceData := spacedb.GetServerSpacePackData(env)
	space := spaceData.GetSpace(int(reqMsg.SpaceId)).(*spacedb.SpaceDb)

	spaceTime := space.GetSpaceTimeAt(m3space.DistAndTime(reqMsg.CurrentTime)).(*spacedb.SpaceTime)
	err := spaceTime.Populate()
	if err != nil {
		SendResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	activeEvents := spaceTime.GetActiveEvents()
	nbActiveNodes := spaceTime.GetNbActiveNodes()
	resMsg := &m3api.SpaceTimeResponseMsg{
		SpaceId:       int32(space.GetId()),
		CurrentTime:   int32(spaceTime.GetCurrentTime()),
		ActiveEvents:  make([]*m3api.EventMsg, len(activeEvents)),
		NbActiveNodes: int32(nbActiveNodes),
		FilteredNodes: nil,
	}
	for i, evt := range activeEvents {
		resMsg.ActiveEvents[i], err = createEventMsg(evt)
		if err != nil {
			SendResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	nodesMsgBuilder, err := makeNodeMsgBuilder(reqMsg, spaceTime)
	if err != nil {
		// Nothing found in this space time may be due to too strong filter
		SendResponse(w, http.StatusNotFound, err.Error())
		return
	}
	spaceTime.VisitNodes(nodesMsgBuilder)
	if nodesMsgBuilder.buildError != nil {
		SendResponse(w, http.StatusInternalServerError, nodesMsgBuilder.buildError.Error())
		return
	}
	resMsg.FilteredNodes = nodesMsgBuilder.foundNodes

	WriteResponseMsg(w, r, resMsg)
}

type NodeToMsgBuilder struct {
	buildError        error
	minNbEventsFilter int
	colorMaskFilter   uint8
	isFiltering       bool
	foundNodes        []*m3api.SpaceTimeNodeMsg
}

func makeNodeMsgBuilder(reqMsg *m3api.SpaceTimeRequestMsg, spaceTime *spacedb.SpaceTime) (*NodeToMsgBuilder, error) {
	n := &NodeToMsgBuilder{
		minNbEventsFilter: int(reqMsg.MinNbEventsFilter),
		colorMaskFilter:   uint8(reqMsg.ColorMaskFilter),
		isFiltering:       false,
	}

	activeEvents := spaceTime.GetActiveEvents()
	nbActiveNodes := spaceTime.GetNbActiveNodes()

	// Try to find the ratio from total active nodes to actual sent back based on filter
	activeNodesRatio := 1.0

	// First check on events present and demanded minimum on a node
	nbActiveEvents := len(activeEvents)
	if nbActiveEvents == 0 || nbActiveEvents < n.minNbEventsFilter {
		return nil, m3util.MakeQsmErrorf("At space time %s there not enough active events %d < %d",
			spaceTime.String(), nbActiveEvents, n.minNbEventsFilter)
	} else {
		if n.minNbEventsFilter > 1 {
			// ratio of 5% per min number above 1
			activeNodesRatio = math.Pow(0.05, float64(n.minNbEventsFilter-1))
			n.isFiltering = true
		}
	}

	// If no filter applied above do the same with color
	if !n.isFiltering {
		// TODO: Create an actual mask of the active events, may be there is no filter here
		maskCount := m3util.CountTheOnes(n.colorMaskFilter)
		if int(maskCount) < len(m3space.AllColors) {
			// ratio of 5% per min number above 1
			activeNodesRatio = math.Pow(0.05, float64(len(m3space.AllColors)-int(maskCount)))
			n.isFiltering = true
		}
	}

	n.foundNodes = make([]*m3api.SpaceTimeNodeMsg, 0, int(activeNodesRatio*float64(nbActiveNodes)))

	return n, nil
}

func (n *NodeToMsgBuilder) VisitNode(node m3space.SpaceTimeNodeIfc) {
	if n.buildError != nil {
		// Stop on error
		return
	}
	if !n.isFiltering ||
		((n.minNbEventsFilter <= len(node.GetEventIds()) || node.HasRoot()) &&
			(n.colorMaskFilter&node.GetColorMask() != uint8(0))) {
		spaceTimeNodeMsg, err := createSpaceTimeNodeMsg(node)
		if err != nil {
			n.buildError = err
			return
		}
		n.foundNodes = append(n.foundNodes, spaceTimeNodeMsg)
	}
}

func createSpaceTimeNodeMsg(node m3space.SpaceTimeNodeIfc) (*m3api.SpaceTimeNodeMsg, error) {
	point, err := node.GetPoint()
	if err != nil {
		return nil, err
	}

	nodeDb := node.(*spacedb.SpaceTimeNode)
	// TODO: This should be done using the link list of directly NodeEventDb object instead of creating an array of interfaces
	evtNodes := nodeDb.GetEventNodes()
	res := &m3api.SpaceTimeNodeMsg{
		PointId:   int64(node.GetPointId()),
		Point:     m3api.PointToPointMsg(*point),
		Nodes:     make([]*m3api.SpaceTimeNodeEventMsg, len(evtNodes)),
		HasRoot:   node.HasRoot(),
		ColorMask: uint32(node.GetColorMask()),
	}
	for i, evtNode := range evtNodes {
		res.Nodes[i] = &m3api.SpaceTimeNodeEventMsg{
			EventId:        int32(evtNode.GetEventId()),
			CreationTime:   int32(evtNode.GetCreationTime()),
			D:              int32(evtNode.GetD()),
			TrioId:         int32(evtNode.GetTrioIndex()),
			ConnectionMask: uint32(evtNode.(*spacedb.NodeEventDb).GetConnectionMask()),
		}
	}
	return res, nil
}
