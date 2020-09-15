package m3server

import (
	"fmt"
	"github.com/freddy33/qsm-go/backend/pathdb"
	"github.com/freddy33/qsm-go/backend/pointdb"
	"github.com/freddy33/qsm-go/model/m3api"
	"github.com/golang/protobuf/proto"
	"io/ioutil"
	"net/http"
)

func createPathContext(w http.ResponseWriter, r *http.Request) {
	Log.Infof("Receive createPathContext")

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		Log.Errorf("receive wrong message in createPathContext: %s", err.Error())
		SendResponse(w, http.StatusBadRequest, "req body could not be read")
		return
	}
	pMsg := &m3api.PathContextMsg{}
	err = proto.Unmarshal(b, pMsg)
	if err != nil {
		Log.Errorf("Could not parse body in createPathContext: %s", err.Error())
		SendResponse(w, http.StatusBadRequest, "req body could not be parsed")
		return
	}
	env := GetEnvironment(r)
	pointData := pointdb.GetPointPackData(env)
	pathData := pathdb.GetServerPathPackData(env)
	newPathCtx := pathData.CreatePathCtxFromAttributes(
		pointData.GetGrowthContextById(int(pMsg.GetGrowthContextId())),
		int(pMsg.GetGrowthOffset()),
		m3api.PointMsgToPoint(pMsg.Center))

	resMsg := m3api.PathContextMsg{
		PathCtxId:       int32(newPathCtx.GetId()),
		GrowthContextId: int32(newPathCtx.GetGrowthCtx().GetId()),
		GrowthOffset:    int32(newPathCtx.GetGrowthOffset()),
		// TODO: Uncomment the one line below once init node done
		//Center: m3api.PointToPointMsg(newPathCtx.GetRootPathNode().P()),
		Center: pMsg.Center,
	}

	data, err := proto.Marshal(&resMsg)
	if err != nil {
		Log.Warnf("Failed to marshal PathContextMsg due to: %q", err.Error())
		w.WriteHeader(500)
		_, err = fmt.Fprintf(w, "Failed to marshal PathContextMsg due to:\n%s\n", err.Error())
		if err != nil {
			Log.Errorf("failed to send error message to response due to %q", err.Error())
		}
	}
	_, err = w.Write(data)
	if err != nil {
		Log.Errorf("failed to send data to response due to %q", err.Error())
	}
}

func initRootNode(w http.ResponseWriter, r *http.Request) {
	Log.Infof("Receive initRootNode")

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		Log.Errorf("receive wrong message in initRootNode: %s", err.Error())
		SendResponse(w, http.StatusBadRequest, "req body could not be read")
		return
	}
	pMsg := &m3api.PathContextMsg{}
	err = proto.Unmarshal(b, pMsg)
	if err != nil {
		Log.Errorf("Could not parse body in initRootNode: %s", err.Error())
		SendResponse(w, http.StatusBadRequest, "req body could not be parsed")
		return
	}
	env := GetEnvironment(r)
	pathData := pathdb.GetServerPathPackData(env)

	pathCtx := pathData.GetPathCtx(int(pMsg.GetPathCtxId()))
	if pathCtx == nil {
		Log.Errorf("Could not find path context with ID: %d", pMsg.GetPathCtxId())
		SendResponse(w, http.StatusBadRequest, "path context id %d does not exists", pMsg.GetPathCtxId())
		return
	}

	pathCtx.InitRootNode(m3api.PointMsgToPoint(pMsg.GetCenter()))

	pn := pathCtx.GetRootPathNode().(*pathdb.PathNodeDb)
	resMsg := m3api.PathNodeMsg{
		PathNodeId:        pn.GetId(),
		PathCtxId:         int32(pn.GetPathContext().GetId()),
		Point:             m3api.PointToPointMsg(pn.P()),
		D:                 int64(pn.D()),
		TrioId:            int32(pn.GetTrioIndex()),
		ConnectionMask:    uint32(pn.GetConnectionMask()),
		LinkedPathNodeIds: pn.GetConnsDataForMsg(),
	}
	data, err := proto.Marshal(&resMsg)
	if err != nil {
		Log.Warnf("Failed to marshal PathNodeMsg due to: %q", err.Error())
		w.WriteHeader(500)
		_, err = fmt.Fprintf(w, "Failed to marshal PathNodeMsg due to:\n%s\n", err.Error())
		if err != nil {
			Log.Errorf("failed to send error message to response due to %q", err.Error())
		}
	}
	_, err = w.Write(data)
	if err != nil {
		Log.Errorf("failed to send data to response due to %q", err.Error())
	}
}

func moveToNextNode(w http.ResponseWriter, r *http.Request) {
	Log.Infof("Receive moveToNextNode")

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		Log.Errorf("receive wrong message in moveToNextNode: %s", err.Error())
		SendResponse(w, http.StatusBadRequest, "req body could not be read")
		return
	}
	pMsg := &m3api.PathContextMsg{}
	err = proto.Unmarshal(b, pMsg)
	if err != nil {
		Log.Errorf("Could not parse body in moveToNextNode: %s", err.Error())
		SendResponse(w, http.StatusBadRequest, "req body could not be parsed")
		return
	}
	env := GetEnvironment(r)
	pathData := pathdb.GetServerPathPackData(env)

	pathCtx := pathData.GetPathCtx(int(pMsg.GetPathCtxId()))
	if pathCtx == nil {
		Log.Errorf("Could not find path context with ID: %d", pMsg.GetPathCtxId())
		SendResponse(w, http.StatusBadRequest, "path context id %d does not exists", pMsg.GetPathCtxId())
		return
	}

	pathCtx.MoveToNextNodes()

	resMsg := m3api.NextMoveRespMsg{}
	openPathNodes := pathCtx.GetAllOpenPathNodes()
	Log.Infof("Sending back %d path nodes back on move to next", len(openPathNodes))
	resMsg.PathNodes = make([]*m3api.PathNodeMsg, len(openPathNodes))
	for i, pni := range openPathNodes {
		pn := pni.(*pathdb.PathNodeDb)
		resMsg.PathNodes[i] = &m3api.PathNodeMsg{
			PathNodeId:        pn.GetId(),
			PathCtxId:         int32(pn.GetPathContext().GetId()),
			Point:             m3api.PointToPointMsg(pn.P()),
			D:                 int64(pn.D()),
			TrioId:            int32(pn.GetTrioIndex()),
			ConnectionMask:    uint32(pn.GetConnectionMask()),
			LinkedPathNodeIds: pn.GetConnsDataForMsg(),
		}
	}

	data, err := proto.Marshal(&resMsg)
	if err != nil {
		Log.Warnf("Failed to marshal NextMoveRespMsg due to: %q", err.Error())
		w.WriteHeader(500)
		_, err = fmt.Fprintf(w, "Failed to marshal NextMoveRespMsg due to:\n%s\n", err.Error())
		if err != nil {
			Log.Errorf("failed to send error message to response due to %q", err.Error())
		}
	}
	_, err = w.Write(data)
	if err != nil {
		Log.Errorf("failed to send data to response due to %q", err.Error())
	}
}
