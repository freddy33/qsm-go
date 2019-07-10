package m3path

import (
	"database/sql"
	"fmt"
	"github.com/freddy33/qsm-go/m3point"
	"sync"
)

type ConnectionState uint8

const (
	ConnectionNoSet ConnectionState = iota
	ConnectionFrom
	ConnectionNext
	ConnectionBlocked
)

type PathContextDb struct {
	id              int
	growthCtx       m3point.GrowthContext
	growthOffset    int
	rootNode        *PathNodeDb
	pathNodeMap     PathNodeMap
	openNodeBuilder *OpenNodeBuilder
}

type OpenNodeBuilder struct {
	pathCtx   *PathContextDb
	d         int
	openNodes []*PathNodeDb
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

	openNodeBuilder := OpenNodeBuilder{pathCtx, 0, make([]*PathNodeDb, 1)}
	openNodeBuilder.openNodes[0] = rootNode

	pathCtx.openNodeBuilder = &openNodeBuilder
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
	return len(pathCtx.openNodeBuilder.openNodes)
}

func (pathCtx *PathContextDb) GetAllOpenPathNodes() []PathNode {
	col := pathCtx.openNodeBuilder.openNodes
	res := make([]PathNode, len(col))
	for i, n := range col {
		res[i] = n
	}
	return res
}

var pathNodeDbPool = sync.Pool{
	New: func() interface{} {
		return new(PathNodeDb)
	},
}

func (onb *OpenNodeBuilder) fillOpenPathNodes() []*PathNodeDb {
	pathCtx := onb.pathCtx
	te, err := GetPathEnv().GetOrCreateTableExec(PathNodesTable)
	if err != nil {
		Log.Fatal(err)
	}
	rows, err := te.Query(SelectPathNodeByCtxAndDistance, pathCtx.id, onb.d)
	if err != nil {
		Log.Fatal(err)
	}
	res := make([]*PathNodeDb, 0, 100)
	for rows.Next() {
		pn := pathNodeDbPool.Get().(*PathNodeDb)
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
				link.node = pn
				link.index = i
				if from[i].Valid && next[i].Valid {
					Log.Errorf("Node DB entry for %d is invalid! link %d is both from and next", pn.id, i)
				}
				link.linkedNode = nil
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
		res = append(res, pn)
	}
	return res
}

func (pathCtx *PathContextDb) MoveToNextNodes() {
	current := pathCtx.openNodeBuilder
	next := new(OpenNodeBuilder)
	next.d = current.d + 1
	next.openNodes = make([]*PathNodeDb, 0, current.nextOpenNodesLen())
	for _, on := range current.openNodes {
		if Log.DoAssert() {
			if on.IsEnd() {
				Log.Errorf("An open end node builder is a dead end at %v", on.P())
				continue
			}
			if !on.IsLatest() {
				if Log.IsTrace() {
					Log.Errorf("An open end node builder has no more active links at %v", on.P())
				}
				continue
			}
		}
		td := on.trioDetails
		if td == nil {
			Log.Fatalf("reached a node without trio %s %s", on.String(), on.GetTrioIndex())
			continue
		}
		nbFrom := 0
		pnb := on.pathBuilder
		for i, pl := range on.links {
			switch pl.connState {
			case ConnectionNext:
				Log.Warnf("executing move to next at %d on open node %s that already has next link at %d!", next.d, on.String(), i)
			case ConnectionFrom:
				nbFrom++
			case ConnectionNoSet:
				cd := td.GetConnections()[i]
				npnb, np := pnb.GetNextPathNodeBuilder(on.P(), cd.GetId(), pathCtx.GetGrowthOffset())

				pId := GetOrCreatePoint(np)

				// TODO: Find node by pathCtx and pId
				// If exists link to it or create dead end

				// Create new node
				pn := pathNodeDbPool.Get().(*PathNodeDb)
				pn.pathCtx = pl.node.pathCtx
				pn.pathBuilder = npnb
				pn.trioDetails = m3point.GetTrioDetails(npnb.GetTrioIndex())
				pn.point = &np
				pn.pointId = pId
				pn.d = next.d
				pn.initLinks()

				// Link the destination node to this link
				pl.linkedNodeId = pn.id
				pl.linkedNode = pn

				// Set one from entry to open node on and check still open node
				isOpen := false
				setFrom := false
				for _, nl := range pn.links {
					if nl.connState == ConnectionNoSet {
						if setFrom {
							isOpen = true
						} else {
							nl.connState = ConnectionFrom
							nl.linkedNode = on
							nl.linkedNodeId = on.id
							setFrom = true
						}
					}
				}
				if !setFrom {
					// link is actually a dead end the dest node cannot accept incoming
					pl.connState = ConnectionBlocked
					pl.linkedNode = nil
					pl.linkedNodeId = -1
				}
				if isOpen {
					next.openNodes = append(next.openNodes, pn)
				}
			}
		}

	}
}

func (pathCtx *PathContextDb) PredictedNextOpenNodesLen() int {
	return pathCtx.openNodeBuilder.nextOpenNodesLen()
}

func (onb *OpenNodeBuilder) nextOpenNodesLen() int {
	return calculatePredictedSize(onb.d, len(onb.openNodes))
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

func (pn *PathNodeDb) String() string {
	return fmt.Sprintf("PNDB%v-%d-%d", pn.point, pn.d, pn.trioDetails.GetId())
}

func (pn *PathNodeDb) GetPathContext() PathContext {
	return pn.pathCtx
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

func (pn *PathNodeDb) addPathLink(connId m3point.ConnectionId) (PathLink, bool) {
	Log.Fatalf("in DB path node %s never call addPathLink", pn.String())
	return nil, false
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
	linkedNode   *PathNodeDb
}

func (pl *PathLinkDb) String() string {
	return fmt.Sprintf("PLDB-%d-%d", pl.connState, pl.index)
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

func (pl *PathLinkDb) createDstNode(pathBuilder m3point.PathNodeBuilder) (PathNode, bool, m3point.PathNodeBuilder) {
	Log.Fatalf("in DB path link %s never call createDstNode", pl.String())
	return nil, false, nil
}

func (*PathLinkDb) dumpInfo(ident int) string {
	panic("implement me")
}
