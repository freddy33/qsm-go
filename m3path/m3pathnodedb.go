package m3path

import (
	"database/sql"
	"fmt"
	"github.com/freddy33/qsm-go/m3db"
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

type PathNodeDb struct {
	id int64

	pathCtxId int
	pathCtx   *PathContextDb

	pathBuilderId int
	pathBuilder   m3point.PathNodeBuilder

	trioId      m3point.TrioIndex
	trioDetails *m3point.TrioDetails

	pointId int64
	point   *m3point.Point

	d     int
	links [3]PathLinkDb
}

type PathLinkDb struct {
	node  *PathNodeDb
	index int

	connState ConnectionState

	linkedNodeId int64
	linkedNode   *PathNodeDb
}

var pathNodeDbPool = sync.Pool{
	New: func() interface{} {
		return new(PathNodeDb)
	},
}

func getNewPathNodeDb() *PathNodeDb {
	pn := pathNodeDbPool.Get().(*PathNodeDb)
	// Make sure all id are -1 and pointer nil
	pn.setToNil()
	return pn
}

func releasePathNodeDb(pn *PathNodeDb) {
	// Make sure it's clean before resending to pool
	pn.setToNil()
	pathNodeDbPool.Put(pn)
}

// Set a p[ath node to nil empty state
func (pn *PathNodeDb) setToNil() {
	pn.id = -1
	pn.pathCtxId = -1
	pn.pathCtx = nil
	pn.pathBuilderId = -1
	pn.pathBuilder = nil
	pn.trioId = m3point.NilTrioIndex
	pn.trioDetails = nil
	pn.pointId = -1
	pn.point = nil
	pn.d = -1
	for i, pl := range pn.links {
		pl.node = pn
		pl.index = i
		pl.connState = ConnectionNoSet
		pl.linkedNodeId = -1
		pl.linkedNode = nil
	}
}

func (pn *PathNodeDb) getConnectionsForDb() (uint16, [3]sql.NullInt64, [3]sql.NullInt64) {
	blockedMask := uint16(0)
	from := [3]sql.NullInt64{}
	next := [3]sql.NullInt64{}
	for i, pl := range pn.links {
		switch pl.connState {
		case ConnectionNoSet:
			from[i].Valid = false
			from[i].Int64 = 0
			next[i].Valid = false
			next[i].Int64 = 0
		case ConnectionFrom:
			from[i].Valid = true
			from[i].Int64 = pl.linkedNodeId
			next[i].Valid = false
			next[i].Int64 = 0
		case ConnectionNext:
			from[i].Valid = false
			from[i].Int64 = 0
			next[i].Valid = true
			next[i].Int64 = pl.linkedNodeId
		case ConnectionBlocked:
			from[i].Valid = false
			from[i].Int64 = 0
			next[i].Valid = false
			next[i].Int64 = 0
			blockedMask |= 1 << uint16(i)
		}
	}
	return blockedMask, from, next
}

func (pn *PathNodeDb) setBlockedMask(blockedMask uint16) {
	for i, pl := range pn.links {
		if blockedMask & (1 << uint16(i)) != 0 {
			pl.connState = ConnectionBlocked
			pl.linkedNodeId = -1
			pl.linkedNode = nil
		}
	}
}

func (pn *PathNodeDb) insertInDb() error {
	te, err := GetPathEnv().GetOrCreateTableExec(PathNodesTable)
	if err != nil {
		return err
	}
	blockedMask, from, next := pn.getConnectionsForDb()
	pn.id, err = te.InsertReturnId(pn.pathCtxId, pn.pathBuilderId, pn.trioId, pn.pointId, pn.d,
		blockedMask,
		from[0], from[1], from[2],
		next[0], next[1], next[2])
	return err
}

func getOrCreatePathNodeDb(pn *PathNodeDb) (bool, int64) {
	return getOrCreatePathNodeDbEnv(GetPathEnv(), pn)
}

func getPathNodeDb(id int64) *PathNodeDb {
	return getPathNodeDbEnv(GetPathEnv(), id)
}

func getPathNodeDbByPointEnv(env *m3db.QsmEnvironment, pointId int64) []*PathNodeDb {
	return nil
}

func getPathNodeDbEnv(env *m3db.QsmEnvironment, pathNodeId int64) *PathNodeDb {
	te, err := env.GetOrCreateTableExec(PathNodesTable)
	if err != nil {
		Log.Errorf("could not get path node table exec due to %v", err)
		return nil
	}
	rows, err := te.Query(SelectPathNodesById, pathNodeId)
	if err != nil {
		Log.Errorf("could not select path node table for id %d exec due to %v", pathNodeId, err)
		return nil
	}
	defer te.CloseRows(rows)
	if rows.Next() {
		pn, err := readRowOnlyIds(rows)
		if err != nil {
			Log.Errorf("Could not read row of %s due to %v", PathNodesTable, err)
		}
		return pn
	}
	return nil
}

func readRowOnlyIds(rows *sql.Rows) (*PathNodeDb, error) {
	pn := getNewPathNodeDb()
	blockedMask := uint16(0)
	from := [3]sql.NullInt64{}
	next := [3]sql.NullInt64{}
	err := rows.Scan(&pn.id, &pn.pathCtxId, &pn.pathBuilderId, &pn.trioId, &pn.pointId, &pn.d,
		&blockedMask,
		&from[0], &from[1], &from[2],
		&next[0], &next[1], &next[2])
	if err != nil {
		return nil, err
	}
	pn.setBlockedMask(blockedMask)
	for i, pl := range pn.links {
		if from[i].Valid && next[i].Valid {
			return nil, m3db.MakeQsmErrorf("Node DB entry for %d is invalid! link %d is both from and next", pn.id, i)
		}
		if from[i].Valid {
			pl.connState = ConnectionFrom
			pl.linkedNodeId = from[i].Int64
		} else if next[i].Valid {
			pl.connState = ConnectionNext
			pl.linkedNodeId = next[i].Int64
		}
	}
	return pn, nil
}

func getOrCreatePathNodeDbEnv(env *m3db.QsmEnvironment, pn *PathNodeDb) (bool, int64) {
	return false, -1
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

func (pn *PathNodeDb) TrioDetails() *m3point.TrioDetails {
	if pn.trioDetails == nil {
		pn.trioDetails = m3point.GetTrioDetails(pn.trioId)
	}
	return pn.trioDetails
}

func (pn *PathNodeDb) SetTrioDetails(trioDetails *m3point.TrioDetails) {
	pn.trioId = trioDetails.GetId()
	pn.trioDetails = trioDetails
}

func (pn *PathNodeDb) String() string {
	return fmt.Sprintf("PNDB%d-%d-%d-%d-%d", pn.id, pn.pathCtxId, pn.pointId, pn.d, pn.trioId)
}

func (pn *PathNodeDb) GetPathContext() PathContext {
	return pn.pathCtx
}

func (pn *PathNodeDb) IsEnd() bool {
	return pn.id <= 0
}

func (pn *PathNodeDb) IsRoot() bool {
	return pn.d == 0
}

func (pn *PathNodeDb) IsLatest() bool {
	onb := pn.PathCtx().openNodeBuilder
	if onb == nil {
		Log.Errorf("asking for latest flag on non initialize path context %s for %s", pn.pathCtx.String(), pn.String())
		return false
	}
	return pn.d >= onb.d
}

func (pn *PathNodeDb) P() m3point.Point {
	if pn.point == nil {
		pn.point = getPoint(pn.pointId)
	}
	return *pn.point
}

func (pn *PathNodeDb) D() int {
	return pn.d
}

func (pn *PathNodeDb) GetTrioIndex() m3point.TrioIndex {
	return pn.trioId
}

func (pn *PathNodeDb) GetFrom() PathLink {
	for _, pl := range pn.links {
		if pl.connState == ConnectionFrom {
			return pl
		}
	}
	return nil
}

func (pn *PathNodeDb) GetOtherFrom() PathLink {
	firstFound := false
	for _, pl := range pn.links {
		if pl.connState == ConnectionFrom {
			if firstFound {
				return pl
			}
			firstFound = true
		}
	}
	return nil
}

func (pn *PathNodeDb) GetNext(i int) PathLink {
	count := 0
	for _, pl := range pn.links {
		if pl.connState == ConnectionNext {
			if count == i {
				return pl
			}
			count++
		}
	}
	return nil
}

func (pn *PathNodeDb) GetNextConnection(connId m3point.ConnectionId) PathLink {
	td := pn.TrioDetails()
	for i, cd := range td.GetConnections() {
		if cd.GetId() == connId && pn.links[i].connState == ConnectionNext {
			return pn.links[i]
		}
	}
	return nil
}

func (pn *PathNodeDb) calcDist() int {
	from := pn.GetFrom()
	if from == nil {
		return 0
	}
	return from.GetSrc().calcDist() + 1
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

func (pl PathLinkDb) String() string {
	return fmt.Sprintf("PLDB-%d-%d", pl.connState, pl.index)
}

func (pl PathLinkDb) GetSrc() PathNode {
	if pl.connState == ConnectionFrom {
		if pl.linkedNode == nil {
			pl.linkedNode = getPathNodeDb(pl.linkedNodeId)
		}
		return pl.linkedNode
	}
	if pl.connState == ConnectionNext {
		return pl.node
	}
	return nil
}

func (pl PathLinkDb) GetDst() PathNode {
	if pl.connState == ConnectionNext {
		if pl.linkedNode == nil {
			pl.linkedNode = getPathNodeDb(pl.linkedNodeId)
		}
		return pl.linkedNode
	}
	if pl.connState == ConnectionFrom {
		return pl.node
	}
	return nil
}

func (pl PathLinkDb) GetConnId() m3point.ConnectionId {
	panic("implement me")
}

func (pl PathLinkDb) HasDestination() bool {
	panic("implement me")
}

func (pl PathLinkDb) IsDeadEnd() bool {
	panic("implement me")
}

func (pl PathLinkDb) SetDeadEnd() {
	panic("implement me")
}

func (pl PathLinkDb) createDstNode(pathBuilder m3point.PathNodeBuilder) (PathNode, bool, m3point.PathNodeBuilder) {
	Log.Fatalf("in DB path link %s never call createDstNode", pl.String())
	return nil, false, nil
}

func (pl PathLinkDb) dumpInfo(ident int) string {
	panic("implement me")
}
