package m3server

import (
	"github.com/freddy33/qsm-go/backend/pathdb"
	"github.com/freddy33/qsm-go/backend/pointdb"
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
	pointData := pointdb.GetPointPackData(env)
	pathData := pathdb.GetServerPathPackData(env)
	center := m3point.Origin
	newPathCtx := pathData.CreatePathCtxFromAttributes(
		pointData.GetGrowthContextByTypeAndIndex(m3point.GrowthType(reqMsg.GetGrowthType()), int(reqMsg.GetGrowthIndex())),
		int(reqMsg.GetGrowthOffset()), center)
	pathNodeDb := newPathCtx.GetRootPathNode().(*pathdb.PathNodeDb)
	resMsg := &m3api.PathContextResponseMsg{
		PathCtxId:       int32(newPathCtx.GetId()),
		GrowthContextId: int32(newPathCtx.GetGrowthCtx().GetId()),
		GrowthOffset:    int32(newPathCtx.GetGrowthOffset()),
		RootPathNode:    pathNodeToMsg(pathNodeDb),
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

func moveToNextNode(w http.ResponseWriter, r *http.Request) {
	Log.Infof("Receive moveToNextNode")

	reqMsg := &m3api.NextMoveRequestMsg{}
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

	currentDist := pathCtx.(*pathdb.PathContextDb).GetCurrentDist()
	if int(reqMsg.CurrentDist) != currentDist {
		SendResponse(w, http.StatusBadRequest, "Path context %d current dist is %d not %d", reqMsg.GetPathCtxId(), currentDist, reqMsg.CurrentDist)
		return
	}

	pathCtx.MoveToNextNodes()

	resMsg := &m3api.NextMoveResponseMsg{}
	openPathNodes := pathCtx.GetAllOpenPathNodes()
	Log.Infof("Sending back %d path nodes back on move to next for %d", len(openPathNodes), pathCtx.GetId())
	resMsg.PathCtxId = int32(pathCtx.GetId())
	resMsg.NextDist = int32(pathCtx.(*pathdb.PathContextDb).GetCurrentDist())
	resMsg.NewPathNodes = make([]*m3api.PathNodeMsg, len(openPathNodes))
	for i, pni := range openPathNodes {
		pn := pni.(*pathdb.PathNodeDb)
		resMsg.NewPathNodes[i] = pathNodeToMsg(pn)
	}

	// TODO: Check how to return the modified path nodes

	WriteResponseMsg(w, r, resMsg)
}
