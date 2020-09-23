package m3server

import (
	"bytes"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3api"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

var apps = make(map[m3util.QsmEnvID]*QsmApp, 20)

func getApp(envId m3util.QsmEnvID) *QsmApp {
	_, ok := apps[envId]
	if !ok {
		apps[envId] = MakeApp(envId)
	}
	return apps[envId]
}

func TestHome(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err, "Could create request")
	rr := httptest.NewRecorder()
	getApp(m3util.PointTestEnv).Router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Result().StatusCode, "Fail to call /")
	response := rr.Body.String()
	assert.True(t, strings.HasPrefix(response, "Using env id="+m3util.PointTestEnv.String()), "fail on response="+response)
}

func verifyProtobufContentType(t *testing.T, rr *httptest.ResponseRecorder, typeName string) {
	assert.Equal(t, http.StatusOK, rr.Result().StatusCode)
	contentType := rr.Header().Get("Content-Type")
	contentTypeSplit := strings.Split(contentType, ";")
	assert.Equal(t, 2, len(contentTypeSplit), "fail on "+contentType)
	assert.Equal(t, contentTypeSplit[0], "application/x-protobuf", "fail on "+contentType)
	mt := strings.TrimSpace(contentTypeSplit[1])
	mtSplit := strings.Split(mt, "=")
	assert.Equal(t, 2, len(mtSplit), "fail on="+mt+" source="+contentType)
	assert.Equal(t, "messageType", mtSplit[0], "fail on="+mt+" source="+contentType)
	assert.Equal(t, "m3api."+typeName, mtSplit[1], "fail on="+mt+" source="+contentType+" expect="+typeName)
}

func TestReadPointData(t *testing.T) {
	Log.SetInfo()
	req, err := http.NewRequest("GET", "/point-data", nil)
	assert.NoError(t, err, "Could create request")
	rr := httptest.NewRecorder()
	getApp(m3util.PointTestEnv).Router.ServeHTTP(rr, req)
	verifyProtobufContentType(t, rr, "PointPackDataMsg")

	b, err := ioutil.ReadAll(rr.Body)
	assert.NoError(t, err, "Fail to read bytes of /point-data")
	pMsg := &m3api.PointPackDataMsg{}
	err = proto.Unmarshal(b, pMsg)
	assert.NoError(t, err, "Fail to marshall bytes of /point-data")
	assert.Equal(t, 50, len(pMsg.AllConnections))
	assert.Equal(t, 200, len(pMsg.AllTrios))
	assert.Equal(t, 52, len(pMsg.AllGrowthContexts))
}

func TestLogLevelSetter(t *testing.T) {
	Log.SetDebug()
	assert.True(t, Log.IsDebug())
	assert.True(t, Log.IsInfo())

	req, err := http.NewRequest("POST", "/log?m3server=INFO", nil)
	assert.NoError(t, err, "Could create request")
	rr := httptest.NewRecorder()
	getApp(m3util.PointTestEnv).Router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Result().StatusCode, "Fail to call /log")
	contentType := rr.Header().Get("Content-Type")
	assert.Equal(t, "text/plain", contentType, "fail on "+contentType)

	assert.False(t, Log.IsDebug())
	assert.True(t, Log.IsInfo())
}

func TestCreatePathContext(t *testing.T) {
	m3util.SetToTestMode()
	Log.SetInfo()
	router := getApp(m3util.PathTestEnv).Router

	initDB(t, router)

	pathCtxId := callCreatePathContext(t, router)

	callNextMove(t, pathCtxId, router)
}

func callNextMove(t *testing.T, pathCtxId int, router *mux.Router) {
	reqMsg := &m3api.NextMoveRequestMsg{
		PathCtxId:   int32(pathCtxId),
		CurrentDist: 0,
	}
	reqBytes, err := proto.Marshal(reqMsg)
	req, err := http.NewRequest("POST", "/next-nodes", bytes.NewReader(reqBytes))
	assert.NoError(t, err, "Could create request")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	verifyProtobufContentType(t, rr, "NextMoveResponseMsg")

	b, err := ioutil.ReadAll(rr.Body)
	assert.NoError(t, err, "Fail to read bytes of /next-nodes")
	nextMoveResponse := &m3api.NextMoveResponseMsg{}
	err = proto.Unmarshal(b, nextMoveResponse)
	assert.NoError(t, err, "Fail to marshall bytes of /next-nodes")
	assert.Equal(t, int32(pathCtxId), nextMoveResponse.GetPathCtxId())
	assert.Equal(t, int32(1), nextMoveResponse.GetNextDist())
	// TODO: Check how to return this
	//assert.Equal(t, 1, len(nextMoveResponse.GetModifiedPathNodes()))
	assert.Equal(t, 3, len(nextMoveResponse.GetNewPathNodes()))
}

func callCreatePathContext(t *testing.T, router *mux.Router) int {
	reqMsg := &m3api.PathContextRequestMsg{
		GrowthType:   8,
		GrowthIndex:  2,
		GrowthOffset: 1,
	}
	reqBytes, err := proto.Marshal(reqMsg)
	req, err := http.NewRequest("PUT", "/create-path-ctx", bytes.NewReader(reqBytes))
	assert.NoError(t, err, "Could create request")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	verifyProtobufContentType(t, rr, "PathContextResponseMsg")

	b, err := ioutil.ReadAll(rr.Body)
	assert.NoError(t, err, "Fail to read bytes of /create-path-ctx")
	pMsg := &m3api.PathContextResponseMsg{}
	err = proto.Unmarshal(b, pMsg)
	assert.NoError(t, err, "Fail to marshall bytes of /create-path-ctx")
	pathCtxId := int(pMsg.PathCtxId)
	assert.True(t, pMsg.PathCtxId > 0, "Did not get path ctx id but "+strconv.Itoa(pathCtxId))
	assert.Equal(t, int32(42), pMsg.GrowthContextId)
	assert.Equal(t, int32(1), pMsg.GrowthOffset)
	assert.Equal(t, m3point.Point{0, 0, 0}, m3api.PointMsgToPoint(pMsg.RootPathNode.Point))
	assert.Equal(t, int32(0), pMsg.RootPathNode.D)
	assert.Equal(t, int32(5), pMsg.RootPathNode.TrioId)
	assert.True(t, pMsg.RootPathNode.PathNodeId > 0, "Did not get path node id id but "+strconv.Itoa(int(pMsg.RootPathNode.PathNodeId)))
	return pathCtxId
}

func initDB(t *testing.T, router *mux.Router) {
	req, err := http.NewRequest("POST", "/test-init", nil)
	assert.NoError(t, err, "Could create request")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusCreated, rr.Result().StatusCode, "Fail to call /test-init")
	contentType := rr.Header().Get("Content-Type")
	assert.Equal(t, "text/plain", contentType, "fail on "+contentType)
}
