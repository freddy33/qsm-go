package m3path

import (
	"database/sql"
	"fmt"
	"github.com/freddy33/qsm-go/m3point"
)

type ConnectionState uint8

const (
	ConnectionNoSet ConnectionState = iota
	ConnectionFrom
	ConnectionNext
)

type PathContextDb struct {
	id           int
	growthCtx    m3point.GrowthContext
	growthOffset int
	rootNode     *PathNodeDb
	pathNodeMap  PathNodeMap
	maxD         int
}

func (pathCtx *PathContextDb) String() string {
	return fmt.Sprintf("PathDB-%s-%d", pathCtx.growthCtx.String(), pathCtx.growthOffset)
}

func (pathCtx *PathContextDb) GetGrowthCtx() m3point.GrowthContext {
	return pathCtx.growthCtx
}

func (pathCtx *PathContextDb) GetGrowthOffset() int {
	return pathCtx.growthOffset
}

func (pathCtx *PathContextDb) GetGrowthType() m3point.GrowthType {
	return pathCtx.growthCtx.GetGrowthType()
}

func (pathCtx *PathContextDb) GetGrowthIndex() int {
	return pathCtx.growthCtx.GetGrowthIndex()
}

func (pathCtx *PathContextDb) GetPathNodeMap() PathNodeMap {
	return pathCtx.pathNodeMap
}

func (pathCtx *PathContextDb) InitRootNode(center m3point.Point) {
	// the path builder enforce origin as the center
	nodeBuilder := m3point.GetPathNodeBuilder(pathCtx.growthCtx, pathCtx.growthOffset, m3point.Origin)

	rootNode := new(PathNodeDb)
	rootNode.pathCtx = pathCtx
	rootNode.pathBuilder = nodeBuilder
	rootNode.trioDetails = m3point.GetTrioDetails(nodeBuilder.GetTrioIndex())
	// But the path node here points to real points in space
	rootNode.pointId = GetOrCreatePoint(center)
	rootNode.point = &center
	rootNode.d = 0
	rootNode.initLinks()

	pathCtx.rootNode = rootNode
	pathCtx.maxD = 0
}

func (pn *PathNodeDb) initLinks() {
	for i := 0; i < 3; i++ {
		link := new(PathLinkDb)
		link.node = pn
		link.index = i
		link.connState = ConnectionNoSet
		link.linkedNodeId = -1
		pn.links[i] = link
	}
}

func (pathCtx *PathContextDb) GetRootPathNode() PathNode {
	return pathCtx.rootNode
}

func (pathCtx *PathContextDb) GetNumberOfOpenNodes() int {
	te, err := GetPathEnv().GetOrCreateTableExec(PathNodesTable)
	if err != nil {
		Log.Fatal(err)
	}
	rows, err := te.Query(SelectPathNodeByCtxAndDistance, pathCtx.id, pathCtx.maxD)
	if err != nil {
		Log.Fatal(err)
	}
	res := make([]PathNode, 0, 100)
	for rows.Next() {
		pn := PathNodeDb{}
		var pathBuilderId, trioId int
		from := [3]sql.NullInt64{}
		next := [3]sql.NullInt64{}
		err = rows.Scan(&pn.id, &pathBuilderId, &trioId, &pn.pointId, &pn.d,
			&from[0], &from[1], &from[2],
			&next[0], &next[1], &next[2])
		if err != nil {
			Log.Errorf("Could not read row of %s due to %v", PathNodesTable, err)
		} else {
			pn.point = GetPoint(pn.pointId)
			pn.pathCtx = pathCtx
			pn.pathBuilder = m3point.GetPathNodeBuilderById(pathBuilderId)
			pn.trioDetails = m3point.GetTrioDetails(m3point.TrioIndex(trioId))
			for i := 0; i < 3; i++ {
				link := new(PathLinkDb)
				link.node = &pn
				link.index = i
				if from[i].Valid && next[i].Valid {
					Log.Errorf("Node DB entry for %d is invalid! link %d is both from and next", pn.id, i)
				}
				if from[i].Valid {
					link.connState = ConnectionFrom
					link.linkedNodeId = int(from[i].Int64)
				} else if next[i].Valid {
					link.connState = ConnectionNext
					link.linkedNodeId = int(next[i].Int64)
				} else {
					link.connState = ConnectionNoSet
					link.linkedNodeId = -1
				}
				pn.links[i] = link
			}
		}
		res = append(res, &pn)
	}
	return len(res)
}

func (*PathContextDb) GetAllOpenPathNodes() []PathNode {
	panic("implement me")
}

func (*PathContextDb) MoveToNextNodes() {
	panic("implement me")
}

func (*PathContextDb) PredictedNextOpenNodesLen() int {
	panic("implement me")
}

func (*PathContextDb) dumpInfo() string {
	panic("implement me")
}

type PathNodeDb struct {
	id          int
	pathCtx     *PathContextDb
	pathBuilder m3point.PathNodeBuilder
	trioDetails *m3point.TrioDetails
	pointId     int64
	point       *m3point.Point
	d           int
	links       [3]*PathLinkDb
}

func (*PathNodeDb) String() string {
	panic("implement me")
}

func (*PathNodeDb) GetPathContext() *BasePathContext {
	panic("implement me")
}

func (*PathNodeDb) IsEnd() bool {
	panic("implement me")
}

func (*PathNodeDb) IsRoot() bool {
	panic("implement me")
}

func (*PathNodeDb) IsLatest() bool {
	panic("implement me")
}

func (*PathNodeDb) P() m3point.Point {
	panic("implement me")
}

func (*PathNodeDb) D() int {
	panic("implement me")
}

func (*PathNodeDb) GetTrioIndex() m3point.TrioIndex {
	panic("implement me")
}

func (*PathNodeDb) GetFrom() PathLink {
	panic("implement me")
}

func (*PathNodeDb) GetOtherFrom() PathLink {
	panic("implement me")
}

func (*PathNodeDb) GetNext(i int) PathLink {
	panic("implement me")
}

func (*PathNodeDb) GetNextConnection(connId m3point.ConnectionId) PathLink {
	panic("implement me")
}

func (*PathNodeDb) calcDist() int {
	panic("implement me")
}

func (*PathNodeDb) addPathLink(connId m3point.ConnectionId) (PathLink, bool) {
	panic("implement me")
}

func (*PathNodeDb) setOtherFrom(pl PathLink) {
	panic("implement me")
}

func (*PathNodeDb) dumpInfo(ident int) string {
	panic("implement me")
}

type PathLinkDb struct {
	node         *PathNodeDb
	index        int
	connState    ConnectionState
	linkedNodeId int
}

func (*PathLinkDb) String() string {
	panic("implement me")
}

func (*PathLinkDb) GetSrc() PathNode {
	panic("implement me")
}

func (*PathLinkDb) GetConnId() m3point.ConnectionId {
	panic("implement me")
}

func (*PathLinkDb) HasDestination() bool {
	panic("implement me")
}

func (*PathLinkDb) IsDeadEnd() bool {
	panic("implement me")
}

func (*PathLinkDb) SetDeadEnd() {
	panic("implement me")
}

func (*PathLinkDb) createDstNode(pathBuilder m3point.PathNodeBuilder) (PathNode, bool, m3point.PathNodeBuilder) {
	panic("implement me")
}

func (*PathLinkDb) dumpInfo(ident int) string {
	panic("implement me")
}
