package m3server

import (
	"github.com/freddy33/qsm-go/backend/spacedb"
	"github.com/freddy33/qsm-go/model/m3api"
	"github.com/freddy33/qsm-go/model/m3space"
	"net/http"
)

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

	resMsg := &m3api.SpaceMsg{
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
	WriteResponseMsg(w, r, resMsg)
}
