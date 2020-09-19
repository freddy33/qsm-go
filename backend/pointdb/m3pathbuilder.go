package pointdb

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3point"
	"strings"
)

type PathNodeBuilder interface {
	fmt.Stringer
	GetEnv() m3util.QsmEnvironment
	GetCubeId() int
	GetTrioIndex() m3point.TrioIndex
	GetNextPathNodeBuilder(from m3point.Point, connId m3point.ConnectionId, offset int) (PathNodeBuilder, m3point.Point)
	DumpInfo() string
	Verify()
}

// The Ctx for each main point start point that gives in the global map the root path node builder
type PathBuilderContext struct {
	GrowthCtx m3point.GrowthContext
	CubeId    int
}

func (ctx *PathBuilderContext) String() string {
	return fmt.Sprintf("PBC-%02d-%03d", ctx.GrowthCtx.GetId(), ctx.CubeId)
}

type BasePathNodeBuilder struct {
	Ctx   *PathBuilderContext
	TrIdx m3point.TrioIndex
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
	NextMainConnId  m3point.ConnectionId
	NextInterConnId m3point.ConnectionId
}

type PathLinkBuilder struct {
	ConnId   m3point.ConnectionId
	PathNode PathNodeBuilder
}

/***************************************************************/
// PathLink Functions
/***************************************************************/

func (pl *PathLinkBuilder) dumpInfo() string {
	return fmt.Sprintf("%s %s", pl.ConnId.String(), pl.PathNode.DumpInfo())
}

func (pl *PathLinkBuilder) GetConnectionId() m3point.ConnectionId {
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

func (pnb *BasePathNodeBuilder) getPointPackData() ServerPointPackDataIfc {
	return pnb.Ctx.GrowthCtx.GetEnv().GetData(m3util.PointIdx).(ServerPointPackDataIfc)
}

func (pnb *BasePathNodeBuilder) GetCubeId() int {
	return pnb.Ctx.CubeId
}

func (pnb *BasePathNodeBuilder) GetTrioIndex() m3point.TrioIndex {
	return pnb.TrIdx
}

/***************************************************************/
// RootPathNodeBuilder Functions
/***************************************************************/

func (rpnb *RootPathNodeBuilder) String() string {
	return fmt.Sprintf("RNB-%s-%s", rpnb.Ctx.String(), rpnb.TrIdx.String())
}

func (rpnb *RootPathNodeBuilder) DumpInfo() string {
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

func (rpnb *RootPathNodeBuilder) GetNextPathNodeBuilder(from m3point.Point, connId m3point.ConnectionId, offset int) (PathNodeBuilder, m3point.Point) {
	for _, plb := range rpnb.PathLinks {
		if plb.ConnId == connId {
			return plb.PathNode, from.Add(rpnb.getPointPackData().GetConnDetailsById(connId).Vector)
		}
	}
	Log.Fatalf("trying to get next path node builder on connection %s which does not exists in %s", connId.String(), rpnb.String())
	return nil, Origin
}

func (rpnb *RootPathNodeBuilder) Verify() {
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
	Ud       m3point.UnitDirection
	Lip      m3point.Point
	BackConn *m3point.ConnectionDetails
	Lipnb    *LastPathNodeBuilder
}

/***************************************************************/
// IntermediatePathNodeBuilder Functions
/***************************************************************/

func (ipnb *IntermediatePathNodeBuilder) String() string {
	return fmt.Sprintf("INB-%s-%s", ipnb.Ctx.String(), ipnb.TrIdx.String())
}

func (ipnb *IntermediatePathNodeBuilder) DumpInfo() string {
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

func (ipnb *IntermediatePathNodeBuilder) GetNextPathNodeBuilder(from m3point.Point, connId m3point.ConnectionId, offset int) (PathNodeBuilder, m3point.Point) {
	for _, plb := range ipnb.PathLinks {
		if plb.ConnId == connId {
			return plb.PathNode, from.Add(ipnb.getPointPackData().GetConnDetailsById(connId).Vector)
		}
	}
	Log.Fatalf("trying to get next path node builder on connection %s which does not exists in %s trio %s", connId.String(), ipnb.String(), ipnb.TrIdx.String())
	return nil, Origin
}

func (ipnb *IntermediatePathNodeBuilder) Verify() {
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

func (lipnb *LastPathNodeBuilder) DumpInfo() string {
	return fmt.Sprintf("LINB-%s %s %s", lipnb.TrIdx.String(), lipnb.NextMainConnId, lipnb.NextInterConnId)
}

func (lipnb *LastPathNodeBuilder) GetNextMainConnId() m3point.ConnectionId {
	return lipnb.NextMainConnId
}

func (lipnb *LastPathNodeBuilder) GetNextInterConnId() m3point.ConnectionId {
	return lipnb.NextInterConnId
}

func (lipnb *LastPathNodeBuilder) GetNextPathNodeBuilder(from m3point.Point, connId m3point.ConnectionId, offset int) (PathNodeBuilder, m3point.Point) {
	ppd := lipnb.getPointPackData()
	nextMainPoint := from.GetNearMainPoint()
	if Log.DoAssert() {
		oNextMainPoint := from.Add(ppd.GetConnDetailsById(lipnb.NextMainConnId).Vector)
		if nextMainPoint != oNextMainPoint {
			Log.Fatalf("last inter main path node %s (%s) does give a main point using %v and %s", lipnb.String(), lipnb.DumpInfo(), from, lipnb.NextMainConnId)
		}
	}
	nextMainPnb := ppd.GetPathNodeBuilder(lipnb.Ctx.GrowthCtx, offset, nextMainPoint)
	if lipnb.NextMainConnId == connId {
		return nextMainPnb, nextMainPoint
	} else if lipnb.NextInterConnId == connId {
		nextInterPnbBack, oInterPoint := nextMainPnb.GetNextPathNodeBuilder(nextMainPoint, lipnb.NextMainConnId.GetNegId(), offset)
		if Log.DoAssert() {
			if from != oInterPoint {
				Log.Fatalf("back calculation on last inter path node %s (%s) failed %v != %v", lipnb.String(), lipnb.DumpInfo(), from, oInterPoint)
			}
		}
		return nextInterPnbBack.GetNextPathNodeBuilder(from, connId, offset)
	}
	Log.Fatalf("trying to get next path node builder on connection %s which does not exists in %s", connId.String(), lipnb.String())
	return nil, Origin
}

func (lipnb *LastPathNodeBuilder) Verify() {
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