package client

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3api"
	"github.com/freddy33/qsm-go/model/m3path"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/golang/protobuf/proto"
	"io/ioutil"
)

type ClientPointPackData struct {
	m3point.BasePointPackData
	env                 *QsmApiEnvironment
	ValidNextTrio       [12][2]m3point.TrioIndex
	AllMod4Permutations [12][4]m3point.TrioIndex
	AllMod8Permutations [12][8]m3point.TrioIndex
}

type ClientPathPackData struct {
	m3path.BasePathPackData
	env *QsmApiEnvironment
}

type PathContextCl struct {
	env       *QsmApiEnvironment
	pointData *ClientPointPackData

	id           int
	growthCtx    m3point.GrowthContext
	growthOffset int

	rootNode *PathNodeCl
	pathNodeMap m3path.PathNodeMap
	pathNodes map[int64]*PathNodeCl

	latestD int
}

type PathNodeCl struct {
	pathCtx        *PathContextCl
	id             int64
	d              int
	point          m3point.Point
	trioDetails    *m3point.TrioDetails
	connectionMask uint16
	linkNodes      [m3path.NbConnections]int64
}

/***************************************************************/
// ClientConnection Functions
/***************************************************************/

func GetClientPointPackData(env m3util.QsmEnvironment) *ClientPointPackData {
	if env.GetData(m3util.PointIdx) == nil {
		ppd := new(ClientPointPackData)
		ppd.EnvId = env.GetId()
		ppd.env = env.(*QsmApiEnvironment)
		env.SetData(m3util.PointIdx, ppd)
		// do not return ppd but always the pointer in Env data array
	}
	return env.GetData(m3util.PointIdx).(*ClientPointPackData)
}

func GetClientPathPackData(env m3util.QsmEnvironment) *ClientPathPackData {
	if env.GetData(m3util.PathIdx) == nil {
		ppd := new(ClientPathPackData)
		ppd.EnvId = env.GetId()
		env.SetData(m3util.PathIdx, ppd)
		ppd.env = env.(*QsmApiEnvironment)
		// do not return ppd but always the pointer in Env data array
	}
	return env.GetData(m3util.PathIdx).(*ClientPathPackData)
}

/***************************************************************/
// ClientPointPackData Functions for GetTrioDetails
/***************************************************************/

func (ppd *ClientPointPackData) GetValidNextTrio() [12][2]m3point.TrioIndex {
	return ppd.ValidNextTrio
}

func (ppd *ClientPointPackData) GetAllMod4Permutations() [12][4]m3point.TrioIndex {
	return ppd.AllMod4Permutations
}

func (ppd *ClientPointPackData) GetAllMod8Permutations() [12][8]m3point.TrioIndex {
	return ppd.AllMod8Permutations
}

func (ppd *ClientPointPackData) GetPathNodeBuilder(growthCtx m3point.GrowthContext, offset int, c m3point.Point) m3point.PathNodeBuilder {
	ppd.CheckPathBuildersInitialized()
	// TODO: Verify the key below stay local and is not staying in memory
	key := m3point.CubeKeyId{GrowthCtxId: growthCtx.GetId(), Cube: m3point.CreateTrioCube(ppd, growthCtx, offset, c)}
	cubeId := ppd.GetCubeIdByKey(key)
	return ppd.GetPathNodeBuilderById(cubeId)
}

/***************************************************************/
// PathContextCl Functions
/***************************************************************/

func (pathCtx *PathContextCl) String() string {
	return fmt.Sprintf("PathCL%d-%s-%d", pathCtx.id, pathCtx.growthCtx.String(), pathCtx.growthOffset)
}

func (pathCtx *PathContextCl) GetId() int {
	return pathCtx.id
}

func (pathCtx *PathContextCl) GetGrowthCtx() m3point.GrowthContext {
	return pathCtx.growthCtx
}

func (pathCtx *PathContextCl) GetGrowthOffset() int {
	return pathCtx.growthOffset
}

func (pathCtx *PathContextCl) GetGrowthType() m3point.GrowthType {
	return pathCtx.growthCtx.GetGrowthType()
}

func (pathCtx *PathContextCl) GetGrowthIndex() int {
	return pathCtx.growthCtx.GetGrowthIndex()
}

func (pathCtx *PathContextCl) GetPathNodeMap() m3path.PathNodeMap {
	return pathCtx.pathNodeMap
}

func (pathCtx *PathContextCl) CountAllPathNodes() int {
	return len(pathCtx.pathNodes)
}

func (pathCtx *PathContextCl) InitRootNode(center m3point.Point) {
	uri := "init-root-node"
	reqMsg := &m3api.PathContextMsg{
		PathCtxId:       int32(pathCtx.GetId()),
		GrowthContextId: int32(pathCtx.GetGrowthCtx().GetId()),
		GrowthOffset:    int32(pathCtx.GetGrowthOffset()),
		Center:          m3api.PointToPointMsg(center),
	}
	body := pathCtx.env.clConn.ExecReq("PUT", uri, reqMsg)
	defer m3util.CloseBody(body)
	b, err := ioutil.ReadAll(body)
	if err != nil {
		Log.Fatalf("Could not read body from REST API end point %q due to %s", uri, err.Error())
		return
	}
	pMsg := new(m3api.PathNodeMsg)
	err = proto.Unmarshal(b, pMsg)
	if err != nil {
		Log.Fatalf("Could not marshall body from REST API end point %q due to %s", uri, err.Error())
		return
	}
	pn := new(PathNodeCl)
	pn.id = pMsg.GetPathNodeId()
	pn.pathCtx = pathCtx
	pn.d = int(pMsg.D)
	pn.point = m3api.PointMsgToPoint(pMsg.GetPoint())
	pn.trioDetails = pathCtx.pointData.GetTrioDetails(m3point.TrioIndex(pMsg.GetTrioId()))
	pn.connectionMask = uint16(pMsg.GetConnectionMask())
	for i, lnId := range pMsg.GetLinkedPathNodeIds() {
		pn.linkNodes[i] = lnId
	}
	pathCtx.rootNode = pn
	pathCtx.pathNodeMap.AddPathNode(pn)
	pathCtx.pathNodes[pn.id] = pn
}

func (pathCtx *PathContextCl) GetRootPathNode() m3path.PathNode {
	return pathCtx.rootNode
}

func (pathCtx *PathContextCl) GetNumberOfOpenNodes() int {
	return len(pathCtx.GetAllOpenPathNodes())
}

func (pathCtx *PathContextCl) GetAllOpenPathNodes() []m3path.PathNode {
	res := make([]m3path.PathNode, 0, 100)
	for _, pn := range pathCtx.pathNodes {
		if pn.IsLatest() {
			res = append(res, pn)
		}
	}
	return res
}

func (pathCtx *PathContextCl) MoveToNextNodes() {
	panic("implement me")
}

func (pathCtx *PathContextCl) PredictedNextOpenNodesLen() int {
	return m3path.CalculatePredictedSize(pathCtx.latestD, pathCtx.GetNumberOfOpenNodes())
}

func (pathCtx *PathContextCl) DumpInfo() string {
	return pathCtx.String()
}

/***************************************************************/
// PathNodeCl Functions
/***************************************************************/

func (pn *PathNodeCl) String() string {
	return fmt.Sprintf("PNCL%d-%d-%d-%d-%v", pn.id, pn.pathCtx.id, pn.d, pn.trioDetails.GetId(), pn.point)
}

func (pn *PathNodeCl) GetId() int64 {
	return pn.id
}

func (pn *PathNodeCl) GetPathContext() m3path.PathContext {
	return pn.pathCtx
}

func (pn *PathNodeCl) IsRoot() bool {
	return pn.d == 0
}

func (pn *PathNodeCl) IsLatest() bool {
	return pn.pathCtx.latestD == pn.d
}

func (pn *PathNodeCl) P() m3point.Point {
	return pn.point
}

func (pn *PathNodeCl) D() int {
	return pn.d
}

func (pn *PathNodeCl) GetTrioIndex() m3point.TrioIndex {
	return pn.trioDetails.GetId()
}

func (pn *PathNodeCl) getConnectionState(connIdx int) m3path.ConnectionState {
	return m3path.GetConnectionState(pn.connectionMask, connIdx)
}

func (pn *PathNodeCl) HasOpenConnections() bool {
	for i := 0; i < m3path.NbConnections; i++ {
		if pn.getConnectionState(i) == m3path.ConnectionNotSet {
			return true
		}
	}
	return false
}

func (pn *PathNodeCl) IsFrom(connIdx int) bool {
	return pn.getConnectionState(connIdx) == m3path.ConnectionFrom
}

func (pn *PathNodeCl) IsNext(connIdx int) bool {
	return pn.getConnectionState(connIdx) == m3path.ConnectionNext
}

func (pn *PathNodeCl) IsDeadEnd(connIdx int) bool {
	return pn.getConnectionState(connIdx) == m3path.ConnectionBlocked
}

func (pn *PathNodeCl) GetTrioDetails() *m3point.TrioDetails {
	return pn.trioDetails
}
