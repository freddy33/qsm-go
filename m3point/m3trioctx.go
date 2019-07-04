package m3point

import (
	"fmt"
)

/*
Define how outgrowth and path evolve from the center. There are 6 types of growth depending of the value of ctxType:
TODO: Create trio index for non nextMainPoint points base on growth context
1. type = 0 : Type not yet existing TODO: Main points will not be covered. In here trio index switch from trio to next that has neg conn
2. type = 1 : All nextMainPoint points have the same base trio index
3. type = 3 : Rotate between valid trios depending on starting index in modulo 3
4. type = 2 : Use the modulo 2 permutation => Specific index valid next trio back and forth
5. type = 4 : Use the modulo 4 permutation => Specific index line in AllMod4Permutations cycling through the 4 values
6. type = 8 : Use the modulo 8 permutation => Specific index line in AllMod8Permutations cycling through the 8 values
*/
type ContextType uint8

var allContextTypes = [5]ContextType{1, 2, 3, 4, 8}
var totalNbContexts = 8 + 12 + 8 + 12 + 12

type TrioContext struct {
	// A generate id used in arrays and db
	id int
	// The context type for this flow context
	ctxType ContextType
	// Index in the permutations to choose from. For type 1 and 3 [0,7] for the other in the 12 list [0,11]
	// Max number of indexes returned by ContextType.GetNbIndexes()
	ctxIndex int
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

var allTrioContexts []*TrioContext

func calculateAllTrioContexts() []*TrioContext {
	res := make([]*TrioContext, totalNbContexts)
	idx := 0
	for _, ctxType := range GetAllContextTypes() {
		nbIndexes := ctxType.GetNbIndexes()
		for pIdx := 0; pIdx < nbIndexes; pIdx++ {
			trCtx := TrioContext{idx, ctxType, pIdx}
			res[idx] = &trCtx
			idx++
		}
	}
	return res
}

/***************************************************************/
// ContextType Functions
/***************************************************************/
func GetTotalNbTrioContexts() int {
	return totalNbContexts
}

func GetAllTrioContexts() []*TrioContext {
	checkTrioContextsInitialized()
	return allTrioContexts
}

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

func (t ContextType) GetMaxOffset() int {
	return maxOffsetPerType[t]
}

/***************************************************************/
// TrioContext Functions
/***************************************************************/

func GetTrioContextById(id int) *TrioContext {
	checkTrioContextsInitialized()
	return allTrioContexts[id]
}

func GetTrioContextByTypeAndIdx(ctxType ContextType, index int) *TrioContext {
	checkTrioContextsInitialized()
	for _, trCtx := range allTrioContexts {
		if trCtx.ctxType == ctxType && trCtx.ctxIndex == index {
			return trCtx
		}
	}
	Log.Fatalf("could not find Trio Context for %d %d", ctxType, index)
	return nil
}

func (trCtx *TrioContext) String() string {
	return fmt.Sprintf("TrioCtx%d-%d-Idx%02d", trCtx.id, trCtx.ctxType, trCtx.ctxIndex)
}

func (trCtx *TrioContext) GetId() int {
	return trCtx.id
}

func (trCtx *TrioContext) GetType() ContextType {
	return trCtx.ctxType
}

func (trCtx *TrioContext) GetIndex() int {
	return trCtx.ctxIndex
}

func (trCtx *TrioContext) SetIndex(idx int) {
	trCtx.ctxIndex = idx
}

func (trCtx *TrioContext) GetBaseTrioDetails(mainPoint Point, offset int) *TrioDetails {
	return GetTrioDetails(trCtx.GetBaseTrioIndex(trCtx.GetBaseDivByThree(mainPoint), offset))
}

func (trCtx *TrioContext) GetBaseTrio(mainPoint Point, offset int) Trio {
	return GetBaseTrio(trCtx.GetBaseTrioIndex(trCtx.GetBaseDivByThree(mainPoint), offset))
}

func (trCtx *TrioContext) GetBaseDivByThree(mainPoint Point) uint64 {
	if !mainPoint.IsMainPoint() {
		Log.Fatalf("cannot ask for Trio index on non nextMainPoint Pos %v in context %v!", mainPoint, trCtx.String())
	}
	return uint64(AbsDIntFromC(mainPoint[0])/3 + AbsDIntFromC(mainPoint[1])/3 + AbsDIntFromC(mainPoint[2])/3)
}

func (trCtx *TrioContext) GetBaseTrioIndex(divByThree uint64, offset int) TrioIndex {
	ctxTrIdx := TrioIndex(trCtx.ctxIndex)
	if trCtx.ctxType == 1 {
		// Always same value
		return ctxTrIdx
	}
	if trCtx.ctxType == 3 {
		// Center on Trio index ctx.GetTrioContextIndex() and then use X, Y, Z where conn are 1
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
		Log.Fatalf("did not find valid Trio for div by three value %d in context %s-%d!", divByThree, trCtx.String(), offset)
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
	Log.Fatalf("event permutation type %d in context %s-%d is invalid!", trCtx.ctxIndex, trCtx.String(), offset)
	return NilTrioIndex
}

// Out of a nextMainPoint point with the given trio details, what is the trio details of the point at the end of connection connId
// npes: The next path elements saved during calculation and returned in this method
func (trCtx *TrioContext) GetForwardTrioFromMain(mainPoint Point, trioDetails *TrioDetails, connId ConnectionId, offset int) (p Point, td *TrioDetails, npes [2]*NextPathElement) {

	Initialize()

	p = Origin
	if Log.DoAssert() {
		// mainPoint should be nextMainPoint
		if !mainPoint.IsMainPoint() {
			Log.Errorf("in context %s current point %v is not nextMainPoint, while looking on %s for %s", trCtx.String(), mainPoint, trioDetails.String(), connId.String())
			return
		}
		// Trio Details should have the connection connId given
		if !trioDetails.HasConnection(connId) {
			Log.Errorf("in context %s-%d trio details %s, does not have the given connection %s", trCtx.String(), offset, trioDetails.String(), connId.String())
			return
		}
		// The trio details index should be the one of this context
		indexFromContext := trCtx.GetBaseTrioDetails(mainPoint, offset).id
		if indexFromContext != trioDetails.id {
			Log.Errorf("in context %s-%d current point %v has a trio index %d from context, which not the one in %s", trCtx.String(), offset, mainPoint, indexFromContext, trioDetails.String())
			return
		}
	}
	// The connection details from nextMainPoint point
	cd := GetConnDetailsById(connId)
	// The actual point that we work on
	cVec := cd.Vector
	p = mainPoint.Add(cVec)

	// We calculate part of the path out of a nextMainPoint point in one go and the output will be PathId for all the way to next nextMainPoint points
	// nmp and nip are create out of which of the 6 connections possible +X, -X, +Y, -Y, +Z, -Z vectors
	cNpe := makeNewNpe(offset)

	idx := 0
	switch cVec.X() {
	case 0:
		// Nothing connect
	case 1:
		cNpe.fillPlusX(trCtx, mainPoint)
	case -1:
		cNpe.fillMinusX(trCtx, mainPoint)
	}
	if cNpe.IsValid() {
		npes[idx] = cNpe
		idx++
		cNpe = makeNewNpe(offset)
	}

	switch cVec.Y() {
	case 0:
		// Nothing connect
	case 1:
		cNpe.fillPlusY(trCtx, mainPoint)
	case -1:
		cNpe.fillMinusY(trCtx, mainPoint)
	}
	if cNpe.IsValid() {
		npes[idx] = cNpe
		idx++
		cNpe = makeNewNpe(offset)
	}

	switch cVec.Z() {
	case 0:
		// Nothing connect
	case 1:
		cNpe.fillPlusZ(trCtx, mainPoint)
	case -1:
		cNpe.fillMinusZ(trCtx, mainPoint)
	}
	if cNpe.IsValid() {
		npes[idx] = cNpe
		idx++
	}

	// First fill the connection details between p and the npe intermediate points
	for _, npe := range npes {
		npe.p2iConn = GetConnDetailsByPoints(p, npe.ipNearNm)
	}

	// We have all we need to find the actual trio of the point of interest p
	for _, possTd := range allTrioDetails {
		if possTd.HasConnections(-cd.Id, npes[0].p2iConn.Id, npes[1].p2iConn.Id) {
			td = possTd
			break
		}
	}
	if td == nil {
		Log.Errorf("did not find any trio details matching %s %s %s in %s offset %d", -cd.Id, npes[0].p2iConn.Id, npes[1].p2iConn.Id, trCtx.String(), offset)
	}

	return
}

// Find the trio index that apply to the intermediate point near next main point
func (trCtx *TrioContext) GetBackTrioOnInterPoint(npe *NextPathElement) (*TrioDetails, [2]*NextPathElement) {
	// Use the GetForwardTrioFromMain() method on the next main point
	checkIP, resultTD, backNpel := trCtx.GetForwardTrioFromMain(npe.nextMainPoint, npe.nextMainTd, npe.nmp2ipConn.GetId(), npe.offset)
	if checkIP != npe.ipNearNm {
		Log.Errorf("Did not find same point %v != %v in brute force for %s on npe=%v", checkIP, npe.ipNearNm, trCtx.String(), *npe)
		return nil, [2]*NextPathElement{nil, nil}
	}
	return resultTD, backNpel
}

/***************************************************************/
// NextPathElement Functions
/***************************************************************/

func makeNewNpe(offset int) *NextPathElement {
	res := NextPathElement{}
	res.valid = false
	res.offset = offset
	return &res
}

func (npe *NextPathElement) IsValid() bool {
	return npe.valid
}

func (npe *NextPathElement) GetNextMainPoint() Point {
	return npe.nextMainPoint
}

func (npe *NextPathElement) GetNextMainTrioDetails() *TrioDetails {
	return npe.nextMainTd
}

func (npe *NextPathElement) GetNextMainTrioId() TrioIndex {
	return npe.nextMainTd.id
}

func (npe *NextPathElement) GetIntermediatePoint() Point {
	return npe.ipNearNm
}

func (npe *NextPathElement) GetP2IConn() *ConnectionDetails {
	return npe.p2iConn
}

func (npe *NextPathElement) GetNmp2IConn() *ConnectionDetails {
	return npe.nmp2ipConn
}

// This is +X, find the -X on the other side
func (npe *NextPathElement) fillPlusX(trCtx *TrioContext, mainPoint Point) {
	npe.nextMainPoint = mainPoint.Add(XFirst)
	npe.nextMainTd = trCtx.GetBaseTrioDetails(npe.nextMainPoint, npe.offset)
	npe.nmp2ipConn = npe.nextMainTd.getMinusXConn()
	npe.ipNearNm = npe.nextMainPoint.Add(npe.nmp2ipConn.Vector)
	npe.valid = true
}

// This is -X, find the +X on the other side
func (npe *NextPathElement) fillMinusX(trCtx *TrioContext, mainPoint Point) {
	npe.nextMainPoint = mainPoint.Add(XFirst.Neg())
	npe.nextMainTd = trCtx.GetBaseTrioDetails(npe.nextMainPoint, npe.offset)
	npe.nmp2ipConn = npe.nextMainTd.getPlusXConn()
	npe.ipNearNm = npe.nextMainPoint.Add(npe.nmp2ipConn.Vector)
	npe.valid = true
}

// This is +Y, find the -Y on the other side
func (npe *NextPathElement) fillPlusY(trCtx *TrioContext, mainPoint Point) {
	npe.nextMainPoint = mainPoint.Add(YFirst)
	npe.nextMainTd = trCtx.GetBaseTrioDetails(npe.nextMainPoint, npe.offset)
	npe.nmp2ipConn = npe.nextMainTd.getMinusYConn()
	npe.ipNearNm = npe.nextMainPoint.Add(npe.nmp2ipConn.Vector)
	npe.valid = true
}

// This is -Y, find the +Y on the other side
func (npe *NextPathElement) fillMinusY(trCtx *TrioContext, mainPoint Point) {
	npe.nextMainPoint = mainPoint.Add(YFirst.Neg())
	npe.nextMainTd = trCtx.GetBaseTrioDetails(npe.nextMainPoint, npe.offset)
	npe.nmp2ipConn = npe.nextMainTd.getPlusYConn()
	npe.ipNearNm = npe.nextMainPoint.Add(npe.nmp2ipConn.Vector)
	npe.valid = true
}

// This is +Z, find the -Z on the other side
func (npe *NextPathElement) fillPlusZ(trCtx *TrioContext, mainPoint Point) {
	npe.nextMainPoint = mainPoint.Add(ZFirst)
	npe.nextMainTd = trCtx.GetBaseTrioDetails(npe.nextMainPoint, npe.offset)
	npe.nmp2ipConn = npe.nextMainTd.getMinusZConn()
	npe.ipNearNm = npe.nextMainPoint.Add(npe.nmp2ipConn.Vector)
	npe.valid = true
}

// This is -Z, find the +Z on the other side
func (npe *NextPathElement) fillMinusZ(trCtx *TrioContext, mainPoint Point) {
	npe.nextMainPoint = mainPoint.Add(ZFirst.Neg())
	npe.nextMainTd = trCtx.GetBaseTrioDetails(npe.nextMainPoint, npe.offset)
	npe.nmp2ipConn = npe.nextMainTd.getPlusZConn()
	npe.ipNearNm = npe.nextMainPoint.Add(npe.nmp2ipConn.Vector)
	npe.valid = true
}
