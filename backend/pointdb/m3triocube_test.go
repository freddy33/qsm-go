package pointdb

import (
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTrioCubeMaps(t *testing.T) {
	Log.SetInfo()
	Log.SetAssert(true)

	ppd := getPointTestData()

	for _, ctxType := range m3point.GetAllGrowthTypes() {
		nbIndexes := ctxType.GetNbIndexes()
		for pIdx := 0; pIdx < nbIndexes; pIdx++ {
			growthCtx := ppd.GetGrowthContextByTypeAndIndex(ctxType, pIdx)
			Log.Debugf("Running test for %s", growthCtx.String())
			max, cubes := ppd.findNbCubes(growthCtx)
			// Test way above
			nbCubesBig := ppd.distinctCubes(growthCtx, max*3)
			assert.Equal(t, len(cubes), len(nbCubesBig), "failed test big for %s for max=%d", growthCtx.String(), max)
			cl := ppd.getCubeList(growthCtx)
			assert.Equal(t, len(cubes), len(cl.allCubes), "failed test big for %s for max=%d", growthCtx.String(), max)
			maxOffset := ctxType.GetMaxOffset()
			for offset := 0; offset < maxOffset; offset++ {
				if !assertWithOffset(t, cl, max+1, offset) {
					return
				}
			}
		}
	}
}

func (pointData *ServerPointPackData) findNbCubes(growthCtx m3point.GrowthContext) (m3point.CInt, map[CubeOfTrioIndex]int) {
	nbCubes := 0
	max := m3point.CInt(1)
	var newCubes map[CubeOfTrioIndex]int
	for ; max < 30; max++ {
		newCubes = pointData.distinctCubes(growthCtx, max)
		if nbCubes == len(newCubes) {
			Log.Debugf("Found max for %s = %d at %d", growthCtx.String(), nbCubes, max-1)
			break
		}
		nbCubes = len(newCubes)
	}
	return max - 1, newCubes
}

func (pointData *ServerPointPackData) distinctCubes(growthCtx m3point.GrowthContext, max m3point.CInt) map[CubeOfTrioIndex]int {
	allCubes := make(map[CubeOfTrioIndex]int)
	maxOffset := growthCtx.GetGrowthType().GetMaxOffset()
	for offset := 0; offset < maxOffset; offset++ {
		cube := CreateTrioCube(pointData, growthCtx, offset, m3point.Origin)
		allCubes[cube]++
	}
	for x := -max; x <= max; x++ {
		for y := -max; y <= max; y++ {
			for z := -max; z <= max; z++ {
				cube := CreateTrioCube(pointData, growthCtx, 0, m3point.Point{x, y, z}.Mul(m3point.THREE))
				allCubes[cube]++
			}
		}
	}
	return allCubes
}

func assertWithOffset(t *testing.T, cl *CubeListBuilder, max m3point.CInt, offset int) bool {
	for x := -max; x <= max; x++ {
		for y := -max; y <= max; y++ {
			for z := -max; z <= max; z++ {
				mp := m3point.Point{x, y, z}.Mul(m3point.THREE)
				cubeExists := cl.exists(offset, mp)
				assert.True(t, cubeExists, "did not find cube for %s at %d and %v", cl.growthCtx.String(), offset, mp)
				if !cubeExists {
					return false
				}
			}
		}
	}
	return true
}

