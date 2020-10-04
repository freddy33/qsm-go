package pathdb

import (
	"database/sql"
	"fmt"
	"github.com/freddy33/qsm-go/backend/m3db"
	"github.com/freddy33/qsm-go/backend/pointdb"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3path"
	"github.com/freddy33/qsm-go/model/m3point"
)

type PathContextDb struct {
	pathData  *ServerPathPackData
	pointData *pointdb.ServerPointPackData

	id           int
	growthCtx    m3point.GrowthContext
	growthOffset int
	maxDist      int

	rootNode *PathNodeDb
}

func (pathCtx *PathContextDb) createRootNode() error {
	if pathCtx.id <= 0 {
		return m3util.MakeQsmErrorf("trying to init root node on not inserted in DB path context %s", pathCtx.String())
	}

	// the path builder enforce origin as the center
	nodeBuilder := pathCtx.pointData.GetPathNodeBuilder(pathCtx.growthCtx, pathCtx.growthOffset, m3point.Origin)

	rootNode := getNewPathNodeDb()

	rootNode.pathCtxId = pathCtx.id
	rootNode.pathCtx = pathCtx

	rootNode.SetPathBuilder(nodeBuilder)

	rootNode.SetTrioId(nodeBuilder.GetTrioIndex())

	// But the path node here points to real points in space
	center := m3point.Origin
	rootNode.pointId = getOrCreatePointTe(pathCtx.pointsTe(), center)
	rootNode.point = &center
	rootNode.d = 0

	err := rootNode.syncInDb()
	if err != nil {
		return m3util.MakeWrapQsmErrorf(err, "could not insert the root node %s of path context %s due to %v", rootNode.String(), pathCtx.String(), err)
	}

	pathCtx.rootNode = rootNode

	rowAffected, err := pathCtx.pathData.pathCtxTe.Update(UpdatePathBuilderId, pathCtx.id, rootNode.pathBuilderId)
	if err != nil {
		return m3util.MakeWrapQsmErrorf(err, "could not update path context %s with new path builder id %d due to %v", pathCtx.String(), rootNode.pathBuilderId, err)
	}
	if rowAffected != 1 {
		return m3util.MakeQsmErrorf("updating path context %s with new path builder id %d returned wrong rows %d", pathCtx.String(), rootNode.pathBuilderId, rowAffected)
	}

	return nil
}

func (pathCtx *PathContextDb) pathNodesTe() *m3db.TableExec {
	return pathCtx.pathData.pathNodesTe
}

func (pathCtx *PathContextDb) pointsTe() *m3db.TableExec {
	return pathCtx.pathData.pointsTe
}

func (pathCtx *PathContextDb) insertInDb() error {
	id64, err := pathCtx.pathData.pathCtxTe.InsertReturnId(pathCtx.GetGrowthCtx().GetId(), pathCtx.GetGrowthOffset())
	if err != nil {
		return err
	}
	pathCtx.id = int(id64)
	return nil
}

func (pathCtx *PathContextDb) String() string {
	return fmt.Sprintf("PathDB%d-%s-%d-%d", pathCtx.id, pathCtx.growthCtx.String(), pathCtx.growthOffset, pathCtx.maxDist)
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

func (pathCtx *PathContextDb) GetMaxDist() int {
	return pathCtx.maxDist
}

func (pathCtx *PathContextDb) GetPathNodeMap() m3path.PathNodeMap {
	Log.Fatalf("in DB path context %s never call GetPathNodeMap", pathCtx.String())
	return nil
}

func (pathCtx *PathContextDb) GetRootPathNode() m3path.PathNode {
	return pathCtx.rootNode
}

func (pathCtx *PathContextDb) GetNumberOfNodesBetween(fromDist int, toDist int) int {
	row := pathCtx.pathData.pathNodesTe.QueryRow(CountPathNodesByCtxAndBetweenDistance, pathCtx.GetId(), fromDist, toDist)
	var count int
	err := row.Scan(&count)
	if err != nil {
		Log.Error(err)
		return -1
	}
	return count
}

func (pathCtx *PathContextDb) GetNumberOfNodesAt(dist int) int {
	row := pathCtx.pathData.pathNodesTe.QueryRow(CountPathNodesByCtxAndDistance, pathCtx.GetId(), dist)
	var count int
	err := row.Scan(&count)
	if err != nil {
		Log.Error(err)
		return -1
	}
	return count
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
		fromNode.SetConnectionState(connIdx, m3path.ConnectionNext)
		if nextPathNode.id <= 0 {
			fromNode.LinkIds[connIdx] = NextLinkIdNotAssigned
		} else {
			fromNode.LinkIds[connIdx] = nextPathNode.id
		}
		fromNode.linkNodes[connIdx] = nextPathNode
	}
	if fromNode.state == SyncInDbPathNode {
		fromNode.state = ModifiedNode
	}
}

func (pathCtx *PathContextDb) makeNewNodes(current, next *OpenNodeBuilder, on *PathNodeDb, td *m3point.TrioDetails) {
	nbFrom := 0
	nbBlocked := 0
	pnb := on.PathBuilder()
	for i := 0; i < m3path.NbConnections; i++ {
		switch on.GetConnectionState(i) {
		case m3path.ConnectionNext:
			Log.Warnf("executing move to next at %d on open node %s that already has next link at %d!", next.d, on.String(), i)
		case m3path.ConnectionFrom:
			nbFrom++
		case m3path.ConnectionBlocked:
			nbBlocked++
		case m3path.ConnectionNotSet:
			cd := td.GetConnections()[i]
			npnb, np := pnb.GetNextPathNodeBuilder(on.P(), cd.GetId(), pathCtx.GetGrowthOffset())

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
						pn.TrioDetails = nil
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

func (pathCtx *PathContextDb) GetPathNodesAt(dist int) ([]m3path.PathNode, error) {
	if dist == 0 && pathCtx.rootNode != nil {
		res := make([]m3path.PathNode, 1)
		res[0] = pathCtx.rootNode
		return res, nil
	}

	if dist > 0 {
		Log.Debugf("Retrieving all path nodes of %s for dist %d", pathCtx.String(), dist)
	}
	te := pathCtx.pathData.pathNodesTe
	rows, err := te.Query(SelectPathNodesByCtxAndDistance, pathCtx.GetId(), dist)
	if err != nil {
		return nil, err
	}
	defer te.CloseRows(rows)
	res := make([]m3path.PathNode, 0, m3path.CalculatePredictedSize(pathCtx.GetGrowthType(), dist))
	for rows.Next() {
		pn, err := createPathNodeFromDbRows(rows)
		if err != nil {
			return nil, m3util.MakeWrapQsmErrorf(err, "Could not read row of %s due to %v", PathNodesTable, err)
		} else {
			if pn.pathCtxId != pathCtx.id {
				return nil, m3util.MakeQsmErrorf("While retrieving all path nodes got a node with context id %d instead of %d",
					pn.pathCtxId, pathCtx.id)
			}
			pn.pathCtx = pathCtx
			res = append(res, pn)
		}
	}
	if dist > 0 {
		Log.Debugf("Returning %d path nodes of %s for dist %d", len(res), pathCtx.String(), dist)
	}
	return res, nil
}

func (pathCtx *PathContextDb) GetPathNodesBetween(fromDist, toDist int) ([]m3path.PathNode, error) {
	te := pathCtx.pathData.pathNodesTe
	rows, err := te.Query(SelectPathNodesByCtxAndBetweenDistance, pathCtx.GetId(), fromDist, toDist)
	if err != nil {
		return nil, err
	}
	totalSize := 0
	for d := fromDist; d <= toDist; d++ {
		totalSize += m3path.CalculatePredictedSize(pathCtx.GetGrowthType(), d)
	}
	defer te.CloseRows(rows)
	res := make([]m3path.PathNode, 0, totalSize)
	for rows.Next() {
		pn, err := createPathNodeFromDbRows(rows)
		if err != nil {
			return nil, m3util.MakeWrapQsmErrorf(err, "Could not read row of %s due to %v", PathNodesTable, err)
		} else {
			if pn.pathCtxId != pathCtx.id {
				return nil, m3util.MakeQsmErrorf("While retrieving all path nodes got a node with context id %d instead of %d",
					pn.pathCtxId, pathCtx.id)
			}
			pn.pathCtx = pathCtx
			res = append(res, pn)
		}
	}
	return res, nil
}

type RangeErrorCollector struct {
	lastError     error
	collectErrors chan error
	done          chan int
}

func MakeRangeErrorCollector() *RangeErrorCollector {
	return &RangeErrorCollector{
		lastError:     nil,
		collectErrors: make(chan error),
		done:          make(chan int),
	}
}

func (ec *RangeErrorCollector) reset() {
	ec.lastError = nil
}

func (ec *RangeErrorCollector) listen() {
	for {
		select {
		case err := <-ec.collectErrors:
			ec.lastError = err
			Log.Error(err)
		case nbDone := <-ec.done:
			if Log.IsTrace() {
				Log.Tracef("Done %d", nbDone)
			}
			break
		}
	}
}

// TODO: This should be in path data entry of the env
var nbParallelProcesses = 8

func (pathCtx *PathContextDb) RequestNewMaxDist(requestDist int) error {
	if requestDist <= pathCtx.GetMaxDist() {
		return nil
	}
	Log.Debugf("Path context %s will set to new dist %d from %d", pathCtx.String(), requestDist, pathCtx.GetMaxDist())
	nbExecution := 0
	for d := pathCtx.GetMaxDist() + 1; d <= requestDist; d++ {
		err := pathCtx.calculateNextMaxDist()
		if err != nil {
			return err
		}
		nbExecution++
	}
	if requestDist > pathCtx.GetMaxDist() {
		return m3util.MakeQsmErrorf("After executing %d next max dist on path context %d the max dist %d still not the requested value %d",
			nbExecution, pathCtx.GetId(), pathCtx.GetMaxDist(), requestDist)
	}
	Log.Infof("Path context %s max dist set to %d", pathCtx.String(), pathCtx.GetMaxDist())
	return nil
}

func (pathCtx *PathContextDb) calculateNextMaxDist() error {
	current, err := createCurrentNodeBuilder(pathCtx)
	if err != nil {
		return err
	}
	next := createNextNodeBuilder(current)

	Log.Debugf("Moving %s from %d to %d", pathCtx.String(), current.d, next.d)

	ec := MakeRangeErrorCollector()
	go ec.listen()

	current.openNodesMap.Range(func(point m3point.Point, pn m3path.PathNode) bool {
		on := pn.(*PathNodeDb)
		if on.id < 0 {
			ec.collectErrors <- m3util.MakeQsmErrorf("An open end path node %s is a not saved node", on.String())
			return false
		}
		if on.IsNew() {
			ec.collectErrors <- m3util.MakeQsmErrorf("An open end path node %s is new!", on.String())
			return false
		}
		if !on.HasOpenConnections() {
			if Log.IsDebug() {
				Log.Debugf("An open end path node %s has no more active links", on.String())
			}
			return false
		}
		if on.TrioId == m3point.NilTrioIndex {
			ec.collectErrors <- m3util.MakeQsmErrorf("reached a node without trio id %s", on.String())
			return true
		}
		td := on.GetTrioDetails(pathCtx.pointData)
		if td == nil {
			ec.collectErrors <- m3util.MakeQsmErrorf("reached a node without trio %s %s", on.String(), on.GetTrioIndex())
			return true
		}
		pathCtx.makeNewNodes(current, next, on, td)
		return false
	}, nbParallelProcesses)

	ec.done <- next.openNodesMap.Size()
	if ec.lastError != nil {
		return ec.lastError
	}
	ec.reset()

	// Save all the new path node to DB
	next.openNodesMap.Range(func(point m3point.Point, pn m3path.PathNode) bool {
		on := pn.(*PathNodeDb)
		err := on.syncInDb()
		if err != nil {
			ec.collectErrors <- err
		} else {
			if on.state == InConflictNode {
				next.insertConflict++
			}
		}
		return false
	}, nbParallelProcesses)

	ec.done <- next.openNodesMap.Size()
	if ec.lastError != nil {
		return ec.lastError
	}
	ec.reset()

	// Update all the previous path node to DB
	// TODO: The update nodes may not be those only
	current.openNodesMap.Range(func(point m3point.Point, pn m3path.PathNode) bool {
		on := pn.(*PathNodeDb)
		err := on.syncInDb()
		if err != nil {
			ec.collectErrors <- err
		} else {
			if on.state == InConflictNode {
				ec.collectErrors <- m3util.MakeQsmErrorf("current path node %s cannot be in conflict!", on.String())
				current.insertConflict++
			}
		}
		return false
	}, nbParallelProcesses)

	ec.done <- next.openNodesMap.Size()
	if ec.lastError != nil {
		return ec.lastError
	}
	ec.reset()

	Log.Infof("%s from=%d to=%d : move from %d to %d nodes with %d %d conflicts", pathCtx.String(), current.d, next.d, current.openNodesSize(), next.openNodesSize(), next.selectConflict, next.insertConflict)

	pathCtx.maxDist = next.d
	rowAffected, err := pathCtx.pathData.pathCtxTe.Update(UpdateMaxDist, pathCtx.id, pathCtx.maxDist)
	if err != nil {
		return m3util.MakeWrapQsmErrorf(err, "could not update path context %s with new max dist %d due to %v", pathCtx.String(), pathCtx.maxDist, err)
	}
	if rowAffected != 1 {
		return m3util.MakeQsmErrorf("updating path context %s with new max dist %d returned wrong rows %d", pathCtx.String(), pathCtx.maxDist, rowAffected)
	}
	return nil
}

func (pathCtx *PathContextDb) DumpInfo() string {
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

func (pathCtx *PathContextDb) GetPathNodeDb(id int64) (*PathNodeDb, error) {
	row := pathCtx.pathData.pathNodesTe.QueryRow(SelectPathNodesById, id)
	pn, err := createPathNodeFromDbRow(row)
	if err != nil {
		return nil, m3util.MakeWrapQsmErrorf(err, "Could not read row of %s due to %s", PathNodesTable, err.Error())
	}
	if pn.pathCtxId != pathCtx.id {
		return nil, m3util.MakeQsmErrorf("While retrieving path node id %d got a node with context id %d instead of %d",
			id, pn.pathCtxId, pathCtx.id)
	}
	pn.pathCtx = pathCtx
	return pn, nil
}

func createPathCtxFromDbRows(rows *sql.Rows, pathData *ServerPathPackData) (*PathContextDb, error) {
	pathCtx := new(PathContextDb)
	pathCtx.pathData = pathData
	pathCtx.pointData = pointdb.GetServerPointPackData(pathData.env)
	var growthCtxId, pathBuilderId int
	err := rows.Scan(&pathCtx.id, &growthCtxId, &pathCtx.growthOffset, &pathBuilderId, &pathCtx.maxDist)
	if err != nil {
		return nil, err
	}
	pathCtx.growthCtx = pathCtx.pointData.GetGrowthContextById(growthCtxId)

	pathData.pathCtxMap[pathCtx.GetId()] = pathCtx

	rootNodes, err := pathCtx.GetPathNodesAt(0)
	if err != nil {
		return nil, err
	}
	if len(rootNodes) != 1 {
		return nil, m3util.MakeQsmErrorf("There should be only one root node at %s not %d", pathCtx.String(), len(rootNodes))
	}
	pathCtx.rootNode = rootNodes[0].(*PathNodeDb)
	if pathCtx.rootNode.pathBuilderId != pathBuilderId {
		return nil, m3util.MakeQsmErrorf("The path builder id at %s do not match %d != %d", pathCtx.String(), pathCtx.rootNode.pathBuilderId, pathBuilderId)
	}

	return pathCtx, nil
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
