package m3point

import (
	"fmt"
)

type GrowthContext struct {
	TrioIndexContext
	// Where this growth context starts from
	center             Point
	// Offset in permutation to start with. Basically index in perm pos at center
	offset int
}

type PathContext struct {
	ctx           *GrowthContext
	trioSequences []PathContextElement
}

type PathContextElement struct {
	srcTrio   *TrioDetails
	nextTrios []*TrioDetails
}

func CreateGrowthContext(center Point, permType ContextType, index int, offset int) *GrowthContext {
	return &GrowthContext{TrioIndexContext{permType, index,}, center, offset}
}

var maxOffsetPerType = map[ContextType]int{
	ContextType(1): 1,
	ContextType(3): 4,
	ContextType(2): 2,
	ContextType(4): 4,
	ContextType(8): 8,
}

func CreateFromRoot(trioIndexCtx *TrioIndexContext, center Point, offset int) *GrowthContext {
	if offset < 0 || offset >= maxOffsetPerType[trioIndexCtx.ctxType] {
		Log.Error("Offset value %d invalid for context type", offset, trioIndexCtx.ctxType)
		return nil
	}
	return CreateGrowthContext(center, trioIndexCtx.ctxType, trioIndexCtx.ctxIndex, offset)
}

func (ctx *GrowthContext) SetIndexOffset(idx, offset int) {
	ctx.ctxIndex = idx
	ctx.offset = offset
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
		ctx.ctxType, ctx.ctxIndex)
}

func (ctx *GrowthContext) String() string {
	return fmt.Sprintf("GrowthType%d-Idx%d-Offset%d", ctx.ctxType, ctx.ctxIndex, ctx.offset)
}

func (ctx *GrowthContext) GetTrioIndex(divByThree uint64) int {
	if ctx.ctxType == 1 {
		// Always same value
		return ctx.ctxIndex
	}
	if ctx.ctxType == 3 {
		// Center on Trio index ctx.ctxIndex and then use X, Y, Z where conn are 1
		mod2 := PosMod2(divByThree)
		if mod2 == 0 {
			return ctx.ctxIndex
		}
		mod3 := int(((divByThree-1)/2 + uint64(ctx.offset)) % 3)
		if mod3 < 0 {
			mod3 += 3
		}
		if ctx.ctxIndex < 4 {
			return ValidNextTrio[3*ctx.ctxIndex+mod3][1]
		}
		count := 0
		for _, validTrio := range ValidNextTrio {
			if validTrio[1] == ctx.ctxIndex {
				if count == mod3 {
					return validTrio[0]
				}
				count++
			}
		}
		panic(fmt.Sprintf("did not find valid Trio for div by three value %d in context %v!", divByThree, ctx))
	}

	divByThreeWithOffset := uint64(ctx.offset) + divByThree
	switch ctx.ctxType {
	case 2:
		permutationMap := ValidNextTrio[ctx.ctxIndex]
		idx := int(PosMod2(divByThreeWithOffset))
		return permutationMap[idx]
	case 4:
		permutationMap := AllMod4Permutations[ctx.ctxIndex]
		idx := int(PosMod4(divByThreeWithOffset))
		return permutationMap[idx]
	case 8:
		permutationMap := AllMod8Permutations[ctx.ctxIndex]
		idx := int(PosMod8(divByThreeWithOffset))
		return permutationMap[idx]
	}
	panic(fmt.Sprintf("event permutation type %d in context %v is invalid!", ctx.ctxType, ctx))
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
