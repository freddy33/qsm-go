package client

import (
	"fmt"
	"github.com/freddy33/qsm-go/model/m3api"
	"github.com/freddy33/qsm-go/model/m3path"
	"github.com/freddy33/qsm-go/model/m3point"
)

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

	maxDist int
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

func (pathCtx *PathContextCl) mapGetPathNodesAt(dist int) []m3path.PathNode {
	res := make([]m3path.PathNode, 0, 100)
	for _, pn := range pathCtx.pathNodes {
		if dist == pn.d {
			res = append(res, pn)
		}
	}
	return res
}

func (pathCtx *PathContextCl) mapGetPathNodesBetween(fromDist, toDist int) []m3path.PathNode {
	res := make([]m3path.PathNode, 0, 100)
	for _, pn := range pathCtx.pathNodes {
		if pn.d >= fromDist && pn.d <= toDist {
			res = append(res, pn)
		}
	}
	return res
}

func (pathCtx *PathContextCl) GetMaxDist() int {
	return pathCtx.maxDist
}

func (pathCtx *PathContextCl) RequestNewMaxDist(requestDist int) error {
	if pathCtx.GetMaxDist() >= requestDist {
		// Already done
		return nil
	}
	uri := "max-dist"
	reqMsg := &m3api.PathNodesRequestMsg{
		PathCtxId:   int32(pathCtx.GetId()),
		Dist: int32(requestDist),
	}
	pMsg := new(m3api.PathNodesResponseMsg)
	_, err := pathCtx.env.clConn.ExecReq("PUT", uri, reqMsg, pMsg)
	if err != nil {
		return err
	}
	pathCtx.maxDist = int(pMsg.MaxDist)
	Log.Infof("New max dist is %d for %d", pathCtx.GetMaxDist(), pathCtx.GetId())
	return nil
}

func (pathCtx *PathContextCl) GetPathNodesAt(dist int) ([]m3path.PathNode, error) {
	uri := "path-nodes"
	reqMsg := &m3api.PathNodesRequestMsg{
		PathCtxId:   int32(pathCtx.GetId()),
		Dist: int32(dist),
		ToDist: int32(0),
	}
	pMsg := new(m3api.PathNodesResponseMsg)
	_, err := pathCtx.env.clConn.ExecReq("GET", uri, reqMsg, pMsg)
	if err != nil {
		return nil, err
	}
	pathNodes := pMsg.GetPathNodes()
	pathCtx.maxDist = int(pMsg.MaxDist)
	Log.Infof("Received back %d path nodes back at %d for %d", len(pathNodes), dist, pathCtx.GetId())
	for _, pMsg := range pathNodes {
		pathCtx.addPathNodeFromMsg(pMsg)
	}
	return pathCtx.mapGetPathNodesAt(dist), nil
}

func (pathCtx *PathContextCl) GetNumberOfNodesAt(dist int) int {
	uri := "nb-path-nodes"
	reqMsg := &m3api.PathNodesRequestMsg{
		PathCtxId:   int32(pathCtx.GetId()),
		Dist: int32(dist),
		ToDist: int32(0),
	}
	pMsg := new(m3api.PathNodesResponseMsg)
	_, err := pathCtx.env.clConn.ExecReq("GET", uri, reqMsg, pMsg)
	if err != nil {
		Log.Error(err)
		return -1
	}
	nbPathNodes := int(pMsg.GetNbPathNodes())
	pathCtx.maxDist = int(pMsg.MaxDist)
	Log.Infof("Received back nb path nodes = %d at %d for %d", nbPathNodes, dist, pathCtx.GetId())
	return nbPathNodes
}

func (pathCtx *PathContextCl) GetNumberOfNodesBetween(fromDist int, toDist int) int {
	uri := "nb-path-nodes"
	reqMsg := &m3api.PathNodesRequestMsg{
		PathCtxId:   int32(pathCtx.GetId()),
		Dist: int32(fromDist),
		ToDist: int32(toDist),
	}
	pMsg := new(m3api.PathNodesResponseMsg)
	_, err := pathCtx.env.clConn.ExecReq("GET", uri, reqMsg, pMsg)
	if err != nil {
		Log.Error(err)
		return -1
	}
	nbPathNodes := int(pMsg.GetNbPathNodes())
	pathCtx.maxDist = int(pMsg.MaxDist)
	Log.Infof("Received back nb path nodes = %d from %d to %d for %d", nbPathNodes, fromDist, toDist, pathCtx.GetId())
	return nbPathNodes
}

func (pathCtx *PathContextCl) GetPathNodesBetween(fromDist, toDist int) ([]m3path.PathNode, error) {
	uri := "path-nodes"
	reqMsg := &m3api.PathNodesRequestMsg{
		PathCtxId:   int32(pathCtx.GetId()),
		Dist: int32(fromDist),
		ToDist: int32(toDist),
	}
	pMsg := new(m3api.PathNodesResponseMsg)
	_, err := pathCtx.env.clConn.ExecReq("GET", uri, reqMsg, pMsg)
	if err != nil {
		return nil, err
	}
	pathNodes := pMsg.GetPathNodes()
	pathCtx.maxDist = int(pMsg.MaxDist)
	Log.Infof("Received back %d path nodes back from %d to %d for %d", len(pathNodes), fromDist, toDist, pathCtx.GetId())
	for _, pMsg := range pathNodes {
		pathCtx.addPathNodeFromMsg(pMsg)
	}
	return pathCtx.mapGetPathNodesBetween(fromDist, toDist), nil
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
