package m3path

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3point"
	"github.com/freddy33/qsm-go/m3util"
	"strings"
)

var Log = m3util.NewLogger("m3path", m3util.INFO)

type PathContextIfc interface {
	fmt.Stringer
	GetTrioCtx() *m3point.TrioContext
	GetOffset() int
	GetTrioContextType() m3point.ContextType
	GetTrioContextIndex() int
	GetPathNodeMap() PathNodeMap
	InitRootNode(center m3point.Point)
	GetRootPathNode() PathNode
	GetNumberOfOpenNodes() int
	GetAllOpenPathNodes() []PathNode
	MoveToNextNodes()
	PredictedNextOpenNodesLen() int
	dumpInfo() string
}

type BasePathContext struct {
	ctx    *m3point.TrioContext
	offset int

	rootPathNode PathNode
	openEndNodes []OpenEndPath

	pathNodeMap PathNodeMap
}

// Struct left at the end of a path builder where next round of building should be done
type OpenEndPath struct {
	// The path node with trio index and next left to be build
	pn PathNode
	// The path builder element used to build the path node above
	pnb m3point.PathNodeBuilder
}

type PathLink interface {
	fmt.Stringer
	GetSrc() PathNode
	GetConnId() m3point.ConnectionId
	HasDestination() bool
	IsDeadEnd() bool
	SetDeadEnd()
	createDstNode(pathBuilder m3point.PathNodeBuilder) (PathNode, bool, m3point.PathNodeBuilder)
	dumpInfo(ident int) string
}

// A single path link between *src* node to one of the next path node *dst* using the connection Id
type BasePathLink struct {
	// After travelling the connId of the above cur.connId there will be 2 new path possible for
	src PathNode
	// The connection used by the link path
	connId m3point.ConnectionId
	// After travelling the connId the pointer to the next path node
	// point to EndPathNode if this link is a dead end
	dst PathNode
}

type PathNode interface {
	fmt.Stringer
	GetPathContext() *BasePathContext
	IsEnd() bool
	IsRoot() bool
	IsLatest() bool
	P() m3point.Point
	D() int
	GetTrioIndex() m3point.TrioIndex
	GetFrom() PathLink
	GetOtherFrom() PathLink
	GetNext(i int) PathLink
	GetNextConnection(connId m3point.ConnectionId) PathLink

	calcDist() int
	addPathLink(connId m3point.ConnectionId) (PathLink, bool)
	setOtherFrom(pl PathLink)
	dumpInfo(ident int) string
}

type BasePathNode struct {
	// The path context the link belongs to
	pathCtx *BasePathContext
	// The point of this path node
	p m3point.Point
	// The current trio index of the path point
	trioId m3point.TrioIndex
}

// The link graph node of a path, representing one point on the graph
// Points to the 2 path links usable from here
type RootPathNode struct {
	BasePathNode
	// After travelling the connId of the above cur.connId there will be 2 new path possible for
	next [3]PathLink
}

// The link graph node of a path, representing one point on the graph
// Points to the 2 path links usable from here
type OutPathNode struct {
	BasePathNode
	// Distance from root
	d int
	// From which link this node came from
	from PathLink
	// If this node came from a combined link
	otherFrom PathLink
	// After travelling the connId of the above cur.connId there will be 2 new path possible for
	next [2]PathLink
}

type EndPathNodeT struct {
}

var EndPathNode = &EndPathNodeT{}
var EndPathLink = &BasePathLink{EndPathNode, m3point.NilConnectionId, EndPathNode }

/***************************************************************/
// BasePathContext Functions
/***************************************************************/

func MakePathContext(ctxType m3point.ContextType, pIdx int, offset int, pnm PathNodeMap) PathContextIfc {
	return MakePathContextFromTrioContext(m3point.GetTrioContextByTypeAndIdx(ctxType, pIdx), offset, pnm)
}

func MakePathContextFromTrioContext(trCtx *m3point.TrioContext, offset int, pnm PathNodeMap) PathContextIfc {
	pathCtx := BasePathContext{}
	pathCtx.ctx = trCtx
	pathCtx.offset = offset
	pathCtx.pathNodeMap = pnm

	return &pathCtx
}

func (pathCtx *BasePathContext) GetTrioCtx() *m3point.TrioContext {
	return pathCtx.ctx
}

func (pathCtx *BasePathContext) GetOffset() int {
	return pathCtx.offset
}

func (pathCtx *BasePathContext) GetTrioContextType() m3point.ContextType {
	return pathCtx.ctx.GetType()
}

func (pathCtx *BasePathContext) GetTrioContextIndex() int {
	return pathCtx.ctx.GetIndex()
}

func (pathCtx *BasePathContext) GetPathNodeMap() PathNodeMap {
	return pathCtx.pathNodeMap
}

func (pathCtx *BasePathContext) GetRootPathNode() PathNode {
	return pathCtx.rootPathNode
}

func (pathCtx *BasePathContext) GetNumberOfOpenNodes() int {
	return len(pathCtx.openEndNodes)
}

func (pathCtx *BasePathContext) GetAllOpenPathNodes() []PathNode {
	res := make([]PathNode, len(pathCtx.openEndNodes))
	for i, oep := range pathCtx.openEndNodes {
		res[i] = oep.pn
	}
	return res
}

func (pathCtx *BasePathContext) InitRootNode(center m3point.Point) {
	// the path builder enforce origin as the center
	nodeBuilder := m3point.GetPathNodeBuilder(pathCtx.ctx, pathCtx.offset, m3point.Origin)

	rootNode := RootPathNode{}
	rootNode.pathCtx = pathCtx
	// But the path node here points to real points in space
	rootNode.p = center
	rootNode.trioId = nodeBuilder.GetTrioIndex()

	pathCtx.rootPathNode = &rootNode
	oep := OpenEndPath{pathCtx.rootPathNode, nodeBuilder}
	pathCtx.openEndNodes = make([]OpenEndPath, 1)
	pathCtx.openEndNodes[0] = oep

	pathCtx.pathNodeMap.AddPathNode(pathCtx.rootPathNode)
}

func (pathCtx *BasePathContext) PredictedNextOpenNodesLen() int {
	d := 0
	for _, non := range pathCtx.openEndNodes {
		if !non.pn.IsEnd() {
			d = non.pn.D()
			break
		}
	}
	if d == 0 {
		return 3
	}
	if d == 1 {
		return 6
	}
	// from sphere area growth of d to d+1 the ratio should be 1 + 2/d + 1/d^2
	origLen := float64(len(pathCtx.openEndNodes))
	df := float64(d)
	predictedRatio := 1.0 + 2.0/df + 1.0/(df*df)
	if d <= 16 {
		predictedRatio = predictedRatio * 1.11
	} else if d <= 32 {
		predictedRatio = predictedRatio * 1.04
	} else {
		predictedRatio = predictedRatio * 1.02
	}
	predictedLen := origLen*predictedRatio
	return int(predictedLen)
}

func (pathCtx *BasePathContext) MoveToNextNodes() {
	newOpenNodes := make([]OpenEndPath, 0, pathCtx.PredictedNextOpenNodesLen())
	for _, oen := range pathCtx.openEndNodes {
		pathNode := oen.pn
		if pathNode.IsEnd() {
			Log.Errorf("An open end node builder is a dead end at %v", oen.pn.P())
			continue
		}
		if !pathNode.IsLatest() {
			if Log.IsTrace() {
				Log.Errorf("An open end node builder has no more active links at %v", oen.pn.P())
			}
			continue
		}
		if pathNode.IsRoot() {
			td := m3point.GetTrioDetails(pathNode.GetTrioIndex())
			for _, c := range td.GetConnections() {
				pl, created := pathNode.addPathLink(c.GetId())
				if created {
					pn, pnc, npnb := pl.createDstNode(oen.pnb)
					if pnc {
						newOpenNodes = append(newOpenNodes, OpenEndPath{pn, npnb})
					}
				}
			}
		} else {
			td := m3point.GetTrioDetails(pathNode.GetTrioIndex())
			if td == nil {
				Log.Fatalf("reached a node without trio %s %s", pathNode.String(), pathNode.GetTrioIndex())
				continue
			}
			from := pathNode.GetFrom()
			if from == nil {
				Log.Fatalf("reached a node without a from %s", pathNode.String())
				continue
			}
			if pathNode.GetOtherFrom() != nil {
				lastConn := td.LastOtherConnection(from.GetConnId().GetNegId(), pathNode.GetOtherFrom().GetConnId().GetNegId())
				pl, created := pathNode.addPathLink(lastConn.GetId())
				if created {
					pn, pnc, npnb := pl.createDstNode(oen.pnb)
					if pnc {
						newOpenNodes = append(newOpenNodes, OpenEndPath{pn, npnb})
					}
				}
			} else {
				nextConns := td.OtherConnectionsFrom(from.GetConnId().GetNegId())
				for _, c := range nextConns {
					pl, created := pathNode.addPathLink(c.GetId())
					if created {
						pn, pnc, npnb := pl.createDstNode(oen.pnb)
						if pnc {
							newOpenNodes = append(newOpenNodes, OpenEndPath{pn, npnb})
						}
					}
				}
			}
		}
	}
	pathCtx.openEndNodes = newOpenNodes
}

func (pathCtx *BasePathContext) String() string {
	return fmt.Sprintf("Path-%s-%d", pathCtx.ctx.String(), pathCtx.offset)
}

func (pathCtx *BasePathContext) dumpInfo() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s\n%s: [", pathCtx.ctx.String(), pathCtx.rootPathNode.String()))
	for i := 0; i < 3; i++ {
		pl := pathCtx.rootPathNode.GetNext(i)
		sb.WriteString("\n")
		if pl != nil {
			sb.WriteString(fmt.Sprintf("%d:%s,", i, pl.dumpInfo(0)))
		} else {
			sb.WriteString(fmt.Sprintf("%d:nil,", i))
		}
	}
	sb.WriteString("]")
	return sb.String()
}

/***************************************************************/
// BasePathLink Functions
/***************************************************************/

func (pl *BasePathLink) GetSrc() PathNode {
	return pl.src
}

func (pl *BasePathLink) GetConnId() m3point.ConnectionId {
	return pl.connId
}

func (pl *BasePathLink) GetDst() PathNode {
	return pl.dst
}

func (pl *BasePathLink) HasDestination() bool {
	return pl.dst != nil && !pl.dst.IsEnd()
}

func (pl *BasePathLink) IsDeadEnd() bool {
	return pl == EndPathLink || pl.dst == nil || pl.dst.IsEnd()
}

func (pl *BasePathLink) SetDeadEnd() {
	pl.dst = EndPathNode
}

func (pl *BasePathLink) createDstNode(pathBuilder m3point.PathNodeBuilder) (PathNode, bool, m3point.PathNodeBuilder) {
	from := pl.src
	dstDistance := from.D() + 1
	pathCtx := from.GetPathContext()
	center := pathCtx.rootPathNode.P()
	npnb, np := pathBuilder.GetNextPathNodeBuilder(from.P().Sub(center), pl.connId, pathCtx.offset)
	realP := center.Add(np)
	existingPn, ok := pathCtx.pathNodeMap.GetPathNode(realP)

	if ok {
		if Log.IsTrace() {
			Log.Trace("adding node at %v to path link %v with path builder %s which already has node %s",
				realP, *pl, pathBuilder.String(), existingPn.String())
		}
		if existingPn.GetTrioIndex() == m3point.NilTrioIndex {
			Log.Fatalf("setting a new node at %v to path link %v using path builder %s on existing one %s with which has a nil trio idx",
				realP, *pl, pathBuilder.String(), existingPn.String())
		}
		if existingPn.GetTrioIndex() != npnb.GetTrioIndex() {
			Log.Fatalf("setting a new node at %v to path link %v using path builder %s on existing one %s with not same trio id %s != %s",
				realP, *pl, pathBuilder.String(), existingPn.String(), existingPn.GetTrioIndex(), npnb.GetTrioIndex())
		}
		distDiff := existingPn.D() - dstDistance
		if distDiff == 0 {
			// Merging path
			existingPn.setOtherFrom(pl)
			pl.dst = existingPn
			return existingPn, false, npnb
		} else if distDiff < 0 {
			// Ignoring longer path
			Log.Infof("Ignoring setting a new node at %v to path link %s on existing one %s since existing dist %d shorter than %d",
				realP, pl.String(), existingPn.String(), existingPn.D(), dstDistance)
			pl.SetDeadEnd()
			return nil, false, npnb
		} else {
			Log.Fatalf("setting a new node at %v to path link %s on existing one %s fatal since existing dist %d is in the future of %d",
				realP, pl.String(), existingPn.String(), existingPn.D(), dstDistance)
			pl.SetDeadEnd()
			return nil, false, npnb
		}
	}

	// Create the new node
	res := OutPathNode{}
	res.pathCtx = pathCtx
	res.p = realP
	res.d = dstDistance
	res.trioId = npnb.GetTrioIndex()
	res.from = pl

	pl.dst = &res

	pathCtx.pathNodeMap.AddPathNode(pl.dst)

	return pl.dst, true, npnb
}

func (pl *BasePathLink) String() string {
	return fmt.Sprintf("PL-%s-%v", pl.connId.String(), pl.IsDeadEnd())
}

func (pl *BasePathLink) dumpInfo(ident int) string {
	if !pl.HasDestination() {
		return pl.String()
	}
	var sb strings.Builder
	sb.WriteString(pl.String())
	if pl.dst != nil {
		sb.WriteString(":{")
		sb.WriteString(pl.dst.dumpInfo(ident + 1))
		sb.WriteString("}")
	} else {
		sb.WriteString(":{nil}")
	}
	return sb.String()
}

/***************************************************************/
// BasePathNode Functions
/***************************************************************/

func (bpn *BasePathNode) GetPathContext() *BasePathContext {
	return bpn.pathCtx
}

func (bpn *BasePathNode) IsEnd() bool {
	return false
}

func (bpn *BasePathNode) P() m3point.Point {
	return bpn.p
}

func (bpn *BasePathNode) GetTrioIndex() m3point.TrioIndex {
	return bpn.trioId
}

/***************************************************************/
// RootPathNode Functions
/***************************************************************/

func (rpn *RootPathNode) String() string {
	return fmt.Sprintf("RPN%v-%s", rpn.p, rpn.trioId)
}

func (rpn *RootPathNode) D() int {
	return 0
}

func (rpn *RootPathNode) IsRoot() bool {
	return true
}

func (rpn *RootPathNode) IsLatest() bool {
	// Latest means some next link open
	for _, pl := range rpn.next {
		if pl == nil {
			return true
		}
	}
	return false
}

func (rpn *RootPathNode) calcDist() int {
	return 0
}

func (rpn *RootPathNode) addPathLink(connId m3point.ConnectionId) (PathLink, bool) {
	if Log.DoAssert() {
		if rpn.trioId == m3point.NilTrioIndex {
			Log.Fatalf("creating a path link on root node %s pointing to non existent trio index", rpn.String())
			return nil, false
		}
		td := m3point.GetTrioDetails(rpn.trioId)
		if td == nil {
			Log.Errorf("creating a path link on root node %s pointing to non existent trio index", rpn.String())
			return nil, false
		}
		if !td.HasConnections(connId) {
			Log.Errorf("creating a path link on root node %s using connections %v not present in trio", rpn.String(), connId)
			return nil, false
		}
	}
	if rpn.next[0] != nil && rpn.next[1] != nil && rpn.next[2] != nil {
		// Check if one already match
		for _, npl := range rpn.next {
			if npl.GetConnId() == connId {
				if Log.IsTrace() {
					Log.Tracef("creating a path link for conn %s on node %s having next %s already matching", connId.String(), rpn.String(), npl.String())
				}
				return npl, false
			}
		}
		Log.Fatalf("creating a path link for conn %s on root node %s having no next left %s, %s, %s", connId.String(), rpn.String(), rpn.next[0].String(), rpn.next[1].String(), rpn.next[2].String())
		return nil, false
	}

	res := BasePathLink{}
	res.src = rpn
	res.connId = connId

	if rpn.next[0] == nil {
		rpn.next[0] = &res
	} else if rpn.next[1] == nil {
		rpn.next[1] = &res
	} else {
		rpn.next[2] = &res
	}

	return &res, true
}

func (rpn *RootPathNode) GetNextConnection(connId m3point.ConnectionId) PathLink {
	for _, pl := range rpn.next {
		if pl != nil && pl.GetConnId() == connId {
			return pl
		}
	}
	return nil
}

func (rpn *RootPathNode) GetNext(i int) PathLink {
	return rpn.next[i]
}

func (rpn *RootPathNode) GetFrom() PathLink {
	return nil
}

func (rpn *RootPathNode) GetOtherFrom() PathLink {
	return nil
}

func (rpn *RootPathNode) setOtherFrom(pl PathLink) {
	Log.Fatalf("cannot set other from on a root node")
}

func (rpn *RootPathNode) dumpInfo(ident int) string {
	panic("implement me")
}

/***************************************************************/
// OutPathNode Functions
/***************************************************************/

func (opn *OutPathNode) String() string {
	return fmt.Sprintf("OPN%v-%s-%s:%04d", opn.p, opn.trioId, opn.from, opn.d)
}

func (opn *OutPathNode) IsRoot() bool {
	return false
}

func (opn *OutPathNode) IsLatest() bool {
	if opn == nil {
		return false
	}
	// Latest means some next link still open
	for _, pl := range opn.next {
		if pl == nil {
			return true
		}
	}
	return false
}

func (opn *OutPathNode) D() int {
	return opn.d
}

func (opn *OutPathNode) GetFrom() PathLink {
	return opn.from
}

func (opn *OutPathNode) GetOtherFrom() PathLink {
	return opn.otherFrom
}

func (opn *OutPathNode) GetNextConnection(connId m3point.ConnectionId) PathLink {
	for _, pl := range opn.next {
		if pl != nil && !pl.IsDeadEnd() && pl.GetConnId() == connId {
			return pl
		}
	}
	return nil
}

func (opn *OutPathNode) GetNext(i int) PathLink {
	return opn.next[i]
}

func (opn *OutPathNode) addPathLink(connId m3point.ConnectionId) (PathLink, bool) {
	if opn.from.GetConnId() == connId.GetNegId() {
		if Log.IsTrace() {
			Log.Tracef("creating a path link for conn %s on node %s having from already matching", connId.String(), opn.String())
		}
		return opn.from, false
	}
	if opn.otherFrom != nil && opn.otherFrom.GetConnId() == connId.GetNegId() {
		if Log.IsTrace() {
			Log.Tracef("creating a path link for conn %s on node %s having other from already matching", connId.String(), opn.String())
		}
		return opn.otherFrom, false
	}
	if Log.DoAssert() {
		if opn.trioId == m3point.NilTrioIndex {
			Log.Fatalf("creating a path link with source node %s pointing to non existent trio index", opn.String())
			return nil, false
		}
		td := m3point.GetTrioDetails(opn.trioId)
		if td == nil {
			Log.Fatalf("creating a path link with source node %s pointing to non existent trio index", opn.String())
			return nil, false
		}
		if !td.HasConnections(connId) {
			Log.Errorf("creating a path link with source node %s using connections %v not present in trio", opn.String(), connId)
			return nil, false
		}
	}
	if opn.next[0] != nil && opn.next[1] != nil {
		// Check if one already match
		for _, npl := range opn.next {
			if npl.GetConnId() == connId {
				if Log.IsTrace() {
					Log.Tracef("creating a path link for conn %s on node %s having next %s already matching", connId.String(), opn.String(), npl.String())
				}
				return npl, false
			}
		}
		Log.Fatalf("creating a path link for conn %s on node %s having no next left %s, %s", connId.String(), opn.String(), opn.next[0].String(), opn.next[1].String())
		return nil, false
	}

	res := BasePathLink{}
	res.src = opn
	res.connId = connId

	if opn.next[0] == nil {
		opn.next[0] = &res
	} else {
		opn.next[1] = &res
	}
	return &res, true
}

func (opn *OutPathNode) calcDist() int {
	return opn.from.GetSrc().calcDist() + 1
}

func (opn *OutPathNode) setOtherFrom(pl PathLink) {
	opn.otherFrom = pl
	// block one of next that cannot be used anymore
	if opn.next[0] == nil {
		opn.next[0] = EndPathLink
	} else if opn.next[1] == nil {
		opn.next[1] = EndPathLink
	}
}

func (opn *OutPathNode) dumpInfo(ident int) string {
	var sb strings.Builder
	sb.WriteString(opn.String())
	if opn.trioId != m3point.NilTrioIndex && (opn.next[0] != nil || opn.next[1] != nil) {
		sb.WriteString("[")
		for i, pl := range opn.next {
			sb.WriteString("\n")
			for k := 0; k < ident; k++ {
				sb.WriteString("  ")
			}
			if pl != nil {
				sb.WriteString(fmt.Sprintf("%d:%s,", i, pl.dumpInfo(ident)))
			} else {
				sb.WriteString(fmt.Sprintf("%d:nil,", i))
			}
		}
		sb.WriteString("]")
	} else {
		sb.WriteString("[]")
	}
	return sb.String()
}

/***************************************************************/
// EndPathNodeT Functions
/***************************************************************/

func (epn *EndPathNodeT) String() string {
	return "END-NODE"
}

func (epn *EndPathNodeT) dumpInfo(ident int) string {
	return epn.String()
}

func (epn *EndPathNodeT) IsEnd() bool {
	return true
}

func (epn *EndPathNodeT) IsRoot() bool {
	return false
}

func (epn *EndPathNodeT) IsLatest() bool {
	return false
}

func (epn *EndPathNodeT) GetPathContext() *BasePathContext {
	Log.Fatalf("trying to get path context from end node")
	return nil
}

func (epn *EndPathNodeT) P() m3point.Point {
	Log.Fatalf("trying to get point from end node")
	return m3point.Origin
}

func (epn *EndPathNodeT) GetTrioIndex() m3point.TrioIndex {
	Log.Fatalf("trying to get trio context from end node")
	return m3point.NilTrioIndex
}

func (epn *EndPathNodeT) D() int {
	Log.Fatalf("trying to get distance from end node")
	return -1
}

func (epn *EndPathNodeT) GetFrom() PathLink {
	Log.Fatalf("trying to get from en node")
	return nil
}

func (epn *EndPathNodeT) GetOtherFrom() PathLink {
	Log.Fatalf("trying to get other from end node")
	return nil
}

func (epn *EndPathNodeT) GetNextConnection(connId m3point.ConnectionId) PathLink {
	Log.Fatalf("trying to get next connection from end node")
	return nil
}

func (epn *EndPathNodeT) GetNext(i int) PathLink {
	Log.Fatalf("trying to get next from end node")
	return nil
}

func (epn *EndPathNodeT) calcDist() int {
	Log.Fatalf("trying to calcDist from end node")
	return 0
}

func (epn *EndPathNodeT) addPathLink(connId m3point.ConnectionId) (PathLink, bool) {
	Log.Fatalf("trying to add path link to end node")
	return nil, false
}

func (epn *EndPathNodeT) setOtherFrom(pl PathLink) {
	Log.Fatalf("trying to set other from to end node")
}
