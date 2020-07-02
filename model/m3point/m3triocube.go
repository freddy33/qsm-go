package m3point

import (
	"fmt"
)

type CubeOfTrioIndex struct {
	// Index of cube center
	Center TrioIndex
	// Indexes of center of cube face ordered by +X, -X, +Y, -Y, +Z, -Z
	CenterFaces [6]TrioIndex
	// Indexes of middle of edges of the cube ordered by +X+Y, +X-Y, +X+Z, +X-Z, -X+Y, -X-Y, -X+Z, -X-Z, +Y+Z, +Y-Z, -Y+Z, -Y-Z
	MiddleEdges [12]TrioIndex
}

type CubeKeyId struct {
	GrowthCtxId int
	Cube        CubeOfTrioIndex
}

func (c CubeKeyId) GetGrowthCtxId() int {
	return c.GrowthCtxId
}

func (c CubeKeyId) GetCube() CubeOfTrioIndex {
	return c.Cube
}

const (
	TotalNumberOfCubes = 5192
)

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
func (ppd *BasePointPackData) CreateTrioCube(growthCtx GrowthContext, offset int, c Point) CubeOfTrioIndex {
	res := CubeOfTrioIndex{}
	res.Center = growthCtx.GetBaseTrioIndex(ppd, growthCtx.GetBaseDivByThree(c), offset)

	res.CenterFaces[PlusX] = growthCtx.GetBaseTrioIndex(ppd, growthCtx.GetBaseDivByThree(c.Add(XFirst)), offset)
	res.CenterFaces[MinusX] = growthCtx.GetBaseTrioIndex(ppd, growthCtx.GetBaseDivByThree(c.Sub(XFirst)), offset)
	res.CenterFaces[PlusY] = growthCtx.GetBaseTrioIndex(ppd, growthCtx.GetBaseDivByThree(c.Add(YFirst)), offset)
	res.CenterFaces[MinusY] = growthCtx.GetBaseTrioIndex(ppd, growthCtx.GetBaseDivByThree(c.Sub(YFirst)), offset)
	res.CenterFaces[PlusZ] = growthCtx.GetBaseTrioIndex(ppd, growthCtx.GetBaseDivByThree(c.Add(ZFirst)), offset)
	res.CenterFaces[MinusZ] = growthCtx.GetBaseTrioIndex(ppd, growthCtx.GetBaseDivByThree(c.Sub(ZFirst)), offset)

	res.MiddleEdges[GetMiddleEdgeIndex(PlusX, PlusY)] = growthCtx.GetBaseTrioIndex(ppd, growthCtx.GetBaseDivByThree(c.Add(XFirst).Add(YFirst)), offset)
	res.MiddleEdges[GetMiddleEdgeIndex(PlusX, MinusY)] = growthCtx.GetBaseTrioIndex(ppd, growthCtx.GetBaseDivByThree(c.Add(XFirst).Sub(YFirst)), offset)
	res.MiddleEdges[GetMiddleEdgeIndex(PlusX, PlusZ)] = growthCtx.GetBaseTrioIndex(ppd, growthCtx.GetBaseDivByThree(c.Add(XFirst).Add(ZFirst)), offset)
	res.MiddleEdges[GetMiddleEdgeIndex(PlusX, MinusZ)] = growthCtx.GetBaseTrioIndex(ppd, growthCtx.GetBaseDivByThree(c.Add(XFirst).Sub(ZFirst)), offset)

	res.MiddleEdges[GetMiddleEdgeIndex(MinusX, PlusY)] = growthCtx.GetBaseTrioIndex(ppd, growthCtx.GetBaseDivByThree(c.Sub(XFirst).Add(YFirst)), offset)
	res.MiddleEdges[GetMiddleEdgeIndex(MinusX, MinusY)] = growthCtx.GetBaseTrioIndex(ppd, growthCtx.GetBaseDivByThree(c.Sub(XFirst).Sub(YFirst)), offset)
	res.MiddleEdges[GetMiddleEdgeIndex(MinusX, PlusZ)] = growthCtx.GetBaseTrioIndex(ppd, growthCtx.GetBaseDivByThree(c.Sub(XFirst).Add(ZFirst)), offset)
	res.MiddleEdges[GetMiddleEdgeIndex(MinusX, MinusZ)] = growthCtx.GetBaseTrioIndex(ppd, growthCtx.GetBaseDivByThree(c.Sub(XFirst).Sub(ZFirst)), offset)

	res.MiddleEdges[GetMiddleEdgeIndex(PlusY, PlusZ)] = growthCtx.GetBaseTrioIndex(ppd, growthCtx.GetBaseDivByThree(c.Add(YFirst).Add(ZFirst)), offset)
	res.MiddleEdges[GetMiddleEdgeIndex(PlusY, MinusZ)] = growthCtx.GetBaseTrioIndex(ppd, growthCtx.GetBaseDivByThree(c.Add(YFirst).Sub(ZFirst)), offset)
	res.MiddleEdges[GetMiddleEdgeIndex(MinusY, PlusZ)] = growthCtx.GetBaseTrioIndex(ppd, growthCtx.GetBaseDivByThree(c.Sub(YFirst).Add(ZFirst)), offset)
	res.MiddleEdges[GetMiddleEdgeIndex(MinusY, MinusZ)] = growthCtx.GetBaseTrioIndex(ppd, growthCtx.GetBaseDivByThree(c.Sub(YFirst).Sub(ZFirst)), offset)

	return res
}

func (cube CubeOfTrioIndex) String() string {
	return fmt.Sprintf("CK-%s-%s-%s-%s", cube.Center.String(), cube.CenterFaces[0].String(), cube.CenterFaces[1].String(), cube.MiddleEdges[0].String())
}

func (cube CubeOfTrioIndex) GetCenter() TrioIndex {
	return cube.Center
}

func (cube CubeOfTrioIndex) GetCenterFaces() [6]TrioIndex {
	return cube.CenterFaces
}

func (cube CubeOfTrioIndex) GetMiddleEdges() [12]TrioIndex {
	return cube.MiddleEdges
}

func (cube CubeOfTrioIndex) GetCenterFaceTrio(ud UnitDirection) TrioIndex {
	return cube.CenterFaces[ud]
}

func (cube CubeOfTrioIndex) GetMiddleEdgeTrio(ud1 UnitDirection, ud2 UnitDirection) TrioIndex {
	return cube.MiddleEdges[GetMiddleEdgeIndex(ud1, ud2)]
}

func (ppd *BasePointPackData) GetCubeById(cubeId int) CubeKeyId {
	ppd.CheckCubesInitialized()
	for cubeKey, id := range ppd.CubeIdsPerKey {
		if id == cubeId {
			return cubeKey
		}
	}
	Log.Fatalf("trying to find cube by id %d which does not exists", cubeId)
	return CubeKeyId{-1, CubeOfTrioIndex{}}
}

func (ppd *BasePointPackData) GetCubeIdByKey(cubeKey CubeKeyId) int {
	ppd.CheckCubesInitialized()
	id, ok := ppd.CubeIdsPerKey[cubeKey]
	if !ok {
		Log.Fatalf("trying to find cube %v which does not exists", cubeKey)
		return -1
	}
	return id
}
