package m3space

import "fmt"

// TODO: Create trio index for non main points base on growth context

type GrowthContext struct {
	center             Point
	permutationType    uint8 // 1,2,4, or 8
	permutationIndex   int   // Index in the permutations to choose from. For type 1 [0,7] for the other in the 12 list [0,11]
	permutationNegFlow bool  // true for backward flow in permutation
	permutationOffset  int   // Offset in perm modulo
}

var reverse2Map = [2]int{1,0}
var reverse3Map = [3]int{2,1,0}
var reverse4Map = [4]int{3,2,1,0}
var reverse8Map = [8]int{7,6,5,4,3,2,1,0}

func PosMod2(i uint64) uint64 {
	return i & 0x0000000000000001
}

func PosMod4(i uint64) uint64 {
	return i & 0x0000000000000003
}

func PosMod8(i uint64) uint64 {
	return i & 0x0000000000000007
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
