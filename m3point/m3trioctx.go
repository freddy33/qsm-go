package m3point

import "fmt"

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
	ctxType  ContextType
	// Index in the permutations to choose from. For type 1 and 3 [0,7] for the other in the 12 list [0,11]
	// Max number of indexes returned by ContextType.GetNbIndexes()
	ctxIndex int
}

var trioIndexContexts map[ContextType][]*TrioIndexContext

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
	Log.Info(count)
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
	return &TrioIndexContext{permType, index, }
}

func (trCtx *TrioIndexContext) String() string {
	return fmt.Sprintf("TrioCtx%d-Idx%02d", trCtx.ctxType, trCtx.ctxIndex)
}

