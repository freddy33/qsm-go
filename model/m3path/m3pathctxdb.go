package m3path

import (
	"database/sql"
	"fmt"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/freddy33/qsm-go/utils/m3db"
)

type PathContextDb struct {
	env             *m3db.QsmDbEnvironment
	ppd             *m3point.BasePointPackData
	pathNodes       *m3db.TableExec
	points          *m3db.TableExec
	id              int
	growthCtx       m3point.GrowthContext
	growthOffset    int
	rootNode        *PathNodeDb
	openNodeBuilder *OpenNodeBuilder
}

func MakePathContextDBFromGrowthContext(env *m3db.QsmDbEnvironment, growthCtx m3point.GrowthContext, offset int) PathContext {
	pathCtx := PathContextDb{}
	pathCtx.env = env
	pathCtx.ppd = m3point.GetPointPackData(env)
	pathCtx.growthCtx = growthCtx
	pathCtx.growthOffset = offset
	pathCtx.rootNode = nil
	pathCtx.openNodeBuilder = nil

	err := pathCtx.insertInDb()
	if err != nil {
		Log.Errorf("could not save new path context %s due to %v", pathCtx.String(), err)
		return nil
	}

	GetPathPackData(env).addPathCtx(&pathCtx)
	return &pathCtx
}

func (pathCtx *PathContextDb) pathNodesTe() *m3db.TableExec {
	if pathCtx.pathNodes != nil {
		return pathCtx.pathNodes
	}
	var err error
	pathCtx.pathNodes, err = pathCtx.env.GetOrCreateTableExec(PathNodesTable)
	if err != nil {
		Log.Fatalf("could not get %s out of %d env due to '%s'", PathNodesTable, pathCtx.env.GetId(), err.Error())
		return nil
	}
	return pathCtx.pathNodes
}

func (pathCtx *PathContextDb) pointsTe() *m3db.TableExec {
	if pathCtx.points != nil {
		return pathCtx.points
	}
	var err error
	pathCtx.points, err = pathCtx.env.GetOrCreateTableExec(PointsTable)
	if err != nil {
		Log.Fatalf("could not get %s out of %d env due to '%s'", PointsTable, pathCtx.env.GetId(), err.Error())
		return nil
	}
	return pathCtx.points
}

func (pathCtx *PathContextDb) insertInDb() error {
	te, err := pathCtx.env.GetOrCreateTableExec(PathContextsTable)
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

func (pathCtx *PathContextDb) GetId() int {
	return pathCtx.id
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
	Log.Fatalf("in DB path context %s never call GetPathNodeMap", pathCtx.String())
	return nil
}

func (pathCtx *PathContextDb) InitRootNode(center m3point.Point) {
	if pathCtx.id <= 0 {
		Log.Fatalf("trying to init root node on not inserted in DB path context %s", pathCtx.String())
		return
	}

	ppd := m3point.GetPointPackData(pathCtx.GetGrowthCtx().GetEnv())

	// the path builder enforce origin as the center
	nodeBuilder := ppd.GetPathNodeBuilder(pathCtx.growthCtx, pathCtx.growthOffset, m3point.Origin)

	rootNode := getNewPathNodeDb()

	rootNode.pathCtxId = pathCtx.id
	rootNode.pathCtx = pathCtx

	rootNode.SetPathBuilder(nodeBuilder)

	rootNode.SetTrioId(nodeBuilder.GetTrioIndex())

	// But the path node here points to real points in space
	rootNode.pointId = getOrCreatePointTe(pathCtx.pointsTe(), center)
	rootNode.point = &center
	rootNode.d = 0

	err := rootNode.syncInDb()
	if err != nil {
		Log.Fatalf("could not insert the root node %s of path context %s due to %v", rootNode.String(), pathCtx.String(), err)
	}

	pathCtx.rootNode = rootNode

	te, err := pathCtx.env.GetOrCreateTableExec(PathContextsTable)
	if err != nil {
		Log.Errorf("could not get path context table exec on init root node of path context %s due to %v", pathCtx.String(), err)
		return
	}
	rowAffected, err := te.Update(UpdatePathBuilderId, pathCtx.id, rootNode.pathBuilderId)
	if rowAffected != 1 {
		Log.Errorf("could not update path context %s with new path builder id %d due to %v", pathCtx.String(), rootNode.pathBuilderId, err)
		return
	}

	onb := createNewNodeBuilder(nil)
	onb.pathCtx = pathCtx
	onb.addPathNode(rootNode)

	pathCtx.openNodeBuilder = onb
}

func (pathCtx *PathContextDb) GetRootPathNode() PathNode {
	return pathCtx.rootNode
}

func (pathCtx *PathContextDb) GetNumberOfOpenNodes() int {
	onb := pathCtx.openNodeBuilder
	if onb == nil {
		return 0
	}
	return onb.openNodesMap.Size()
}

// TODO: Remove the need for this
func (pathCtx *PathContextDb) GetAllOpenPathNodes() []PathNode {
	pnm := pathCtx.openNodeBuilder.openNodesMap
	res := make([]PathNode, pnm.Size())
	idx := 0
	pnm.Range(func(point m3point.Point, pn PathNode) bool {
		res[idx] = pn
		idx++
		return false
	}, 1)
	return res
}

func (pathCtx *PathContextDb) createConnection(currentD int, fromNode *PathNodeDb, cd *m3point.ConnectionDetails, connIdx int, nextPathNode *PathNodeDb) {
	if nextPathNode.d != fromNode.d+1 {
		Log.Errorf("Got path node %s p=%v but not correct distance since %d != %d + 1!", nextPathNode.String(), nextPathNode.P(), nextPathNode.d, fromNode.d)
		// Blocking link
		fromNode.setDeadEnd(connIdx)
		return
	}
	modelError := nextPathNode.setFrom(cd.GetNegId(), fromNode)
	// Check if connection open on the other side for adding other from
	if modelError != nil {
		// from cannot be set => this is blocked
		fromNode.setDeadEnd(connIdx)
	} else {
		// Link the destination node to this link
		fromNode.setConnectionState(connIdx, ConnectionNext)
		if nextPathNode.id <= 0 {
			fromNode.linkNodeIds[connIdx] = NextLinkIdNotAssigned
		} else {
			fromNode.linkNodeIds[connIdx] = nextPathNode.id
		}
		fromNode.linkNodes[connIdx] = nextPathNode
	}
}

func (pathCtx *PathContextDb) makeNewNodes(current, next *OpenNodeBuilder, on *PathNodeDb, td *m3point.TrioDetails) {
	nbFrom := 0
	nbBlocked := 0
	pnb := on.PathBuilder()
	for i := 0; i < NbConnections; i++ {
		switch on.getConnectionState(i) {
		case ConnectionNext:
			Log.Warnf("executing move to next at %d on open node %s that already has next link at %d!", next.d, on.String(), i)
		case ConnectionFrom:
			nbFrom++
		case ConnectionBlocked:
			nbBlocked++
		case ConnectionNotSet:
			cd := td.GetConnections()[i]
			center := pathCtx.rootNode.P()
			npnb, np := pnb.GetNextPathNodeBuilder(on.P().Sub(center), cd.GetId(), pathCtx.GetGrowthOffset())
			np = np.Add(center)

			pId := getOrCreatePointTe(pathCtx.pointsTe(), np)

			inCurrent := current.openNodesMap.GetPathNode(np)
			if inCurrent != nil {
				// point back to previous distance outgrowth so d + 1 != d => dead end
				on.setDeadEnd(i)
			} else {
				var pn *PathNodeDb
				pn1 := next.openNodesMap.GetPathNode(np)
				if pn1 != nil {
					pn = pn1.(*PathNodeDb)
				}
				if pn == nil {
					// Find if there is a old path node
					pnIdInDB := pathCtx.getPathNodeIdByPoint(pId)
					if pnIdInDB > 0 {
						next.selectConflict++
						// point back to old distance outgrowth so dead end
						on.setDeadEnd(i)
					} else {
						// Create new node
						pn = getNewPathNodeDb()
						pn.pathCtxId = pathCtx.id
						pn.pathCtx = pathCtx
						pn.SetPathBuilder(npnb)
						pn.SetTrioId(npnb.GetTrioIndex())
						pn.point = &np
						pn.pointId = pId
						pn.d = next.d

						fromMap, inserted := next.openNodesMap.AddPathNode(pn)
						if !inserted {
							pn.release()
							pn = fromMap.(*PathNodeDb)
						}
					}
				}
				if pn != nil {
					// The pn may not be in DB yet be careful using id
					pathCtx.createConnection(next.d, on, cd, i, pn)
				}
			}
		}
	}
}

// TODO: This should be in path data entry of the env
var nbParallelProcesses = 8

func (pathCtx *PathContextDb) MoveToNextNodes() {
	current := pathCtx.openNodeBuilder
	next := createNewNodeBuilder(current)

	current.openNodesMap.Range(func(point m3point.Point, pn PathNode) bool {
		on := pn.(*PathNodeDb)
		if on.id < 0 {
			Log.Errorf("An open end path node %s is a not saved node", on.String())
			return false
		}
		if on.IsNew() {
			Log.Errorf("An open end path node %s is new!", on.String())
			return false
		}
		if !on.HasOpenConnections() {
			if Log.IsDebug() {
				Log.Debugf("An open end path node %s has no more active links", on.String())
			}
			return false
		}
		if on.trioId == m3point.NilTrioIndex {
			Log.Fatalf("reached a node without trio id %s %s", on.String())
			return true
		}
		td := on.GetTrioDetails()
		if td == nil {
			Log.Fatalf("reached a node without trio %s %s", on.String(), on.GetTrioIndex())
			return true
		}
		pathCtx.makeNewNodes(current, next, on, td)
		return false
	}, nbParallelProcesses)
	// Save all the new path node to DB
	next.openNodesMap.Range(func(point m3point.Point, pn PathNode) bool {
		on := pn.(*PathNodeDb)
		err := on.syncInDb()
		if err != nil {
			Log.Error(err)
		} else {
			if on.state == InConflictNode {
				next.insertConflict++
			}
		}
		return false
	}, nbParallelProcesses)
	// Update all the previous path node to DB
	// TODO: The update nodes may not be those only
	current.openNodesMap.Range(func(point m3point.Point, pn PathNode) bool {
		on := pn.(*PathNodeDb)
		err := on.syncInDb()
		if err != nil {
			Log.Error(err)
		} else {
			if on.state == InConflictNode {
				Log.Errorf("current path node %s cannot be in conflict!", on.String())
				current.insertConflict++
			}
		}
		return false
	}, nbParallelProcesses)
	Log.Infof("%s dist=%d : move from %d to %d open nodes with %d %d conflicts", pathCtx.String(), next.d, current.openNodesSize(), next.openNodesSize(), next.selectConflict, next.insertConflict)
	pathCtx.openNodeBuilder = next
	current.clear()
}

func (pathCtx *PathContextDb) PredictedNextOpenNodesLen() int {
	return pathCtx.openNodeBuilder.nextOpenNodesLen()
}

func (pathCtx *PathContextDb) dumpInfo() string {
	return pathCtx.String()
}

func (pathCtx *PathContextDb) CountAllPathNodes() int {
	row := pathCtx.pathNodesTe().QueryRow(CountPathNodesByCtx, pathCtx.id)
	var res int
	err := row.Scan(&res)
	if err != nil {
		Log.Errorf("could not count path node for id %d exec due to %v", pathCtx.id, err)
		return -1
	}
	return res
}

func (pathCtx *PathContextDb) getPathNodeDb(id int64) *PathNodeDb {
	te := pathCtx.pathNodesTe()
	row := te.QueryRow(SelectPathNodesById, id)
	pn, err := fetchSingleDbRow(row)
	if err != nil {
		Log.Fatalf("Could not read row of %s due to %v", PathNodesTable, err)
		return nil
	}
	if pn.pathCtxId != pathCtx.id {
		Log.Fatalf("While retrieving path node id %d got a node with context id %d instead of %d",
			id, pn.pathCtxId, pathCtx.id)
		return nil
	}
	pn.pathCtx = pathCtx
	return pn
}

func (pathCtx *PathContextDb) getPathNodeIdByPoint(pointId int64) int64 {
	te := pathCtx.pathNodesTe()
	row := te.QueryRow(SelectPathNodeIdByCtxAndPointId, pathCtx.id, pointId)
	res := int64(-1)
	err := row.Scan(&res)
	if err == sql.ErrNoRows {
		return -1
	}
	if err != nil {
		Log.Fatalf("Could not read row of %s for %s %d due to '%v'", PathNodesTable, pathCtx.String(), pointId, err)
		return -1
	}
	return res
}
