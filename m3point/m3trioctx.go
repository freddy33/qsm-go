package m3point

import (
	"fmt"
	"sort"
)

/*
Define how outgrowth and path evolve from the center. There are 6 types of growth depending of the value of ctxType:
TODO: Create trio index for non main points base on growth context
1. type = 0 : Type not yet existing TODO: Main points will not be covered. In here trio index switch from trio to next that has neg conn
2. type = 1 : All main points have the same base trio index
3. type = 3 : Rotate between valid trios depending on starting index in modulo 3
4. type = 2 : Use the modulo 2 permutation => Specific index valid next trio back and forth
5. type = 4 : Use the modulo 4 permutation => Specific index line in AllMod4Permutations cycling through the 4 values
6. type = 8 : Use the modulo 8 permutation => Specific index line in AllMod8Permutations cycling through the 8 values
*/
type ContextType uint8

var allContextTypes = [5]ContextType{1, 2, 3, 4, 8}

type TrioIndexContext struct {
	ctxType ContextType
	// Index in the permutations to choose from. For type 1 and 3 [0,7] for the other in the 12 list [0,11]
	// Max number of indexes returned by ContextType.GetNbIndexes()
	ctxIndex int
}

var trioIndexContexts map[ContextType][]*TrioIndexContext
var trioDetailsPerContext map[ContextType][]*TrioDetailList

func init() {
	count := make(map[ContextType]int)
	trioIndexContexts = make(map[ContextType][]*TrioIndexContext)
	for _, ctxType := range GetAllContextTypes() {
		nbIndexes := ctxType.GetNbIndexes()
		trioIndexContexts[ctxType] = make([]*TrioIndexContext, nbIndexes)
		for pIdx := 0; pIdx < nbIndexes; pIdx++ {
			trioIndexContexts[ctxType][pIdx] = createTrioIndexContext(ctxType, pIdx)
			count[ctxType]++
		}
	}
	trioDetailsPerContext = make(map[ContextType][]*TrioDetailList)
	Log.Debug(count)
}

/***************************************************************/
// ContextType Functions
/***************************************************************/
func GetAllContextTypes() [5]ContextType {
	return allContextTypes
}

func (t ContextType) String() string {
	return fmt.Sprintf("CtxType%d", t)
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

/***************************************************************/
// TrioIndexContext Functions
/***************************************************************/

func GetTrioIndexContext(ctxType ContextType, index int) *TrioIndexContext {
	return trioIndexContexts[ctxType][index]
}

func createTrioIndexContext(permType ContextType, index int) *TrioIndexContext {
	return &TrioIndexContext{permType, index,}
}

func (trCtx *TrioIndexContext) String() string {
	return fmt.Sprintf("TrioCtx%d-Idx%02d", trCtx.ctxType, trCtx.ctxIndex)
}

func (trCtx *TrioIndexContext) GetType() ContextType {
	return trCtx.ctxType
}

func (trCtx *TrioIndexContext) GetIndex() int {
	return trCtx.ctxIndex
}

func (trCtx *TrioIndexContext) SetIndex(idx int) {
	trCtx.ctxIndex = idx
}

func (trCtx *TrioIndexContext) GetPossibleTrioList() *TrioDetailList {
	var r *TrioDetailList
	l, ok := trioDetailsPerContext[trCtx.ctxType]
	if !ok {
		l = make([]*TrioDetailList, trCtx.ctxType.GetNbIndexes())
		trioDetailsPerContext[trCtx.ctxType] = l
	}

	r = l[trCtx.ctxIndex]
	if r != nil {
		return r
	}
	r = trCtx.makePossibleTrioList()
	l[trCtx.ctxIndex] = r
	return r
}

func (trCtx *TrioIndexContext) makePossibleTrioList() *TrioDetailList {
	res := TrioDetailList{}
	var tlToFind TrioLink
	if trCtx.ctxType == 1 {
		// Always same index so all details where links are this
		tlToFind = MakeTrioLink(trCtx.ctxIndex,trCtx.ctxIndex,trCtx.ctxIndex)
		for i, td := range allTrioDetails {
			if td.Links.Exists(&tlToFind) {
				// Add the pointer from list not from iteration
				toAdd := allTrioDetails[i]
				res.addUnique(toAdd)
			}
		}
		sort.Sort(&res)
		return &res
	}

	// Now use context to create possible trio links
	possLinks := TrioLinkList{}
	for divByThree := uint64(1); divByThree < 23; divByThree++ {
		a := trCtx.GetBaseTrioIndex(divByThree-1, 0)
		b := trCtx.GetBaseTrioIndex(divByThree, 0)
		c := trCtx.GetBaseTrioIndex(divByThree+1, 0)
		toAdd1 := MakeTrioLink(a, b, b)
		possLinks.addUnique(&toAdd1)
		toAdd2 := MakeTrioLink(a, b, c)
		possLinks.addUnique(&toAdd2)
		toAdd3 := MakeTrioLink(b, a, a)
		possLinks.addUnique(&toAdd3)
		toAdd4 := MakeTrioLink(b, a, c)
		possLinks.addUnique(&toAdd4)
	}

	// Now extract all trio details associated with given links
	for _, tl := range possLinks {
		for i, td := range allTrioDetails {
			if td.Links.Exists(tl) {
				// Add the pointer from list not from iteration
				res.addUnique(allTrioDetails[i])
			}
		}
	}

	sort.Sort(&res)

	return &res
}

func (trCtx *TrioIndexContext) GetBaseTrio(p Point, offset int) Trio {
	return GetBaseTrio(trCtx.GetBaseTrioIndex(trCtx.GetBaseDivByThree(p), offset))
}

func (trCtx *TrioIndexContext) GetBaseDivByThree(p Point) uint64 {
	if !p.IsMainPoint() {
		panic(fmt.Sprintf("cannot ask for Trio index on non main Pos %v in context %v!", p, trCtx.String()))
	}
	return uint64(Abs64(p[0])/3 + Abs64(p[1])/3 + Abs64(p[2])/3)
}

func (trCtx *TrioIndexContext) GetBaseTrioIndex(divByThree uint64, offset int) int {
	if trCtx.ctxType == 1 {
		// Always same value
		return trCtx.ctxIndex
	}
	if trCtx.ctxType == 3 {
		// Center on Trio index ctx.GetIndex() and then use X, Y, Z where conn are 1
		mod2 := PosMod2(divByThree)
		if mod2 == 0 {
			return trCtx.ctxIndex
		}
		mod3 := int(((divByThree-1)/2 + uint64(offset)) % 3)
		if trCtx.ctxIndex < 4 {
			return validNextTrio[3*trCtx.ctxIndex+mod3][1]
		}
		count := 0
		for _, validTrio := range validNextTrio {
			if validTrio[1] == trCtx.ctxIndex {
				if count == mod3 {
					return validTrio[0]
				}
				count++
			}
		}
		panic(fmt.Sprintf("did not find valid Trio for div by three value %d in context %v!", divByThree, trCtx.String()))
	}

	divByThreeWithOffset := uint64(offset) + divByThree
	switch trCtx.ctxType {
	case 2:
		permutationMap := validNextTrio[trCtx.ctxIndex]
		idx := int(PosMod2(divByThreeWithOffset))
		return permutationMap[idx]
	case 4:
		permutationMap := AllMod4Permutations[trCtx.ctxIndex]
		idx := int(PosMod4(divByThreeWithOffset))
		return permutationMap[idx]
	case 8:
		permutationMap := AllMod8Permutations[trCtx.ctxIndex]
		idx := int(PosMod8(divByThreeWithOffset))
		return permutationMap[idx]
	}
	panic(fmt.Sprintf("event permutation type %d in context %v is invalid!", trCtx.ctxIndex, trCtx.String()))
}

// Stupid reverse engineering of trio index that works for main and non main points
func FindTrioIndex(c Point, np [3]Point, ctx *TrioIndexContext, offset int) (int, TrioLink) {
	link := MakeTrioLink(getTrioIdxNearestMain(c, ctx, offset), getTrioIdxNearestMain(np[1], ctx, offset), getTrioIdxNearestMain(np[2], ctx, offset))
	toFind := MakeTrioDetails(MakeVector(c, np[0]), MakeVector(c, np[1]), MakeVector(c, np[2]))
	for trIdx, td := range allTrioDetails {
		if toFind.GetTrio() == td.GetTrio() {
			return trIdx, link
		}
	}
	Log.Errorf("did not find any trio for %v %v %v", c, np, toFind)
	Log.Errorf("All trio index %s", link.String())
	return -1, link
}

func getTrioIdxNearestMain(p Point, ctx *TrioIndexContext, offset int) int {
	return ctx.GetBaseTrioIndex(ctx.GetBaseDivByThree(p.GetNearMainPoint()), offset)
}

