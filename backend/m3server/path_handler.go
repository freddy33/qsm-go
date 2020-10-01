package m3server

import (
	"github.com/freddy33/qsm-go/backend/pathdb"
	"github.com/freddy33/qsm-go/model/m3api"
	"github.com/freddy33/qsm-go/model/m3path"
	"github.com/freddy33/qsm-go/model/m3point"
	"net/http"
)

func createPathContext(w http.ResponseWriter, r *http.Request) {
	Log.Infof("Receive createPathContext")

	reqMsg := &m3api.PathContextRequestMsg{}
	if !ReadRequestMsg(w, r, reqMsg) {
		return
	}

	env := GetEnvironment(r)
	pathData := pathdb.GetServerPathPackData(env)
	pathCtx, err := pathData.GetPathCtxDb(
		m3point.GrowthType(reqMsg.GetGrowthType()),
		int(reqMsg.GetGrowthIndex()),
		int(reqMsg.GetGrowthOffset()))
	if err != nil {
		SendResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	pathNodeDb := pathCtx.GetRootPathNode().(*pathdb.PathNodeDb)
	resMsg := &m3api.PathContextResponseMsg{
		PathCtxId:       int32(pathCtx.GetId()),
		GrowthContextId: int32(pathCtx.GetGrowthCtx().GetId()),
		GrowthOffset:    int32(pathCtx.GetGrowthOffset()),
		RootPathNode:    pathNodeToMsg(pathNodeDb),
		MaxDist:         int32(pathCtx.GetMaxDist()),
	}
	WriteResponseMsg(w, r, resMsg)
}

func increaseMaxDist(w http.ResponseWriter, r *http.Request) {
	Log.Infof("Receive increaseMaxDist")

	reqMsg, pathCtx := extractPathNodeRequest(w, r)
	if reqMsg == nil {
		return
	}

	reqDist := int(reqMsg.Dist)
	initialMaxDist := pathCtx.GetMaxDist()
	if reqDist <= initialMaxDist {
		SendResponse(w, http.StatusAccepted, "Path context %d already has max dist %d above the requested dist %d", pathCtx.GetId(), initialMaxDist, reqDist)
		return
	}

	// Below distance 25 all allowed, above only increases of 3 are allowed
	if reqDist > 25 && reqDist-initialMaxDist > 3 {
		SendResponse(w, http.StatusRequestEntityTooLarge, "Path context %d has max dist %d which is too far away from the requested dist %d.\n"+
			"Please request smaller increment in max distance.", pathCtx.GetId(), initialMaxDist, reqDist)
		return
	}
	err := pathCtx.RequestNewMaxDist(reqDist)
	if err != nil {
		SendResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	finalMaxDist := pathCtx.GetMaxDist()
	nbPathNodes := pathCtx.GetNumberOfNodesBetween(initialMaxDist+1, finalMaxDist)
	resMsg := createPathNodesResponse(pathCtx, initialMaxDist, finalMaxDist, nbPathNodes)

	WriteResponseMsg(w, r, resMsg)
}

func getPathNodes(w http.ResponseWriter, r *http.Request) {
	Log.Infof("Receive getPathNodes")

	reqMsg, pathCtx := extractPathNodeRequest(w, r)
	if reqMsg == nil {
		return
	}

	valid, fromDist, toDist := verifyMaxDist(w, r, pathCtx, reqMsg)
	if !valid {
		return
	}
	var pathNodes []m3path.PathNode
	var err error
	if toDist <= 0 {
		pathNodes, err = pathCtx.GetPathNodesAt(fromDist)
	} else {
		pathNodes, err = pathCtx.GetPathNodesBetween(fromDist, toDist)
	}
	if err != nil {
		SendResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	nbPathNodes := len(pathNodes)
	resMsg := createPathNodesResponse(pathCtx, fromDist, toDist, nbPathNodes)

	resMsg.PathNodes = make([]*m3api.PathNodeMsg, nbPathNodes)
	for i, pn := range pathNodes {
		resMsg.PathNodes[i] = pathNodeToMsg(pn.(*pathdb.PathNodeDb))
	}

	WriteResponseMsg(w, r, resMsg)
}

func getNbPathNodes(w http.ResponseWriter, r *http.Request) {
	Log.Infof("Receive getNbPathNodes")

	reqMsg, pathCtx := extractPathNodeRequest(w, r)
	if reqMsg == nil {
		return
	}

	valid, fromDist, toDist := verifyMaxDist(w, r, pathCtx, reqMsg)
	if !valid {
		return
	}

	var nbPathNodes int
	if toDist <= 0 {
		nbPathNodes = pathCtx.GetNumberOfNodesAt(fromDist)
	} else {
		nbPathNodes = pathCtx.GetNumberOfNodesBetween(fromDist, toDist)
	}
	if nbPathNodes < 1 {
		SendResponse(w, http.StatusInternalServerError, "Could not retrieve the count of %s between dist %d and %d", pathCtx.String(), fromDist, toDist)
		return
	}
	resMsg := createPathNodesResponse(pathCtx, fromDist, toDist, nbPathNodes)

	WriteResponseMsg(w, r, resMsg)
}

func pathNodeToMsg(pathNodeDb *pathdb.PathNodeDb) *m3api.PathNodeMsg {
	return &m3api.PathNodeMsg{
		PathNodeId:        pathNodeDb.GetId(),
		Point:             m3api.PointToPointMsg(pathNodeDb.P()),
		D:                 int32(pathNodeDb.D()),
		TrioId:            int32(pathNodeDb.GetTrioIndex()),
		ConnectionMask:    uint32(pathNodeDb.GetConnectionMask()),
		LinkedPathNodeIds: pathNodeDb.GetConnsDataForMsg(),
	}
}

func verifyMaxDist(w http.ResponseWriter, r *http.Request, pathCtx m3path.PathContext, reqMsg *m3api.PathNodesRequestMsg) (bool, int, int) {
	fromDist := int(reqMsg.Dist)
	toDist := int(reqMsg.ToDist)
	if toDist <= 0 {
		if fromDist > pathCtx.GetMaxDist() {
			SendResponse(w, http.StatusUnprocessableEntity, "path context id %d has a max dist %d which is below the requested %d",
				reqMsg.GetPathCtxId(), pathCtx.GetMaxDist(), fromDist)
			return false, -1, -1
		}
	} else {
		if toDist < fromDist {
			SendResponse(w, http.StatusBadRequest, "request to_dist value %d not zero but below the starting requested dist %d for request on path context id %d",
				toDist, fromDist, reqMsg.GetPathCtxId())
			return false, -1, -1
		}
		if toDist > pathCtx.GetMaxDist() {
			SendResponse(w, http.StatusUnprocessableEntity, "path context id %d has a max dist %d which is below the requested %d",
				reqMsg.GetPathCtxId(), pathCtx.GetMaxDist(), toDist)
			return false, -1, -1
		}
	}
	return true, fromDist, toDist
}

func extractPathNodeRequest(w http.ResponseWriter, r *http.Request) (*m3api.PathNodesRequestMsg, m3path.PathContext) {
	reqMsg := &m3api.PathNodesRequestMsg{}
	if !ReadRequestMsg(w, r, reqMsg) {
		return nil, nil
	}

	env := GetEnvironment(r)
	pathData := pathdb.GetServerPathPackData(env)

	pathCtx := pathData.GetPathCtx(int(reqMsg.GetPathCtxId()))
	if pathCtx == nil {
		SendResponse(w, http.StatusBadRequest, "path context id %d does not exists", reqMsg.GetPathCtxId())
		return nil, nil
	}
	return reqMsg, pathCtx
}

func createPathNodesResponse(pathCtx m3path.PathContext, fromDist int, toDist int, nbPathNodes int) *m3api.PathNodesResponseMsg {
	resMsg := &m3api.PathNodesResponseMsg{
		PathCtxId:   int32(pathCtx.GetId()),
		Dist:        int32(fromDist),
		ToDist:      int32(toDist),
		MaxDist:     int32(pathCtx.GetMaxDist()),
		NbPathNodes: int32(nbPathNodes),
	}
	return resMsg
}
