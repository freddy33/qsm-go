package m3point

import (
	"fmt"
	"github.com/freddy33/qsm-go/model/m3api"
	"github.com/freddy33/qsm-go/utils/m3util"
	"github.com/golang/protobuf/proto"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const (
	prodPort    = "8063"
	testPort    = "8877"
	rootUrl     = "http://localhost:" + prodPort + "/"
	testRootUrl = "http://localhost:" + testPort + "/"
)

func GetRootUrl() string {
	if m3util.TestMode {
		return testRootUrl
	} else {
		return rootUrl
	}
}

func ExecGetReq(envId m3util.QsmEnvID, uri string) io.ReadCloser {
	url := GetRootUrl()
	client := http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url+uri, nil)
	if err != nil {
		Log.Errorf("Could not request for REST API end point %q due to: %s", url, err.Error())
		return nil
	}
	if req == nil {
		Log.Errorf("Got a nil request for REST API end point %q", url)
		return nil
	}
	if envId != m3util.NoEnv {
		req.Header.Add("QsmEnvId", envId.String())
	}
	resp, err := client.Do(req)
	if err != nil {
		Log.Errorf("Could not retrieve data from REST API end point %q due to: %s", url, err.Error())
		return nil
	}
	if resp == nil {
		Log.Errorf("Got a nil response from REST API end point %q", url)
		return nil
	}
	return resp.Body
}

func CheckServerUp() bool {
	body := ExecGetReq(m3util.NoEnv, "")
	if body == nil {
		return false
	}
	defer m3util.CloseBody(body)
	bytes, err := ioutil.ReadAll(body)
	if err != nil {
		return true
	}
	response := string(bytes)
	Log.Debugf("All good on home response %q", response)
	return true
}

func getApiFullTestEnv(envId m3util.QsmEnvID) m3util.QsmEnvironment {
	if !m3util.TestMode {
		Log.Fatalf("Cannot use GetFullTestDb in non test mode!")
	}
	if !CheckServerUp() {
		m3util.RunQsm(envId, "run", "server", "-test", "-port", testPort)
	}
	body := ExecGetReq(envId, "/test-init")
	defer m3util.CloseBody(body)
	b, err := ioutil.ReadAll(body)
	if err != nil {
		Log.Errorf("Could not read body from REST API end point %q due to %s", "test-init", err.Error())
		return nil
	}
	response := string(b)
	substr := fmt.Sprintf("env id %d was initialized", envId)
	if strings.Contains(response, substr) {
		Log.Debugf("All good on home response %q", response)
	} else {
		Log.Errorf("The response from REST API end point %q did not have %s in %q", "test-init", substr, response)
		return nil
	}
	m3api.SetEnvironmentCreator()
	env := m3util.GetEnvironment(envId)
	InitializeEnv(env)
	return env
}

func InitializeEnv(env m3util.QsmEnvironment) {
	var ppd *LoadedPointPackData
	ppdIfc := env.GetData(m3util.PointIdx)
	if ppdIfc != nil {
		ppd = ppdIfc.(*LoadedPointPackData)
		if ppd.PathBuildersLoaded {
			Log.Debugf("Env %d already loaded", env.GetId())
			return
		}
	}
	if ppdIfc == nil {
		ppd = new(LoadedPointPackData)
		ppd.EnvId = env.GetId()
		env.SetData(m3util.PointIdx, ppd)
	}
	if ppd == nil {
		Log.Fatalf("Something wrong above")
		return
	}
	body := ExecGetReq(env.GetId(), "point-data")
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

	ppd.AllConnections = make([]*ConnectionDetails, len(pMsg.AllConnections))
	ppd.AllConnectionsByVector = make(map[Point]*ConnectionDetails, len(pMsg.AllConnections))
	for idx, c := range pMsg.AllConnections {
		vector := c.GetVector()
		point := Point{CInt(vector.GetX()), CInt(vector.GetY()), CInt(vector.GetZ())}
		cd := &ConnectionDetails{
			Id:     ConnectionId(c.GetConnId()),
			Vector: point,
			ConnDS: DInt(c.GetDs()),
		}
		ppd.AllConnections[idx] = cd
		ppd.AllConnectionsByVector[point] = cd
	}
	ppd.ConnectionsLoaded = true
	Log.Debugf("loaded %d connections", len(ppd.AllConnections))

	ppd.AllTrioDetails = make([]*TrioDetails, len(pMsg.AllTrios))
	for idx, tr := range pMsg.AllTrios {
		ppd.AllTrioDetails[idx] = &TrioDetails{
			Id: TrioIndex(tr.GetTrioId()),
			Conns: [3]*ConnectionDetails{ppd.GetConnDetailsById(ConnectionId(tr.ConnIds[0])),
				ppd.GetConnDetailsById(ConnectionId(tr.ConnIds[1])),
				ppd.GetConnDetailsById(ConnectionId(tr.ConnIds[2]))},
		}
	}
	ppd.TrioDetailsLoaded = true
	Log.Debugf("loaded %d trios", len(ppd.AllTrioDetails))

	for i := 0; i < 12; i++ {
		ppd.ValidNextTrio[i][0] = TrioIndex(pMsg.ValidNextTrioIds[i*2])
		ppd.ValidNextTrio[i][1] = TrioIndex(pMsg.ValidNextTrioIds[i*2+1])
		for k := 0; k < 4; k++ {
			ppd.AllMod4Permutations[i][k] = TrioIndex(pMsg.Mod4PermutationsTrioIds[i*4+k])
		}
		for k := 0; k < 8; k++ {
			ppd.AllMod8Permutations[i][k] = TrioIndex(pMsg.Mod8PermutationsTrioIds[i*8+k])
		}
	}
	Log.Debugf("loaded all valid next and permutation trios")

	ppd.AllGrowthContexts = make([]GrowthContext, len(pMsg.AllGrowthContexts))
	for idx, gc := range pMsg.AllGrowthContexts {
		ppd.AllGrowthContexts[idx] = &BaseGrowthContext{
			Env:         env,
			Id:          int(gc.GetGrowthContextId()),
			GrowthType:  GrowthType(gc.GetGrowthType()),
			GrowthIndex: int(gc.GetGrowthIndex()),
		}
	}
	ppd.GrowthContextsLoaded = true
	Log.Debugf("loaded %d growth context", len(ppd.AllGrowthContexts))

	ppd.CubeIdsPerKey = make(map[CubeKeyId]int, len(pMsg.AllCubes))
	growthCtxByCubeId := make(map[int]int, len(pMsg.AllCubes))
	for id, cube := range pMsg.AllCubes {
		// Do not load dummy cube
		if id != 0 {
			key := CubeKeyId{
				GrowthCtxId: int(cube.GetGrowthContextId()),
				Cube: CubeOfTrioIndex{
					Center:      TrioIndex(cube.GetCenterTrioId()),
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

	ppd.PathBuilders = make([]*RootPathNodeBuilder, len(pMsg.AllPathNodeBuilders))
	for idx, pnd := range pMsg.AllPathNodeBuilders {
		if idx == 0 {
			// Dummy cube and path loader
			continue
		}
		cubeId := int(pnd.GetCubeId())
		trIdx := TrioIndex(pnd.GetTrioId())
		tr := ppd.GetTrioDetails(trIdx)
		ppd.PathBuilders[idx] = &RootPathNodeBuilder{
			BasePathNodeBuilder: BasePathNodeBuilder{Ctx: &PathBuilderContext{
				GrowthCtx: ppd.GetGrowthContextById(growthCtxByCubeId[cubeId]),
				CubeId:    cubeId},
				TrIdx: trIdx},
			PathLinks: convertToInterPathBuilders(ppd, growthCtxByCubeId, tr, pnd),
		}
	}
	ppd.PathBuildersLoaded = true
	Log.Debugf("loaded %d path builders", len(ppd.PathBuilders))
}

func convertToInterPathBuilders(ppd *LoadedPointPackData, growthCtxByCubeId map[int]int, tr *TrioDetails, pnd *m3api.RootPathNodeBuilderMsg) [3]PathLinkBuilder {
	res := [3]PathLinkBuilder{}
	interNodeBuilders := pnd.GetInterNodeBuilders()
	for idx, cd := range tr.Conns {
		res[idx] = PathLinkBuilder{
			ConnId:   cd.Id,
			PathNode: convertToInterPathBuilder(ppd, growthCtxByCubeId, interNodeBuilders[idx]),
		}
	}
	return res
}

func convertToInterPathBuilder(ppd *LoadedPointPackData, growthCtxByCubeId map[int]int, pnd *m3api.IntermediatePathNodeBuilderMsg) *IntermediatePathNodeBuilder {
	cubeId := int(pnd.GetCubeId())
	trIdx := TrioIndex(pnd.GetTrioId())
	return &IntermediatePathNodeBuilder{
		BasePathNodeBuilder: BasePathNodeBuilder{Ctx: &PathBuilderContext{
			GrowthCtx: ppd.GetGrowthContextById(growthCtxByCubeId[cubeId]),
			CubeId:    cubeId},
			TrIdx: trIdx},
		PathLinks: [2]PathLinkBuilder{
			{
				ConnId:   ConnectionId(pnd.Link1ConnId),
				PathNode: convertToLastPathBuilder(ppd, growthCtxByCubeId, pnd.LastNodeBuilder1),
			},
			{
				ConnId:   ConnectionId(pnd.Link2ConnId),
				PathNode: convertToLastPathBuilder(ppd, growthCtxByCubeId, pnd.LastNodeBuilder2),
			},
		},
	}
}

func convertToLastPathBuilder(ppd *LoadedPointPackData, growthCtxByCubeId map[int]int, pnd *m3api.LastPathNodeBuilderMsg) *LastPathNodeBuilder {
	cubeId := int(pnd.GetCubeId())
	trIdx := TrioIndex(pnd.GetTrioId())
	return &LastPathNodeBuilder{
		BasePathNodeBuilder: BasePathNodeBuilder{Ctx: &PathBuilderContext{
			GrowthCtx: ppd.GetGrowthContextById(growthCtxByCubeId[cubeId]),
			CubeId:    cubeId},
			TrIdx: trIdx},
		NextMainConnId:  ConnectionId(pnd.NextMainConnId),
		NextInterConnId: ConnectionId(pnd.NextInterConnId),
	}
}

func get6TrioIndex(s []int32) [6]TrioIndex {
	if len(s) != 6 {
		Log.Fatalf("cannot convert slice of size %d to 6", len(s))
	}
	res := [6]TrioIndex{}
	for idx, i := range s {
		res[idx] = TrioIndex(i)
	}
	return res
}
func get12TrioIndex(s []int32) [12]TrioIndex {
	if len(s) != 12 {
		Log.Fatalf("cannot convert slice of size %d to 12", len(s))
	}
	res := [12]TrioIndex{}
	for idx, i := range s {
		res[idx] = TrioIndex(i)
	}
	return res
}
