package m3point

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTrioCubeMaps(t *testing.T) {
	Log.SetInfo()
	Log.SetAssert(true)

	for _, ctxType := range GetAllContextTypes() {
		nbIndexes := ctxType.GetNbIndexes()
		for pIdx := 0; pIdx < nbIndexes; pIdx++ {
			trCtx := GetTrioContextByTypeAndIdx(ctxType, pIdx)
			max, cubes := findNbCubes(trCtx)
			// Test way above
			nbCubesBig := distinctCubes(trCtx, max*3)
			assert.Equal(t, len(cubes), len(nbCubesBig), "failed test big for %s for max=%d", trCtx.String(), max)
			cl := GetCubeList(trCtx)
			assert.Equal(t, len(cubes), len(cl.allCubes), "failed test big for %s for max=%d", trCtx.String(), max)
			maxOffset := ctxType.GetMaxOffset()
			for offset := 0; offset < maxOffset; offset++ {
				assertWithOffset(t, cl, max + 1, offset)
			}
		}
	}
}

func findNbCubes(trCtx *TrioContext) (CInt, map[CubeKey]int) {
	nbCubes := 0
	max := CInt(1)
	var newCubes map[CubeKey]int
	for ; max < 30; max++ {
		newCubes = distinctCubes(trCtx, max)
		if nbCubes == len(newCubes) {
			Log.Debugf("Found max for %s = %d at %d", trCtx.String(), nbCubes, max-1)
			break
		}
		nbCubes = len(newCubes)
	}
	return max-1, newCubes
}

func distinctCubes(trCtx *TrioContext, max CInt) map[CubeKey]int {
	allCubes := make(map[CubeKey]int)
	maxOffset := trCtx.ctxType.GetMaxOffset()
	for offset := 0; offset < maxOffset; offset++ {
		cube := createTrioCube(trCtx, offset, Origin)
		allCubes[cube]++
	}
	for x := -max; x <= max; x++ {
		for y := -max; y <= max; y++ {
			for z := -max; z <= max; z++ {
				cube := createTrioCube(trCtx, 0, Point{x,y,z}.Mul(THREE))
				allCubes[cube]++
			}
		}
	}
	return allCubes
}

func assertWithOffset(t *testing.T, cl *CubeListPerContext, max CInt, offset int) {
	for x := -max; x <= max; x++ {
		for y := -max; y <= max; y++ {
			for z := -max; z <= max; z++ {
				mp := Point{x, y, z}.Mul(THREE)
				assert.True(t, cl.exists(offset, mp), "did not find cube for %s at %d and %v", cl.trCtx.String(), offset, mp)
			}
		}
	}
}

