package m3server

import (
	"encoding/json"
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

func TestGetAllPathContext(t *testing.T) {
	m3util.SetToTestMode()
	Log.SetInfo()
	qsmApp := getTestServerApp(t)

	pathCtxId, _ := callCreatePathContext(t, qsmApp, 2, 0, 0, 8, 0)
	if pathCtxId <= 0 {
		return
	}

	nbFound, passed := callGetAllPathContext(t, qsmApp)
	if nbFound <= 0 || !passed {
		return
	}
	Log.Infof("Found %d path context", nbFound)
}

func TestPathContextCreateAndIncrease(t *testing.T) {
	m3util.SetToTestMode()
	Log.SetInfo()
	qsmApp := getTestServerApp(t)
	router := qsmApp.Router

	pathCtxId, maxDist := callCreatePathContext(t, qsmApp, 8, 2, 1, 42, 5)
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

func callGetAllPathContext(t *testing.T, qsmApp *QsmApp) (int, bool) {
	resMsg := &m3api.PathContextListMsg{}
	if !sendAndReceive(t, &requestTest{
		router:              qsmApp.Router,
		requestContentType:  "",
		responseContentType: "proto",
		typeName:            "PathContextListMsg",
		methodName:          "GET",
		uri:                 "/path-context",
	}, nil, resMsg) {
		return -1, false
	}

	pointData := pointdb.GetServerPointPackData(qsmApp.Env)

	nbFound := 0
	for i, pathMsg := range resMsg.PathContexts {
		jsonBytes, err := json.Marshal(pathMsg)
		if !assert.NoError(t, err, "fail at index %d", i) {
			return nbFound, false
		}
		msg := string(jsonBytes)
		growthType := m3point.GrowthType(pathMsg.GrowthType)
		growthIndex := int(pathMsg.GrowthIndex)
		growthOffset := int(pathMsg.GrowthOffset)
		growthContextFromDb := pointData.GetGrowthContextByTypeAndIndex(growthType, growthIndex)
		expectedTrioId := growthContextFromDb.GetBaseTrioIndex(pointData, 0, growthOffset)
		pathCtxId, maxDist := assertPathContextMsg(t, qsmApp, pathMsg, growthType, growthIndex, growthOffset, growthContextFromDb.GetId(), expectedTrioId)
		if pathCtxId <= 0 || maxDist < 0 {
			return nbFound, assert.Fail(t, "path context wrong", "fail for index %d : %q", i, msg)
		}
		nbFound++
	}
	return nbFound, true
}

func callCreatePathContext(t *testing.T, qsmApp *QsmApp,
	growthType m3point.GrowthType, growthIndex int, growthOffset int, expectedGrowthContextId int, expectedTrioId m3point.TrioIndex) (int, int) {
	reqMsg := &m3api.PathContextRequestMsg{
		GrowthType:   int32(growthType),
		GrowthIndex:  int32(growthIndex),
		GrowthOffset: int32(growthOffset),
	}
	resMsg := &m3api.PathContextMsg{}
	if !sendAndReceive(t, &requestTest{
		router:              qsmApp.Router,
		requestContentType:  "proto",
		responseContentType: "proto",
		typeName:            "PathContextMsg",
		methodName:          "POST",
		uri:                 "/path-context",
	}, reqMsg, resMsg) {
		return -1, -1
	}

	return assertPathContextMsg(t, qsmApp, resMsg, growthType, growthIndex, growthOffset, expectedGrowthContextId, expectedTrioId)
}

func assertPathContextMsg(t *testing.T, qsmApp *QsmApp, pathMsg *m3api.PathContextMsg,
	growthType m3point.GrowthType, growthIndex int, growthOffset int,
	expectedGrowthContextId int, expectedTrioId m3point.TrioIndex) (int, int) {
	pointData := pointdb.GetServerPointPackData(qsmApp.Env)
	pathCtxId := int(pathMsg.PathCtxId)
	maxDist := int(pathMsg.MaxDist)
	growthContextId := int(pathMsg.GrowthContextId)

	assert.True(t, pathMsg.PathCtxId > 0, "Did not get path ctx id but "+strconv.Itoa(pathCtxId))
	assert.Equal(t, growthType, m3point.GrowthType(pathMsg.GrowthType))

	if growthIndex > 0 {
		if !assert.Equal(t, growthIndex, int(pathMsg.GrowthIndex)) ||
			!assert.Equal(t, growthOffset, int(pathMsg.GrowthOffset)) {
			return -1, -1
		}
	}

	if expectedGrowthContextId > 0 {
		if !assert.Equal(t, expectedGrowthContextId, growthContextId) {
			return -1, -1
		}
	}
	growthContextFromDb := pointData.GetGrowthContextById(growthContextId)
	good := assert.NotNil(t, growthContextFromDb, "did not find growth context id %d", pathMsg.GrowthContextId) &&
		assert.Equal(t, m3point.Point{0, 0, 0}, m3api.PointMsgToPoint(pathMsg.RootPathNode.Point)) &&
		assert.Equal(t, growthType, growthContextFromDb.GetGrowthType()) &&
		assert.Equal(t, int32(0), pathMsg.RootPathNode.D) &&
		assert.True(t, pathMsg.RootPathNode.PathNodeId > 0, "Did not get path node id id but "+strconv.Itoa(int(pathMsg.RootPathNode.PathNodeId)))
	if !good {
		return -1, -1
	}
	if growthIndex > 0 {
		if !assert.Equal(t, growthIndex, growthContextFromDb.GetGrowthIndex()) {
			return -1, -1
		}
	}
	if expectedTrioId != m3point.NilTrioIndex {
		if !assert.Equal(t, expectedTrioId, m3point.TrioIndex(pathMsg.RootPathNode.TrioId)) {
			return -1, -1
		}
	}

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
			router:              router,
			requestContentType:  "query",
			responseContentType: "proto",
			typeName:            "PathNodesResponseMsg",
			methodName:          "PUT",
			uri:                 "/max-dist",
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
		router:              router,
		requestContentType:  "query",
		responseContentType: "proto",
		typeName:            "PathNodesResponseMsg",
		methodName:          "GET",
		uri:                 "/nb-path-nodes",
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
		router:              router,
		requestContentType:  "query",
		responseContentType: "json",
		typeName:            "PathNodesResponseMsg",
		methodName:          "GET",
		uri:                 "/path-nodes",
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
