package m3point

import "fmt"

type CubeKey struct {
	// Index of cube center
	center TrioIndex
	// Indexes of center of cube face ordered by +X, -X, +Y, -Y, +Z, -Z
	centerFaces [6]TrioIndex
	// Indexes of middle of edges of the cube ordered by +X+Y, +X-Y, +X+Z, +X-Z, -X+Y, -X-Y, -X+Z, -X-Z, +Y+Z, +Y-Z, -Y+Z, -Y-Z
	middleEdges [12]TrioIndex
}

type CubeListPerContext struct {
	trCtx    *TrioContext
	allCubes []CubeKey
}

var allTrioCubes [][]*CubeListPerContext

/***************************************************************/
// CubeKey Functions
/***************************************************************/

func GetMiddleEdgeIndex(ud1 UnitDirection, ud2 UnitDirection) int {
	if ud1 == ud2 {
		Log.Fatalf("Cannot find middle edge for 2 identical unit direction %d", ud1)
		return -1
	}
	// Order the 2
	if ud1 > ud2 {
		ud1, ud2 = ud2, ud1
	}
	if ud1%2 == 0 && ud1 == ud2-1 {
		Log.Fatalf("Cannot find middle edge for unit directions %d and %d since they are on same axis", ud1, ud2)
		return -1
	}
	switch ud1 {
	case PlusX:
		return int(ud2 - 2)
	case MinusX:
		return int(4 + ud2 - 2)
	case PlusY:
		return int(8 + ud2 - 4)
	case MinusY:
		return int(8 + 2 + ud2 - 4)
	}
	Log.Fatalf("Cannot find middle edge for unit directions %d and %d since they are incoherent", ud1, ud2)
	return -1
}

// Fill all the indexes assuming the distance of c from origin used in div by three
func createTrioCube(trCtx *TrioContext, offset int, c Point) CubeKey {
	res := CubeKey{}
	res.center = trCtx.GetBaseTrioIndex(trCtx.GetBaseDivByThree(c), offset)

	res.centerFaces[PlusX] = trCtx.GetBaseTrioIndex(trCtx.GetBaseDivByThree(c.Add(XFirst)), offset)
	res.centerFaces[MinusX] = trCtx.GetBaseTrioIndex(trCtx.GetBaseDivByThree(c.Sub(XFirst)), offset)
	res.centerFaces[PlusY] = trCtx.GetBaseTrioIndex(trCtx.GetBaseDivByThree(c.Add(YFirst)), offset)
	res.centerFaces[MinusY] = trCtx.GetBaseTrioIndex(trCtx.GetBaseDivByThree(c.Sub(YFirst)), offset)
	res.centerFaces[PlusZ] = trCtx.GetBaseTrioIndex(trCtx.GetBaseDivByThree(c.Add(ZFirst)), offset)
	res.centerFaces[MinusZ] = trCtx.GetBaseTrioIndex(trCtx.GetBaseDivByThree(c.Sub(ZFirst)), offset)

	res.middleEdges[GetMiddleEdgeIndex(PlusX, PlusY)] = trCtx.GetBaseTrioIndex(trCtx.GetBaseDivByThree(c.Add(XFirst).Add(YFirst)), offset)
	res.middleEdges[GetMiddleEdgeIndex(PlusX, MinusY)] = trCtx.GetBaseTrioIndex(trCtx.GetBaseDivByThree(c.Add(XFirst).Sub(YFirst)), offset)
	res.middleEdges[GetMiddleEdgeIndex(PlusX, PlusZ)] = trCtx.GetBaseTrioIndex(trCtx.GetBaseDivByThree(c.Add(XFirst).Add(ZFirst)), offset)
	res.middleEdges[GetMiddleEdgeIndex(PlusX, MinusZ)] = trCtx.GetBaseTrioIndex(trCtx.GetBaseDivByThree(c.Add(XFirst).Sub(ZFirst)), offset)

	res.middleEdges[GetMiddleEdgeIndex(MinusX, PlusY)] = trCtx.GetBaseTrioIndex(trCtx.GetBaseDivByThree(c.Sub(XFirst).Add(YFirst)), offset)
	res.middleEdges[GetMiddleEdgeIndex(MinusX, MinusY)] = trCtx.GetBaseTrioIndex(trCtx.GetBaseDivByThree(c.Sub(XFirst).Sub(YFirst)), offset)
	res.middleEdges[GetMiddleEdgeIndex(MinusX, PlusZ)] = trCtx.GetBaseTrioIndex(trCtx.GetBaseDivByThree(c.Sub(XFirst).Add(ZFirst)), offset)
	res.middleEdges[GetMiddleEdgeIndex(MinusX, MinusZ)] = trCtx.GetBaseTrioIndex(trCtx.GetBaseDivByThree(c.Sub(XFirst).Sub(ZFirst)), offset)

	res.middleEdges[GetMiddleEdgeIndex(PlusY, PlusZ)] = trCtx.GetBaseTrioIndex(trCtx.GetBaseDivByThree(c.Add(YFirst).Add(ZFirst)), offset)
	res.middleEdges[GetMiddleEdgeIndex(PlusY, MinusZ)] = trCtx.GetBaseTrioIndex(trCtx.GetBaseDivByThree(c.Add(YFirst).Sub(ZFirst)), offset)
	res.middleEdges[GetMiddleEdgeIndex(MinusY, PlusZ)] = trCtx.GetBaseTrioIndex(trCtx.GetBaseDivByThree(c.Sub(YFirst).Add(ZFirst)), offset)
	res.middleEdges[GetMiddleEdgeIndex(MinusY, MinusZ)] = trCtx.GetBaseTrioIndex(trCtx.GetBaseDivByThree(c.Sub(YFirst).Sub(ZFirst)), offset)

	return res
}

func (ck CubeKey) String() string {
	return fmt.Sprintf("CK-%s-%s-%s-%s", ck.center.String(), ck.centerFaces[0].String(), ck.centerFaces[1].String(), ck.middleEdges[0].String())
}

func (ck CubeKey) GetCenterTrio() TrioIndex {
	return ck.center
}

func (ck CubeKey) GetCenterFaceTrio(ud UnitDirection) TrioIndex {
	return ck.centerFaces[ud]
}

func (ck CubeKey) GetMiddleEdgeTrio(ud1 UnitDirection, ud2 UnitDirection) TrioIndex {
	return ck.middleEdges[GetMiddleEdgeIndex(ud1, ud2)]
}

/***************************************************************/
// CubeListPerContext Functions
/***************************************************************/

func createAllTrioCubes() {
	if len(allTrioCubes) != 0 {
		// done
		return
	}
	allTrioCubes = make([][]*CubeListPerContext, 9)
	for _, ctxType := range GetAllContextTypes() {
		nbIndexes := ctxType.GetNbIndexes()
		allTrioCubes[ctxType] = make([]*CubeListPerContext, nbIndexes)
		for pIdx := 0; pIdx < nbIndexes; pIdx++ {
			trCtx := GetTrioIndexContext(ctxType, pIdx)
			cl := CubeListPerContext{trCtx, nil,}
			switch ctxType {
			case 1:
				cl.populate(1)
			case 3:
				cl.populate(6)
			case 2:
				cl.populate(1)
			case 4:
				cl.populate(4)
			case 8:
				cl.populate(8)
			}
			allTrioCubes[ctxType][pIdx] = &cl
		}
	}
}

func (cl *CubeListPerContext) populate(max CInt) {
	allCubesMap := make(map[CubeKey]int)
	// For center populate for all offsets
	maxOffset := cl.trCtx.ctxType.GetMaxOffset()
	for offset := 0; offset < maxOffset; offset++ {
		cube := createTrioCube(cl.trCtx, offset, Origin)
		allCubesMap[cube]++
	}
	// Go through space
	for x := CInt(-max); x <= max; x++ {
		for y := CInt(-max); y <= max; y++ {
			for z := CInt(-max); z <= max; z++ {
				cube := createTrioCube(cl.trCtx, 0, Point{x, y, z}.Mul(THREE))
				allCubesMap[cube]++
			}
		}
	}
	cl.allCubes = make([]CubeKey, len(allCubesMap))
	idx := 0
	for c := range allCubesMap {
		cl.allCubes[idx] = c
		idx++
	}
}

func (cl *CubeListPerContext) exists(offset int, c Point) bool {
	toFind := createTrioCube(cl.trCtx, offset, c)
	for _, c := range cl.allCubes {
		if c == toFind {
			return true
		}
	}
	return false
}

func GetCubeList(ctxType ContextType, index int) *CubeListPerContext {
	createAllTrioCubes()
	return allTrioCubes[ctxType][index]
}
