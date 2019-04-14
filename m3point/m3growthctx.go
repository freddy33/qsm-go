package m3point

import (
	"fmt"
)

// TODO: Create trio index for non main points base on growth context

/*
Define how outgrowth and path evolve from the center. There are 6 types of growth depending of the value of permutationType:
1. type = 0 : Main points will not be covered. TODO Switch from trio to next that has neg conn
2. type = 1 : All main points have the same base trio index
3. type = 3 : Rotate between valid trios depending on starting index in modulo 3
4. type = 2 : Use the modulo 2 permutation => Specific index valid next trio back and forth
5. type = 4 : Use the modulo 4 permutation => Specific index line in AllMod4Permutations cycling through the 4 values
6. type = 8 : Use the modulo 8 permutation => Specific index line in AllMod8Permutations cycling through the 8 values
*/
type ContextType uint8 // 0,1,2,3,4, or 8

var allContextTypes = [5]ContextType{1, 2, 3, 4, 8}

type GrowthContext struct {
	center             Point
	permutationType    ContextType
	permutationIndex   int  // Index in the permutations to choose from. For type 1 [0,7] for the other in the 12 list [0,11]
	permutationNegFlow bool // true for backward flow in permutation
	permutationOffset  int  // Offset in permutation to start with
}

var reverse2Map = [2]int{1, 0}
var reverse3Map = [3]int{2, 1, 0}
var reverse4Map = [4]int{3, 2, 1, 0}
var reverse8Map = [8]int{7, 6, 5, 4, 3, 2, 1, 0}

var allRootContexts map[ContextType][]*GrowthContext

func init() {
	count := make(map[ContextType]int)
	allRootContexts = make(map[ContextType][]*GrowthContext)
	for _, ctxType := range GetAllContextTypes() {
		nbIndexes := ctxType.GetNbIndexes()
		allRootContexts[ctxType] = make([]*GrowthContext, nbIndexes)
		for pIdx := 0; pIdx < nbIndexes; pIdx++ {
			allRootContexts[ctxType][pIdx] = CreateRootGrowthContext(ctxType, pIdx)
			count[ctxType]++
		}
	}
	Log.Info(count)
}

type PathContext struct {
	ctx           *GrowthContext
	trioSequences []PathContextElement
}

type PathContextElement struct {
	srcTrio   *TrioDetails
	nextTrios []*TrioDetails
}

func PosMod2(i uint64) uint64 {
	return i & 0x0000000000000001
}

func PosMod4(i uint64) uint64 {
	return i & 0x0000000000000003
}

func PosMod8(i uint64) uint64 {
	return i & 0x0000000000000007
}

func CreateGrowthContext(center Point, permType ContextType, index int, flow bool, offset int) *GrowthContext {
	return &GrowthContext{center, permType, index, flow, offset}
}

func CreateRootGrowthContext(permType ContextType, index int) *GrowthContext {
	return CreateGrowthContext(Origin, permType, index, false, 0)
}

func CreateFromRoot(rootCtx *GrowthContext, center Point, flow bool, offset int) *GrowthContext {
	if offset < 0 || offset >= maxOffsetPerType[rootCtx.permutationType] {
		Log.Error("Offset value %d invalid for context type", offset, rootCtx.permutationType)
		return nil
	}
	newCtx := *rootCtx
	newCtx.center = center
	newCtx.permutationNegFlow = flow
	newCtx.permutationOffset = offset
	return &newCtx
}

func GetAllContextTypes() [5]ContextType {
	return allContextTypes
}

func (t ContextType) IsPermutation() bool {
	return t == ContextType(2) || t == ContextType(4) || t == ContextType(8)
}

func (t ContextType) GetModulo() int {
	return int(t)
}

func (t ContextType) GetNbIndexes() int {
	if t.IsPermutation() {
		return 12
	}
	return 8
}

func GetRootContext(ctxType ContextType, index int) *GrowthContext {
	return allRootContexts[ctxType][index]
}

var maxOffsetPerType = map[ContextType]int{
	ContextType(1): 1,
	ContextType(3): 4,
	ContextType(2): 2,
	ContextType(4): 4,
	ContextType(8): 8,
}

func (ctx *GrowthContext) SetIndexOffset(idx, offset int) {
	ctx.permutationIndex = idx
	ctx.permutationOffset = offset
}

func (ctx *GrowthContext) SetCenter(c Point) {
	ctx.center = c
}

func (ctx *GrowthContext) GetCenter() Point {
	return ctx.center
}

func (ctx *GrowthContext) GetFileName() string {
	return fmt.Sprintf("C_%03d_%03d_%03d_G_%d_%d",
		ctx.center[0], ctx.center[1], ctx.center[2],
		ctx.permutationType, ctx.permutationIndex)
}

func (ctx *GrowthContext) GetContextString() string {
	return fmt.Sprintf("Type %d, Idx %d, Neg %v, Offset %d", ctx.permutationType, ctx.permutationIndex, ctx.permutationNegFlow, ctx.permutationOffset)
}

func (ctx *GrowthContext) GetTrioIndex(divByThree uint64) int {
	if ctx.permutationType == 1 {
		// Always same value
		return ctx.permutationIndex
	}
	if ctx.permutationType == 3 {
		// Center on Trio index ctx.permutationIndex and then use X, Y, Z where conn are 1
		mod2 := PosMod2(divByThree)
		if mod2 == 0 {
			return ctx.permutationIndex
		}
		mod3 := int(((divByThree-1)/2 + uint64(ctx.permutationOffset)) % 3)
		if mod3 < 0 {
			mod3 += 3
		}
		if ctx.permutationIndex < 4 {
			return ValidNextTrio[3*ctx.permutationIndex+mod3][1]
		}
		count := 0
		for _, validTrio := range ValidNextTrio {
			if validTrio[1] == ctx.permutationIndex {
				if count == mod3 {
					return validTrio[0]
				}
				count++
			}
		}
		panic(fmt.Sprintf("did not find valid Trio for div by three value %d in context %v!", divByThree, ctx))
	}

	divByThreeWithOffset := uint64(ctx.permutationOffset) + divByThree
	switch ctx.permutationType {
	case 2:
		permutMap := ValidNextTrio[ctx.permutationIndex]
		idx := int(PosMod2(divByThreeWithOffset))
		if ctx.permutationNegFlow {
			return permutMap[reverse2Map[idx]]
		} else {
			return permutMap[idx]
		}
	case 4:
		permutMap := AllMod4Permutations[ctx.permutationIndex]
		idx := int(PosMod4(divByThreeWithOffset))
		if ctx.permutationNegFlow {
			return permutMap[reverse4Map[idx]]
		} else {
			return permutMap[idx]
		}
	case 8:
		permutMap := AllMod8Permutations[ctx.permutationIndex]
		idx := int(PosMod8(divByThreeWithOffset))
		if ctx.permutationNegFlow {
			return permutMap[reverse8Map[idx]]
		} else {
			return permutMap[idx]
		}
	}
	panic(fmt.Sprintf("event permutation type %d in context %v is invalid!", ctx.permutationType, ctx))
}

func (ctx *GrowthContext) GetDivByThree(p Point) uint64 {
	if !p.IsMainPoint() {
		panic(fmt.Sprintf("cannot ask for Trio index on non main Pos %v in context %v!", p, ctx))
	}
	return uint64(Abs64(p[0]-ctx.center[0])/3 + Abs64(p[1]-ctx.center[1])/3 + Abs64(p[2]-ctx.center[2])/3)
}

func (ctx *GrowthContext) GetTrio(p Point) Trio {
	return AllBaseTrio[ctx.GetTrioIndex(ctx.GetDivByThree(p))]
}
