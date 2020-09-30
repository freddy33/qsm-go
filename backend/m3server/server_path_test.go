package m3server

import (
	"github.com/freddy33/qsm-go/backend/pointdb"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3api"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestPathContextMove(t *testing.T) {
	m3util.SetToTestMode()
	Log.SetInfo()
	qsmApp := getApp(m3util.PathTestEnv)
	router := qsmApp.Router

	initDB(t, router)

	pathCtxId := callCreatePathContext(t, qsmApp, 8, 2, 1, 42)

	callGetPathNodes(t, pathCtxId, router, 3, 12)
	callGetPathNodes(t, pathCtxId, router, 2, 6)
	callGetPathNodes(t, pathCtxId, router, 4, 22)
}

func callCreatePathContext(t *testing.T, qsmApp *QsmApp,
	growthType m3point.GrowthType, growthIndex int, growthOffset int, expectedId int) int {
	reqMsg := &m3api.PathContextRequestMsg{
		GrowthType:   int32(growthType),
		GrowthIndex:  int32(growthIndex),
		GrowthOffset: int32(growthOffset),
	}
	resMsg := &m3api.PathContextResponseMsg{}
	sendAndReceive(t, &requestTest{
		router:      qsmApp.Router,
		contentType: "proto",
		typeName:    "PathContextResponseMsg",
		methodName:  "POST",
		uri:         "/path-context",
	}, reqMsg, resMsg)

	pathCtxId := int(resMsg.PathCtxId)
	assert.True(t, resMsg.PathCtxId > 0, "Did not get path ctx id but "+strconv.Itoa(pathCtxId))
	assert.Equal(t, int32(expectedId), resMsg.GrowthContextId)
	pointData := pointdb.GetPointPackData(qsmApp.Env)
	growthContextFromDb := pointData.GetGrowthContextById(int(resMsg.GrowthContextId))
	assert.Equal(t, growthType, growthContextFromDb.GetGrowthType())
	assert.Equal(t, growthIndex, growthContextFromDb.GetGrowthIndex())
	assert.Equal(t, int32(1), resMsg.GrowthOffset)
	assert.Equal(t, m3point.Point{0, 0, 0}, m3api.PointMsgToPoint(resMsg.RootPathNode.Point))
	assert.Equal(t, int32(0), resMsg.RootPathNode.D)
	assert.Equal(t, int32(5), resMsg.RootPathNode.TrioId)
	assert.True(t, resMsg.RootPathNode.PathNodeId > 0, "Did not get path node id id but "+strconv.Itoa(int(resMsg.RootPathNode.PathNodeId)))
	return pathCtxId
}

func callGetPathNodes(t *testing.T, pathCtxId int, router *mux.Router, dist int, expectedActiveNodes int) {
	reqMsg := &m3api.PathNodesRequestMsg{
		PathCtxId: int32(pathCtxId),
		Dist:      int32(dist),
	}
	nextMoveResponse := &m3api.PathNodesResponseMsg{}
	sendAndReceive(t, &requestTest{
		router:      router,
		contentType: "proto",
		typeName:    "PathNodesResponseMsg",
		methodName:  "GET",
		uri:         "/path-nodes",
	}, reqMsg, nextMoveResponse)

	assert.Equal(t, int32(pathCtxId), nextMoveResponse.GetPathCtxId())
	assert.Equal(t, int32(dist), nextMoveResponse.GetDist())
	// TODO: Check how to return this
	//assert.Equal(t, 1, len(nextMoveResponse.GetModifiedPathNodes()))
	assert.Equal(t, expectedActiveNodes, len(nextMoveResponse.GetPathNodes()))
}
