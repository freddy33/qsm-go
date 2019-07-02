package m3point

import "fmt"

type CubeOfTrioIndex struct {
	// Index of cube center
	center TrioIndex
	// Indexes of center of cube face ordered by +X, -X, +Y, -Y, +Z, -Z
	centerFaces [6]TrioIndex
	// Indexes of middle of edges of the cube ordered by +X+Y, +X-Y, +X+Z, +X-Z, -X+Y, -X-Y, -X+Z, -X-Z, +Y+Z, +Y-Z, -Y+Z, -Y-Z
	middleEdges [12]TrioIndex
}

type CubeKeyId struct {
	trCtxId int
	cube    CubeOfTrioIndex
}

type CubeListBuilder struct {
	trCtx    *TrioContext
	allCubes []CubeOfTrioIndex
}

const (
	TotalNumberOfCubes = 5192
)

var cubeIdsPerKey map[CubeKeyId]int

/***************************************************************/
// CubeOfTrioIndex Functions
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
func createTrioCube(trCtx *TrioContext, offset int, c Point) CubeOfTrioIndex {
	res := CubeOfTrioIndex{}
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

func (cube CubeOfTrioIndex) String() string {
	return fmt.Sprintf("CK-%s-%s-%s-%s", cube.center.String(), cube.centerFaces[0].String(), cube.centerFaces[1].String(), cube.middleEdges[0].String())
}

func (cube CubeOfTrioIndex) GetCenterTrio() TrioIndex {
	return cube.center
}

func (cube CubeOfTrioIndex) GetCenterFaceTrio(ud UnitDirection) TrioIndex {
	return cube.centerFaces[ud]
}

func (cube CubeOfTrioIndex) GetMiddleEdgeTrio(ud1 UnitDirection, ud2 UnitDirection) TrioIndex {
	return cube.middleEdges[GetMiddleEdgeIndex(ud1, ud2)]
}

/***************************************************************/
// CubeListBuilder Functions
/***************************************************************/

func calculateAllContextCubes() map[CubeKeyId]int {
	res := make(map[CubeKeyId]int, TotalNumberOfCubes)
	cubeIdx := 1
	for _, trCtx := range GetAllTrioContexts() {
		cl := CubeListBuilder{trCtx, nil,}
		switch trCtx.GetType() {
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
		for _, cube := range cl.allCubes {
			key := CubeKeyId{trCtx.id, cube}
			res[key] = cubeIdx
			cubeIdx++
		}
	}
	return res
}

func (cl *CubeListBuilder) populate(max CInt) {
	allCubesMap := make(map[CubeOfTrioIndex]int)
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
	cl.allCubes = make([]CubeOfTrioIndex, len(allCubesMap))
	idx := 0
	for c := range allCubesMap {
		cl.allCubes[idx] = c
		idx++
	}
}

func (cl *CubeListBuilder) exists(offset int, c Point) bool {
	toFind := createTrioCube(cl.trCtx, offset, c)
	for _, c := range cl.allCubes {
		if c == toFind {
			return true
		}
	}
	return false
}

func GetCubeList(trCtx *TrioContext) *CubeListBuilder {
	checkCubesInitialized()
	res := CubeListBuilder{}
	res.trCtx = trCtx
	res.allCubes = make([]CubeOfTrioIndex, 0, 100)
	for cubeKey := range cubeIdsPerKey {
		if cubeKey.trCtxId == trCtx.GetId() {
			res.allCubes = append(res.allCubes, cubeKey.cube)
		}
	}
	return &res
}

func GetCubeById(cubeId int) CubeKeyId {
	checkCubesInitialized()
	for cubeKey, id := range cubeIdsPerKey {
		if id == cubeId {
			return cubeKey
		}
	}
	Log.Errorf("trying to find cube by id %d which does not exists", cubeId)
	return CubeKeyId{-1, CubeOfTrioIndex{}}
}

func GetCubeIdByKey(cubeKey CubeKeyId) int {
	checkCubesInitialized()
	id, ok := cubeIdsPerKey[cubeKey]
	if !ok {
		Log.Errorf("trying to find cube %v which does not exists", cubeKey)
		return -1
	}
	return id
}
