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
	if ReadRequestMsg(w, r, reqMsg) {
		return
	}

	env := GetEnvironment(r)

	space, err := spacedb.CreateSpace(env, reqMsg.SpaceName, m3space.DistAndTime(reqMsg.ActivePathNodeThreshold),
		int(reqMsg.MaxTriosPerPoint), int(reqMsg.MaxPathNodesPerPoint))
	if err != nil {
		SendResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteResponseMsg(w, r, spaceDbToMsg(space))
}

func spaceDbToMsg(space *spacedb.SpaceDb) *m3api.SpaceMsg {
	return &m3api.SpaceMsg{
		SpaceId:                 int32(space.GetId()),
		SpaceName:               space.GetName(),
		ActivePathNodeThreshold: int32(space.GetActivePathNodeThreshold()),
		MaxTriosPerPoint:        int32(space.GetMaxTriosPerPoint()),
		MaxPathNodesPerPoint:    int32(space.GetMaxPathNodesPerPoint()),
		MaxTime:                 int32(space.GetMaxTime()),
		CurrentTime:             int32(space.GetCurrentTime()),
		MaxCoord:                int32(space.GetMaxCoord()),
		EventIds:                space.GetEventIdsForMsg(),
		NbActiveNodes:           int32(space.GetNbActiveNodes()),
	}
}

func createEvent(w http.ResponseWriter, r *http.Request) {
	Log.Infof("Receive createEvent")

	reqMsg := &m3api.EventMsg{}
	if ReadRequestMsg(w, r, reqMsg) {
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
	resMsg := &m3api.EventResponseMsg{
		EventId:   eventId,
		PathCtxId: int32(event.GetPathContext().GetId()),
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
