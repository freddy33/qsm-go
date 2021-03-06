package m3server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/c2h5oh/datasize"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3api"
	"github.com/freddy33/urlquery"
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
	router              *mux.Router
	requestContentType  string
	responseContentType string
	typeName            string
	methodName          string
	uri                 string
}

func (req *requestTest) String() string {
	return fmt.Sprintf("%s:%q %s->%s", req.methodName, req.uri, req.requestContentType, req.responseContentType)
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

func TestListEnv(t *testing.T) {
	Log.SetInfo()
	router := getApp(m3util.PointTestEnv).Router
	pMsg := &m3api.EnvListMsg{}
	if !sendAndReceive(t, &requestTest{
		router:              router,
		requestContentType:  "",
		responseContentType: "proto",
		typeName:            "EnvListMsg",
		methodName:          "GET",
		uri:                 "/list-env",
	}, nil, pMsg) {
		return
	}

	assert.True(t, len(pMsg.Envs) > 0)
	for i, envMsg := range pMsg.Envs {
		Log.Infof("Index %d : Env %d %q %s %02.2f", i, envMsg.EnvId,
			envMsg.SchemaName,
			datasize.ByteSize(envMsg.SchemaSize).HR(),
			envMsg.SchemaSizePercent)
	}
}

func TestReadPointData(t *testing.T) {
	Log.SetInfo()
	router := getApp(m3util.PointTestEnv).Router
	initDB(t, router)
	pMsg := &m3api.PointPackDataMsg{}
	if !sendAndReceive(t, &requestTest{
		router:              router,
		requestContentType:  "",
		responseContentType: "proto",
		typeName:            "PointPackDataMsg",
		methodName:          "GET",
		uri:                 "/point-data",
	}, nil, pMsg) {
		return
	}

	assert.Equal(t, 50, len(pMsg.AllConnections))
	assert.Equal(t, 200, len(pMsg.AllTrios))
	assert.Equal(t, 52, len(pMsg.AllGrowthContexts))
}

func verifyStatus(t *testing.T, rr *httptest.ResponseRecorder, req *requestTest) bool {
	statusCode := rr.Result().StatusCode
	if !assert.Equal(t, http.StatusOK, statusCode, "fail on %v", req) {
		msg := "Content not text/plain"
		if rr.Header().Get("Content-Type") == "text/plain" {
			b, err := ioutil.ReadAll(rr.Body)
			if !assert.NoError(t, err, "Fail to read bytes for %v", req) {
				return false
			}
			msg = string(b)
		}
		return assert.Fail(t, "Received wrong code", "Got %d with message: %q", statusCode, msg)
	}
	return true
}

func verifyResponsePlainText(t *testing.T, rr *httptest.ResponseRecorder, req *requestTest) bool {
	if !verifyStatus(t, rr, req) {
		return false
	}
	contentType := rr.Header().Get("Content-Type")
	return assert.Equal(t, "text/plain", contentType, "fail on %q for %v", contentType, req)
}

func verifyResponseContentType(t *testing.T, rr *httptest.ResponseRecorder, req *requestTest) bool {
	if !verifyStatus(t, rr, req) {
		return false
	}
	contentType := rr.Header().Get("Content-Type")
	contentTypeSplit := strings.Split(contentType, ";")
	if !assert.Equal(t, 2, len(contentTypeSplit), "fail on %q for %v", contentType, req) {
		return false
	}
	var firstPart string
	if req.responseContentType == "json" {
		firstPart = "application/json"
	} else if req.responseContentType == "proto" {
		firstPart = "application/x-protobuf"
	}
	if !assert.Equal(t, firstPart, contentTypeSplit[0], "fail on %q for %v", contentType, req) {
		return false
	}

	mt := strings.TrimSpace(contentTypeSplit[1])
	mtSplit := strings.Split(mt, "=")
	good := assert.Equal(t, 2, len(mtSplit), "fail on=%q source=%q for %v", mt, contentType, req)
	good = good && assert.Equal(t, "messageType", mtSplit[0], "fail on=%q source=%q for %v", mt, contentType, req)
	good = good && assert.Equal(t, "m3api."+req.typeName, mtSplit[1], "fail on=%q source=%q for %v", mt, contentType, req)
	return good
}

func sendAndReceive(t *testing.T, req *requestTest, reqMsg proto.Message, resMsg proto.Message) bool {
	var err error
	var httpReq *http.Request

	if reqMsg != nil {
		var reqBytes []byte
		if req.requestContentType == "query" {
			reqBytes, err = urlquery.Marshal(reqMsg)
		} else if req.requestContentType == "json" {
			reqBytes, err = json.Marshal(reqMsg)
		} else if req.requestContentType == "proto" {
			reqBytes, err = proto.Marshal(reqMsg)
		} else {
			return assert.Fail(t, "Invalid content type %q for %v", req.responseContentType, req)
		}

		if !assert.NoError(t, err, "could not marshal %v", req) {
			return false
		}

		if req.requestContentType == "query" {
			httpReq, err = http.NewRequest(req.methodName, req.uri, nil)
			if !assert.NoError(t, err, "Could create request %v", req) {
				return false
			}
			httpReq.URL.RawQuery = string(reqBytes)
		} else {
			httpReq, err = http.NewRequest(req.methodName, req.uri, bytes.NewReader(reqBytes))
			if !assert.NoError(t, err, "Could create request %v", req) {
				return false
			}
			if req.requestContentType == "json" {
				httpReq.Header.Set("Content-Type", "application/json")
			} else if req.requestContentType == "proto" {
				httpReq.Header.Set("Content-Type", "application/x-protobuf")
			}
		}
	} else {
		if !assert.Equal(t, "", req.requestContentType, "The request %v should have empty request content type", req) {
			return false
		}
		httpReq, err = http.NewRequest(req.methodName, req.uri, nil)
	}
	if !assert.NoError(t, err, "Could create request %v", req) {
		return false
	}

	if resMsg != nil && req.responseContentType != req.requestContentType {
		if req.responseContentType == "json" {
			httpReq.Header.Set("Accept", "application/json")
		} else if req.responseContentType == "proto" {
			httpReq.Header.Set("Accept", "application/x-protobuf")
		} else {
			return assert.Fail(t, "Invalid content type %q for %v", req.responseContentType, req)
		}
	}

	rr := httptest.NewRecorder()
	req.router.ServeHTTP(rr, httpReq)

	if resMsg != nil {
		if !verifyResponseContentType(t, rr, req) {
			return false
		}
		b, err := ioutil.ReadAll(rr.Body)
		if !assert.NoError(t, err, "Fail to read bytes for %v", req) {
			return false
		}
		if req.responseContentType == "json" {
			err = json.Unmarshal(b, resMsg)
		} else if req.responseContentType == "proto" {
			err = proto.Unmarshal(b, resMsg)
		} else {
			return assert.Fail(t, "Invalid content type %q for %v", req.responseContentType, req)
		}
		return assert.NoError(t, err, "Fail to marshall bytes of %v", req)
	} else {
		return verifyResponsePlainText(t, rr, req)
	}
}

func initDB(t *testing.T, router *mux.Router) bool {
	req, err := http.NewRequest("POST", "/init-env", nil)
	if !assert.NoError(t, err, "Could create request") {
		return false
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	contentType := rr.Header().Get("Content-Type")

	good := assert.Equal(t, http.StatusCreated, rr.Result().StatusCode, "Fail to call /init-env") &&
		assert.Equal(t, "text/plain", contentType, "fail on "+contentType)

	return good
}

var testServerAlreadyInit = false

func getTestServerApp(t *testing.T) *QsmApp {
	qsmApp := getApp(m3util.TestServerEnv)
	if !testServerAlreadyInit {
		if !initDB(t, qsmApp.Router) {
			return nil
		}
		testServerAlreadyInit = true
	}
	return qsmApp
}
