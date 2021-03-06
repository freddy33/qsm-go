package client

import (
	"bytes"
	"encoding/json"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3api"
	"github.com/freddy33/qsm-go/model/m3path"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/freddy33/urlquery"
	"github.com/golang/protobuf/proto"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	ExecOK              = "OK"
	ExecFailed          = "FAIL"
	ContentTypeProtobuf = "application/x-protobuf"
	ContentTypeJson     = "application/json"
)

func (cl *ClientConnection) ExecReq(method string, uri string, reqMsg proto.Message, respMsg proto.Message, useQueryParams bool) (string, error) {
	uri = strings.TrimPrefix(uri, "/")

	var err error
	var reqBytes []byte
	if reqMsg != nil {
		if useQueryParams {
			reqBytes, err = urlquery.Marshal(reqMsg)
		} else {
			reqBytes, err = proto.Marshal(reqMsg)
		}
		if err != nil {
			return ExecFailed, m3util.MakeWrapQsmErrorf(err, "Failed marshalling message using query %v in %s:%s for REST API end point %q due to: %v", useQueryParams, method, uri, cl.backendRootURL, err)
		}
	}
	req, err := http.NewRequest(method, cl.backendRootURL+uri, bytes.NewReader(reqBytes))
	if err != nil {
		return ExecFailed, m3util.MakeWrapQsmErrorf(err, "Could not request %s:%s for REST API end point %q due to: %v", method, uri, cl.backendRootURL, err)
	}
	if req == nil {
		return ExecFailed, m3util.MakeQsmErrorf("Got a nil request %s:%s for REST API end point %q", method, uri, cl.backendRootURL)
	}
	req.Header.Add(m3api.HttpEnvIdKey, cl.envId.String())
	if reqMsg != nil {
		if useQueryParams {
			req.URL.RawQuery = string(reqBytes)
		} else {
			req.Header.Add("Content-Type", ContentTypeProtobuf)
		}
	}
	if respMsg != nil {
		req.Header.Add("Accept", ContentTypeProtobuf)
	}
	resp, err := cl.httpClient.Do(req)
	if err != nil {
		return ExecFailed, m3util.MakeWrapQsmErrorf(err, "Could not retrieve data from REST API %s:%s end point %q due to: %v", method, uri, cl.backendRootURL, err)
	}
	if resp == nil {
		return ExecFailed, m3util.MakeQsmErrorf("Got a nil response from REST API %s:%s end point %q", method, uri, cl.backendRootURL)
	}

	// Always read the whole body even on error
	respBody := resp.Body
	defer m3util.CloseBody(respBody)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusCreated {
		return ExecFailed, m3util.MakeQsmErrorf("Query %s:%s received wrong status %d", method, uri, resp.StatusCode)
	}

	respBytes, err := ioutil.ReadAll(respBody)
	if err != nil {
		return ExecFailed, m3util.MakeWrapQsmErrorf(err, "Could not read body from REST API end point %q due to %v", uri, err)
	}

	responseContentType := resp.Header.Get("Content-Type")
	if strings.HasPrefix(responseContentType, "text/plain") {
		return string(respBytes), nil
	}

	if respMsg != nil {
		// TODO: Verify content type and resp object type match
		if strings.HasPrefix(responseContentType, ContentTypeJson) {
			err = json.Unmarshal(respBytes, respMsg)
		} else if strings.HasPrefix(responseContentType, ContentTypeProtobuf) {
			err = proto.Unmarshal(respBytes, respMsg)
		} else {
			return ExecFailed, m3util.MakeWrapQsmErrorf(err, "REST API end point %q returned unknown content type %q", uri, responseContentType)
		}
		if err != nil {
			return ExecFailed, m3util.MakeWrapQsmErrorf(err, "Could not unmarshal from REST API end point %q due to %v", uri, err)
		}
		return ExecOK, nil
	}

	return string(respBytes), nil
}

func (cl *ClientConnection) CheckServerUp() bool {
	response, err := cl.ExecReq(http.MethodGet, "", nil, nil, false)
	if err != nil {
		Log.Error(err)
		return false
	}
	Log.Debugf("All good on home response %q", response)
	return true
}

func (env *QsmApiEnvironment) initializeSpaceData() {
	var spaceData *ClientSpacePackData
	spaceIfc := env.GetData(m3util.SpaceIdx)
	if spaceIfc == nil {
		spaceData = new(ClientSpacePackData)
		spaceData.Env = env
		spaceData.allSpaces = make(map[int]*SpaceCl)
		env.SetData(m3util.SpaceIdx, spaceData)
	} else {
		spaceData = spaceIfc.(*ClientSpacePackData)
		if spaceData.Env != env {
			Log.Fatalf("Something wrong with env setup")
		}
	}
}

func (env *QsmApiEnvironment) initializePathData() {
	var pathData *ClientPathPackData
	ppdIfc := env.GetData(m3util.PathIdx)
	if ppdIfc == nil {
		pathData = new(ClientPathPackData)
		pathData.env = env
		pathData.pathCtxMap = m3point.MakeNonBlockConcurrentMap(25)
		env.SetData(m3util.PathIdx, pathData)
	} else {
		pathData = ppdIfc.(*ClientPathPackData)
		if pathData.env != env {
			Log.Fatalf("Something wrong with env setup")
		}
	}
}

func (env *QsmApiEnvironment) initializePointData() {
	var pointData *ClientPointPackData
	ppdIfc := env.GetData(m3util.PointIdx)
	if ppdIfc != nil {
		pointData = ppdIfc.(*ClientPointPackData)
		if pointData.GrowthContextsLoaded {
			Log.Debugf("env %d already loaded", env.GetId())
			return
		}
	}
	if ppdIfc == nil {
		pointData = new(ClientPointPackData)
		pointData.EnvId = env.GetId()
		pointData.env = env
		env.SetData(m3util.PointIdx, pointData)
	}

	if pointData == nil {
		Log.Fatalf("Something wrong above")
		return
	}
	pMsg := &m3api.PointPackDataMsg{}
	_, err := env.clConn.ExecReq(http.MethodGet, "point-data", nil, pMsg, false)
	if err != nil {
		Log.Fatal(err)
		return
	}

	pointData.AllConnections = make([]*m3point.ConnectionDetails, len(pMsg.AllConnections))
	pointData.AllConnectionsByVector = make(map[m3point.Point]*m3point.ConnectionDetails, len(pMsg.AllConnections))
	for idx, c := range pMsg.AllConnections {
		vector := c.GetVector()
		point := m3point.Point{m3point.CInt(vector.GetX()), m3point.CInt(vector.GetY()), m3point.CInt(vector.GetZ())}
		cd := &m3point.ConnectionDetails{
			Id:     m3point.ConnectionId(c.GetConnId()),
			Vector: point,
			ConnDS: m3point.DInt(c.GetDs()),
		}
		pointData.AllConnections[idx] = cd
		pointData.AllConnectionsByVector[point] = cd
	}
	pointData.ConnectionsLoaded = true
	Log.Debugf("loaded %d connections", len(pointData.AllConnections))

	pointData.AllTrioDetails = make([]*m3point.TrioDetails, len(pMsg.AllTrios))
	for idx, tr := range pMsg.AllTrios {
		pointData.AllTrioDetails[idx] = &m3point.TrioDetails{
			Id: m3point.TrioIndex(tr.GetTrioId()),
			Conns: [3]*m3point.ConnectionDetails{pointData.GetConnDetailsById(m3point.ConnectionId(tr.ConnIds[0])),
				pointData.GetConnDetailsById(m3point.ConnectionId(tr.ConnIds[1])),
				pointData.GetConnDetailsById(m3point.ConnectionId(tr.ConnIds[2]))},
		}
	}
	pointData.TrioDetailsLoaded = true
	Log.Debugf("loaded %d trios", len(pointData.AllTrioDetails))

	for i := 0; i < 12; i++ {
		pointData.ValidNextTrio[i][0] = m3point.TrioIndex(pMsg.ValidNextTrioIds[i*2])
		pointData.ValidNextTrio[i][1] = m3point.TrioIndex(pMsg.ValidNextTrioIds[i*2+1])
		for k := 0; k < 4; k++ {
			pointData.AllMod4Permutations[i][k] = m3point.TrioIndex(pMsg.Mod4PermutationsTrioIds[i*4+k])
		}
		for k := 0; k < 8; k++ {
			pointData.AllMod8Permutations[i][k] = m3point.TrioIndex(pMsg.Mod8PermutationsTrioIds[i*8+k])
		}
	}
	Log.Debugf("loaded all valid next and permutation trios")

	pointData.AllGrowthContexts = make([]m3point.GrowthContext, len(pMsg.AllGrowthContexts))
	for idx, gc := range pMsg.AllGrowthContexts {
		pointData.AllGrowthContexts[idx] = &m3point.BaseGrowthContext{
			Env:         env,
			Id:          int(gc.GetGrowthContextId()),
			GrowthType:  m3point.GrowthType(gc.GetGrowthType()),
			GrowthIndex: int(gc.GetGrowthIndex()),
		}
	}
	pointData.GrowthContextsLoaded = true
	Log.Debugf("loaded %d growth context", len(pointData.AllGrowthContexts))
}

func (pathData *ClientPathPackData) GetEnvId() m3util.QsmEnvID {
	if pathData == nil {
		return m3util.NoEnv
	}
	return pathData.env.GetId()
}

func (pathData *ClientPathPackData) GetPathCtx(id m3path.PathContextId) m3path.PathContext {
	pathCtx, ok := pathData.pathCtxMap.Load(id)
	if ok {
		return (*PathContextCl)(pathCtx)
	}

	uri := "path-context"
	reqMsg := &m3api.PathContextIdMsg{PathCtxId: int32(id)}
	pMsg := new(m3api.PathContextMsg)
	_, err := pathData.env.clConn.ExecReq(http.MethodGet, uri, reqMsg, pMsg, true)
	if err != nil {
		Log.Fatal(err)
		return nil
	}
	return MakePatchContextClient(pathData, pMsg)
}

func (pathData *ClientPathPackData) GetPathCtxFromAttributes(growthType m3point.GrowthType, growthIndex int, offset int) (m3path.PathContext, error) {
	uri := "path-context"
	reqMsg := &m3api.PathContextRequestMsg{
		GrowthType:   int32(growthType),
		GrowthIndex:  int32(growthIndex),
		GrowthOffset: int32(offset),
	}
	pMsg := new(m3api.PathContextMsg)
	_, err := pathData.env.clConn.ExecReq(http.MethodPost, uri, reqMsg, pMsg, true)
	if err != nil {
		Log.Fatal(err)
		return nil, nil
	}

	return MakePatchContextClient(pathData, pMsg), nil
}
