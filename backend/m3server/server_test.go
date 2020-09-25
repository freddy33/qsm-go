package m3server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3api"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type requestTest struct {
	router      *mux.Router
	contentType string
	typeName    string
	methodName  string
	uri         string
}

func (req *requestTest) String() string {
	return fmt.Sprintf("%s:%q", req.methodName, req.uri)
}

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

func TestReadPointData(t *testing.T) {
	Log.SetInfo()
	router := getApp(m3util.PointTestEnv).Router
	pMsg := &m3api.PointPackDataMsg{}
	sendAndReceive(t, &requestTest{
		router:      router,
		contentType: "proto",
		typeName:    "PointPackDataMsg",
		methodName:  "GET",
		uri:         "/point-data",
	}, nil, pMsg)

	assert.Equal(t, 50, len(pMsg.AllConnections))
	assert.Equal(t, 200, len(pMsg.AllTrios))
	assert.Equal(t, 52, len(pMsg.AllGrowthContexts))
}

func verifyResponsePlainText(t *testing.T, rr *httptest.ResponseRecorder, req *requestTest) {
	assert.Equal(t, http.StatusOK, rr.Result().StatusCode, "fail on %v", req)
	contentType := rr.Header().Get("Content-Type")
	assert.Equal(t, "text/plain", contentType, "fail on %q for %v", contentType, req)
}

func verifyResponseContentType(t *testing.T, rr *httptest.ResponseRecorder, req *requestTest) {
	assert.Equal(t, http.StatusOK, rr.Result().StatusCode, "fail on %v", req)
	contentType := rr.Header().Get("Content-Type")
	contentTypeSplit := strings.Split(contentType, ";")
	assert.Equal(t, 2, len(contentTypeSplit), "fail on %q for %v", contentType, req)
	if req.contentType == "json" {
		assert.Equal(t, contentTypeSplit[0], "application/json", "fail on %q for %v", contentType, req)
	} else if req.contentType == "proto" {
		assert.Equal(t, contentTypeSplit[0], "application/x-protobuf", "fail on %q for %v", contentType, req)
	}
	mt := strings.TrimSpace(contentTypeSplit[1])
	mtSplit := strings.Split(mt, "=")
	assert.Equal(t, 2, len(mtSplit), "fail on=%q source=%q for %v", mt, contentType, req)
	assert.Equal(t, "messageType", mtSplit[0], "fail on=%q source=%q for %v", mt, contentType, req)
	assert.Equal(t, "m3api."+req.typeName, mtSplit[1], "fail on=%q source=%q for %v", mt, contentType, req)
}

func sendAndReceive(t *testing.T, req *requestTest, reqMsg proto.Message, resMsg proto.Message) {
	var err error
	var httpReq *http.Request
	if reqMsg != nil {
		var reqBytes []byte
		if req.contentType == "json" {
			reqBytes, err = json.Marshal(reqMsg)
		} else if req.contentType == "proto" {
			reqBytes, err = proto.Marshal(reqMsg)
		} else {
			assert.Fail(t, "Invalid content type %q for %v", req.contentType, req)
		}
		assert.NoError(t, err, "could not marshal %v", req)
		httpReq, err = http.NewRequest(req.methodName, req.uri, bytes.NewReader(reqBytes))
	} else {
		httpReq, err = http.NewRequest(req.methodName, req.uri, nil)
	}
	assert.NoError(t, err, "Could create request %v", req)

	if req.contentType == "json" {
		httpReq.Header.Set("Content-Type", "application/json")
	} else if req.contentType == "proto" {
		httpReq.Header.Set("Content-Type", "application/x-protobuf")
	} else {
		assert.Fail(t, "Invalid content type %q for %v", req.contentType, req)
	}
	rr := httptest.NewRecorder()
	req.router.ServeHTTP(rr, httpReq)

	b, err := ioutil.ReadAll(rr.Body)
	assert.NoError(t, err, "Fail to read bytes for %v", req)

	if resMsg != nil {
		verifyResponseContentType(t, rr, req)
		var err error
		if req.contentType == "json" {
			err = json.Unmarshal(b, resMsg)
		} else if req.contentType == "proto" {
			err = proto.Unmarshal(b, resMsg)
		} else {
			assert.Fail(t, "Invalid content type %q for %v", req.contentType, req)
		}
		assert.NoError(t, err, "Fail to marshall bytes of %v", req)
	} else {
		verifyResponsePlainText(t, rr, req)
	}
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
