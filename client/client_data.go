package client

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3path"
	"github.com/freddy33/qsm-go/model/m3point"
)

type ClientPointPackData struct {
	m3point.BasePointPackData
	env *QsmApiEnvironment
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
}

type PathNodeCl struct {
	pathCtx        *PathContextCl
	id             int64
	d              int
	point          m3point.Point
	trioDetails    *m3point.TrioDetails
	connectionMask uint16
	linkNodes      [m3path.NbConnections]*PathNodeCl
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
	Log.Fatalf("in NEW path context %s never call GetPathNodeMap", pathCtx.String())
	return nil
}

func (pathCtx *PathContextCl) CountAllPathNodes() int {
	panic("implement me")
}

func (pathCtx *PathContextCl) InitRootNode(center m3point.Point) {
	panic("implement me")
}

func (pathCtx *PathContextCl) GetRootPathNode() m3path.PathNode {
	return pathCtx.rootNode
}

func (pathCtx *PathContextCl) GetNumberOfOpenNodes() int {
	panic("implement me")
}

func (pathCtx *PathContextCl) GetAllOpenPathNodes() []m3path.PathNode {
	panic("implement me")
}

func (pathCtx *PathContextCl) MoveToNextNodes() {
	panic("implement me")
}

func (pathCtx *PathContextCl) PredictedNextOpenNodesLen() int {
	panic("implement me")
}

func (pathCtx *PathContextCl) DumpInfo() string {
	panic("implement me")
}

/***************************************************************/
// PathNodeCl Functions
/***************************************************************/

func (pn *PathNodeCl) String() string {
	panic("implement me")
}

func (pn *PathNodeCl) GetId() int64 {
	panic("implement me")
}

func (pn *PathNodeCl) GetPathContext() m3path.PathContext {
	panic("implement me")
}

func (pn *PathNodeCl) IsRoot() bool {
	panic("implement me")
}

func (pn *PathNodeCl) IsLatest() bool {
	panic("implement me")
}

func (pn *PathNodeCl) P() m3point.Point {
	panic("implement me")
}

func (pn *PathNodeCl) D() int {
	panic("implement me")
}

func (pn *PathNodeCl) GetTrioIndex() m3point.TrioIndex {
	panic("implement me")
}

func (pn *PathNodeCl) HasOpenConnections() bool {
	panic("implement me")
}

func (pn *PathNodeCl) IsFrom(connIdx int) bool {
	panic("implement me")
}

func (pn *PathNodeCl) IsNext(connIdx int) bool {
	panic("implement me")
}

func (pn *PathNodeCl) IsDeadEnd(connIdx int) bool {
	panic("implement me")
}

func (pn *PathNodeCl) GetTrioDetails() *m3point.TrioDetails {
	panic("implement me")
}
