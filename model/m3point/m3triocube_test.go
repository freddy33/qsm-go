package m3point

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTrioCubeMaps(t *testing.T) {
	Log.SetInfo()
	Log.SetAssert(true)

	ppd := getPointTestData()

	for _, ctxType := range GetAllContextTypes() {
		nbIndexes := ctxType.GetNbIndexes()
		for pIdx := 0; pIdx < nbIndexes; pIdx++ {
			growthCtx := ppd.GetGrowthContextByTypeAndIndex(ctxType, pIdx)
			max, cubes := findNbCubes(growthCtx)
			// Test way above
			nbCubesBig := distinctCubes(growthCtx, max*3)
			assert.Equal(t, len(cubes), len(nbCubesBig), "failed test big for %s for max=%d", growthCtx.String(), max)
			cl := ppd.getCubeList(growthCtx)
			assert.Equal(t, len(cubes), len(cl.allCubes), "failed test big for %s for max=%d", growthCtx.String(), max)
			maxOffset := ctxType.GetMaxOffset()
			for offset := 0; offset < maxOffset; offset++ {
				assertWithOffset(t, cl, max + 1, offset)
			}
		}
	}
}

func findNbCubes(growthCtx GrowthContext) (CInt, map[CubeOfTrioIndex]int) {
	nbCubes := 0
	max := CInt(1)
	var newCubes map[CubeOfTrioIndex]int
	for ; max < 30; max++ {
		newCubes = distinctCubes(growthCtx, max)
		if nbCubes == len(newCubes) {
			Log.Debugf("Found max for %s = %d at %d", growthCtx.String(), nbCubes, max-1)
			break
		}
		nbCubes = len(newCubes)
	}
	return max-1, newCubes
}

func distinctCubes(growthCtx GrowthContext, max CInt) map[CubeOfTrioIndex]int {
	allCubes := make(map[CubeOfTrioIndex]int)
	maxOffset := growthCtx.GetGrowthType().GetMaxOffset()
	for offset := 0; offset < maxOffset; offset++ {
		cube := createTrioCube(growthCtx, offset, Origin)
		allCubes[cube]++
	}
	for x := -max; x <= max; x++ {
		for y := -max; y <= max; y++ {
			for z := -max; z <= max; z++ {
				cube := createTrioCube(growthCtx, 0, Point{x,y,z}.Mul(THREE))
				allCubes[cube]++
			}
		}
	}
	return allCubes
}

func assertWithOffset(t *testing.T, cl *CubeListBuilder, max CInt, offset int) {
	for x := -max; x <= max; x++ {
		for y := -max; y <= max; y++ {
			for z := -max; z <= max; z++ {
				mp := Point{x, y, z}.Mul(THREE)
				assert.True(t, cl.exists(offset, mp), "did not find cube for %s at %d and %v", cl.growthCtx.String(), offset, mp)
			}
		}
	}
}

