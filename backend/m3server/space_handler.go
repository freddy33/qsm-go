package m3server

import (
	"github.com/freddy33/qsm-go/backend/pathdb"
	"github.com/freddy33/qsm-go/backend/spacedb"
	"github.com/freddy33/qsm-go/model/m3api"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/freddy33/qsm-go/model/m3space"
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
		SendResponse(w, http.StatusNotFound, "SpaceTime id %d does not exists", reqMsg.SpaceId)
		return
	}
	event, err := space.CreateEvent(m3point.GrowthType(reqMsg.GrowthType), int(reqMsg.GrowthIndex), int(reqMsg.GrowthOffset),
		m3space.DistAndTime(reqMsg.CreationTime), m3api.PointMsgToPoint(reqMsg.Center), m3space.EventColor(reqMsg.Color))
	if err != nil {
		SendResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	eventNode := event.GetCenterNode()
	eventId := int32(event.GetId())
	point, err := eventNode.GetPoint()
	if err != nil {
		SendResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	pathNode, err := eventNode.GetPathNode()
	if err != nil {
		SendResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	pathNodeDb := pathNode.(*pathdb.PathNodeDb)
	pathContext := event.GetPathContext()
	resMsg := &m3api.EventMsg{
		EventId:      eventId,
		SpaceId:      int32(space.GetId()),
		GrowthType:   int32(pathContext.GetGrowthType()),
		GrowthIndex:  int32(pathContext.GetGrowthIndex()),
		GrowthOffset: int32(pathContext.GetGrowthOffset()),
		CreationTime: int32(event.GetCreationTime()),
		PathCtxId:    int32(pathContext.GetId()),
		Color:        uint32(event.GetColor()),
		RootNode: &m3api.NodeEventMsg{
			EventNodeId:    eventNode.GetId(),
			EventId:        eventId,
			Point:          m3api.PointToPointMsg(*point),
			CreationTime:   int32(eventNode.GetCreationTime()),
			D:              int32(eventNode.GetD()),
			TrioId:         int32(pathNode.GetTrioIndex()),
			ConnectionMask: uint32(pathNodeDb.GetConnectionMask()),
		},
	}
	WriteResponseMsg(w, r, resMsg)
}
