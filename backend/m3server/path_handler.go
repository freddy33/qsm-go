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
	if ReadRequestMsg(w, r, reqMsg) {
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

func getPathNodes(w http.ResponseWriter, r *http.Request) {
	Log.Infof("Receive getPathNodes")

	reqMsg := &m3api.PathNodesRequestMsg{}
	if ReadRequestMsg(w, r, reqMsg) {
		return
	}

	env := GetEnvironment(r)
	pathData := pathdb.GetServerPathPackData(env)

	pathCtx := pathData.GetPathCtx(int(reqMsg.GetPathCtxId()))
	if pathCtx == nil {
		SendResponse(w, http.StatusBadRequest, "path context id %d does not exists", reqMsg.GetPathCtxId())
		return
	}

	dist := int(reqMsg.Dist)
	toDist := int(reqMsg.ToDist)
	var pathNodes []m3path.PathNode
	var err error
	if toDist <= 0 {
		pathNodes, err = pathCtx.GetPathNodesAt(dist)
	} else {
		pathNodes, err = pathCtx.GetPathNodesBetween(dist, toDist)
	}
	if err != nil {
		SendResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	nbPathNodes := len(pathNodes)
	resMsg := &m3api.PathNodesResponseMsg{
		PathCtxId:   int32(pathCtx.GetId()),
		Dist:        int32(dist),
		ToDist:      int32(toDist),
		MaxDist:     int32(pathCtx.GetMaxDist()),
		NbPathNodes: int32(nbPathNodes),
		PathNodes:   make([]*m3api.PathNodeMsg, nbPathNodes),
	}
	for i, pn := range pathNodes {
		resMsg.PathNodes[i] = pathNodeToMsg(pn.(*pathdb.PathNodeDb))
	}

	WriteResponseMsg(w, r, resMsg)
}

func getNbPathNodes(w http.ResponseWriter, r *http.Request) {
	Log.Infof("Receive getNbPathNodes")

	reqMsg := &m3api.PathNodesRequestMsg{}
	if ReadRequestMsg(w, r, reqMsg) {
		return
	}

	env := GetEnvironment(r)
	pathData := pathdb.GetServerPathPackData(env)

	pathCtx := pathData.GetPathCtx(int(reqMsg.GetPathCtxId()))
	if pathCtx == nil {
		SendResponse(w, http.StatusBadRequest, "path context id %d does not exists", reqMsg.GetPathCtxId())
		return
	}

	var nbPathNodes int
	dist := int(reqMsg.Dist)
	toDist := int(reqMsg.ToDist)
	if toDist <= 0 {
		nbPathNodes = pathCtx.GetNumberOfNodesAt(dist)
	} else {
		nbPathNodes = pathCtx.GetNumberOfNodesBetween(dist, toDist)
	}
	if nbPathNodes < 1 {
		SendResponse(w, http.StatusInternalServerError, "Could not retrieve the count of %s between dist %d and %d", pathCtx.String(), dist, toDist)
		return
	}
	resMsg := &m3api.PathNodesResponseMsg{
		PathCtxId:   int32(pathCtx.GetId()),
		Dist:        int32(dist),
		ToDist:      int32(toDist),
		MaxDist:     int32(pathCtx.GetMaxDist()),
		NbPathNodes: int32(nbPathNodes),
	}

	WriteResponseMsg(w, r, resMsg)
}
