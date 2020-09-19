package pointdb

import (
	"fmt"
	"github.com/freddy33/qsm-go/model/m3point"
)

type CubeOfTrioIndex struct {
	// Index of cube center
	Center m3point.TrioIndex
	// Indexes of center of cube face ordered by +X, -X, +Y, -Y, +Z, -Z
	CenterFaces [6]m3point.TrioIndex
	// Indexes of middle of edges of the cube ordered by +X+Y, +X-Y, +X+Z, +X-Z, -X+Y, -X-Y, -X+Z, -X-Z, +Y+Z, +Y-Z, -Y+Z, -Y-Z
	MiddleEdges [12]m3point.TrioIndex
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

func GetMiddleEdgeIndex(ud1 m3point.UnitDirection, ud2 m3point.UnitDirection) int {
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
	case m3point.PlusX:
		return int(ud2 - 2)
	case m3point.MinusX:
		return int(4 + ud2 - 2)
	case m3point.PlusY:
		return int(8 + ud2 - 4)
	case m3point.MinusY:
		return int(8 + 2 + ud2 - 4)
	}
	Log.Fatalf("Cannot find middle edge for unit directions %d and %d since they are incoherent", ud1, ud2)
	return -1
}

// Fill all the indexes assuming the distance of c from origin used in div by three
func CreateTrioCube(ppd ServerPointPackDataIfc, growthCtx m3point.GrowthContext, offset int, c m3point.Point) CubeOfTrioIndex {
	res := CubeOfTrioIndex{}
	res.Center = growthCtx.GetBaseTrioIndex(ppd, growthCtx.GetBaseDivByThree(c), offset)

	res.CenterFaces[m3point.PlusX] = growthCtx.GetBaseTrioIndex(ppd, growthCtx.GetBaseDivByThree(c.Add(XFirst)), offset)
	res.CenterFaces[m3point.MinusX] = growthCtx.GetBaseTrioIndex(ppd, growthCtx.GetBaseDivByThree(c.Sub(XFirst)), offset)
	res.CenterFaces[m3point.PlusY] = growthCtx.GetBaseTrioIndex(ppd, growthCtx.GetBaseDivByThree(c.Add(YFirst)), offset)
	res.CenterFaces[m3point.MinusY] = growthCtx.GetBaseTrioIndex(ppd, growthCtx.GetBaseDivByThree(c.Sub(YFirst)), offset)
	res.CenterFaces[m3point.PlusZ] = growthCtx.GetBaseTrioIndex(ppd, growthCtx.GetBaseDivByThree(c.Add(ZFirst)), offset)
	res.CenterFaces[m3point.MinusZ] = growthCtx.GetBaseTrioIndex(ppd, growthCtx.GetBaseDivByThree(c.Sub(ZFirst)), offset)

	res.MiddleEdges[GetMiddleEdgeIndex(m3point.PlusX, m3point.PlusY)] = growthCtx.GetBaseTrioIndex(ppd, growthCtx.GetBaseDivByThree(c.Add(XFirst).Add(YFirst)), offset)
	res.MiddleEdges[GetMiddleEdgeIndex(m3point.PlusX, m3point.MinusY)] = growthCtx.GetBaseTrioIndex(ppd, growthCtx.GetBaseDivByThree(c.Add(XFirst).Sub(YFirst)), offset)
	res.MiddleEdges[GetMiddleEdgeIndex(m3point.PlusX, m3point.PlusZ)] = growthCtx.GetBaseTrioIndex(ppd, growthCtx.GetBaseDivByThree(c.Add(XFirst).Add(ZFirst)), offset)
	res.MiddleEdges[GetMiddleEdgeIndex(m3point.PlusX, m3point.MinusZ)] = growthCtx.GetBaseTrioIndex(ppd, growthCtx.GetBaseDivByThree(c.Add(XFirst).Sub(ZFirst)), offset)

	res.MiddleEdges[GetMiddleEdgeIndex(m3point.MinusX, m3point.PlusY)] = growthCtx.GetBaseTrioIndex(ppd, growthCtx.GetBaseDivByThree(c.Sub(XFirst).Add(YFirst)), offset)
	res.MiddleEdges[GetMiddleEdgeIndex(m3point.MinusX, m3point.MinusY)] = growthCtx.GetBaseTrioIndex(ppd, growthCtx.GetBaseDivByThree(c.Sub(XFirst).Sub(YFirst)), offset)
	res.MiddleEdges[GetMiddleEdgeIndex(m3point.MinusX, m3point.PlusZ)] = growthCtx.GetBaseTrioIndex(ppd, growthCtx.GetBaseDivByThree(c.Sub(XFirst).Add(ZFirst)), offset)
	res.MiddleEdges[GetMiddleEdgeIndex(m3point.MinusX, m3point.MinusZ)] = growthCtx.GetBaseTrioIndex(ppd, growthCtx.GetBaseDivByThree(c.Sub(XFirst).Sub(ZFirst)), offset)

	res.MiddleEdges[GetMiddleEdgeIndex(m3point.PlusY, m3point.PlusZ)] = growthCtx.GetBaseTrioIndex(ppd, growthCtx.GetBaseDivByThree(c.Add(YFirst).Add(ZFirst)), offset)
	res.MiddleEdges[GetMiddleEdgeIndex(m3point.PlusY, m3point.MinusZ)] = growthCtx.GetBaseTrioIndex(ppd, growthCtx.GetBaseDivByThree(c.Add(YFirst).Sub(ZFirst)), offset)
	res.MiddleEdges[GetMiddleEdgeIndex(m3point.MinusY, m3point.PlusZ)] = growthCtx.GetBaseTrioIndex(ppd, growthCtx.GetBaseDivByThree(c.Sub(YFirst).Add(ZFirst)), offset)
	res.MiddleEdges[GetMiddleEdgeIndex(m3point.MinusY, m3point.MinusZ)] = growthCtx.GetBaseTrioIndex(ppd, growthCtx.GetBaseDivByThree(c.Sub(YFirst).Sub(ZFirst)), offset)

	return res
}

func (cube CubeOfTrioIndex) String() string {
	return fmt.Sprintf("CK-%s-%s-%s-%s", cube.Center.String(), cube.CenterFaces[0].String(), cube.CenterFaces[1].String(), cube.MiddleEdges[0].String())
}

func (cube CubeOfTrioIndex) GetCenter() m3point.TrioIndex {
	return cube.Center
}

func (cube CubeOfTrioIndex) GetCenterFaces() [6]m3point.TrioIndex {
	return cube.CenterFaces
}

func (cube CubeOfTrioIndex) GetMiddleEdges() [12]m3point.TrioIndex {
	return cube.MiddleEdges
}

func (cube CubeOfTrioIndex) GetCenterFaceTrio(ud m3point.UnitDirection) m3point.TrioIndex {
	return cube.CenterFaces[ud]
}

func (cube CubeOfTrioIndex) GetMiddleEdgeTrio(ud1 m3point.UnitDirection, ud2 m3point.UnitDirection) m3point.TrioIndex {
	return cube.MiddleEdges[GetMiddleEdgeIndex(ud1, ud2)]
}
