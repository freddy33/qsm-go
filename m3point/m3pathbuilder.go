package m3point

import (
	"fmt"
	"strings"
)

// The ctx for each main point start point that gives in the global map the root path node builder
type PathBuilderContext struct {
	trCtxId int
	cubeId  int
}

func (ctx *PathBuilderContext) String() string {
	return fmt.Sprintf("PBC-%02d-%03d", ctx.trCtxId, ctx.cubeId)
}

type PathNodeBuilder interface {
	fmt.Stringer
	GetTrioIndex() TrioIndex
	GetNextPathNodeBuilder(from Point, connId ConnectionId, offset int) (PathNodeBuilder, Point)
	dumpInfo() string
	verify()
}

type BasePathNodeBuilder struct {
	ctx   PathBuilderContext
	trIdx TrioIndex
}

type RootPathNodeBuilder struct {
	BasePathNodeBuilder
	pathLinks [3]PathLinkBuilder
}

type IntermediatePathNodeBuilder struct {
	BasePathNodeBuilder
	pathLinks [2]PathLinkBuilder
}

type LastIntermediatePathNodeBuilder struct {
	BasePathNodeBuilder
	nextMainConnId  ConnectionId
	nextInterConnId ConnectionId
}

type PathLinkBuilder struct {
	connId   ConnectionId
	pathNode PathNodeBuilder
}

// The index of this slice is the cube id
var pathBuilders []*RootPathNodeBuilder

var maxOffsetPerType = map[ContextType]int{
	ContextType(1): 1,
	ContextType(3): 4,
	ContextType(2): 2,
	ContextType(4): 4,
	ContextType(8): 8,
}

func calculateAllPathBuilders() []*RootPathNodeBuilder {
	checkCubesInitialized()
	res := make([]*RootPathNodeBuilder, TotalNumberOfCubes+1)
	res[0] = nil
	for cubeKey, cubeId := range cubeIdsPerKey {
		key := PathBuilderContext{cubeKey.trCtxId, cubeId}
		root := RootPathNodeBuilder{}
		root.ctx = key
		root.populate()
		res[cubeId] = &root
	}
	return res
}

func GetPathNodeBuilder(trCtx *TrioContext, offset int, c Point) PathNodeBuilder {
	checkPathBuildersInitialized()
	key := CubeKeyId{trCtx.id, createTrioCube(trCtx, offset, c)}
	cubeId := GetCubeIdByKey(key)
	return pathBuilders[cubeId]
}

/***************************************************************/
// PathLink Functions
/***************************************************************/

func (pl *PathLinkBuilder) dumpInfo() string {
	return fmt.Sprintf("%s %s", pl.connId.String(), pl.pathNode.dumpInfo())
}

/***************************************************************/
// BasePathNodeBuilder Functions
/***************************************************************/

func (pnb *BasePathNodeBuilder) GetTrioIndex() TrioIndex {
	return pnb.trIdx
}

/***************************************************************/
// RootPathNodeBuilder Functions
/***************************************************************/

func (rpnb *RootPathNodeBuilder) String() string {
	return fmt.Sprintf("RNB-%s-%s", rpnb.ctx.String(), rpnb.trIdx.String())
}

func (rpnb *RootPathNodeBuilder) dumpInfo() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\n%s", rpnb.String()))
	for i, pl := range rpnb.pathLinks {
		sb.WriteString(fmt.Sprintf("\n\t%d : ", i))
		sb.WriteString(pl.dumpInfo())
	}
	return sb.String()
}

func (rpnb *RootPathNodeBuilder) GetNextPathNodeBuilder(from Point, connId ConnectionId, offset int) (PathNodeBuilder, Point) {
	for _, plb := range rpnb.pathLinks {
		if plb.connId == connId {
			return plb.pathNode, from.Add(GetConnDetailsById(connId).Vector)
		}
	}
	Log.Fatalf("trying to get next path node builder on connection %s which does not exists in %s", connId.String(), rpnb.String())
	return nil, Origin
}

func (rpnb *RootPathNodeBuilder) verify() {
	td := GetTrioDetails(rpnb.trIdx)
	if !td.HasConnection(rpnb.pathLinks[0].connId) {
		Log.Errorf("%s failed checking next path link 0 %s part of trio", rpnb.String(), rpnb.pathLinks[0].connId)
	}
	if !td.HasConnection(rpnb.pathLinks[1].connId) {
		Log.Errorf("%s failed checking next path link 1 %s part of trio", rpnb.String(), rpnb.pathLinks[1].connId)
	}
	if !td.HasConnection(rpnb.pathLinks[2].connId) {
		Log.Errorf("%s failed checking next path link 2 %s part of trio", rpnb.String(), rpnb.pathLinks[2].connId)
	}
	if rpnb.pathLinks[0].connId == rpnb.pathLinks[1].connId {
		Log.Errorf("%s failed checking next path links 0 and 1 connections are different", rpnb.String(), rpnb.pathLinks[0].connId, rpnb.pathLinks[1].connId)
	}
	if rpnb.pathLinks[0].connId == rpnb.pathLinks[2].connId {
		Log.Errorf("%s failed checking next path links 0 and 2 connections are different", rpnb.String(), rpnb.pathLinks[0].connId, rpnb.pathLinks[2].connId)
	}
	if rpnb.pathLinks[1].connId == rpnb.pathLinks[2].connId {
		Log.Errorf("%s failed checking next path links 1 and 2 connections are different", rpnb.String(), rpnb.pathLinks[1].connId, rpnb.pathLinks[2].connId)
	}
}

type NextMainPathNode struct {
	ud       UnitDirection
	lip      Point
	backConn *ConnectionDetails
	lipnb    *LastIntermediatePathNodeBuilder
}

func (rpnb *RootPathNodeBuilder) populate() {
	trCtx := GetTrioContextById(rpnb.ctx.trCtxId)
	cubeKey := GetCubeById(rpnb.ctx.cubeId)
	cube := cubeKey.cube
	rpnb.trIdx = cube.center
	td := GetTrioDetails(rpnb.trIdx)
	for i, cd := range td.conns {
		// We are talking about the intermediate point here
		ip := cd.Vector

		// From each center out connection there 2 last PNB
		// They can be filled from the 2 unit directions of the base vector
		nextMains := [2]NextMainPathNode{}
		for j, ud := range cd.GetDirections() {
			nextMains[j].ud = ud
			nmp := ud.GetFirstPoint()
			nextTrIdx := cube.GetCenterFaceTrio(ud)
			nextTd := GetTrioDetails(nextTrIdx)
			backConn := nextTd.getOppositeConn(ud)
			nextMains[j].lip = nmp.Add(backConn.Vector)
			nextMains[j].backConn = backConn
			lipnb := LastIntermediatePathNodeBuilder{}
			lipnb.ctx = rpnb.ctx
			lipnb.nextMainConnId = backConn.GetNegId()
			nextMains[j].lipnb = &lipnb
		}

		// We have all the last nodes let's create the intermediate one
		// We have the three connections from ip to find the correct trio
		var iTd *TrioDetails
		ipConns := [2]*ConnectionDetails{GetConnDetailsByPoints(ip, nextMains[0].lip), GetConnDetailsByPoints(ip, nextMains[1].lip)}
		for _, possTd := range allTrioDetails {
			if possTd.HasConnections(cd.GetNegId(), ipConns[0].GetId(), ipConns[1].GetId()) {
				iTd = possTd
				break
			}
		}
		if iTd == nil {
			Log.Fatalf("did not find any trio details matching %s %s %s in %s cube %s", cd.GetNegId(), ipConns[0].GetId(), ipConns[1].GetId(), trCtx.String(), cube.String())
			return
		}

		ipnb := IntermediatePathNodeBuilder{}
		ipnb.ctx = rpnb.ctx
		ipnb.trIdx = iTd.GetId()

		// Find the trio index for filling the last intermediate
		for j, nm := range nextMains {
			backUds := nm.backConn.GetDirections()
			foundUd := false
			for _, backUd := range backUds {
				if backUd.GetOpposite() == nm.ud {
					foundUd = true
				} else {
					nextInterTrIdx := cube.GetMiddleEdgeTrio(nm.ud, backUd)
					nextInterTd := GetTrioDetails(nextInterTrIdx)
					nextInterBackConn := nextInterTd.getOppositeConn(backUd)
					nextInterNearMainPoint := nm.ud.GetFirstPoint().Add(backUd.GetFirstPoint()).Add(nextInterBackConn.Vector)
					lipToOtherConn := GetConnDetailsByPoints(nm.lip, nextInterNearMainPoint)
					nm.lipnb.nextInterConnId = lipToOtherConn.GetId()

					var liTd *TrioDetails
					for _, possTd := range allTrioDetails {
						if possTd.HasConnections(ipConns[j].GetNegId(), nm.lipnb.nextInterConnId, nm.lipnb.nextMainConnId) {
							liTd = possTd
							break
						}
					}
					if liTd == nil {
						Log.Fatalf("did not find any trio details matching %s %s %s in %s cube %s", ipConns[j].GetNegId(), nm.lipnb.nextInterConnId, nm.lipnb.nextMainConnId, trCtx.String(), cube.String())
						return
					}
					nm.lipnb.trIdx = liTd.GetId()
				}
			}
			if !foundUd {
				Log.Fatalf("direction mess between trio details %s %s and %d %v", td.String(), iTd.String(), nm.ud, backUds)
			}
			nm.lipnb.verify()
			ipnb.pathLinks[j] = PathLinkBuilder{ipConns[j].GetId(), nm.lipnb}
		}
		ipnb.verify()

		rpnb.pathLinks[i] = PathLinkBuilder{cd.Id, &ipnb}
	}
	rpnb.verify()
}

/***************************************************************/
// IntermediatePathNodeBuilder Functions
/***************************************************************/

func (ipnb *IntermediatePathNodeBuilder) String() string {
	return fmt.Sprintf("INB-%s-%s", ipnb.ctx.String(), ipnb.trIdx.String())
}

func (ipnb *IntermediatePathNodeBuilder) dumpInfo() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("INB-%s", ipnb.trIdx.String()))
	for i, pl := range ipnb.pathLinks {
		sb.WriteString(fmt.Sprintf("\n\t\t%d : ", i))
		sb.WriteString(pl.dumpInfo())
	}
	return sb.String()
}

func (ipnb *IntermediatePathNodeBuilder) GetNextPathNodeBuilder(from Point, connId ConnectionId, offset int) (PathNodeBuilder, Point) {
	for _, plb := range ipnb.pathLinks {
		if plb.connId == connId {
			return plb.pathNode, from.Add(GetConnDetailsById(connId).Vector)
		}
	}
	Log.Fatalf("trying to get next path node builder on connection %s which does not exists in %s trio %s", connId.String(), ipnb.String(), GetTrioDetails(ipnb.trIdx).String())
	return nil, Origin
}

func (ipnb *IntermediatePathNodeBuilder) verify() {
	td := GetTrioDetails(ipnb.trIdx)
	if !td.HasConnection(ipnb.pathLinks[0].connId) {
		Log.Errorf("%s failed checking next path link 0 %s part of trio", ipnb.String(), ipnb.pathLinks[0].connId)
	}
	if !td.HasConnection(ipnb.pathLinks[1].connId) {
		Log.Errorf("%s failed checking next path link 1 %s part of trio", ipnb.String(), ipnb.pathLinks[1].connId)
	}
	if ipnb.pathLinks[0].connId == ipnb.pathLinks[1].connId {
		Log.Errorf("%s failed checking next path links connections are different", ipnb.String(), ipnb.pathLinks[0].connId, ipnb.pathLinks[1].connId)
	}
}

/***************************************************************/
// LastIntermediatePathNodeBuilder Functions
/***************************************************************/

func (lipnb *LastIntermediatePathNodeBuilder) String() string {
	return fmt.Sprintf("LINB-%s-%s", lipnb.ctx.String(), lipnb.trIdx.String())
}

func (lipnb *LastIntermediatePathNodeBuilder) dumpInfo() string {
	return fmt.Sprintf("LINB-%s %s %s", lipnb.trIdx.String(), lipnb.nextMainConnId, lipnb.nextInterConnId)
}

func (lipnb *LastIntermediatePathNodeBuilder) GetNextPathNodeBuilder(from Point, connId ConnectionId, offset int) (PathNodeBuilder, Point) {
	nextMainPoint := from.GetNearMainPoint()
	if Log.DoAssert() {
		oNextMainPoint := from.Add(GetConnDetailsById(lipnb.nextMainConnId).Vector)
		if nextMainPoint != oNextMainPoint {
			Log.Fatalf("last inter main path node %s (%s) does give a main point using %v and %s", lipnb.String(), lipnb.dumpInfo(), from, lipnb.nextMainConnId)
		}
	}
	nextMainPnb := GetPathNodeBuilder(GetTrioContextById(lipnb.ctx.trCtxId), offset, nextMainPoint)
	if lipnb.nextMainConnId == connId {
		return nextMainPnb, nextMainPoint
	} else if lipnb.nextInterConnId == connId {
		nextInterPnbBack, oInterPoint := nextMainPnb.GetNextPathNodeBuilder(nextMainPoint, lipnb.nextMainConnId.GetNegId(), offset)
		if Log.DoAssert() {
			if from != oInterPoint {
				Log.Fatalf("back calculation on last inter path node %s (%s) failed %v != %v", lipnb.String(), lipnb.dumpInfo(), from, oInterPoint)
			}
		}
		return nextInterPnbBack.GetNextPathNodeBuilder(from, connId, offset)
	}
	Log.Fatalf("trying to get next path node builder on connection %s which does not exists in %s", connId.String(), lipnb.String())
	return nil, Origin
}

func (lipnb *LastIntermediatePathNodeBuilder) verify() {
	td := GetTrioDetails(lipnb.trIdx)
	if !td.HasConnection(lipnb.nextMainConnId) {
		Log.Errorf("%s %s %s failed checking next main connection part of trio", lipnb.String(), lipnb.nextMainConnId, lipnb.nextInterConnId)
	}
	if !td.HasConnection(lipnb.nextInterConnId) {
		Log.Errorf("%s %s %s failed checking next intermediate connection part of trio", lipnb.String(), lipnb.nextMainConnId, lipnb.nextInterConnId)
	}
	if lipnb.nextMainConnId == lipnb.nextInterConnId {
		Log.Errorf("%s %s %s failed checking next main and intermediate connections are different", lipnb.String(), lipnb.nextMainConnId, lipnb.nextInterConnId)
	}
}
