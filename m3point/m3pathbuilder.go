package m3point

import (
	"fmt"
	"strings"
)

// The ctx for each main point start point that gives in the global map the root path node builder
type PathBuilderContext struct {
	trCtx  *TrioIndexContext
	offset int
	div    uint64
}

func (ctx *PathBuilderContext) String() string {
	return fmt.Sprintf("PBC-%s-%d-%d", ctx.trCtx.String(), ctx.offset, ctx.div)
}

type PathNodeBuilder interface {
	fmt.Stringer
	GetTrioIndex() TrioIndex
	GetNextPathNodeBuilder(connId ConnectionId) PathNodeBuilder
	dumpInfo() string
	verify()
}

type BasePathNodeBuilder struct {
	ctx   *PathBuilderContext
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

var pathBuilders = make(map[PathBuilderContext]*RootPathNodeBuilder)

var MaxOffsetPerType = map[ContextType]int{
	ContextType(1): 1,
	ContextType(3): 4,
	ContextType(2): 2,
	ContextType(4): 4,
	ContextType(8): 8,
}

func createAllPathBuilders() int {
	if len(pathBuilders) != 0 {
		return len(pathBuilders)
	}
	for _, ctxType := range GetAllContextTypes() {
		nbIndexes := ctxType.GetNbIndexes()
		for pIdx := 0; pIdx < nbIndexes; pIdx++ {
			trioCtx := GetTrioIndexContext(ctxType, pIdx)
			maxOffset := MaxOffsetPerType[ctxType]
			for offset := 0; offset < maxOffset; offset++ {
				for div := uint64(0); div < 8; div++ {
					key := PathBuilderContext{trioCtx, offset, div}
					trIdx := trioCtx.GetBaseTrioIndex(div, offset)
					root := RootPathNodeBuilder{}
					root.ctx = &key
					root.trIdx = trIdx
					root.populate()
					pathBuilders[key] = &root
				}
			}
		}
	}
	return len(pathBuilders)
}

func GetPathNodeBuilder(trCtx *TrioIndexContext, offset int, divByThree uint64) PathNodeBuilder {
	return pathBuilders[PathBuilderContext{trCtx, offset, PosMod8(divByThree)}]
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

func (rpnb *RootPathNodeBuilder) GetNextPathNodeBuilder(connId ConnectionId) PathNodeBuilder {
	for _, plb := range rpnb.pathLinks {
		if plb.connId == connId {
			return plb.pathNode
		}
	}
	Log.Fatalf("trying to get next path node builder on connection %s which does not exists in %s", connId.String(), rpnb.String())
	return nil
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

func (rpnb *RootPathNodeBuilder) populate() {
	trCtx := rpnb.ctx.trCtx
	td := GetTrioDetails(rpnb.trIdx)
	for i, cd := range td.conns {
		_, ntd, npes := trCtx.GetForwardTrioFromMain(Origin.Add(XFirst.Mul(int64(rpnb.ctx.div))), td, cd.Id, rpnb.ctx.offset)
		ipnb := IntermediatePathNodeBuilder{}
		ipnb.ctx = rpnb.ctx
		ipnb.trIdx = ntd.GetId()
		for j, npe := range npes {
			lipnb := LastIntermediatePathNodeBuilder{}
			lipnb.ctx = rpnb.ctx
			ipTd, _ := trCtx.GetBackTrioOnInterPoint(npe)
			lipnb.trIdx = ipTd.GetId()
			lipnb.nextMainConnId = npe.nmp2ipConn.GetNegId()
			lipnb.nextInterConnId = ipTd.LastOtherConnection(lipnb.nextMainConnId, npe.GetP2IConn().GetNegId()).GetId()
			lipnb.verify()
			ipnb.pathLinks[j] = PathLinkBuilder{npe.GetP2IConn().Id, &lipnb}
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

func (ipnb *IntermediatePathNodeBuilder) GetNextPathNodeBuilder(connId ConnectionId) PathNodeBuilder {
	for _, plb := range ipnb.pathLinks {
		if plb.connId == connId {
			return plb.pathNode
		}
	}
	Log.Fatalf("trying to get next path node builder on connection %s which does not exists in %s trio %s", connId.String(), ipnb.String(), GetTrioDetails(ipnb.trIdx).String())
	return nil
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

func (lipnb *LastIntermediatePathNodeBuilder) GetNextPathNodeBuilder(connId ConnectionId) PathNodeBuilder {
	nextMainPnb := GetPathNodeBuilder(lipnb.ctx.trCtx, lipnb.ctx.offset, PosMod8(lipnb.ctx.div+1))
	if lipnb.nextMainConnId == connId {
		return nextMainPnb
	} else if lipnb.nextInterConnId == connId {
		return nextMainPnb.GetNextPathNodeBuilder(lipnb.nextMainConnId.GetNegId()).GetNextPathNodeBuilder(connId)
	}
	Log.Fatalf("trying to get next path node builder on connection %s which does not exists in %s", connId.String(), lipnb.String())
	return nil
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
