package m3path

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3point"
	"github.com/freddy33/qsm-go/m3util"
	"strings"
)

var Log = m3util.NewLogger("m3path", m3util.INFO)

type PathContext struct {
	ctx               *m3point.TrioIndexContext
	offset            int
	rootTrioId        m3point.TrioIndex
	rootPathLinks     [3]*PathLink
	openEndPaths      []OpenEndPath
	possiblePathIds   map[PathIdKey][2]NextPathLink
	pathNodesPerPoint map[m3point.Point]*PathNode
}

type PathIdKey struct {
	previousMainDivThree uint64
	previousMainTrioId   m3point.TrioIndex
	previousMainConnId   m3point.ConnectionId
	previousTrioId       m3point.TrioIndex
	previousConnId       m3point.ConnectionId
}

type NextPathLink struct {
	connId     m3point.ConnectionId
	nextTrioId m3point.TrioIndex
}

// A single path link between *src* node to one of the next path node *dst* using the connection Id
type PathLink struct {
	// The path context the link belongs to
	pathCtx *PathContext
	// After travelling the connId of the above cur.connId there will be 2 new path possible for
	src *PathNode
	// The connection used by the link path
	connId m3point.ConnectionId
	// After travelling the connId the pointer to the next path node
	dst *PathNode
}

// The link graph node of a path, representing one point on the graph
// Points to the 2 path links usable from here
type PathNode struct {
	p m3point.Point
	// Distance from root
	d int
	// From which link this node came from
	from *PathLink
	// The current trio index of the path point
	trioId m3point.TrioIndex
	// After travelling the connId of the above cur.connId there will be 2 new path possible for
	next [2]*PathLink
	// If this node came from a combined link
	otherFrom *PathLink
}

type OpenPathType int8

const (
	RootOpenPath OpenPathType = iota
	MainPointOpenPath
	InterPointOpenPath
	NilOpenPath
)

// Struct left at the end of a path builder where next round of building should be done
type OpenEndPath struct {
	// The type of open path
	kind OpenPathType
	// The path node with trio index and next left to be build
	pn *PathNode
	// The next path element used to build the path node above
	npel *m3point.NextPathElement
}

var NilOpenEndPath = OpenEndPath{NilOpenPath, nil, nil,}

/***************************************************************/
// PathContext Functions
/***************************************************************/

func MakePathContext(ctxType m3point.ContextType, pIdx int, offset int) *PathContext {
	return MakePathContextFromTrioContext(m3point.GetTrioIndexContext(ctxType, pIdx), offset)
}

func MakePathContextFromTrioContext(trCtx *m3point.TrioIndexContext, offset int) *PathContext {
	pathCtx := PathContext{}
	pathCtx.ctx = trCtx
	pathCtx.offset = offset
	pathCtx.pathNodesPerPoint = make(map[m3point.Point]*PathNode)

	return &pathCtx
}

func (pathCtx *PathContext) initRootLinks() {
	trIdx := pathCtx.ctx.GetBaseTrioIndex(0, pathCtx.offset)
	pathCtx.rootTrioId = trIdx

	td := m3point.GetTrioDetails(trIdx)
	for i, c := range td.GetConnections() {
		pathCtx.makeRootPathLink(i, c.GetId())
	}

	// Hack since only node with three next set all to nil, but still need to be filled in the map
	rootPathNode := PathNode{}
	rootPathNode.p = m3point.Origin
	rootPathNode.d = 0
	rootPathNode.trioId = trIdx
	pathCtx.pathNodesPerPoint[m3point.Origin] = &rootPathNode

	pathCtx.openEndPaths = make([]OpenEndPath, 1)
	pathCtx.openEndPaths[0] = OpenEndPath{
		RootOpenPath,
		&rootPathNode,
		nil,
	}
}

func (pathCtx *PathContext) makeRootPathLink(idx int, connId m3point.ConnectionId) *PathLink {
	res := PathLink{}
	res.pathCtx = pathCtx
	res.src = nil
	res.connId = connId
	pathCtx.rootPathLinks[idx] = &res
	return &res
}

func (pn *PathNode) addInterOpenEndPath(backNpe *m3point.NextPathElement) OpenEndPath {
	nnpl := pn.addPathLink(backNpe.GetP2IConn().GetId())
	if nnpl == nil || nnpl.dst != nil {
		return NilOpenEndPath
	}
	npn := nnpl.createDstNode(backNpe.GetIntermediatePoint(), m3point.NilTrioIndex)
	if npn == nil {
		return NilOpenEndPath
	}

	newEndPath := OpenEndPath{}
	newEndPath.kind = InterPointOpenPath
	newEndPath.npel = backNpe
	newEndPath.pn = npn
	return newEndPath
}

func (pn *PathNode) addMainOpenEndPath(npel *m3point.NextPathElement) OpenEndPath {
	nnpl := pn.addPathLink(npel.GetNmp2IConn().GetNegId())
	if nnpl == nil || nnpl.dst != nil {
		return NilOpenEndPath
	}
	npn := nnpl.createDstNode(npel.GetNextMainPoint(), npel.GetNextMainTrioId())
	if npn == nil {
		return NilOpenEndPath
	}

	newEndPath := OpenEndPath{}
	newEndPath.kind = MainPointOpenPath
	newEndPath.npel = npel
	newEndPath.pn = npn

	return newEndPath
}

func (pl *PathLink) addAllPaths(mainPoint m3point.Point, td *m3point.TrioDetails) []OpenEndPath {
	res := make([]OpenEndPath, 0, 4)
	trCtx := pl.pathCtx.ctx
	p, nextTrio, nextPathEls := trCtx.GetForwardTrioFromMain(mainPoint, td, pl.connId, pl.pathCtx.offset)
	pn := pl.createDstNode(p, nextTrio.GetId())
	if pn == nil {
		// nothing left
		return res
	}
	for j := 0; j < 2; j++ {
		npel := nextPathEls[j]
		npl := pn.addPathLink(npel.GetP2IConn().GetId())
		if npl == nil || npl.dst != nil {
			break
		}
		ipTd, backPathEls := trCtx.GetBackTrioOnInterPoint(npel)
		npn := npl.createDstNode(npel.GetIntermediatePoint(), ipTd.GetId())
		if npn != nil {
			mop := npn.addMainOpenEndPath(npel)
			if mop != NilOpenEndPath {
				res = append(res, mop)
			}
			for _, backNpe := range backPathEls {
				// One of the back path el should go back to main point => not interesting
				if backNpe.GetNextMainPoint() != mainPoint {
					iop := npn.addInterOpenEndPath(backNpe)
					if iop != NilOpenEndPath {
						res = append(res, iop)
					}
				}
			}
		}
	}
	return res
}

func (pathCtx *PathContext) moveToNextMainPoints() {
	var newOpenPaths []OpenEndPath
	trCtx := pathCtx.ctx
	for _, oep := range pathCtx.openEndPaths {
		if oep.kind == RootOpenPath {
			mpn := oep.pn
			newOpenPaths = make([]OpenEndPath, 12)
			idx := 0
			mainPoint := mpn.p
			td := m3point.GetTrioDetails(mpn.trioId)
			if Log.DoAssert() {
				if len(pathCtx.openEndPaths) != 1 {
					Log.Errorf("Got more than one (%d) open path and one is a root open path for %s", len(pathCtx.openEndPaths), trCtx.String())
					return
				}
				if !mainPoint.IsMainPoint() {
					Log.Errorf("The root open path has a non main point %v for %s", mainPoint, trCtx.String())
					return
				}
			}
			for _, pl := range pathCtx.rootPathLinks {
				oeps := pl.addAllPaths(mainPoint, td)
				for _, oep := range oeps {
					newOpenPaths[idx] = oep
					idx++
				}
			}
		} else if oep.kind == MainPointOpenPath {
			if cap(newOpenPaths) == 0 {
				newOpenPaths = make([]OpenEndPath, 0, 2*len(pathCtx.openEndPaths))
			}
			mpn := oep.pn
			mainPoint := mpn.p
			td := m3point.GetTrioDetails(mpn.trioId)
			if Log.DoAssert() {
				if !mainPoint.IsMainPoint() {
					Log.Errorf("The main open path has a non main point %v for %s", mainPoint, trCtx.String())
					return
				}
			}
			if mpn.otherFrom == nil {
				ocs := td.OtherConnectionsFrom(mpn.from.connId)
				for _, oc := range ocs {
					pl := mpn.addPathLink(oc.GetId())
					if pl != nil && pl.dst == nil {
						oeps := pl.addAllPaths(mainPoint, td)
						for _, oep := range oeps {
							newOpenPaths = append(newOpenPaths, oep)
						}
					}
				}
			} else {
				oc := td.LastOtherConnection(oep.pn.from.connId.GetNegId(), oep.pn.otherFrom.connId.GetNegId())
				pl := mpn.addPathLink(oc.GetId())
				if pl != nil && pl.dst == nil {
					oeps := pl.addAllPaths(mainPoint, td)
					for _, oep := range oeps {
						newOpenPaths = append(newOpenPaths, oep)
					}
				}
			}
		} else if oep.kind == InterPointOpenPath {
			ipn := oep.pn
			intPoint := ipn.p
			if Log.DoAssert() {
				if intPoint.IsMainPoint() {
					Log.Errorf("The intermediate open path has a main point %v for %s", intPoint, trCtx.String())
					return
				}
			}
			ipTd, backPathEls := trCtx.GetBackTrioOnInterPoint(oep.npel)
			if oep.pn.trioId == m3point.NilTrioIndex {
				oep.pn.trioId = ipTd.GetId()
			} else {
				if oep.pn.trioId != ipTd.GetId() {
					Log.Fatalf("Open end path inter %v has tid %s diff from %s using %v",
						oep, oep.pn.trioId.String(), ipTd.String(), *oep.npel)
				}
			}
			mop := ipn.addMainOpenEndPath(oep.npel)
			if mop != NilOpenEndPath {
				newOpenPaths = append(newOpenPaths, mop)
			}
			for _, backNpe := range backPathEls {
				if backNpe.GetNextMainPoint() != oep.npel.GetNextMainPoint() {
					iop := ipn.addInterOpenEndPath(backNpe)
					if iop != NilOpenEndPath {
						newOpenPaths = append(newOpenPaths, iop)
					}
				}
			}
		} else {
			Log.Errorf("got a wrong kind of open path %v", oep)
		}
	}
	pathCtx.openEndPaths = newOpenPaths
}

func (pathCtx *PathContext) String() string {
	return fmt.Sprintf("Path-%s-%d", pathCtx.ctx.String(), pathCtx.offset)
}

func (pathCtx *PathContext) dumpInfo() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s\n%s: [", pathCtx.ctx.String(), pathCtx.rootTrioId.String()))
	for i, pl := range pathCtx.rootPathLinks {
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
// PathLink Functions
/***************************************************************/

func (pl *PathLink) createDstNode(p m3point.Point, tdId m3point.TrioIndex) *PathNode {
	var dstDistance int
	if pl.src != nil {
		dstDistance = pl.src.d + 1
	} else {
		dstDistance = 1
	}
	existingPn, ok := pl.pathCtx.pathNodesPerPoint[p]

	if ok {
		if Log.IsTrace() {
			Log.Trace("adding node at %v to path link %v with tdId %s which already has node %v", p, *pl, tdId, *existingPn)
		}
		if existingPn.trioId == m3point.NilTrioIndex {
			existingPn.trioId = tdId
		}
		if tdId == m3point.NilTrioIndex {
			tdId = existingPn.trioId
		}
		if existingPn.trioId != tdId {
			Log.Fatalf("setting a new node at %v to path link %v on existing one %v with not same trio id %d",
				p, *pl, *existingPn, tdId)
		}
		distDiff := existingPn.d - dstDistance
		if distDiff == 0 {
			// Merging path
			existingPn.otherFrom = pl
		} else if distDiff == 2 {
			// Taking over since this is a shorter path
			Log.Infof("Resetting path link %s setting a new node at %v on existing one %s since existing dist %d longer than %d",
				pl.String(), p, existingPn.String(), existingPn.d, dstDistance)
			return nil
		} else if distDiff == -2 {
			// Ignoring longer path
			Log.Infof("Ignoring setting a new node at %v to path link %s on existing one %s since existing dist %d shorter than %d",
				p, pl.String(), existingPn.String(), existingPn.d, dstDistance)
			return nil
		} else {
			Log.Fatalf("setting a new node at %v to path link %s on existing one %s fatal since dist %d completely out of range of %d",
				p, pl.String(), existingPn.String(), existingPn.d, dstDistance)
			return nil
		}
		return existingPn
	}

	// Create the new node
	res := PathNode{}
	res.p = p
	res.d = dstDistance
	res.from = pl
	res.trioId = tdId
	pl.dst = &res

	pl.pathCtx.pathNodesPerPoint[p] = pl.dst

	return pl.dst
}

func (pl *PathLink) String() string {
	return fmt.Sprintf("PL-%s", pl.connId.String())
}

func (pl *PathLink) dumpInfo(ident int) string {
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
// PathNode Functions
/***************************************************************/

func (pn *PathNode) addPathLink(connId m3point.ConnectionId) *PathLink {
	if pn.from.connId == connId.GetNegId() {
		if Log.IsTrace() {
			Log.Tracef("creating a path link for conn %s on node %s having from already matching", connId.String(), pn.String())
		}
		return pn.from
	}
	if Log.DoAssert() {
		if pn.trioId == m3point.NilTrioIndex {
			Log.Errorf("creating a path link with source node %s pointing to non existent trio index", pn.String())
			return nil
		}
		td := m3point.GetTrioDetails(pn.trioId)
		if td == nil {
			Log.Errorf("creating a path link with source node %s pointing to non existent trio index", pn.String())
			return nil
		}
		if !td.HasConnections(connId) {
			Log.Errorf("creating a path link with source node %s using connections %v not present in trio", pn.String(), connId)
			return nil
		}
	}
	if pn.next[0] != nil && pn.next[1] != nil {
		// Check if one already match
		for _, npl := range pn.next {
			if npl.connId == connId {
				if Log.IsTrace() {
					Log.Tracef("creating a path link for conn %s on node %s having next %s already matching", connId.String(), pn.String(), npl.String())
				}
				return npl
			}
		}
		Log.Fatalf("creating a path link for conn %s on node %s having no next left %s, %s", connId.String(), pn.String(), pn.next[0].String(), pn.next[1].String())
		return nil
	}

	res := PathLink{}
	res.pathCtx = pn.from.pathCtx
	res.src = pn
	res.connId = connId

	if pn.next[0] == nil {
		pn.next[0] = &res
	} else {
		pn.next[1] = &res
	}
	return &res
}

func (pn *PathNode) String() string {
	return fmt.Sprintf("PN%v-%s-%s:%04d", pn.p, pn.trioId, pn.from, pn.d)
}

func (pn *PathNode) calcDist() int {
	if pn.from.src == nil {
		return 1
	}
	return pn.from.src.calcDist() + 1
}

func (pn *PathNode) dumpInfo(ident int) string {
	var sb strings.Builder
	sb.WriteString(pn.String())
	if pn.trioId != m3point.NilTrioIndex && (pn.next[0] != nil || pn.next[1] != nil) {
		sb.WriteString("[")
		for i, pl := range pn.next {
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
