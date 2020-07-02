package m3server

import (
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTrioCubeMaps(t *testing.T) {
	m3point.Log.SetInfo()
	m3point.Log.SetAssert(true)

	ppd := getPointTestData()

	for _, ctxType := range m3point.GetAllContextTypes() {
		nbIndexes := ctxType.GetNbIndexes()
		for pIdx := 0; pIdx < nbIndexes; pIdx++ {
			growthCtx := ppd.GetGrowthContextByTypeAndIndex(ctxType, pIdx)
			max, cubes := ppd.findNbCubes(growthCtx)
			// Test way above
			nbCubesBig := ppd.distinctCubes(growthCtx, max*3)
			assert.Equal(t, len(cubes), len(nbCubesBig), "failed test big for %s for max=%d", growthCtx.String(), max)
			cl := ppd.getCubeList(growthCtx)
			assert.Equal(t, len(cubes), len(cl.allCubes), "failed test big for %s for max=%d", growthCtx.String(), max)
			maxOffset := ctxType.GetMaxOffset()
			for offset := 0; offset < maxOffset; offset++ {
				assertWithOffset(t, cl, max+1, offset)
			}
		}
	}
}

func (ppd *PointPackData) findNbCubes(growthCtx m3point.GrowthContext) (m3point.CInt, map[m3point.CubeOfTrioIndex]int) {
	nbCubes := 0
	max := m3point.CInt(1)
	var newCubes map[m3point.CubeOfTrioIndex]int
	for ; max < 30; max++ {
		newCubes = ppd.distinctCubes(growthCtx, max)
		if nbCubes == len(newCubes) {
			m3point.Log.Debugf("Found max for %s = %d at %d", growthCtx.String(), nbCubes, max-1)
			break
		}
		nbCubes = len(newCubes)
	}
	return max - 1, newCubes
}

func (ppd *PointPackData) distinctCubes(growthCtx m3point.GrowthContext, max m3point.CInt) map[m3point.CubeOfTrioIndex]int {
	allCubes := make(map[m3point.CubeOfTrioIndex]int)
	maxOffset := growthCtx.GetGrowthType().GetMaxOffset()
	for offset := 0; offset < maxOffset; offset++ {
		cube := ppd.CreateTrioCube(growthCtx, offset, m3point.Origin)
		allCubes[cube]++
	}
	for x := -max; x <= max; x++ {
		for y := -max; y <= max; y++ {
			for z := -max; z <= max; z++ {
				cube := ppd.CreateTrioCube(growthCtx, 0, m3point.Point{x, y, z}.Mul(m3point.THREE))
				allCubes[cube]++
			}
		}
	}
	return allCubes
}

func assertWithOffset(t *testing.T, cl *CubeListBuilder, max m3point.CInt, offset int) {
	for x := -max; x <= max; x++ {
		for y := -max; y <= max; y++ {
			for z := -max; z <= max; z++ {
				mp := m3point.Point{x, y, z}.Mul(m3point.THREE)
				assert.True(t, cl.exists(offset, mp), "did not find cube for %s at %d and %v", cl.growthCtx.String(), offset, mp)
			}
		}
	}
}

