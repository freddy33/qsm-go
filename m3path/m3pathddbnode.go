package m3path

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3point"
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

func MakePathContextDBFromGrowthContext(growthCtx m3point.GrowthContext, offset int, pnm PathNodeMap) PathContext {
	pathCtx := PathContextDb{}
	pathCtx.growthCtx = growthCtx
	pathCtx.growthOffset = offset
	pathCtx.rootNode = nil
	pathCtx.pathNodeMap = pnm
	pathCtx.openNodeBuilder = nil

	err := pathCtx.insertInDb()
	if err != nil {
		Log.Errorf("could not save new path context %s due to %v", pathCtx.String(), err)
	}

	return &pathCtx
}

func (pathCtx *PathContextDb) insertInDb() error {
	te, err := GetPathEnv().GetOrCreateTableExec(PathContextsTable)
	if err != nil {
		return err
	}
	id64, err := te.InsertReturnId(pathCtx.GetGrowthCtx().GetId(), pathCtx.GetGrowthOffset())
	if err != nil {
		return err
	}
	pathCtx.id = int(id64)
	return nil
}

func (pathCtx *PathContextDb) String() string {
	return fmt.Sprintf("PathDB%d-%s-%d", pathCtx.id, pathCtx.growthCtx.String(), pathCtx.growthOffset)
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
	if pathCtx.id <= 0 {
		Log.Fatalf("trying to init root node on not inserted in DB path context %s", pathCtx.String())
		return
	}

	// the path builder enforce origin as the center
	nodeBuilder := m3point.GetPathNodeBuilder(pathCtx.growthCtx, pathCtx.growthOffset, m3point.Origin)

	rootNode := getNewPathNodeDb()
	rootNode.pathCtxId = pathCtx.id
	rootNode.pathCtx = pathCtx
	rootNode.pathBuilderId = nodeBuilder.GetCubeId()
	rootNode.pathBuilder = nodeBuilder
	rootNode.trioId = nodeBuilder.GetTrioIndex()
	rootNode.trioDetails = m3point.GetTrioDetails(rootNode.trioId)

	// But the path node here points to real points in space
	rootNode.pointId = getOrCreatePoint(center)
	rootNode.point = &center
	rootNode.d = 0

	err := rootNode.insertInDb()
	if err != nil {
		Log.Fatalf("could not insert the root node %s of path context %s due to %v", rootNode.String(), pathCtx.String(), err)
	}

	pathCtx.rootNode = rootNode

	te, err := GetPathEnv().GetOrCreateTableExec(PathContextsTable)
	if err != nil {
		Log.Errorf("could not get path context table exec on init root node of path context %s due to %v", pathCtx.String(), err)
		return
	}
	rowAffected, err := te.Update(UpdatePathBuilderId, pathCtx.id, rootNode.pathBuilderId)
	if rowAffected != 1 {
		Log.Errorf("could not update path context %s with new path builder id %d due to %v", pathCtx.String(), rootNode.pathBuilderId, err)
		return
	}

	openNodeBuilder := OpenNodeBuilder{pathCtx, 0, make([]*PathNodeDb, 1)}
	openNodeBuilder.openNodes[0] = rootNode

	pathCtx.openNodeBuilder = &openNodeBuilder
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
		pn, err := readRowOnlyIds(rows)
		if err != nil {
			Log.Errorf("Could not read row of %s due to %v", PathNodesTable, err)
		} else {
			res = append(res, pn)
		}
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

				pId := getOrCreatePoint(np)

				// TODO: Find node by pathCtx and pId
				// If exists link to it or create dead end

				// Create new node
				pn := getNewPathNodeDb()
				pn.pathCtx = pl.node.pathCtx
				pn.pathBuilder = npnb
				pn.trioDetails = m3point.GetTrioDetails(npnb.GetTrioIndex())
				pn.point = &np
				pn.pointId = pId
				pn.d = next.d

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
