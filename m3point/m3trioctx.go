package m3point

import (
	"fmt"
)

/*
Define how outgrowth and path evolve from the center. There are 6 types of growth depending of the value of growthType:
TODO: Create trio index for non nextMainPoint points base on growth type
1. type = 0 : Type not yet existing TODO: Main points will not be covered. In here trio index switch from trio to next that has neg conn
2. type = 1 : All nextMainPoint points have the same base trio index
3. type = 3 : Rotate between valid trios depending on starting index in modulo 3
4. type = 2 : Use the modulo 2 permutation => Specific index valid next trio back and forth
5. type = 4 : Use the modulo 4 permutation => Specific index line in AllMod4Permutations cycling through the 4 values
6. type = 8 : Use the modulo 8 permutation => Specific index line in AllMod8Permutations cycling through the 8 values
*/
type GrowthType uint8

var allContextTypes = [5]GrowthType{1, 2, 3, 4, 8}
var totalNbContexts = 8 + 12 + 8 + 12 + 12

type BaseGrowthContext struct {
	// A generate id used in arrays and db
	id int
	// The context type for this flow context
	growthType GrowthType
	// Index in the permutations to choose from. For type 1 and 3 [0,7] for the other in the 12 list [0,11]
	// Max number of indexes returned by GrowthType.GetNbIndexes()
	growthIndex int
}

// A struct representing one next nextMainPoint point where a path is going toward
type NextPathElement struct {
	valid  bool
	offset int
	// The next nextMainPoint points where p path is going to go to
	nextMainPoint Point
	// The trio details for this specific next nextMainPoint point
	nextMainTd *TrioDetails

	// The intermediate point before p path will reach before going to the nextMainPoint point
	ipNearNm Point
	// The connection used on nextMainPoint point to reach the previous intermediate point
	nmp2ipConn *ConnectionDetails

	// The connection used between the 2 intermediate points
	p2iConn *ConnectionDetails
}

var allGrowthContexts []GrowthContext

func calculateAllGrowthContexts() []GrowthContext {
	res := make([]GrowthContext, totalNbContexts)
	idx := 0
	for _, ctxType := range GetAllContextTypes() {
		nbIndexes := ctxType.GetNbIndexes()
		for pIdx := 0; pIdx < nbIndexes; pIdx++ {
			growthCtx := BaseGrowthContext{idx, ctxType, pIdx}
			res[idx] = &growthCtx
			idx++
		}
	}
	return res
}

/***************************************************************/
// GrowthType Functions
/***************************************************************/

func GetAllGrowthContexts() []GrowthContext {
	checkGrowthContextsInitialized()
	return allGrowthContexts
}

func GetAllContextTypes() [5]GrowthType {
	return allContextTypes
}

func (t GrowthType) String() string {
	return fmt.Sprintf("CtxType%d", t)
}

func (t GrowthType) IsPermutation() bool {
	return t == GrowthType(2) || t == GrowthType(4) || t == GrowthType(8)
}

func (t GrowthType) GetModulo() int {
	return int(t)
}

func (t GrowthType) GetNbIndexes() int {
	if t.IsPermutation() {
		return 12
	}
	return 8
}

func (t GrowthType) GetMaxOffset() int {
	return maxOffsetPerType[t]
}

/***************************************************************/
// BaseGrowthContext Functions
/***************************************************************/

func GetGrowthContextById(id int) GrowthContext {
	checkGrowthContextsInitialized()
	return allGrowthContexts[id]
}

func GetGrowthContextByTypeAndIndex(growthType GrowthType, index int) GrowthContext {
	checkGrowthContextsInitialized()
	for _, growthCtx := range allGrowthContexts {
		if growthCtx.GetGrowthType() == growthType && growthCtx.GetGrowthIndex() == index {
			return growthCtx
		}
	}
	Log.Fatalf("could not find trio Context for %d %d", growthType, index)
	return nil
}

func (gowthCtx *BaseGrowthContext) String() string {
	return fmt.Sprintf("GrowthCtx%d-%d-Idx%02d", gowthCtx.id, gowthCtx.growthType, gowthCtx.growthIndex)
}

func (gowthCtx *BaseGrowthContext) GetId() int {
	return gowthCtx.id
}

func (gowthCtx *BaseGrowthContext) GetGrowthType() GrowthType {
	return gowthCtx.growthType
}

func (gowthCtx *BaseGrowthContext) GetGrowthIndex() int {
	return gowthCtx.growthIndex
}

func (gowthCtx *BaseGrowthContext) GetBaseTrioDetails(mainPoint Point, offset int) *TrioDetails {
	return GetTrioDetails(gowthCtx.GetBaseTrioIndex(gowthCtx.GetBaseDivByThree(mainPoint), offset))
}

func (gowthCtx *BaseGrowthContext) GetBaseDivByThree(mainPoint Point) uint64 {
	if !mainPoint.IsMainPoint() {
		Log.Fatalf("cannot ask for trio index on non nextMainPoint Pos %v in context %v!", mainPoint, gowthCtx.String())
	}
	return uint64(AbsDIntFromC(mainPoint[0])/3 + AbsDIntFromC(mainPoint[1])/3 + AbsDIntFromC(mainPoint[2])/3)
}

func (gowthCtx *BaseGrowthContext) GetBaseTrioIndex(divByThree uint64, offset int) TrioIndex {
	ctxTrIdx := TrioIndex(gowthCtx.growthIndex)
	if gowthCtx.growthType == 1 {
		// Always same value
		return ctxTrIdx
	}
	if gowthCtx.growthType == 3 {
		// Center on trio index ctx.GetGrowthIndex() and then use X, Y, Z where conn are 1
		mod2 := PosMod2(divByThree)
		if mod2 == 0 {
			return ctxTrIdx
		}
		mod3 := int(((divByThree-1)/2 + uint64(offset)) % 3)
		if gowthCtx.growthIndex < 4 {
			return TrioIndex(validNextTrio[3*gowthCtx.growthIndex+mod3][1])
		}
		count := 0
		for _, validTrio := range validNextTrio {
			if validTrio[1] == ctxTrIdx {
				if count == mod3 {
					return validTrio[0]
				}
				count++
			}
		}
		Log.Fatalf("did not find valid trio for div by three value %d in context %s-%d!", divByThree, gowthCtx.String(), offset)
	}

	divByThreeWithOffset := uint64(offset) + divByThree
	switch gowthCtx.growthType {
	case 2:
		permutationMap := validNextTrio[gowthCtx.growthIndex]
		idx := int(PosMod2(divByThreeWithOffset))
		return permutationMap[idx]
	case 4:
		permutationMap := AllMod4Permutations[gowthCtx.growthIndex]
		idx := int(PosMod4(divByThreeWithOffset))
		return permutationMap[idx]
	case 8:
		permutationMap := AllMod8Permutations[gowthCtx.growthIndex]
		idx := int(PosMod8(divByThreeWithOffset))
		return permutationMap[idx]
	}
	Log.Fatalf("event permutation type %d in context %s-%d is invalid!", gowthCtx.growthIndex, gowthCtx.String(), offset)
	return NilTrioIndex
}
