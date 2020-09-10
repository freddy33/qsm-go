package pathdb

import (
	"database/sql"
	"fmt"
	"github.com/freddy33/qsm-go/backend/m3db"
	"github.com/freddy33/qsm-go/model/m3path"
	"github.com/freddy33/qsm-go/model/m3point"
	"sync"
)

type ConnectionState uint16

const (
	ConnectionMaskBits   = 4
	ConnectionStateMask  = uint16(0x0003)
	SingleConnectionMask = uint16(0x000F)
)
const (
	ConnectionNotSet  ConnectionState = 0x0000
	ConnectionFrom    ConnectionState = 0x0001
	ConnectionNext    ConnectionState = 0x0002
	ConnectionBlocked ConnectionState = 0x0003
	// Extra states possible as mask
)

const (
	NewPathNodeId         int64 = -1
	InPoolId              int64 = -2
	LinkIdNotSet          int64 = -3
	DeadEndId             int64 = -4
	NextLinkIdNotAssigned int64 = -5
)

type PathNodeDbState int16

const (
	NewPathNode PathNodeDbState = iota
	InPoolNode
	SyncInDbPathNode
	InConflictNode
	ModifiedNode
)

type PathNodeDb struct {
	state PathNodeDbState

	// In most cases this is already filled
	pathCtxId int
	// TODO: Create a map in ServerPathPackData for this
	pathCtx *PathContextDb

	// Just Ids will fill this only
	id      int64
	pointId int64
	d       int

	// Full Id loaded will fill this
	pathBuilderId  int
	trioId         m3point.TrioIndex
	connectionMask uint16
	linkNodeIds    [m3path.NbConnections]int64

	// This is dynamically loaded on demand from DB
	point *m3point.Point

	// This is dynamically loaded on demand from PointPackData
	pathBuilder m3point.PathNodeBuilder
	trioDetails *m3point.TrioDetails

	// This is populated during creation and should not be used for non new node
	linkNodes [m3path.NbConnections]*PathNodeDb
}

// Should be used only inside getNewPathNodeDb() and release() methods
var pathNodeDbPool = sync.Pool{
	New: func() interface{} {
		return new(PathNodeDb)
	},
}

func getNewPathNodeDb() *PathNodeDb {
	pn := pathNodeDbPool.Get().(*PathNodeDb)
	// Make sure all id are negative and pointer nil
	pn.setToNil(NewPathNodeId)
	return pn
}

func (pn *PathNodeDb) release() {
	// Make sure it's clean before resending to pool
	pn.setToNil(InPoolId)
	pathNodeDbPool.Put(pn)
}

func (pn *PathNodeDb) reduceSize() {
	pn.point = nil
	pn.pathBuilder = nil
	pn.trioDetails = nil
	for i := 0; i < m3path.NbConnections; i++ {
		pn.linkNodes[i] = nil
	}
}

// Set a p[ath node to nil empty state
func (pn *PathNodeDb) setToNil(id int64) {
	if id == InPoolId {
		pn.state = InPoolNode
	} else {
		pn.state = NewPathNode
	}
	pn.id = id
	pn.pathCtxId = -1
	pn.pathCtx = nil
	pn.pathBuilderId = -1
	pn.pathBuilder = nil
	pn.trioId = m3point.NilTrioIndex
	pn.trioDetails = nil
	pn.pointId = -1
	pn.point = nil
	pn.d = -1
	pn.connectionMask = uint16(ConnectionNotSet)
	for i := 0; i < m3path.NbConnections; i++ {
		pn.linkNodeIds[i] = LinkIdNotSet
		pn.linkNodes[i] = nil
	}
}

func (pn *PathNodeDb) IsNew() bool {
	return pn.id == NewPathNodeId
}

func (pn *PathNodeDb) IsInPool() bool {
	return pn.id == InPoolId
}

func (pn *PathNodeDb) getConnectionMaskValue(connIdx int) uint16 {
	return (pn.connectionMask >> uint16(connIdx*ConnectionMaskBits)) & SingleConnectionMask
}

func (pn *PathNodeDb) getConnectionState(connIdx int) ConnectionState {
	return ConnectionState(pn.getConnectionMaskValue(connIdx) & ConnectionStateMask)
}

func (pn *PathNodeDb) setConnectionMask(connIdx int, maskValue uint16) {
	allConnsMask := pn.connectionMask
	// Zero the bit mask for this connection
	allConnsMask &^= SingleConnectionMask << uint16(connIdx*ConnectionMaskBits)
	// Add the new mask value
	allConnsMask |= maskValue << uint16(connIdx*ConnectionMaskBits)
	pn.connectionMask = allConnsMask
	if pn.state == SyncInDbPathNode {
		pn.state = ModifiedNode
	}
}

func (pn *PathNodeDb) setConnectionState(connIdx int, state ConnectionState) {
	connMask := pn.getConnectionMaskValue(connIdx)
	// Zero what is not state mask bit
	connMask &^= ConnectionStateMask
	// Set the new state value
	connMask |= uint16(state)
	pn.setConnectionMask(connIdx, connMask)
}

func (pn *PathNodeDb) setPathIdsFromDbData(pathNodeIds [m3path.NbConnections]sql.NullInt64) {
	for i := 0; i < m3path.NbConnections; i++ {
		// Always Nullify actual node pointers when loading from DB
		pn.linkNodes[i] = nil
		switch pn.getConnectionState(i) {
		case ConnectionNotSet:
			if Log.DoAssert() {
				if pathNodeIds[i].Valid {
					Log.Errorf("Not set linked id of %s has wrong state in DB for %d since %v should be NULL",
						pn.String(), i, pathNodeIds[i])
				}
			}
			pn.linkNodeIds[i] = LinkIdNotSet
		case ConnectionFrom:
			if !pathNodeIds[i].Valid {
				Log.Errorf("Linked id of %s has wrong state in DB for %d since %v should be linked",
					pn.String(), i, pathNodeIds[i])
			}
			pn.linkNodeIds[i] = pathNodeIds[i].Int64
		case ConnectionNext:
			if pathNodeIds[i].Valid {
				pn.linkNodeIds[i] = pathNodeIds[i].Int64
			} else {
				pn.linkNodeIds[i] = NextLinkIdNotAssigned
			}
		case ConnectionBlocked:
			if Log.DoAssert() {
				if pathNodeIds[i].Valid {
					Log.Errorf("Blocked linked id of %s has wrong state in DB for %d since %v should be NULL",
						pn.String(), i, pathNodeIds[i])
				}
			}
			pn.linkNodeIds[i] = DeadEndId
		}
	}
}

func (pn *PathNodeDb) getConnsDataForDb() [m3path.NbConnections]sql.NullInt64 {
	pathNodeIds := [m3path.NbConnections]sql.NullInt64{}
	for i := 0; i < m3path.NbConnections; i++ {
		switch pn.getConnectionState(i) {
		case ConnectionNotSet:
			if Log.DoAssert() {
				if pn.linkNodeIds[i] != LinkIdNotSet {
					Log.Errorf("Linked id of %s not set correctly for %d since %d != %d",
						pn.String(), i, pn.linkNodeIds[i], LinkIdNotSet)
				}
				/*
					if pn.linkNodeIds[i] != LinkIdNotSet && pn.linkNodeIds[i] != NextLinkIdNotAssigned {
						Log.Errorf("Linked id of %s not set correctly for %d since %d not in ( %d , %d ) ",
							pn.String(), i, pn.linkNodeIds[i], LinkIdNotSet, NextLinkIdNotAssigned)
					}
				*/
			}
			pathNodeIds[i].Valid = false
			pathNodeIds[i].Int64 = 0
		case ConnectionFrom:
			if pn.linkNodeIds[i] <= 0 {
				Log.Errorf("Linked id for from of %s not set correctly for %d since %d <= 0",
					pn.String(), i, pn.linkNodeIds[i])
			}
			pathNodeIds[i].Valid = true
			pathNodeIds[i].Int64 = pn.linkNodeIds[i]
		case ConnectionNext:
			if pn.linkNodeIds[i] == NextLinkIdNotAssigned {
				pathNodeIds[i].Valid = false
				pathNodeIds[i].Int64 = 0
			} else if pn.linkNodeIds[i] > 0 {
				pathNodeIds[i].Valid = true
				pathNodeIds[i].Int64 = pn.linkNodeIds[i]
			} else {
				Log.Errorf("Linked id for next of %s not set correctly for %d since %d != %d && %d <= 0",
					pn.String(), i, pn.linkNodeIds[i], NextLinkIdNotAssigned, pn.linkNodeIds[i])
			}
		case ConnectionBlocked:
			if Log.DoAssert() {
				if pn.linkNodeIds[i] != DeadEndId {
					Log.Errorf("Linked id of %s not set correctly for %d since %d != %d",
						pn.String(), i, pn.linkNodeIds[i], DeadEndId)
				}
			}
			pathNodeIds[i].Valid = false
			pathNodeIds[i].Int64 = 0
		}
	}
	return pathNodeIds
}

func (pn *PathNodeDb) syncInDb() error {
	switch pn.state {
	case InPoolNode:
		return m3db.MakeQsmErrorf("trying to save path node from Pool!")
	case InConflictNode:
		return m3db.MakeQsmErrorf("trying to save path node %s that is in conflict! Use the other one.", pn.String())
	case NewPathNode:
		// Fetch Ids of next path nodes already synced in DB
		for i := 0; i < m3path.NbConnections; i++ {
			if pn.linkNodes[i] != nil && pn.linkNodes[i].state != SyncInDbPathNode && pn.linkNodeIds[i] == NextLinkIdNotAssigned {
				// The next node was sync in DB using the id
				pn.linkNodeIds[i] = pn.linkNodes[i].id
			}
		}
		if pn.pointId <= 0 {
			if pn.point == nil {
				return m3db.MakeQsmErrorf("cannot sync in DB path node %s with no point info", pn.String())
			}
			pn.pointId = getOrCreatePointTe(pn.PathCtx().pointsTe(), *pn.point)
			if pn.pointId <= 0 {
				return m3db.MakeQsmErrorf("cannot sync in DB path node %s while point insertion %v failed", pn.String(), *pn.point)
			}
		}
		err, filtered := pn.insertInDb()
		if err != nil {
			if filtered {
				pn.state = InConflictNode
				return nil
			} else {
				return m3db.MakeQsmErrorf("Could not save path node %s due to '%s'", pn.String(), err.Error())
			}
		} else {
			pn.state = SyncInDbPathNode
			return nil
		}
	case SyncInDbPathNode:
		// Already sync all good
		if pn.id <= 0 {
			return m3db.MakeQsmErrorf("Path node %s supposed to be DB synced but id=%d", pn.String(), pn.id)
		}
		return nil
	case ModifiedNode:
		// Fetch Ids of next path nodes already synced in DB
		for i := 0; i < m3path.NbConnections; i++ {
			if pn.linkNodes[i] != nil && pn.linkNodes[i].state != SyncInDbPathNode && pn.linkNodeIds[i] == NextLinkIdNotAssigned {
				// The next node was sync in DB using the id
				pn.linkNodeIds[i] = pn.linkNodes[i].id
			}
		}
		return pn.updateInDb()
	}
	return m3db.MakeQsmErrorf("Path node %s has unknown state=%d", pn.String(), pn.state)
}

func (pn *PathNodeDb) insertInDb() (error, bool) {
	if pn.pointId < 0 {
		return m3db.MakeQsmErrorf("cannot insert in DB %s since the point was not inserted", pn.String()), false
	}
	te := pn.pathCtx.pathNodesTe()
	pathNodeIds := pn.getConnsDataForDb()
	var err error
	pn.id, err = te.InsertReturnId(pn.pathCtxId, pn.pathBuilderId, pn.trioId, pn.pointId, pn.d,
		pn.connectionMask,
		pathNodeIds[0], pathNodeIds[1], pathNodeIds[2])
	if err == nil {
		pn.state = SyncInDbPathNode
	}
	return err, te.IsFiltered(err)
}

func (pn *PathNodeDb) updateInDb() error {
	pathNodeIds := pn.getConnsDataForDb()
	updatedRows, err := pn.pathCtx.pathNodesTe().Update(UpdatePathNode, pn.id,
		pn.connectionMask,
		pathNodeIds[0], pathNodeIds[1], pathNodeIds[2])
	if err != nil {
		return err
	}
	if updatedRows != 1 {
		return m3db.MakeQsmErrorf("updating path node id %d did not return 1 row but %d in %s", pn.id, updatedRows, pn.String())
	}
	pn.state = SyncInDbPathNode
	return nil
}

func fetchDbRow(rows *sql.Rows) (*PathNodeDb, error) {
	pn := getNewPathNodeDb()
	pathNodeIds := [m3path.NbConnections]sql.NullInt64{}
	err := rows.Scan(&pn.id, &pn.pathCtxId, &pn.pathBuilderId, &pn.trioId, &pn.pointId, &pn.d,
		&pn.connectionMask,
		&pathNodeIds[0], &pathNodeIds[1], &pathNodeIds[2])
	if err != nil {
		pn.release()
		return nil, err
	}
	pn.setPathIdsFromDbData(pathNodeIds)
	pn.state = SyncInDbPathNode
	return pn, nil
}

func fetchSingleDbRow(row *sql.Row) (*PathNodeDb, error) {
	pn := getNewPathNodeDb()
	pathNodeIds := [m3path.NbConnections]sql.NullInt64{}
	err := row.Scan(&pn.id, &pn.pathCtxId, &pn.pathBuilderId, &pn.trioId, &pn.pointId, &pn.d,
		&pn.connectionMask,
		&pathNodeIds[0], &pathNodeIds[1], &pathNodeIds[2])
	if err != nil {
		pn.release()
		return nil, err
	}
	pn.setPathIdsFromDbData(pathNodeIds)
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

func (pn *PathNodeDb) GetTrioDetails() *m3point.TrioDetails {
	if pn.trioDetails == nil {
		pn.trioDetails = pn.PathCtx().ppd.GetTrioDetails(pn.trioId)
	}
	return pn.trioDetails
}

func (pn *PathNodeDb) SetTrioId(trioId m3point.TrioIndex) {
	pn.trioId = trioId
	pn.trioDetails = nil
}

func (pn *PathNodeDb) SetTrioDetails(trioDetails *m3point.TrioDetails) {
	pn.trioId = trioDetails.GetId()
	pn.trioDetails = trioDetails
}

func (pn *PathNodeDb) PathBuilder() m3point.PathNodeBuilder {
	if pn.pathBuilder == nil {
		pn.pathBuilder = pn.PathCtx().ppd.GetPathNodeBuilderById(pn.pathBuilderId)
	}
	return pn.pathBuilder
}

func (pn *PathNodeDb) SetPathBuilder(pathBuilder m3point.PathNodeBuilder) {
	pn.pathBuilderId = pathBuilder.GetCubeId()
	pn.pathBuilder = pathBuilder
}

func (pn *PathNodeDb) String() string {
	return fmt.Sprintf("PNDB%d-%d-%d-%d-%d", pn.id, pn.pathCtxId, pn.pointId, pn.d, pn.trioId)
}

func (pn *PathNodeDb) GetPathContext() m3path.PathContext {
	pn.check()
	return pn.pathCtx
}

func (pn *PathNodeDb) IsRoot() bool {
	pn.check()
	return pn.d == 0
}

func (pn *PathNodeDb) IsLatest() bool {
	pn.check()
	onb := pn.PathCtx().openNodeBuilder
	if onb == nil {
		Log.Errorf("asking for latest flag on non initialize path context %s for %s", pn.pathCtx.String(), pn.String())
		return false
	}
	return pn.d >= onb.d
}

func (pn *PathNodeDb) HasOpenConnections() bool {
	pn.check()
	for i := 0; i < m3path.NbConnections; i++ {
		if pn.getConnectionState(i) == ConnectionNotSet {
			return true
		}
	}
	return false
}

func (pn *PathNodeDb) IsFrom(connIdx int) bool {
	pn.check()
	return pn.getConnectionState(connIdx) == ConnectionFrom
}

func (pn *PathNodeDb) IsNext(connIdx int) bool {
	pn.check()
	return pn.getConnectionState(connIdx) == ConnectionNext
}

func (pn *PathNodeDb) IsDeadEnd(connIdx int) bool {
	pn.check()
	return pn.getConnectionState(connIdx) == ConnectionBlocked
}

func (pn *PathNodeDb) setDeadEnd(connIdx int) {
	pn.check()
	pn.setConnectionState(connIdx, ConnectionBlocked)
	pn.linkNodeIds[connIdx] = DeadEndId
	pn.linkNodes[connIdx] = nil
}

func (pn *PathNodeDb) check() {
	if pn.IsInPool() {
		Log.Fatalf("Cannot use in pool path node for %s", pn.String())
	}
}

func (pn *PathNodeDb) GetId() int64 {
	return pn.id
}

func (pn *PathNodeDb) P() m3point.Point {
	pn.check()
	if pn.point == nil {
		if pn.pointId <= 0 {
			Log.Fatalf("Cannot retrieve point not already set for %s", pn.String())
			return m3point.Origin
		}
		var err error
		pn.point, err = getPointEnv(pn.pathCtx.env, pn.pointId)
		if err != nil {
			Log.Fatal(err)
			return m3point.Origin
		}
	}
	return *pn.point
}

func (pn *PathNodeDb) D() int {
	pn.check()
	return pn.d
}

func (pn *PathNodeDb) GetTrioIndex() m3point.TrioIndex {
	return pn.trioId
}

func (pn *PathNodeDb) GetNext(connIdx int) int64 {
	if pn.getConnectionState(connIdx) == ConnectionNext {
		return pn.linkNodeIds[connIdx]
	}
	return LinkIdNotSet
}

func (pn *PathNodeDb) GetNextConnection(connId m3point.ConnectionId) int64 {
	td := pn.GetTrioDetails()
	for i, cd := range td.GetConnections() {
		if cd.GetId() == connId {
			if pn.getConnectionState(i) != ConnectionNext {
				Log.Errorf("asked to retrieve next connection for %s on %s but it is a next conn", pn.String(), connId.String())
				return LinkIdNotSet
			}
			return pn.linkNodeIds[i]
		}
	}
	return LinkIdNotSet
}

func (pn *PathNodeDb) setFrom(connId m3point.ConnectionId, fromNode *PathNodeDb) error {
	td := pn.GetTrioDetails()
	for i, cd := range td.GetConnections() {
		if cd.GetId() == connId {
			if pn.getConnectionState(i) == ConnectionNotSet {
				if Log.IsTrace() {
					Log.Tracef("set from %s on node %s at conn %s %d.", fromNode.String(), pn.String(), connId, i)
				}
				pn.setConnectionState(i, ConnectionFrom)
				pn.linkNodeIds[i] = fromNode.id
				if pn.state == SyncInDbPathNode {
					pn.state = ModifiedNode
				}
				return nil
			} else {
				if Log.IsDebug() {
					Log.Debugf("Cannot set from %s on node %s at conn %s %d, since it is already %d to %d.", fromNode.String(), pn.String(), connId, i, pn.getConnectionState(i), pn.linkNodeIds[i])
				}
				// TODO: This is very expensive and happens a lot =>
				return MakeQsmModelErrorf(ConnectionNotAvailable, "Connection %s not available on %d", connId.String(), pn.pointId)
			}
		}
	}
	err := MakeQsmModelErrorf(ConnectionNotFound, "Could not set from on path node %s since connId %s does not exists in %s ", pn.String(), connId.String(), td.String())
	Log.Error(err)
	return err
}

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
