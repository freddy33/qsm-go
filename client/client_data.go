package client

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3api"
	"github.com/freddy33/qsm-go/model/m3path"
	"github.com/freddy33/qsm-go/model/m3point"
)

type ClientPointPackData struct {
	m3point.BasePointPackData
	env                 *QsmApiEnvironment
	ValidNextTrio       [12][2]m3point.TrioIndex
	AllMod4Permutations [12][4]m3point.TrioIndex
	AllMod8Permutations [12][8]m3point.TrioIndex
}

type ClientPathPackData struct {
	env        *QsmApiEnvironment
	pathCtxMap map[int]*PathContextCl
}

type PathContextCl struct {
	env       *QsmApiEnvironment
	pointData *ClientPointPackData

	id           int
	growthCtx    m3point.GrowthContext
	growthOffset int

	rootNode    *PathNodeCl
	pathNodeMap m3path.PathNodeMap
	pathNodes   map[int64]*PathNodeCl

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
	return env.GetData(m3util.PointIdx).(*ClientPointPackData)
}

func GetClientPathPackData(env m3util.QsmEnvironment) *ClientPathPackData {
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
	panic("initRootNode client should not be called")
}

func (pathCtx *PathContextCl) addPathNodeFromMsg(pMsg *m3api.PathNodeMsg) *PathNodeCl {
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
	pathCtx.pathNodeMap.AddPathNode(pn)
	pathCtx.pathNodes[pn.id] = pn
	return pn
}

func (pathCtx *PathContextCl) GetRootPathNode() m3path.PathNode {
	return pathCtx.rootNode
}

func (pathCtx *PathContextCl) GetNumberOfOpenNodes() int {
	return len(pathCtx.GetAllOpenPathNodes())
}

func (pathCtx *PathContextCl) GetAllOpenPathNodes() []m3path.PathNode {
	return pathCtx.mapGetPathNodesAt(pathCtx.latestD)
}

func (pathCtx *PathContextCl) mapGetPathNodesAt(dist int) []m3path.PathNode {
	res := make([]m3path.PathNode, 0, 100)
	for _, pn := range pathCtx.pathNodes {
		if dist == pn.d {
			res = append(res, pn)
		}
	}
	return res
}

func (pathCtx *PathContextCl) GetPathNodesAt(dist int) ([]m3path.PathNode, error) {
	uri := "path-nodes"
	reqMsg := &m3api.PathNodesRequestMsg{
		PathCtxId:   int32(pathCtx.GetId()),
		Dist: int32(dist),
	}
	pMsg := new(m3api.PathNodesResponseMsg)
	_, err := pathCtx.env.clConn.ExecReq("GET", uri, reqMsg, pMsg)
	if err != nil {
		return nil, err
	}
	pathNodes := pMsg.GetPathNodes()
	Log.Infof("Received back %d path nodes back on move to next", len(pathNodes))
	for _, pMsg := range pathNodes {
		pathCtx.addPathNodeFromMsg(pMsg)
	}
	return pathCtx.mapGetPathNodesAt(dist), nil
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
