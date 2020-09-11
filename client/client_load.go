package client

import (
	"bytes"
	"fmt"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3api"
	"github.com/freddy33/qsm-go/model/m3path"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/golang/protobuf/proto"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func (cl *ClientConnection) ExecReq(method string, uri string, m proto.Message) io.ReadCloser {
	uri = strings.TrimPrefix(uri, "/")
	client := http.Client{Timeout: 10 * time.Second}
	var body io.Reader
	if m != nil {
		b, err := proto.Marshal(m)
		if err != nil {
			Log.Errorf("Failed marshalling message in %s:%s for REST API end point %q due to: %s",
				method, uri,
				cl.BackendRootURL, err.Error())
			return nil
		}
		body = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, cl.BackendRootURL+uri, body)
	if err != nil {
		Log.Errorf("Could not request %s:%s for REST API end point %q due to: %s",
			method, uri,
			cl.BackendRootURL, err.Error())
		return nil
	}
	if req == nil {
		Log.Errorf("Got a nil request %s:%s for REST API end point %q",
			method, uri,
			cl.BackendRootURL)
		return nil
	}
	req.Header.Add(m3api.HttpEnvIdKey, cl.EnvId.String())

	resp, err := client.Do(req)
	if err != nil {
		Log.Errorf("Could not retrieve data from REST API %s:%s end point %q due to: %s",
			method, uri,
			cl.BackendRootURL, err.Error())
		return nil
	}
	if resp == nil {
		Log.Errorf("Got a nil response from REST API %s:%s end point %q",
			method, uri,
			cl.BackendRootURL)
		return nil
	}
	return resp.Body
}

func (cl *ClientConnection) CheckServerUp() bool {
	body := cl.ExecReq(http.MethodGet, "", nil)
	if body == nil {
		return false
	}
	defer m3util.CloseBody(body)
	b, err := ioutil.ReadAll(body)
	if err != nil {
		return true
	}
	response := string(b)
	Log.Debugf("All good on home response %q", response)
	return true
}

func GetFullApiTestEnv(envId m3util.QsmEnvID) *QsmApiEnvironment {
	if !m3util.TestMode {
		Log.Fatalf("Cannot use GetFullTestDb in non test mode!")
	}

	env := GetEnvironment(envId)
	cl := env.clConn

	if !cl.CheckServerUp() {
		Log.Fatalf("Test backend server down!")
	}

	// Equivalent of calling filldb job
	body := cl.ExecReq(http.MethodPost, "test-init", nil)
	defer m3util.CloseBody(body)
	b, err := ioutil.ReadAll(body)
	if err != nil {
		Log.Errorf("Could not read body from REST API end point %q due to %s", "test-init", err.Error())
		return nil
	}
	response := string(b)
	substr := fmt.Sprintf("env id %d was initialized", cl.EnvId)
	if strings.Contains(response, substr) {
		Log.Debugf("All good on home response %q", response)
	} else {
		Log.Errorf("The response from REST API end point %q did not have %s in %q", "test-init", substr, response)
		return nil
	}

	env.initialize()
	return env
}

func (env *QsmApiEnvironment) initialize() {
	var ppd *ClientPointPackData
	ppdIfc := env.GetData(m3util.PointIdx)
	if ppdIfc != nil {
		ppd = ppdIfc.(*ClientPointPackData)
		if ppd.PathBuildersLoaded {
			Log.Debugf("Env %d already loaded", env.GetId())
			return
		}
	}
	if ppdIfc == nil {
		ppd = new(ClientPointPackData)
		ppd.EnvId = env.GetId()
		env.SetData(m3util.PointIdx, ppd)
	}
	if ppd == nil {
		Log.Fatalf("Something wrong above")
		return
	}
	body := env.clConn.ExecReq(http.MethodGet, "point-data", nil)
	defer m3util.CloseBody(body)
	b, err := ioutil.ReadAll(body)
	if err != nil {
		Log.Fatalf("Could not read body from REST API end point %q due to %s", "point-data", err.Error())
		return
	}
	pMsg := &m3api.PointPackDataMsg{}
	err = proto.Unmarshal(b, pMsg)
	if err != nil {
		Log.Fatalf("Could not marshall body from REST API end point %q due to %s", "point-data", err.Error())
		return
	}

	ppd.AllConnections = make([]*m3point.ConnectionDetails, len(pMsg.AllConnections))
	ppd.AllConnectionsByVector = make(map[m3point.Point]*m3point.ConnectionDetails, len(pMsg.AllConnections))
	for idx, c := range pMsg.AllConnections {
		vector := c.GetVector()
		point := m3point.Point{m3point.CInt(vector.GetX()), m3point.CInt(vector.GetY()), m3point.CInt(vector.GetZ())}
		cd := &m3point.ConnectionDetails{
			Id:     m3point.ConnectionId(c.GetConnId()),
			Vector: point,
			ConnDS: m3point.DInt(c.GetDs()),
		}
		ppd.AllConnections[idx] = cd
		ppd.AllConnectionsByVector[point] = cd
	}
	ppd.ConnectionsLoaded = true
	Log.Debugf("loaded %d connections", len(ppd.AllConnections))

	ppd.AllTrioDetails = make([]*m3point.TrioDetails, len(pMsg.AllTrios))
	for idx, tr := range pMsg.AllTrios {
		ppd.AllTrioDetails[idx] = &m3point.TrioDetails{
			Id: m3point.TrioIndex(tr.GetTrioId()),
			Conns: [3]*m3point.ConnectionDetails{ppd.GetConnDetailsById(m3point.ConnectionId(tr.ConnIds[0])),
				ppd.GetConnDetailsById(m3point.ConnectionId(tr.ConnIds[1])),
				ppd.GetConnDetailsById(m3point.ConnectionId(tr.ConnIds[2]))},
		}
	}
	ppd.TrioDetailsLoaded = true
	Log.Debugf("loaded %d trios", len(ppd.AllTrioDetails))

	for i := 0; i < 12; i++ {
		ppd.ValidNextTrio[i][0] = m3point.TrioIndex(pMsg.ValidNextTrioIds[i*2])
		ppd.ValidNextTrio[i][1] = m3point.TrioIndex(pMsg.ValidNextTrioIds[i*2+1])
		for k := 0; k < 4; k++ {
			ppd.AllMod4Permutations[i][k] = m3point.TrioIndex(pMsg.Mod4PermutationsTrioIds[i*4+k])
		}
		for k := 0; k < 8; k++ {
			ppd.AllMod8Permutations[i][k] = m3point.TrioIndex(pMsg.Mod8PermutationsTrioIds[i*8+k])
		}
	}
	Log.Debugf("loaded all valid next and permutation trios")

	ppd.AllGrowthContexts = make([]m3point.GrowthContext, len(pMsg.AllGrowthContexts))
	for idx, gc := range pMsg.AllGrowthContexts {
		ppd.AllGrowthContexts[idx] = &m3point.BaseGrowthContext{
			Env:         env,
			Id:          int(gc.GetGrowthContextId()),
			GrowthType:  m3point.GrowthType(gc.GetGrowthType()),
			GrowthIndex: int(gc.GetGrowthIndex()),
		}
	}
	ppd.GrowthContextsLoaded = true
	Log.Debugf("loaded %d growth context", len(ppd.AllGrowthContexts))

	ppd.CubeIdsPerKey = make(map[m3point.CubeKeyId]int, len(pMsg.AllCubes))
	growthCtxByCubeId := make(map[int]int, len(pMsg.AllCubes))
	for id, cube := range pMsg.AllCubes {
		// Do not load dummy cube
		if id != 0 {
			key := m3point.CubeKeyId{
				GrowthCtxId: int(cube.GetGrowthContextId()),
				Cube: m3point.CubeOfTrioIndex{
					Center:      m3point.TrioIndex(cube.GetCenterTrioId()),
					CenterFaces: get6TrioIndex(cube.GetCenterFacesTrioIds()),
					MiddleEdges: get12TrioIndex(cube.GetMiddleEdgesTrioIds()),
				},
			}
			cubeId := int(cube.GetCubeId())
			ppd.CubeIdsPerKey[key] = cubeId
			growthCtxByCubeId[cubeId] = key.GetGrowthCtxId()
		}
	}
	ppd.CubesLoaded = true
	Log.Debugf("loaded %d cubes", len(ppd.CubeIdsPerKey))

	ppd.PathBuilders = make([]*m3point.RootPathNodeBuilder, len(pMsg.AllPathNodeBuilders))
	for idx, pnd := range pMsg.AllPathNodeBuilders {
		if idx == 0 {
			// Dummy cube and path loader
			continue
		}
		cubeId := int(pnd.GetCubeId())
		trIdx := m3point.TrioIndex(pnd.GetTrioId())
		tr := ppd.GetTrioDetails(trIdx)
		ppd.PathBuilders[idx] = &m3point.RootPathNodeBuilder{
			BasePathNodeBuilder: m3point.BasePathNodeBuilder{Ctx: &m3point.PathBuilderContext{
				GrowthCtx: ppd.GetGrowthContextById(growthCtxByCubeId[cubeId]),
				CubeId:    cubeId},
				TrIdx: trIdx},
			PathLinks: convertToInterPathBuilders(ppd, growthCtxByCubeId, tr, pnd),
		}
	}
	ppd.PathBuildersLoaded = true
	Log.Debugf("loaded %d path builders", len(ppd.PathBuilders))
}

func convertToInterPathBuilders(ppd *ClientPointPackData, growthCtxByCubeId map[int]int, tr *m3point.TrioDetails, pnd *m3api.RootPathNodeBuilderMsg) [3]m3point.PathLinkBuilder {
	res := [3]m3point.PathLinkBuilder{}
	interNodeBuilders := pnd.GetInterNodeBuilders()
	for idx, cd := range tr.Conns {
		res[idx] = m3point.PathLinkBuilder{
			ConnId:   cd.Id,
			PathNode: convertToInterPathBuilder(ppd, growthCtxByCubeId, interNodeBuilders[idx]),
		}
	}
	return res
}

func convertToInterPathBuilder(ppd *ClientPointPackData, growthCtxByCubeId map[int]int, pnd *m3api.IntermediatePathNodeBuilderMsg) *m3point.IntermediatePathNodeBuilder {
	cubeId := int(pnd.GetCubeId())
	trIdx := m3point.TrioIndex(pnd.GetTrioId())
	return &m3point.IntermediatePathNodeBuilder{
		BasePathNodeBuilder: m3point.BasePathNodeBuilder{Ctx: &m3point.PathBuilderContext{
			GrowthCtx: ppd.GetGrowthContextById(growthCtxByCubeId[cubeId]),
			CubeId:    cubeId},
			TrIdx: trIdx},
		PathLinks: [2]m3point.PathLinkBuilder{
			{
				ConnId:   m3point.ConnectionId(pnd.Link1ConnId),
				PathNode: convertToLastPathBuilder(ppd, growthCtxByCubeId, pnd.LastNodeBuilder1),
			},
			{
				ConnId:   m3point.ConnectionId(pnd.Link2ConnId),
				PathNode: convertToLastPathBuilder(ppd, growthCtxByCubeId, pnd.LastNodeBuilder2),
			},
		},
	}
}

func convertToLastPathBuilder(ppd *ClientPointPackData, growthCtxByCubeId map[int]int, pnd *m3api.LastPathNodeBuilderMsg) *m3point.LastPathNodeBuilder {
	cubeId := int(pnd.GetCubeId())
	trIdx := m3point.TrioIndex(pnd.GetTrioId())
	return &m3point.LastPathNodeBuilder{
		BasePathNodeBuilder: m3point.BasePathNodeBuilder{Ctx: &m3point.PathBuilderContext{
			GrowthCtx: ppd.GetGrowthContextById(growthCtxByCubeId[cubeId]),
			CubeId:    cubeId},
			TrIdx: trIdx},
		NextMainConnId:  m3point.ConnectionId(pnd.NextMainConnId),
		NextInterConnId: m3point.ConnectionId(pnd.NextInterConnId),
	}
}

func get6TrioIndex(s []int32) [6]m3point.TrioIndex {
	if len(s) != 6 {
		Log.Fatalf("cannot convert slice of size %d to 6", len(s))
	}
	res := [6]m3point.TrioIndex{}
	for idx, i := range s {
		res[idx] = m3point.TrioIndex(i)
	}
	return res
}
func get12TrioIndex(s []int32) [12]m3point.TrioIndex {
	if len(s) != 12 {
		Log.Fatalf("cannot convert slice of size %d to 12", len(s))
	}
	res := [12]m3point.TrioIndex{}
	for idx, i := range s {
		res[idx] = m3point.TrioIndex(i)
	}
	return res
}

func (ppd *ClientPathPackData) CreatePathCtxFromAttributes(growthCtx m3point.GrowthContext, offset int, center m3point.Point) m3path.PathContext {
	uri := "create-path-ctx"
	reqMsg := &m3api.PathContextMsg{
		GrowthContextId: int32(growthCtx.GetId()),
		GrowthOffset:    int32(offset),
		Center:          m3api.PointToPointMsg(center),
	}
	body := ppd.env.clConn.ExecReq(http.MethodPut, uri, reqMsg)
	defer m3util.CloseBody(body)
	b, err := ioutil.ReadAll(body)
	if err != nil {
		Log.Fatalf("Could not read body from REST API end point %q due to %s", uri, err.Error())
		return nil
	}
	pMsg := &m3api.PathContextMsg{}
	err = proto.Unmarshal(b, pMsg)
	if err != nil {
		Log.Fatalf("Could not marshall body from REST API end point %q due to %s", uri, err.Error())
		return nil
	}
	pathCtx := new(PathContextCl)
	pathCtx.id = int(pMsg.GetPathCtxId())
	pathCtx.env = ppd.env
	pointData := GetClientPointPackData(ppd.env)
	pathCtx.pointData = pointData
	pathCtx.growthCtx = pointData.GetGrowthContextById(int(pMsg.GetGrowthContextId()))
	pathCtx.growthOffset = int(pMsg.GetGrowthOffset())
	pathCtx.rootNode = nil
	return pathCtx
}


