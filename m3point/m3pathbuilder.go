package m3point

// The key for each main point start point that gives in the global map the root path node builder
type PathBuilderKey struct {
	trCtx *TrioIndexContext
	offset int
	div uint64
}

type PathNodeBuilder interface {
	GetTrioIndex() TrioIndex
	GetNextPathNodeBuilder(connId ConnectionId) PathNodeBuilder
}

type RootPathNodeBuilder struct {
	trIdx TrioIndex
	pathLinks [3]PathLinkBuilder
}

type EndPathNodeBuilder struct {
	trIdx TrioIndex
	intermediate bool
	nextDiv uint64
}

type IntermediatePathNodeBuilder struct {
	trIdx TrioIndex
	pathLinks [2]PathLinkBuilder
}

type PathLinkBuilder struct {
	connId ConnectionId
	pathNode *PathNodeBuilder
}

var pathBuilders = make(map[PathBuilderKey]*RootPathNodeBuilder)

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
					key := PathBuilderKey{trioCtx, offset, div}
					trIdx := trioCtx.GetBaseTrioIndex(div, offset)
					root := RootPathNodeBuilder{}
					root.trIdx = trIdx
					pathBuilders[key] = &root
				}
			}
		}
	}
	return len(pathBuilders)
}

func GetPathNodeBuilder(trCtx *TrioIndexContext, offset int, divByThree uint64) PathNodeBuilder {
	return pathBuilders[PathBuilderKey{trCtx, offset, PosMod8(divByThree)}]
}

func (rpnb *RootPathNodeBuilder) GetTrioIndex() TrioIndex {
	return rpnb.trIdx
}

func (rpnb *RootPathNodeBuilder) GetNextPathNodeBuilder(connId ConnectionId) PathNodeBuilder {
	return nil
}
