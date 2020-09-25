package m3server

import (
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
	router := getApp(m3util.PathTestEnv).Router

	initDB(t, router)

	pathCtxId := callCreatePathContext(t, router)

	callNextMove(t, pathCtxId, router, 1, 3)
	callNextMove(t, pathCtxId, router, 2, 6)
}

func callCreatePathContext(t *testing.T, router *mux.Router) int {
	reqMsg := &m3api.PathContextRequestMsg{
		GrowthType:   8,
		GrowthIndex:  2,
		GrowthOffset: 1,
	}
	resMsg := &m3api.PathContextResponseMsg{}
	sendAndReceive(t, &requestTest{
		router:      router,
		contentType: "proto",
		typeName:    "PathContextResponseMsg",
		methodName:  "PUT",
		uri:         "/create-path-ctx",
	}, reqMsg, resMsg)

	pathCtxId := int(resMsg.PathCtxId)
	assert.True(t, resMsg.PathCtxId > 0, "Did not get path ctx id but "+strconv.Itoa(pathCtxId))
	assert.Equal(t, int32(42), resMsg.GrowthContextId)
	assert.Equal(t, int32(1), resMsg.GrowthOffset)
	assert.Equal(t, m3point.Point{0, 0, 0}, m3api.PointMsgToPoint(resMsg.RootPathNode.Point))
	assert.Equal(t, int32(0), resMsg.RootPathNode.D)
	assert.Equal(t, int32(5), resMsg.RootPathNode.TrioId)
	assert.True(t, resMsg.RootPathNode.PathNodeId > 0, "Did not get path node id id but "+strconv.Itoa(int(resMsg.RootPathNode.PathNodeId)))
	return pathCtxId
}

func callNextMove(t *testing.T, pathCtxId int, router *mux.Router, dist int, activeNodes int) {
	reqMsg := &m3api.NextMoveRequestMsg{
		PathCtxId:   int32(pathCtxId),
		CurrentDist: int32(dist - 1),
	}
	nextMoveResponse := &m3api.NextMoveResponseMsg{}
	sendAndReceive(t, &requestTest{
		router:      router,
		contentType: "proto",
		typeName:    "NextMoveResponseMsg",
		methodName:  "POST",
		uri:         "/next-nodes",
	}, reqMsg, nextMoveResponse)

	assert.Equal(t, int32(pathCtxId), nextMoveResponse.GetPathCtxId())
	assert.Equal(t, int32(dist), nextMoveResponse.GetNextDist())
	// TODO: Check how to return this
	//assert.Equal(t, 1, len(nextMoveResponse.GetModifiedPathNodes()))
	assert.Equal(t, activeNodes, len(nextMoveResponse.GetNewPathNodes()))
}
