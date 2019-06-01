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
			trCtx := GetTrioIndexContext(ctxType, pIdx)
			max, cubes := findNbCubes(trCtx)
			// Test way above
			nbCubesBig := distinctCubes(trCtx, max*3)
			assert.Equal(t, len(cubes), len(nbCubesBig), "failed test big for %s for max=%d", trCtx.String(), max)
			cl := GetCubeList(ctxType, pIdx)
			assert.Equal(t, len(cubes), len(cl.allCubes), "failed test big for %s for max=%d", trCtx.String(), max)
			maxOffset := MaxOffsetPerType[ctxType]
			for offset := 0; offset < maxOffset; offset++ {
				assertWithOffset(t, cl, max + 1, offset)
			}
		}
	}
}

func findNbCubes(trCtx *TrioIndexContext) (int64, map[TrioIndexCubeKey]int) {
	nbCubes := 0
	max := int64(1)
	var newCubes map[TrioIndexCubeKey]int
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

func distinctCubes(trCtx *TrioIndexContext, max int64) map[TrioIndexCubeKey]int {
	allCubes := make(map[TrioIndexCubeKey]int)
	maxOffset := MaxOffsetPerType[trCtx.ctxType]
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

func assertWithOffset(t *testing.T, cl *CubeListPerContext, max int64, offset int) {
	for x := -max; x <= max; x++ {
		for y := -max; y <= max; y++ {
			for z := -max; z <= max; z++ {
				mp := Point{x, y, z}.Mul(THREE)
				assert.True(t, cl.exists(offset, mp), "did not find cube for %s at %d and %v", cl.trCtx.String(), offset, mp)
			}
		}
	}
}

