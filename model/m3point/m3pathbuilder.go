package m3point

import (
	"fmt"
	"github.com/freddy33/qsm-go/utils/m3util"
	"strings"
)

// The Ctx for each main point start point that gives in the global map the root path node builder
type PathBuilderContext struct {
	GrowthCtx GrowthContext
	CubeId    int
}

func (ctx *PathBuilderContext) String() string {
	return fmt.Sprintf("PBC-%02d-%03d", ctx.GrowthCtx.GetId(), ctx.CubeId)
}

type BasePathNodeBuilder struct {
	Ctx   *PathBuilderContext
	TrIdx TrioIndex
}

type RootPathNodeBuilder struct {
	BasePathNodeBuilder
	PathLinks [3]PathLinkBuilder
}

type IntermediatePathNodeBuilder struct {
	BasePathNodeBuilder
	PathLinks [2]PathLinkBuilder
}

type LastPathNodeBuilder struct {
	BasePathNodeBuilder
	NextMainConnId  ConnectionId
	NextInterConnId ConnectionId
}

type PathLinkBuilder struct {
	ConnId   ConnectionId
	PathNode PathNodeBuilder
}

func (ppd *BasePointPackData) GetPathNodeBuilder(growthCtx GrowthContext, offset int, c Point) PathNodeBuilder {
	ppd.checkPathBuildersInitialized()
	// TODO: Verify the key below stay local and is not staying in memory
	key := CubeKeyId{growthCtx.GetId(), ppd.CreateTrioCube(growthCtx, offset, c)}
	cubeId := ppd.GetCubeIdByKey(key)
	return ppd.GetPathNodeBuilderById(cubeId)
}

func (ppd *BasePointPackData) GetPathNodeBuilderById(cubeId int) PathNodeBuilder {
	return ppd.PathBuilders[cubeId]
}

/***************************************************************/
// PathLink Functions
/***************************************************************/

func (pl *PathLinkBuilder) dumpInfo() string {
	return fmt.Sprintf("%s %s", pl.ConnId.String(), pl.PathNode.dumpInfo())
}

func (pl *PathLinkBuilder) GetConnectionId() ConnectionId {
	return pl.ConnId
}

func (pl *PathLinkBuilder) GetPathNodeBuilder() PathNodeBuilder {
	return pl.PathNode
}

/***************************************************************/
// BasePathNodeBuilder Functions
/***************************************************************/

func (pnb *BasePathNodeBuilder) GetEnv() m3util.QsmEnvironment {
	return pnb.Ctx.GrowthCtx.GetEnv()
}

func (pnb *BasePathNodeBuilder) getPointPackData() *BasePointPackData {
	return GetPointPackData(pnb.Ctx.GrowthCtx.GetEnv())
}

func (pnb *BasePathNodeBuilder) GetCubeId() int {
	return pnb.Ctx.CubeId
}

func (pnb *BasePathNodeBuilder) GetTrioIndex() TrioIndex {
	return pnb.TrIdx
}

/***************************************************************/
// RootPathNodeBuilder Functions
/***************************************************************/

func (rpnb *RootPathNodeBuilder) String() string {
	return fmt.Sprintf("RNB-%s-%s", rpnb.Ctx.String(), rpnb.TrIdx.String())
}

func (rpnb *RootPathNodeBuilder) dumpInfo() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\n%s", rpnb.String()))
	for i, pl := range rpnb.PathLinks {
		sb.WriteString(fmt.Sprintf("\n\t%d : ", i))
		sb.WriteString(pl.dumpInfo())
	}
	return sb.String()
}

func (rpnb *RootPathNodeBuilder) GetPathLinks() []PathLinkBuilder {
	return rpnb.PathLinks[:]
}

func (rpnb *RootPathNodeBuilder) GetNextPathNodeBuilder(from Point, connId ConnectionId, offset int) (PathNodeBuilder, Point) {
	for _, plb := range rpnb.PathLinks {
		if plb.ConnId == connId {
			return plb.PathNode, from.Add(rpnb.getPointPackData().GetConnDetailsById(connId).Vector)
		}
	}
	Log.Fatalf("trying to get next path node builder on connection %s which does not exists in %s", connId.String(), rpnb.String())
	return nil, Origin
}

func (rpnb *RootPathNodeBuilder) verify() {
	td := rpnb.getPointPackData().GetTrioDetails(rpnb.TrIdx)
	if !td.HasConnection(rpnb.PathLinks[0].ConnId) {
		Log.Errorf("%s failed checking next path link 0 %s part of trio", rpnb.String(), rpnb.PathLinks[0].ConnId)
	}
	if !td.HasConnection(rpnb.PathLinks[1].ConnId) {
		Log.Errorf("%s failed checking next path link 1 %s part of trio", rpnb.String(), rpnb.PathLinks[1].ConnId)
	}
	if !td.HasConnection(rpnb.PathLinks[2].ConnId) {
		Log.Errorf("%s failed checking next path link 2 %s part of trio", rpnb.String(), rpnb.PathLinks[2].ConnId)
	}
	if rpnb.PathLinks[0].ConnId == rpnb.PathLinks[1].ConnId {
		Log.Errorf("%s failed checking next path links 0 and 1 connections are different", rpnb.String(), rpnb.PathLinks[0].ConnId, rpnb.PathLinks[1].ConnId)
	}
	if rpnb.PathLinks[0].ConnId == rpnb.PathLinks[2].ConnId {
		Log.Errorf("%s failed checking next path links 0 and 2 connections are different", rpnb.String(), rpnb.PathLinks[0].ConnId, rpnb.PathLinks[2].ConnId)
	}
	if rpnb.PathLinks[1].ConnId == rpnb.PathLinks[2].ConnId {
		Log.Errorf("%s failed checking next path links 1 and 2 connections are different", rpnb.String(), rpnb.PathLinks[1].ConnId, rpnb.PathLinks[2].ConnId)
	}
}

type NextMainPathNode struct {
	ud       UnitDirection
	lip      Point
	backConn *ConnectionDetails
	lipnb    *LastPathNodeBuilder
}

func (rpnb *RootPathNodeBuilder) Populate() {
	growthCtx := rpnb.Ctx.GrowthCtx
	ppd := rpnb.getPointPackData()
	cubeKey := ppd.GetCubeById(rpnb.Ctx.CubeId)
	cube := cubeKey.Cube
	rpnb.TrIdx = cube.Center
	td := ppd.GetTrioDetails(rpnb.TrIdx)
	for i, cd := range td.Conns {
		// We are talking about the intermediate point here
		ip := cd.Vector

		// From each center out connection there 2 last PNB
		// They can be filled from the 2 unit directions of the base vector
		nextMains := [2]NextMainPathNode{}
		for j, ud := range cd.GetDirections() {
			nextMains[j].ud = ud
			nmp := ud.GetFirstPoint()
			nextTrIdx := cube.GetCenterFaceTrio(ud)
			nextTd := ppd.GetTrioDetails(nextTrIdx)
			backConn := nextTd.getOppositeConn(ud)
			nextMains[j].lip = nmp.Add(backConn.Vector)
			nextMains[j].backConn = backConn
			lipnb := LastPathNodeBuilder{}
			lipnb.Ctx = rpnb.Ctx
			lipnb.NextMainConnId = backConn.GetNegId()
			nextMains[j].lipnb = &lipnb
		}

		// We have all the last nodes let's create the intermediate one
		// We have the three connections from ip to find the correct trio
		var iTd *TrioDetails
		ipConns := [2]*ConnectionDetails{ppd.GetConnDetailsByPoints(ip, nextMains[0].lip), ppd.GetConnDetailsByPoints(ip, nextMains[1].lip)}
		for _, possTd := range ppd.AllTrioDetails {
			if possTd.HasConnections(cd.GetNegId(), ipConns[0].GetId(), ipConns[1].GetId()) {
				iTd = possTd
				break
			}
		}
		if iTd == nil {
			Log.Fatalf("did not find any trio details matching %s %s %s in %s cube %s", cd.GetNegId(), ipConns[0].GetId(), ipConns[1].GetId(), growthCtx.String(), cube.String())
			return
		}

		ipnb := IntermediatePathNodeBuilder{}
		ipnb.Ctx = rpnb.Ctx
		ipnb.TrIdx = iTd.GetId()

		// Find the trio index for filling the last intermediate
		for j, nm := range nextMains {
			backUds := nm.backConn.GetDirections()
			foundUd := false
			for _, backUd := range backUds {
				if backUd.GetOpposite() == nm.ud {
					foundUd = true
				} else {
					nextInterTrIdx := cube.GetMiddleEdgeTrio(nm.ud, backUd)
					nextInterTd := ppd.GetTrioDetails(nextInterTrIdx)
					nextInterBackConn := nextInterTd.getOppositeConn(backUd)
					nextInterNearMainPoint := nm.ud.GetFirstPoint().Add(backUd.GetFirstPoint()).Add(nextInterBackConn.Vector)
					lipToOtherConn := ppd.GetConnDetailsByPoints(nm.lip, nextInterNearMainPoint)
					nm.lipnb.NextInterConnId = lipToOtherConn.GetId()

					var liTd *TrioDetails
					for _, possTd := range ppd.AllTrioDetails {
						if possTd.HasConnections(ipConns[j].GetNegId(), nm.lipnb.NextInterConnId, nm.lipnb.NextMainConnId) {
							liTd = possTd
							break
						}
					}
					if liTd == nil {
						Log.Fatalf("did not find any trio details matching %s %s %s in %s cube %s", ipConns[j].GetNegId(), nm.lipnb.NextInterConnId, nm.lipnb.NextMainConnId, growthCtx.String(), cube.String())
						return
					}
					nm.lipnb.TrIdx = liTd.GetId()
				}
			}
			if !foundUd {
				Log.Fatalf("direction mess between trio details %s %s and %d %v", td.String(), iTd.String(), nm.ud, backUds)
			}
			nm.lipnb.verify()
			ipnb.PathLinks[j] = PathLinkBuilder{ipConns[j].GetId(), nm.lipnb}
		}
		ipnb.verify()

		rpnb.PathLinks[i] = PathLinkBuilder{cd.Id, &ipnb}
	}
	rpnb.verify()
}

/***************************************************************/
// IntermediatePathNodeBuilder Functions
/***************************************************************/

func (ipnb *IntermediatePathNodeBuilder) String() string {
	return fmt.Sprintf("INB-%s-%s", ipnb.Ctx.String(), ipnb.TrIdx.String())
}

func (ipnb *IntermediatePathNodeBuilder) dumpInfo() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("INB-%s", ipnb.TrIdx.String()))
	for i, pl := range ipnb.PathLinks {
		sb.WriteString(fmt.Sprintf("\n\t\t%d : ", i))
		sb.WriteString(pl.dumpInfo())
	}
	return sb.String()
}

func (ipnb *IntermediatePathNodeBuilder) GetPathLinks() []PathLinkBuilder {
	return ipnb.PathLinks[:]
}

func (ipnb *IntermediatePathNodeBuilder) GetNextPathNodeBuilder(from Point, connId ConnectionId, offset int) (PathNodeBuilder, Point) {
	for _, plb := range ipnb.PathLinks {
		if plb.ConnId == connId {
			return plb.PathNode, from.Add(ipnb.getPointPackData().GetConnDetailsById(connId).Vector)
		}
	}
	Log.Fatalf("trying to get next path node builder on connection %s which does not exists in %s trio %s", connId.String(), ipnb.String(), ipnb.TrIdx.String())
	return nil, Origin
}

func (ipnb *IntermediatePathNodeBuilder) verify() {
	td := ipnb.getPointPackData().GetTrioDetails(ipnb.TrIdx)
	if !td.HasConnection(ipnb.PathLinks[0].ConnId) {
		Log.Errorf("%s failed checking next path link 0 %s part of trio", ipnb.String(), ipnb.PathLinks[0].ConnId)
	}
	if !td.HasConnection(ipnb.PathLinks[1].ConnId) {
		Log.Errorf("%s failed checking next path link 1 %s part of trio", ipnb.String(), ipnb.PathLinks[1].ConnId)
	}
	if ipnb.PathLinks[0].ConnId == ipnb.PathLinks[1].ConnId {
		Log.Errorf("%s failed checking next path links connections are different", ipnb.String(), ipnb.PathLinks[0].ConnId, ipnb.PathLinks[1].ConnId)
	}
}

/***************************************************************/
// LastPathNodeBuilder Functions
/***************************************************************/

func (lipnb *LastPathNodeBuilder) String() string {
	return fmt.Sprintf("LINB-%s-%s", lipnb.Ctx.String(), lipnb.TrIdx.String())
}

func (lipnb *LastPathNodeBuilder) dumpInfo() string {
	return fmt.Sprintf("LINB-%s %s %s", lipnb.TrIdx.String(), lipnb.NextMainConnId, lipnb.NextInterConnId)
}

func (lipnb *LastPathNodeBuilder) GetNextMainConnId() ConnectionId {
	return lipnb.NextMainConnId
}

func (lipnb *LastPathNodeBuilder) GetNextInterConnId() ConnectionId {
	return lipnb.NextInterConnId
}

func (lipnb *LastPathNodeBuilder) GetNextPathNodeBuilder(from Point, connId ConnectionId, offset int) (PathNodeBuilder, Point) {
	ppd := lipnb.getPointPackData()
	nextMainPoint := from.GetNearMainPoint()
	if Log.DoAssert() {
		oNextMainPoint := from.Add(ppd.GetConnDetailsById(lipnb.NextMainConnId).Vector)
		if nextMainPoint != oNextMainPoint {
			Log.Fatalf("last inter main path node %s (%s) does give a main point using %v and %s", lipnb.String(), lipnb.dumpInfo(), from, lipnb.NextMainConnId)
		}
	}
	nextMainPnb := ppd.GetPathNodeBuilder(lipnb.Ctx.GrowthCtx, offset, nextMainPoint)
	if lipnb.NextMainConnId == connId {
		return nextMainPnb, nextMainPoint
	} else if lipnb.NextInterConnId == connId {
		nextInterPnbBack, oInterPoint := nextMainPnb.GetNextPathNodeBuilder(nextMainPoint, lipnb.NextMainConnId.GetNegId(), offset)
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

func (lipnb *LastPathNodeBuilder) verify() {
	td := lipnb.getPointPackData().GetTrioDetails(lipnb.TrIdx)
	if !td.HasConnection(lipnb.NextMainConnId) {
		Log.Errorf("%s %s %s failed checking next main connection part of trio", lipnb.String(), lipnb.NextMainConnId, lipnb.NextInterConnId)
	}
	if !td.HasConnection(lipnb.NextInterConnId) {
		Log.Errorf("%s %s %s failed checking next intermediate connection part of trio", lipnb.String(), lipnb.NextMainConnId, lipnb.NextInterConnId)
	}
	if lipnb.NextMainConnId == lipnb.NextInterConnId {
		Log.Errorf("%s %s %s failed checking next main and intermediate connections are different", lipnb.String(), lipnb.NextMainConnId, lipnb.NextInterConnId)
	}
}
