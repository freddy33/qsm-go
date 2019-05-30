package m3point

import "fmt"

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
	nextDiv         uint64
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

func (rpnb *RootPathNodeBuilder) GetNextPathNodeBuilder(connId ConnectionId) PathNodeBuilder {
	for _, plb := range rpnb.pathLinks {
		if plb.connId == connId {
			return plb.pathNode
		}
	}
	Log.Fatalf("trying to get next path node builder on connection %s which does not exists in %s", connId.String(), rpnb.String())
	return nil
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
			lipnb.nextDiv = PosMod8(rpnb.ctx.div + 1)
			lipnb.nextMainConnId = npe.nmp2ipConn.GetNegId()
			lipnb.nextInterConnId = ipTd.LastOtherConnection(lipnb.nextMainConnId, npe.GetP2IConn().Id.GetNegId()).GetId()
			ipnb.pathLinks[j] = PathLinkBuilder{npe.GetP2IConn().Id, &lipnb}
		}
		rpnb.pathLinks[i] = PathLinkBuilder{cd.Id, &ipnb}
	}
}

/***************************************************************/
// IntermediatePathNodeBuilder Functions
/***************************************************************/

func (ipnb *IntermediatePathNodeBuilder) String() string {
	return fmt.Sprintf("INB-%s-%s", ipnb.ctx.String(), ipnb.trIdx.String())
}

func (ipnb *IntermediatePathNodeBuilder) GetNextPathNodeBuilder(connId ConnectionId) PathNodeBuilder {
	for _, plb := range ipnb.pathLinks {
		if plb.connId == connId {
			return plb.pathNode
		}
	}
	Log.Fatalf("trying to get next path node builder on connection %s which does not exists in %s", connId.String(), ipnb.String())
	return nil
}

/***************************************************************/
// LastIntermediatePathNodeBuilder Functions
/***************************************************************/

func (lipnb *LastIntermediatePathNodeBuilder) String() string {
	return fmt.Sprintf("LINB-%s-%s-%d", lipnb.ctx.String(), lipnb.trIdx.String(), lipnb.nextDiv)
}

func (lipnb *LastIntermediatePathNodeBuilder) GetNextPathNodeBuilder(connId ConnectionId) PathNodeBuilder {
	nextMainPnb := GetPathNodeBuilder(lipnb.ctx.trCtx, lipnb.ctx.offset, lipnb.nextDiv)
	if lipnb.nextMainConnId == connId {
		return nextMainPnb
	} else if lipnb.nextInterConnId == connId {
		return nextMainPnb.GetNextPathNodeBuilder(lipnb.nextMainConnId.GetNegId()).GetNextPathNodeBuilder(connId)
	}
	Log.Fatalf("trying to get next path node builder on connection %s which does not exists in %s", connId.String(), lipnb.String())
	return nil
}
