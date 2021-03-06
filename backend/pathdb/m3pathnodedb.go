package pathdb

import (
	"database/sql"
	"fmt"
	"github.com/freddy33/qsm-go/backend/pointdb"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3path"
	"github.com/freddy33/qsm-go/model/m3point"
	"sync"
)

type PathNodeDbState int16

const (
	NewPathNode PathNodeDbState = iota
	InPoolNode
	SyncInDbPathNode
	InConflictNode
	ModifiedNode
)

type ConnectionsStateMgr interface {
	m3path.ConnectionStateIfc

	SetTrioId(trioId m3point.TrioIndex)
	SetTrioDetails(trioDetails *m3point.TrioDetails)

	// Connection mask management methods
	GetConnectionMask() uint16
	GetConnectionState(connIdx int) m3path.ConnectionState
	SetConnectionMask(connIdx int, maskValue uint16)
	SetConnectionState(connIdx int, state m3path.ConnectionState)

	// Link Ids management methods
	SetLinkIdsFromDbData(linkIds [m3path.NbConnections]sql.NullInt64)
	GetConnsDataForMsg() []m3point.Int64Id
	GetLinkIdsForDb() [m3path.NbConnections]sql.NullInt64
	SetConnStateToNil()
}

type ConnectionsStateDb struct {
	ConnectionMask uint16
	LinkIds        [m3path.NbConnections]m3point.Int64Id
	TrioId         m3point.TrioIndex
	TrioDetails    *m3point.TrioDetails
}

type PathNodeDb struct {
	ConnectionsStateDb
	state PathNodeDbState

	// In most cases this is already filled
	pathCtxId m3path.PathContextId
	pathCtx   *PathContextDb

	// Just Ids will fill this only
	id        m3path.PathNodeId
	pathPoint m3path.PathPoint
	d         int

	// Full Id loaded will fill this
	pathBuilderId  int
	pathBuilderIdx int

	// This is dynamically loaded on demand from PointPackData
	pathBuilder pointdb.PathNodeBuilder

	// This is populated during creation and should not be used for non new node
	linkNodes [m3path.NbConnections]*PathNodeDb
}

/***************************************************************/
// ConnectionsStateDb Functions
/***************************************************************/

// Set a connection state to nil empty state
func (cn *ConnectionsStateDb) SetConnStateToNil() {
	cn.TrioId = m3point.NilTrioIndex
	cn.TrioDetails = nil
	cn.ConnectionMask = uint16(m3path.ConnectionNotSet)
	for i := 0; i < m3path.NbConnections; i++ {
		cn.LinkIds[i] = m3path.LinkIdNotSet
	}
}

func (cn *ConnectionsStateDb) GetTrioIndex() m3point.TrioIndex {
	return cn.TrioId
}

func (cn *ConnectionsStateDb) GetConnectionMask() uint16 {
	return cn.ConnectionMask
}

func (cn *ConnectionsStateDb) GetConnectionState(connIdx int) m3path.ConnectionState {
	return m3path.GetConnectionState(cn.ConnectionMask, connIdx)
}

func (cn *ConnectionsStateDb) SetFullConnectionMask(maskValue uint16) {
	cn.ConnectionMask = maskValue
}

func (cn *ConnectionsStateDb) SetConnectionMask(connIdx int, maskValue uint16) {
	allConnsMask := cn.ConnectionMask
	// Zero the bit mask for this connection
	allConnsMask &^= m3path.SingleConnectionMask << uint16(connIdx*m3path.ConnectionMaskBits)
	// Add the new mask value
	allConnsMask |= maskValue << uint16(connIdx*m3path.ConnectionMaskBits)
	cn.ConnectionMask = allConnsMask
}

func (cn *ConnectionsStateDb) SetConnectionState(connIdx int, state m3path.ConnectionState) {
	connMask := m3path.SetConnectionState(cn.ConnectionMask, connIdx, state)
	cn.SetConnectionMask(connIdx, connMask)
}

func (cn *ConnectionsStateDb) SetLinkIdsFromDbData(linkIds [m3path.NbConnections]sql.NullInt64) {
	for i := 0; i < m3path.NbConnections; i++ {
		switch cn.GetConnectionState(i) {
		case m3path.ConnectionNotSet:
			if Log.DoAssert() {
				if linkIds[i].Valid {
					Log.Errorf("Not set linked id of %v has wrong state in DB for %d since %v should be NULL",
						cn, i, linkIds[i])
				}
			}
			cn.LinkIds[i] = m3path.LinkIdNotSet
		case m3path.ConnectionFrom:
			if !linkIds[i].Valid {
				Log.Errorf("Linked id of %v has wrong state in DB for %d since %v should be linked",
					cn, i, linkIds[i])
			}
			cn.LinkIds[i] = m3point.Int64Id(linkIds[i].Int64)
		case m3path.ConnectionNext:
			if linkIds[i].Valid {
				cn.LinkIds[i] = m3point.Int64Id(linkIds[i].Int64)
			} else {
				cn.LinkIds[i] = m3path.NextLinkIdNotAssigned
			}
		case m3path.ConnectionBlocked:
			if Log.DoAssert() {
				if linkIds[i].Valid {
					Log.Errorf("Blocked linked id of %v has wrong state in DB for %d since %v should be NULL",
						cn, i, linkIds[i])
				}
			}
			cn.LinkIds[i] = m3path.DeadEndId
		}
	}
}

func (cn *ConnectionsStateDb) GetConnsDataForMsg() []int64 {
	res := make([]int64, len(cn.LinkIds))
	for i, id := range cn.LinkIds {
		res[i] = int64(id)
	}
	return res
}

func (cn *ConnectionsStateDb) GetLinkIdsForDb() [m3path.NbConnections]sql.NullInt64 {
	dbLinkIds := [m3path.NbConnections]sql.NullInt64{}
	for i := 0; i < m3path.NbConnections; i++ {
		switch cn.GetConnectionState(i) {
		case m3path.ConnectionNotSet:
			//if Log.DoAssert() {
			if cn.LinkIds[i] != m3path.LinkIdNotSet {
				Log.Errorf("Linked id of %v not set correctly for %d since %d != %d",
					cn, i, cn.LinkIds[i], m3path.LinkIdNotSet)
			}
			//}
			dbLinkIds[i].Valid = false
			dbLinkIds[i].Int64 = 0
		case m3path.ConnectionFrom:
			if cn.LinkIds[i] <= 0 {
				Log.Fatalf("Linked id for from of %v not set correctly for %d since %d <= 0",
					cn, i, cn.LinkIds[i])
			}
			dbLinkIds[i].Valid = true
			dbLinkIds[i].Int64 = int64(cn.LinkIds[i])
		case m3path.ConnectionNext:
			if cn.LinkIds[i] == m3path.NextLinkIdNotAssigned {
				dbLinkIds[i].Valid = false
				dbLinkIds[i].Int64 = 0
			} else if cn.LinkIds[i] > 0 {
				dbLinkIds[i].Valid = true
				dbLinkIds[i].Int64 = int64(cn.LinkIds[i])
			} else {
				Log.Fatalf("Linked id for next of %v not set correctly for %d since %d != %d && %d <= 0",
					cn, i, cn.LinkIds[i], m3path.NextLinkIdNotAssigned, cn.LinkIds[i])
			}
		case m3path.ConnectionBlocked:
			if cn.LinkIds[i] != m3path.DeadEndId {
				Log.Fatalf("Linked id of %v not set correctly for %d since %d != %d",
					cn, i, cn.LinkIds[i], m3path.DeadEndId)
			}
			dbLinkIds[i].Valid = false
			dbLinkIds[i].Int64 = 0
		}
	}
	return dbLinkIds
}

func (cn *ConnectionsStateDb) GetTrioDetails(pointData m3point.PointPackDataIfc) *m3point.TrioDetails {
	if cn.TrioDetails == nil {
		cn.TrioDetails = pointData.GetTrioDetails(cn.TrioId)
	}
	return cn.TrioDetails
}

func (cn *ConnectionsStateDb) SetTrioId(trioId m3point.TrioIndex) {
	cn.TrioId = trioId
	cn.TrioDetails = nil
}

func (cn *ConnectionsStateDb) SetTrioDetails(trioDetails *m3point.TrioDetails) {
	cn.TrioId = trioDetails.GetId()
	cn.TrioDetails = trioDetails
}

func (cn *ConnectionsStateDb) HasOpenConnections() bool {
	for i := 0; i < m3path.NbConnections; i++ {
		if cn.GetConnectionState(i) == m3path.ConnectionNotSet {
			return true
		}
	}
	return false
}

func (cn *ConnectionsStateDb) IsFrom(connIdx int) bool {
	return cn.GetConnectionState(connIdx) == m3path.ConnectionFrom
}

func (cn *ConnectionsStateDb) IsNext(connIdx int) bool {
	return cn.GetConnectionState(connIdx) == m3path.ConnectionNext
}

func (cn *ConnectionsStateDb) IsDeadEnd(connIdx int) bool {
	return cn.GetConnectionState(connIdx) == m3path.ConnectionBlocked
}

/***************************************************************/
// PathNodeDb Functions
/***************************************************************/

// Should be used only inside getNewPathNodeDb() and release() methods
var pathNodeDbPool = sync.Pool{
	New: func() interface{} {
		return new(PathNodeDb)
	},
}

func getNewPathNodeDb() *PathNodeDb {
	pn := pathNodeDbPool.Get().(*PathNodeDb)
	// Make sure all id are negative and pointer nil
	pn.setToNil(m3path.NewPathNodeId)
	return pn
}

func (pn *PathNodeDb) release() {
	// Cannot release a root node
	if pn.id > 0 && pn.d == 0 {
		return
	}
	if pn.pathCtx != nil && pn.pathCtx.currentNodeBuilder != nil && pn.pathCtx.currentNodeBuilder.hasPathNode(pn) {
		return
	}
	// Make sure it's clean before resending to pool
	pn.setToNil(m3path.InPoolId)
	pathNodeDbPool.Put(pn)
}

func (pn *PathNodeDb) setToNil(id m3point.Int64Id) {
	if id == m3path.InPoolId {
		pn.state = InPoolNode
	} else {
		pn.state = NewPathNode
	}
	pn.id = m3path.PathNodeId(id)
	pn.pathCtxId = -1
	pn.pathCtx = nil
	pn.pathBuilderId = -1
	pn.pathBuilder = nil
	pn.pathPoint = m3path.NilPathPoint
	pn.d = -1
	pn.SetConnStateToNil()
	pn.clearLinkNodes()
}

func (pn *PathNodeDb) IsNew() bool {
	return pn.id == m3path.PathNodeId(m3path.NewPathNodeId)
}

func (pn *PathNodeDb) IsInPool() bool {
	return pn.id == m3path.PathNodeId(m3path.InPoolId)
}
func (pn *PathNodeDb) syncInDb() error {
	switch pn.state {
	case InPoolNode:
		return m3util.MakeQsmErrorf("trying to save path node from Pool!")
	case InConflictNode:
		return m3util.MakeQsmErrorf("trying to save path node %q that is in conflict! Use the other one.", pn.String())
	case NewPathNode:
		// Fetch Ids of next path nodes already synced in DB
		for i := 0; i < m3path.NbConnections; i++ {
			if pn.linkNodes[i] != nil && pn.linkNodes[i].state != SyncInDbPathNode && pn.LinkIds[i] == m3path.NextLinkIdNotAssigned {
				// The next node was sync in DB using the id
				pn.LinkIds[i] = m3point.Int64Id(pn.linkNodes[i].id)
			}
		}
		filtered, err := pn.insertInDb()
		if err != nil {
			if filtered {
				pn.state = InConflictNode
				return nil
			} else {
				return m3util.MakeWrapQsmErrorf(err, "Could not save path node %q due to %v", pn.String(), err)
			}
		} else {
			pn.state = SyncInDbPathNode
			return nil
		}
	case SyncInDbPathNode:
		// Already sync all good
		if pn.id <= 0 {
			return m3util.MakeQsmErrorf("Path node %s supposed to be DB synced but id=%d", pn.String(), pn.id)
		}
		return nil
	case ModifiedNode:
		// Fetch Ids of next path nodes already synced in DB
		for i := 0; i < m3path.NbConnections; i++ {
			if pn.linkNodes[i] != nil && pn.linkNodes[i].state != SyncInDbPathNode && pn.LinkIds[i] == m3path.NextLinkIdNotAssigned {
				// The next node was sync in DB using the id
				pn.LinkIds[i] = m3point.Int64Id(pn.linkNodes[i].id)
			}
		}
		return pn.updateInDb()
	}
	return m3util.MakeQsmErrorf("Path node %s has unknown state=%d", pn.String(), pn.state)
}

func (pn *PathNodeDb) insertInDb() (bool, error) {
	if pn.pathPoint == m3path.NilPathPoint {
		return false, m3util.MakeQsmErrorf("cannot sync in DB path node %s with no point info", pn.String())
	}
	te := pn.pathCtx.pathNodesTe()
	pathNodeIds := pn.GetLinkIdsForDb()
	id, err := te.InsertReturnId(pn.pathCtxId, pn.pathBuilderId, pn.pathBuilderIdx, pn.TrioId, pn.pathPoint.Id, pn.d,
		pn.ConnectionMask,
		pathNodeIds[0], pathNodeIds[1], pathNodeIds[2])
	if err != nil {
		return te.IsFiltered(err), m3util.MakeWrapQsmErrorf(err, "insert in DB %s failed with %v", pn.String(), err)
	}
	pn.id = m3path.PathNodeId(id)
	pn.state = SyncInDbPathNode
	return false, nil
}

func (pn *PathNodeDb) updateInDb() error {
	pathNodeIds := pn.GetLinkIdsForDb()
	updatedRows, err := pn.pathCtx.pathNodesTe().Update(UpdatePathNode, pn.id,
		pn.ConnectionMask,
		pathNodeIds[0], pathNodeIds[1], pathNodeIds[2])
	if err != nil {
		return err
	}
	if updatedRows != 1 {
		return m3util.MakeQsmErrorf("updating path node id %d did not return 1 row but %d in %s", pn.id, updatedRows, pn.String())
	}
	pn.state = SyncInDbPathNode
	return nil
}

func createPathNodeFromDbRows(rows *sql.Rows) (*PathNodeDb, error) {
	pn := getNewPathNodeDb()
	pp := m3path.PathPoint{}
	pathNodeIds := [m3path.NbConnections]sql.NullInt64{}
	err := rows.Scan(&pn.id, &pn.pathCtxId, &pn.pathBuilderId, &pn.pathBuilderIdx, &pn.TrioId, &pp.Id, &pn.d,
		&pn.ConnectionMask,
		&pathNodeIds[0], &pathNodeIds[1], &pathNodeIds[2],
		&pp.P[0], &pp.P[1], &pp.P[2])
	if err != nil {
		pn.release()
		return nil, err
	}
	pn.clearLinkNodes()
	pn.SetLinkIdsFromDbData(pathNodeIds)
	pn.pathPoint = pp
	pn.state = SyncInDbPathNode
	return pn, nil
}

func createPathNodeFromDbRow(row *sql.Row) (*PathNodeDb, error) {
	pn := getNewPathNodeDb()
	pp := m3path.PathPoint{}
	pathNodeIds := [m3path.NbConnections]sql.NullInt64{}
	err := row.Scan(&pn.id, &pn.pathCtxId, &pn.pathBuilderId, &pn.pathBuilderIdx, &pn.TrioId, &pp.Id, &pn.d,
		&pn.ConnectionMask,
		&pathNodeIds[0], &pathNodeIds[1], &pathNodeIds[2])
	if err != nil {
		pn.release()
		return nil, err
	}
	pn.clearLinkNodes()
	pn.SetLinkIdsFromDbData(pathNodeIds)
	pn.state = SyncInDbPathNode
	return pn, nil
}

func (pn *PathNodeDb) PathCtx() *PathContextDb {
	if pn.pathCtx == nil {
		Log.Fatalf("the path context should always been initialized before for %s", pn.String())
	}
	return pn.pathCtx
}

func (pn *PathNodeDb) SetPathCtx(pathCtx *PathContextDb) {
	if pathCtx == nil {
		Log.Fatalf("cannot set a nil path context on %s", pn.String())
		return
	}
	if pn.pathCtxId != -1 && pn.pathCtxId != pathCtx.id {
		Log.Fatalf("trying to set the path context %d on %s which already has a different one!", pathCtx.id, pn.String())
		return
	}
	pn.pathCtxId = pathCtx.id
	pn.pathCtx = pathCtx
}

func (pn *PathNodeDb) PathBuilder() pointdb.PathNodeBuilder {
	if pn.pathBuilder == nil {
		rootPathBuilder := pn.PathCtx().pointData.GetRootPathNodeBuilderById(pn.pathBuilderId)
		// Find in all the linked path builder the one matching
		pn.pathBuilder = rootPathBuilder.GetPathBuilderByIndex(pn.pathBuilderIdx)
	}
	return pn.pathBuilder
}

func (pn *PathNodeDb) SetPathBuilder(pathBuilder pointdb.PathNodeBuilder) {
	pnIdx := -1
	for i := 0; i < pointdb.NbPathBuildersPerContext; i++ {
		if pathBuilder == pathBuilder.GetPathBuilderByIndex(i) {
			pnIdx = i
			break
		}
	}
	if pnIdx < 0 {
		Log.Fatalf("Did not find path builder %s in its own context", pathBuilder.String())
	}
	pn.pathBuilderId = pathBuilder.GetCubeId()
	pn.pathBuilderIdx = pnIdx
	pn.pathBuilder = pathBuilder
}

func (pn *PathNodeDb) String() string {
	return fmt.Sprintf("PNDB%d-%d-%d-%d-%d", pn.id, pn.pathCtxId, pn.pathPoint.Id, pn.d, pn.TrioId)
}

func (pn *PathNodeDb) GetPathContext() m3path.PathContext {
	pn.check()
	return pn.pathCtx
}

func (pn *PathNodeDb) IsRoot() bool {
	pn.check()
	return pn.d == 0
}

func (pn *PathNodeDb) setDeadEnd(connIdx int) {
	pn.check()
	pn.SetConnectionState(connIdx, m3path.ConnectionBlocked)
	pn.LinkIds[connIdx] = m3path.DeadEndId
	pn.linkNodes[connIdx] = nil
	if pn.state == SyncInDbPathNode {
		pn.state = ModifiedNode
	}
}

func (pn *PathNodeDb) check() {
	if pn.IsInPool() {
		Log.Fatalf("Cannot use in pool path node for %s", pn.String())
	}
}

func (pn *PathNodeDb) GetId() m3path.PathNodeId {
	return pn.id
}

func (pn *PathNodeDb) P() m3point.Point {
	pn.check()
	return pn.pathPoint.P
}

func (pn *PathNodeDb) D() int {
	pn.check()
	return pn.d
}

func (pn *PathNodeDb) GetNext(connIdx int) m3path.PathNodeId {
	if pn.GetConnectionState(connIdx) == m3path.ConnectionNext {
		return m3path.PathNodeId(pn.LinkIds[connIdx])
	}
	return m3path.PathNodeId(m3path.LinkIdNotSet)
}

func (pn *PathNodeDb) GetNextConnection(connId m3point.ConnectionId) m3path.PathNodeId {
	td := pn.GetTrioDetails(pn.pathCtx.pointData)
	for i, cd := range td.GetConnections() {
		if cd.GetId() == connId {
			if pn.GetConnectionState(i) != m3path.ConnectionNext {
				Log.Errorf("asked to retrieve next connection for %s on %s but it is a next conn", pn.String(), connId.String())
				return m3path.PathNodeId(m3path.LinkIdNotSet)
			}
			return m3path.PathNodeId(pn.LinkIds[i])
		}
	}
	return m3path.PathNodeId(m3path.LinkIdNotSet)
}

func (pn *PathNodeDb) setFrom(connId m3point.ConnectionId, fromNode *PathNodeDb) error {
	td := pn.GetTrioDetails(pn.pathCtx.pointData)
	for i, cd := range td.GetConnections() {
		if cd.GetId() == connId {
			if pn.GetConnectionState(i) == m3path.ConnectionNotSet {
				if Log.IsTrace() {
					Log.Tracef("set from %s on node %s at conn %s %d.", fromNode.String(), pn.String(), connId, i)
				}
				pn.SetConnectionState(i, m3path.ConnectionFrom)
				pn.LinkIds[i] = m3point.Int64Id(fromNode.id)
				if pn.state == SyncInDbPathNode {
					pn.state = ModifiedNode
				}
				return nil
			} else {
				if Log.IsDebug() {
					Log.Debugf("Cannot set from %s on node %s at conn %s %d, since it is already %d to %d.", fromNode.String(), pn.String(), connId, i, pn.GetConnectionState(i), pn.LinkIds[i])
				}
				// TODO: This is very expensive and happens a lot =>
				return MakeQsmModelErrorf(ConnectionNotAvailable, "Connection %s not available on %d", connId.String(), pn.pathPoint.Id)
			}
		}
	}
	err := MakeQsmModelErrorf(ConnectionNotFound, "Could not set from on path node %s since connId %s does not exists in %s ", pn.String(), connId.String(), td.String())
	Log.Error(err)
	return err
}

func (pn *PathNodeDb) clearLinkNodes() {
	for i := 0; i < m3path.NbConnections; i++ {
		// Always Nullify actual node pointers when loading from DB
		pn.linkNodes[i] = nil
	}
}

/***************************************************************/
// ErrorType Functions
/***************************************************************/

type ErrorType int

const (
	ConnectionNotFound ErrorType = iota
	ConnectionNotAvailable
)

type QsmModelError struct {
	errType ErrorType
	msg     string
}

func (err *QsmModelError) Error() string {
	return err.msg
}

func MakeQsmModelErrorf(errType ErrorType, format string, args ...interface{}) *QsmModelError {
	return &QsmModelError{errType, fmt.Sprintf(format, args...)}
}
