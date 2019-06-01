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
			nbCubes, max := findNbCubes(trCtx)
			// Test way above
			nbCubesBig := distinctCubes(trCtx, max*3)
			assert.Equal(t, nbCubes, nbCubesBig, "failed test big for %s for max=%d", trCtx.String(), max)
		}
	}
}

func findNbCubes(trCtx *TrioIndexContext) (int, int64) {
	nbCubes := 0
	max := int64(1)
	for ; max < 30; max++ {
		newNbCubes := distinctCubes(trCtx, max)
		if nbCubes == newNbCubes {
			Log.Infof("Found max for %s = %d at %d", trCtx.String(), nbCubes, max-1)
			break
		}
		nbCubes = newNbCubes
	}
	return nbCubes, max-1
}

func distinctCubes(trCtx *TrioIndexContext, max int64) int {
	allCubes := make(map[TrioIndexCubeKey]int)
	for x := -max; x <= max; x++ {
		for y := -max; y <= max; y++ {
			for z := -max; z <= max; z++ {
				cube := createTrioCube(trCtx, 0, Point{x,y,z}.Mul(THREE))
				allCubes[cube]++
			}
		}
	}
	return len(allCubes)
}