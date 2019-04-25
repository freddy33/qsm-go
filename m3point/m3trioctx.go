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
		tlToFind = makeTrioLinkFromInt(trCtx.ctxIndex, trCtx.ctxIndex, trCtx.ctxIndex)
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
		toAdd1 := makeTrioLink(a, b, b)
		possLinks.addUnique(&toAdd1)
		toAdd2 := makeTrioLink(a, b, c)
		possLinks.addUnique(&toAdd2)
		toAdd3 := makeTrioLink(b, a, a)
		possLinks.addUnique(&toAdd3)
		toAdd4 := makeTrioLink(b, a, c)
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

// TODO: Change returned value to uint8
func (trCtx *TrioIndexContext) GetBaseTrioIndex(divByThree uint64, offset int) TrioIndex {
	ctxTrIdx := TrioIndex(trCtx.ctxIndex)
	if trCtx.ctxType == 1 {
		// Always same value
		return ctxTrIdx
	}
	if trCtx.ctxType == 3 {
		// Center on Trio index ctx.GetIndex() and then use X, Y, Z where conn are 1
		mod2 := PosMod2(divByThree)
		if mod2 == 0 {
			return ctxTrIdx
		}
		mod3 := int(((divByThree-1)/2 + uint64(offset)) % 3)
		if trCtx.ctxIndex < 4 {
			return TrioIndex(validNextTrio[3*trCtx.ctxIndex+mod3][1])
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

var AssertState = true

func (trCtx *TrioIndexContext) GetNextTrio(mainPoint Point, trioDetails *TrioDetails, connId ConnectionId) TrioIndex {
	if AssertState {
		// mainPoint should be main
		if !mainPoint.IsMainPoint() {
			Log.Errorf("in context %s current point %v is not main, while looking on %s for %s", trCtx.String(), mainPoint, trioDetails.String(), connId.String())
			return NilTrioIndex
		}
	}
	return NilTrioIndex
}

func (trCtx *TrioIndexContext) GetNextTrios(current Point, currentTrioIdx TrioIndex, fromConnId ConnectionId) (nextConnIds [2]ConnectionId, nextTrios [2]TrioIndex) {
	possibleTrios := *trCtx.GetPossibleTrioList()

	td := GetTrioDetails(currentTrioIdx)
	oc := td.OtherConnectionsFrom(-fromConnId)

	for i := 0; i < 2; i++ {
		nextConnIds[i] = oc[i].Id
		np := current.Add(oc[i].Vector)
		if np.IsMainPoint() {
			nextTrios[i] = trCtx.GetBaseTrioIndex(trCtx.GetBaseDivByThree(np), 0)
		} else {
			// Find trio by finding conn id for next point np
			// First connId where np came from
			nextFrommConn := oc[i]
			nextFromConnId := -(nextFrommConn.Id)
			// Second harder to find
			var nextConnId ConnectionId
			// need to find the next next point
			var nnp Point

			if current.IsMainPoint() {
				// So td is a base vector
				if !td.IsBaseTrio() {
					Log.Errorf("current point %v is main and trio associated %s is not base", current, td.String())
				}
				// So oc[i] is a base connection and so has 2 extensions from +X, -X, +Y, -Y, +Z, -Z
				x := nextFrommConn.Vector.X()
				if x != 0 {
					// Use X
					// next point should connect to next main point toward X + base vector there
					var nextMain Point
					if x > 0 {
						nextMain = current.Add(XFirst)
					} else {
						nextMain = current.Sub(XFirst)
					}
					nextMainTrioIdx := trCtx.GetBaseTrioIndex(trCtx.GetBaseDivByThree(nextMain), 0)
					nextMainTd := GetTrioDetails(nextMainTrioIdx)
					if x > 0 {
						// TODO: Should be direct from td
						nnp = nextMain.Sub(nextMainTd.GetTrio().getMinusXVector())
					} else {
						nnp = nextMain.Add(nextMainTd.GetTrio().getPlusXVector())
					}
				} else {
					// Use Y
					y := nextFrommConn.Vector.Y()
					if y == 0 {
						Log.Errorf("something wrong with base vector %v", nextFrommConn.String())
					}
					// next point should connect to next main point toward Y + base vector there
					var nextMain Point
					if y > 0 {
						nextMain = current.Add(YFirst)
					} else {
						nextMain = current.Sub(YFirst)
					}
					nextMainTrioIdx := trCtx.GetBaseTrioIndex(trCtx.GetBaseDivByThree(nextMain), 0)
					nextMainTd := GetTrioDetails(nextMainTrioIdx)
					if y > 0 {
						// TODO: Should be direct from td
						nnp = nextMain.Sub(nextMainTd.GetTrio().getMinusYVector())
					} else {
						nnp = nextMain.Add(nextMainTd.GetTrio().getPlusYVector())
					}
				}
			} else {
				// If current is not main, than np is close to the next main point. Get the nearest main point of np is where it goes
				// TODO: Verify above assumption is correct!
				nnp = np.GetNearMainPoint()
			}

			nextConnId = GetConnDetailsByPoints(np, nnp).Id
			// We have 2 connId let's see if we find one (and only one) in the list of possible trios
			solutions := make([]TrioIndex, 0, 1)
			for _, possTr := range possibleTrios {
				if possTr.HasConnections(nextConnId, nextFromConnId) {
					solutions = append(solutions, possTr.GetId())
				}
			}
			if len(solutions) == 0 {
				Log.Errorf("did not find for context %s any trio with %s and %s in %v", trCtx.String(), nextConnId.String(), nextFromConnId.String(), possibleTrios)
			}
			if len(solutions) > 1 {
				Log.Errorf("found more than one for context %s trio %v with %s and %s in %v", trCtx.String(), solutions, nextConnId.String(), nextFromConnId.String(), possibleTrios)
			}
			nextTrios[i] = solutions[0]
		}
	}
	return
}

// Stupid reverse engineering of trio index that works for main and non main points
func FindTrioIndex(c Point, np [3]Point, ctx *TrioIndexContext, offset int) (TrioIndex, TrioLink) {
	link := makeTrioLink(getTrioIdxNearestMain(c, ctx, offset), getTrioIdxNearestMain(np[1], ctx, offset), getTrioIdxNearestMain(np[2], ctx, offset))
	toFind := MakeTrioDetails(MakeVector(c, np[0]), MakeVector(c, np[1]), MakeVector(c, np[2]))
	for _, td := range allTrioDetails {
		if toFind.GetTrio() == td.GetTrio() {
			return td.id, link
		}
	}
	Log.Errorf("did not find any trio for %v %v %v", c, np, toFind)
	Log.Errorf("All trio index %s", link.String())
	return NilTrioIndex, link
}

func getTrioIdxNearestMain(p Point, ctx *TrioIndexContext, offset int) TrioIndex {
	return ctx.GetBaseTrioIndex(ctx.GetBaseDivByThree(p.GetNearMainPoint()), offset)
}
