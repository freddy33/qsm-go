package m3point

type TrioIndexCubeKey struct {
	// Index of cube center
	center TrioIndex
	// Indexes of center of cube face ordered by +X, -X, +Y, -Y, +Z, -Z
	centerFaces [6]TrioIndex
	// Indexes of middle of edges of the cube ordered by +X+Y, +X-Y, +X+Z, +X-Z, -X+Y, -X-Y, -X+Z, -X-Z, +Y+Z, +Y-Z, -Y+Z, -Y-Z
	middleEdges [12]TrioIndex
}

type CubeListPerContext struct {
	trCtx *TrioIndexContext
	allCubes []TrioIndexCubeKey
}

var allTrioCubes [][]*CubeListPerContext

func createAllTrioCubes() {
	if len(allTrioCubes) != 0 {
		// done
		return
	}
	allTrioCubes = make([][]*CubeListPerContext, 9)

}


// Fill all the indexes assuming the distance of c from origin used in div by three
func createTrioCube(trCtx *TrioIndexContext, offset int, c Point) TrioIndexCubeKey {
	res := TrioIndexCubeKey{}
	res.center = trCtx.GetBaseTrioIndex(trCtx.GetBaseDivByThree(c), offset)

	res.centerFaces[0] = trCtx.GetBaseTrioIndex(trCtx.GetBaseDivByThree(c.Add(XFirst)), offset)
	res.centerFaces[1] = trCtx.GetBaseTrioIndex(trCtx.GetBaseDivByThree(c.Sub(XFirst)), offset)
	res.centerFaces[2] = trCtx.GetBaseTrioIndex(trCtx.GetBaseDivByThree(c.Add(YFirst)), offset)
	res.centerFaces[3] = trCtx.GetBaseTrioIndex(trCtx.GetBaseDivByThree(c.Sub(YFirst)), offset)
	res.centerFaces[4] = trCtx.GetBaseTrioIndex(trCtx.GetBaseDivByThree(c.Add(ZFirst)), offset)
	res.centerFaces[5] = trCtx.GetBaseTrioIndex(trCtx.GetBaseDivByThree(c.Sub(ZFirst)), offset)

	res.middleEdges[0] = trCtx.GetBaseTrioIndex(trCtx.GetBaseDivByThree(c.Add(XFirst).Add(YFirst)), offset)
	res.middleEdges[1] = trCtx.GetBaseTrioIndex(trCtx.GetBaseDivByThree(c.Add(XFirst).Sub(YFirst)), offset)
	res.middleEdges[2] = trCtx.GetBaseTrioIndex(trCtx.GetBaseDivByThree(c.Add(XFirst).Add(ZFirst)), offset)
	res.middleEdges[3] = trCtx.GetBaseTrioIndex(trCtx.GetBaseDivByThree(c.Add(XFirst).Sub(ZFirst)), offset)

	res.middleEdges[4] = trCtx.GetBaseTrioIndex(trCtx.GetBaseDivByThree(c.Sub(XFirst).Add(YFirst)), offset)
	res.middleEdges[5] = trCtx.GetBaseTrioIndex(trCtx.GetBaseDivByThree(c.Sub(XFirst).Sub(YFirst)), offset)
	res.middleEdges[6] = trCtx.GetBaseTrioIndex(trCtx.GetBaseDivByThree(c.Sub(XFirst).Add(ZFirst)), offset)
	res.middleEdges[7] = trCtx.GetBaseTrioIndex(trCtx.GetBaseDivByThree(c.Sub(XFirst).Sub(ZFirst)), offset)

	res.middleEdges[8] = trCtx.GetBaseTrioIndex(trCtx.GetBaseDivByThree(c.Add(YFirst).Add(ZFirst)), offset)
	res.middleEdges[9] = trCtx.GetBaseTrioIndex(trCtx.GetBaseDivByThree(c.Add(YFirst).Sub(ZFirst)), offset)
	res.middleEdges[10] = trCtx.GetBaseTrioIndex(trCtx.GetBaseDivByThree(c.Sub(YFirst).Add(ZFirst)), offset)
	res.middleEdges[11] = trCtx.GetBaseTrioIndex(trCtx.GetBaseDivByThree(c.Sub(YFirst).Sub(ZFirst)), offset)

	return res
}
