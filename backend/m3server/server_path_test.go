package m3server

import (
	"fmt"
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
	qsmApp := getApp(m3util.TestServerEnv)
	router := qsmApp.Router

	initDB(t, router)

	pathCtxId, maxDist := callCreatePathContext(t, qsmApp, 8, 2, 1, 42)
	if pathCtxId <= 0 {
		return
	}

	good := callGetPathNodes(t, pathCtxId, &maxDist, router, 3, 0, 12) &&
		callGetPathNodes(t, pathCtxId, &maxDist, router, 2, 0, 6) &&
		callGetPathNodes(t, pathCtxId, &maxDist, router, 4, 0, 22)
	if !good {
		Log.Info("failed!")
	}
}

func callCreatePathContext(t *testing.T, qsmApp *QsmApp,
	growthType m3point.GrowthType, growthIndex int, growthOffset int, expectedId int) (int, int) {
	reqMsg := &m3api.PathContextRequestMsg{
		GrowthType:   int32(growthType),
		GrowthIndex:  int32(growthIndex),
		GrowthOffset: int32(growthOffset),
	}
	resMsg := &m3api.PathContextResponseMsg{}
	if !sendAndReceive(t, &requestTest{
		router:      qsmApp.Router,
		contentType: "proto",
		typeName:    "PathContextResponseMsg",
		methodName:  "POST",
		uri:         "/path-context",
	}, reqMsg, resMsg) {
		return -1, -1
	}

	pathCtxId := int(resMsg.PathCtxId)
	maxDist := int(resMsg.MaxDist)
	assert.True(t, resMsg.PathCtxId > 0, "Did not get path ctx id but "+strconv.Itoa(pathCtxId))
	assert.Equal(t, int32(expectedId), resMsg.GrowthContextId)
	pointData := pointdb.GetServerPointPackData(qsmApp.Env)
	growthContextFromDb := pointData.GetGrowthContextById(int(resMsg.GrowthContextId))
	assert.Equal(t, growthType, growthContextFromDb.GetGrowthType())
	assert.Equal(t, growthIndex, growthContextFromDb.GetGrowthIndex())
	assert.Equal(t, int32(1), resMsg.GrowthOffset)
	assert.Equal(t, m3point.Point{0, 0, 0}, m3api.PointMsgToPoint(resMsg.RootPathNode.Point))
	assert.Equal(t, int32(0), resMsg.RootPathNode.D)
	assert.Equal(t, int32(5), resMsg.RootPathNode.TrioId)
	assert.True(t, resMsg.RootPathNode.PathNodeId > 0, "Did not get path node id id but "+strconv.Itoa(int(resMsg.RootPathNode.PathNodeId)))
	return pathCtxId, maxDist
}

func callGetPathNodes(t *testing.T, pathCtxId int, origMaxDist *int, router *mux.Router, dist int, toDist int, expectedActiveNodes int) bool {
	// Check and call increase max dist if needed
	newMaxDist := dist
	if toDist > 0 {
		newMaxDist = toDist
	}
	if newMaxDist > *origMaxDist {
		// Need to increase max dist
		reqMaxMsg := &m3api.PathNodesRequestMsg{
			PathCtxId: int32(pathCtxId),
			Dist:      int32(newMaxDist),
		}
		resMaxMsg := &m3api.PathNodesResponseMsg{}
		if !sendAndReceive(t, &requestTest{
			router:      router,
			contentType: "proto",
			typeName:    "PathNodesResponseMsg",
			methodName:  "PUT",
			uri:         "/max-dist",
		}, reqMaxMsg, resMaxMsg) {
			return false
		}
		nbPathNodes := int(resMaxMsg.NbPathNodes)
		incMaxDist := int(resMaxMsg.MaxDist)

		good := assert.Equal(t, int32(pathCtxId), resMaxMsg.GetPathCtxId()) &&
			assert.Equal(t, int32(*origMaxDist), resMaxMsg.GetDist()) &&
			assert.Equal(t, int32(newMaxDist), resMaxMsg.GetToDist()) &&
			assert.True(t, incMaxDist >= toDist) &&
			assert.True(t, incMaxDist >= dist) &&
			assert.True(t, incMaxDist > *origMaxDist) &&
			assert.True(t, nbPathNodes >= 3)
		if !good {
			return false
		}
	} else {
		// Test we get accepted status code
	}

	// First call and check on the number of nodes
	reqNbMsg := &m3api.PathNodesRequestMsg{
		PathCtxId: int32(pathCtxId),
		Dist:      int32(dist),
		ToDist:    int32(toDist),
	}
	resNbMsg := &m3api.PathNodesResponseMsg{}
	if !sendAndReceive(t, &requestTest{
		router:      router,
		contentType: "proto",
		typeName:    "PathNodesResponseMsg",
		methodName:  "GET",
		uri:         "/nb-path-nodes",
	}, reqNbMsg, resNbMsg) {
		return false
	}
	nbPathNodes := int(resNbMsg.NbPathNodes)
	maxDist := int(resNbMsg.MaxDist)
	good := assert.Equal(t, int32(pathCtxId), resNbMsg.GetPathCtxId()) &&
		assert.Equal(t, int32(dist), resNbMsg.GetDist()) &&
		assert.Equal(t, int32(toDist), resNbMsg.GetToDist()) &&
		assert.True(t, maxDist >= toDist) &&
		assert.True(t, maxDist >= dist) &&
		assert.True(t, maxDist >= *origMaxDist) &&
		assert.Equal(t, expectedActiveNodes, nbPathNodes)
	if !good {
		return false
	}

	// Then retrieve the collection of nodes
	reqMsg := &m3api.PathNodesRequestMsg{
		PathCtxId: int32(pathCtxId),
		Dist:      int32(dist),
		ToDist:    int32(toDist),
	}
	pathNodesResp := &m3api.PathNodesResponseMsg{}
	if !sendAndReceive(t, &requestTest{
		router:      router,
		contentType: "proto",
		typeName:    "PathNodesResponseMsg",
		methodName:  "GET",
		uri:         "/path-nodes",
	}, reqMsg, pathNodesResp) {
		return false
	}

	good = assert.Equal(t, int32(pathCtxId), pathNodesResp.GetPathCtxId()) &&
		assert.Equal(t, int32(dist), pathNodesResp.GetDist()) &&
		assert.Equal(t, int32(toDist), pathNodesResp.GetToDist()) &&
		assert.Equal(t, expectedActiveNodes, int(pathNodesResp.NbPathNodes)) &&
		assert.Equal(t, expectedActiveNodes, len(pathNodesResp.GetPathNodes()))
	if !good {
		return false
	}

	for i, pn := range pathNodesResp.GetPathNodes() {
		msg := fmt.Sprintf("Something wrong with i=%d id=%d tr=%d d=%d", i, pn.PathNodeId, pn.TrioId, pn.D)
		good = assert.True(t, pn.PathNodeId > 0, msg) &&
			assert.True(t, pn.TrioId >= 0, msg) &&
			assert.True(t, pn.TrioId != int32(m3point.NilTrioIndex), msg)
		if toDist > 0 {
			good = good && assert.True(t, pn.D >= int32(dist) && pn.D <= int32(toDist), msg)
		} else {
			good = good && assert.Equal(t, pn.D, int32(dist), msg)
		}
		if !good {
			return false
		}
	}
	*origMaxDist = int(pathNodesResp.MaxDist)
	return true
}
