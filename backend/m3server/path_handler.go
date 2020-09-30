package m3server

import (
	"github.com/freddy33/qsm-go/backend/pathdb"
	"github.com/freddy33/qsm-go/model/m3api"
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
	newPathCtx, err := pathData.GetPathCtxDb(m3point.GrowthType(reqMsg.GetGrowthType()),
		int(reqMsg.GetGrowthIndex()), int(reqMsg.GetGrowthOffset()))
	if err != nil {
		SendResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	pathNodeDb := newPathCtx.GetRootPathNode().(*pathdb.PathNodeDb)
	resMsg := &m3api.PathContextResponseMsg{
		PathCtxId:       int32(newPathCtx.GetId()),
		GrowthContextId: int32(newPathCtx.GetGrowthCtx().GetId()),
		GrowthOffset:    int32(newPathCtx.GetGrowthOffset()),
		RootPathNode:    pathNodeToMsg(pathNodeDb),
		MaxDist:         int32(0),
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
	pathNodes, err := pathCtx.GetPathNodesAt(dist)
	if err != nil {
		SendResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	resMsg := &m3api.PathNodesResponseMsg{
		PathCtxId: int32(pathCtx.GetId()),
		Dist:      int32(dist),
		PathNodes: make([]*m3api.PathNodeMsg, len(pathNodes)),
	}
	for i, pn := range pathNodes {
		resMsg.PathNodes[i] = pathNodeToMsg(pn.(*pathdb.PathNodeDb))
	}

	WriteResponseMsg(w, r, resMsg)
}
