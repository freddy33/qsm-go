package m3server

import (
	"fmt"
	"github.com/freddy33/qsm-go/backend/m3api"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/freddy33/qsm-go/utils/m3db"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var apps = make(map[m3db.QsmEnvID]*QsmApp, 20)

func getApp(envId m3db.QsmEnvID) *QsmApp {
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
	getApp(m3db.PointTestEnv).Router.ServeHTTP(rr, req)
	assert.NoError(t, err, "Fail to call /")
	fmt.Println(rr.Body.String())
}

func TestReadPointData(t *testing.T) {
	Log.SetDebug()
	req, err := http.NewRequest("GET", "/point-data", nil)
	assert.NoError(t, err, "Could create request")
	rr := httptest.NewRecorder()
	getApp(m3db.PointTestEnv).Router.ServeHTTP(rr, req)
	assert.NoError(t, err, "Fail to call /point-data")
	contentType := rr.Header().Get("Content-Type")
	contentTypeSplit := strings.Split(contentType, ";")
	assert.Equal(t, 2, len(contentTypeSplit), "fail on "+contentType)
	assert.Equal(t, contentTypeSplit[0], "application/x-protobuf", "fail on "+contentType)
	mt := strings.TrimSpace(contentTypeSplit[1])
	mtSplit := strings.Split(mt, "=")
	assert.Equal(t, 2, len(mtSplit), "fail on="+mt+" source="+contentType)
	assert.Equal(t, "messageType", mtSplit[0], "fail on="+mt+" source="+contentType)
	assert.Equal(t, "backend.m3api.PointPackDataMsg", mtSplit[1], "fail on="+mt+" source="+contentType)
	b, err := ioutil.ReadAll(rr.Body)
	assert.NoError(t, err, "Fail to read bytes of /point-data")
	pMsg := &m3api.PointPackDataMsg{}
	err = proto.Unmarshal(b, pMsg)
	assert.NoError(t, err, "Fail to marshall bytes of /point-data")
	assert.Equal(t, 50, len(pMsg.AllConnections))
	assert.Equal(t, 200, len(pMsg.AllTrios))
	assert.Equal(t, 52, len(pMsg.AllGrowthContexts))
	assert.Equal(t, m3point.TotalNumberOfCubes+1, len(pMsg.AllCubes))
	assert.Equal(t, m3point.TotalNumberOfCubes+1, len(pMsg.AllPathNodeBuilders))
}
