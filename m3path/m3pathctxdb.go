package m3path

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3db"
	"github.com/freddy33/qsm-go/m3point"
)

type PathContextDb struct {
	env             *m3db.QsmEnvironment
	ppd             *m3point.PointPackData
	pathNodes       *m3db.TableExec
	points          *m3db.TableExec
	id              int
	growthCtx       m3point.GrowthContext
	growthOffset    int
	rootNode        *PathNodeDb
	openNodeBuilder *OpenNodeBuilder
}

func MakePathContextDBFromGrowthContext(env *m3db.QsmEnvironment, growthCtx m3point.GrowthContext, offset int) PathContext {
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

	err, _ := rootNode.insertInDb()
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

func (pathCtx *PathContextDb) setLinkOnExistingNode(current, next *OpenNodeBuilder, on *PathNodeDb, cd *m3point.ConnectionDetails, pl *PathLinkDb, pnInMap *PathNodeDb) {
	if pnInMap.d != next.d {
		Log.Errorf("Got entry in map %s p=%v but not same D %d != %d!", pnInMap.String(), pnInMap.P(), pnInMap.d, next.d)
		// Blocking link
		pl.SetDeadEnd()
		return
	}
	modelError := pnInMap.setFrom(cd.GetNegId(), on)
	// Check if connection open on the other side for adding other from
	if modelError != nil {
		// from cannot be set => this is blocked
		pl.SetDeadEnd()
	} else {
		// Link the destination node to this link
		pl.connState = ConnectionNext
		pl.linkedNodeId = pnInMap.id
		pl.linkedNode = pnInMap
	}
	err := pnInMap.updateInDb()
	if err != nil {
		Log.Errorf("Got err updating new from in DB %s when updating %s", err.Error(), pnInMap.String())
	}
}

func (pathCtx *PathContextDb) makeNewNodes(current, next *OpenNodeBuilder, on *PathNodeDb, td *m3point.TrioDetails) {
	nbFrom := 0
	nbBlocked := 0
	pnb := on.PathBuilder()
	for i, pl := range on.links {
		switch pl.connState {
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

			updated := false

			_, inCurrent := current.openNodesMap.GetPathNode(np)
			if inCurrent {
				// point back to previous distance outgrowth so d + 1 != d => dead end
				pl.SetDeadEnd()
				updated = true
			} else {
				pnInMap, _ := next.openNodesMap.GetPathNode(np)
				if pnInMap != nil {
					// TODO: The pnInMap may not be in DB yet =>
					// TODO: 1. Either make the following method works with in memory path node
					// TODO: 2. Create a method on PathNodeDb to sync with DB any time anywhere:
					// TODO:  2.1 : It means keep track of state
					// TODO:  2.2 : Make sure only one in memory object with given ID exists
					pathCtx.setLinkOnExistingNode(current, next, on, cd, pl, pnInMap.(*PathNodeDb))
					updated = true
				}
			}

			if !updated {
				// Create new node
				pId := getOrCreatePointTe(pathCtx.pointsTe(), np)
				pn := getNewPathNodeDb()
				pn.pathCtxId = pl.node.pathCtxId
				pn.pathCtx = pl.node.pathCtx
				pn.SetPathBuilder(npnb)
				pn.SetTrioId(npnb.GetTrioIndex())
				pn.point = &np
				pn.pointId = pId
				pn.d = next.d

				sqlErr, filtered := pn.insertInDb()

				if sqlErr != nil {
					if filtered {
						pnInDB := pathCtx.getPathNodeDbByPoint(pId)
						if pnInDB == nil {
							Log.Errorf("Cannot be!! found same point already at %s %d", pathCtx.String(), pId)
						} else {
							if Log.IsDebug() {
								Log.Debugf("Already checked and still found same point already at %s %d", pathCtx.String(), pId)
							}
							next.insertConflict++
							pathCtx.setLinkOnExistingNode(current, next, on, cd, pl, pnInDB)
						}
					} else {
						Log.Error(sqlErr)
					}
					pn.release()
				} else {
					// TODO: Check on inserted or not => report on conflict number
					pn, _ := next.openNodesMap.AddPathNode(pn)
					pnInDB := pn.(*PathNodeDb)
					modelError := pnInDB.setFrom(cd.GetNegId(), on)
					if modelError != nil {
						// from cannot be set => this is blocked
						pl.SetDeadEnd()
					} else {
						// Link the destination node to this link
						pl.connState = ConnectionNext
						pl.linkedNodeId = pnInDB.id
						pl.linkedNode = pnInDB
					}
				}
			}
		}
	}
	// TODO: Verify that all path node touched are updated in DB
	err := on.updateInDb()
	if err != nil {
		Log.Errorf("Got err in DB %s when updating %s", err.Error(), on.String())
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
			Log.Errorf("An open end node builder is a nil node for %s", pathCtx.String())
			return false
		}
		if Log.DoAssert() {
			if on.IsEnd() {
				Log.Errorf("An open end node builder is a dead end at %v", on.P())
				return false
			}
			if !on.IsLatest() {
				Log.Errorf("An open end node builder has no more active links at %v", on.P())
				return false
			}
		}
		if on.trioId == m3point.NilTrioIndex {
			Log.Fatalf("reached a node without trio id %s %s", on.String())
			return true
		}
		td := on.TrioDetails()
		if td == nil {
			Log.Fatalf("reached a node without trio %s %s", on.String(), on.GetTrioIndex())
			return true
		}
		pathCtx.makeNewNodes(current, next, on, td)
		return false
	}, nbParallelProcesses)
	Log.Debugf("%s dist=%d : move from %d to %d open nodes with %d %d conflicts", pathCtx.String(), next.d, len(current.openNodes), len(next.openNodes), next.selectConflict, next.insertConflict)
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
	rows, err := te.Query(SelectPathNodesById, id)
	if err != nil {
		Log.Errorf("could not select path node for id %d exec due to %v", id, err)
		return nil
	}
	defer te.CloseRows(rows)
	if rows.Next() {
		pn, err := readRowOnlyIds(rows)
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
	return nil
}

func (pathCtx *PathContextDb) getPathNodeDbByPoint(pointId int64) *PathNodeDb {
	te := pathCtx.pathNodesTe()
	rows, err := te.Query(SelectPathNodeByCtxAndPoint, pathCtx.id, pointId)
	if err != nil {
		Log.Errorf("could not select path node for ctx %d and point %d exec due to %v", pathCtx.id, pointId, err)
		return nil
	}
	defer te.CloseRows(rows)
	if rows.Next() {
		pn, err := readRowOnlyIds(rows)
		if err != nil {
			Log.Fatalf("Could not read row of %s due to %v", PathNodesTable, err)
			return nil
		}
		if pn.pathCtxId != pathCtx.id {
			Log.Fatalf("While retrieving path node point id %d got a node with context id %d instead of %d",
				pointId, pn.pathCtxId, pathCtx.id)
			return nil
		}
		pn.pathCtx = pathCtx
		return pn
	}
	return nil
}
