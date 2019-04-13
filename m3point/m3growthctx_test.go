package m3point

import (
	"github.com/freddy33/qsm-go/m3util"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestPosMod2(t *testing.T) {
	Log.Level = m3util.DEBUG
	assert.Equal(t, uint64(1), PosMod2(5))
	assert.Equal(t, uint64(0), PosMod2(4))
	assert.Equal(t, uint64(1), PosMod2(3))
	assert.Equal(t, uint64(0), PosMod2(2))
	assert.Equal(t, uint64(1), PosMod2(1))
	assert.Equal(t, uint64(0), PosMod2(0))
}

func TestPosMod4(t *testing.T) {
	Log.Level = m3util.DEBUG
	assert.Equal(t, uint64(1), PosMod4(5))
	assert.Equal(t, uint64(0), PosMod4(4))
	assert.Equal(t, uint64(3), PosMod4(3))
	assert.Equal(t, uint64(2), PosMod4(2))
	assert.Equal(t, uint64(1), PosMod4(1))
	assert.Equal(t, uint64(0), PosMod4(0))
}

func TestPosMod8(t *testing.T) {
	Log.Level = m3util.DEBUG
	assert.Equal(t, uint64(1), PosMod8(9))
	assert.Equal(t, uint64(0), PosMod8(8))
	assert.Equal(t, uint64(7), PosMod8(7))
	assert.Equal(t, uint64(6), PosMod8(6))
	assert.Equal(t, uint64(5), PosMod8(5))
	assert.Equal(t, uint64(4), PosMod8(4))
	assert.Equal(t, uint64(3), PosMod8(3))
	assert.Equal(t, uint64(2), PosMod8(2))
	assert.Equal(t, uint64(1), PosMod8(1))
	assert.Equal(t, uint64(0), PosMod8(0))
}

const (
	SPLIT          = 4
	BENCH_NB_ROUND = 100
	TEST_NB_ROUND  = 25
)

func BenchmarkCtx1(b *testing.B) {
	Log.Level = m3util.WARN
	runForCtxType(b.N, BENCH_NB_ROUND, 1)
}

func BenchmarkCtx2(b *testing.B) {
	Log.Level = m3util.WARN
	runForCtxType(b.N, BENCH_NB_ROUND, 2)
}

func BenchmarkCtx3(b *testing.B) {
	Log.Level = m3util.WARN
	runForCtxType(b.N, BENCH_NB_ROUND, 3)
}

func BenchmarkCtx4(b *testing.B) {
	Log.Level = m3util.WARN
	runForCtxType(b.N, BENCH_NB_ROUND, 4)
}

func BenchmarkCtx8(b *testing.B) {
	Log.Level = m3util.WARN
	runForCtxType(b.N, BENCH_NB_ROUND, 8)
}

func TestCtx2(t *testing.T) {
	Log.Level = m3util.INFO
	runForCtxType(1, TEST_NB_ROUND, 2)
}

func TestCtxPerType(t *testing.T) {
	Log.Level = m3util.INFO
	for _, pType := range [5]uint8{1, 2, 3, 4, 8} {
		runForCtxType(1, TEST_NB_ROUND, pType)
	}
}

func runForCtxType(N, nbRound int, pType uint8) {
	allCtx := getAllContexts()
	for r := 0; r < N; r++ {
		maxUsed := 0
		maxLatest := 0
		for _, ctx := range allCtx[pType] {
			nU, nL := runNextPoints(&ctx, nbRound)
			if nU > maxUsed {
				maxUsed = nU
			}
			if nL > maxLatest {
				maxLatest = nL
			}
		}
		Log.Infof("Max size for all context of type %d: %d, %d with %d runs", pType, maxUsed, maxLatest, nbRound)
	}
}

func BenchmarkAllGrowth(b *testing.B) {
	Log.Level = m3util.WARN
	nbRound := 50
	allCtx := getAllContexts()
	for r := 0; r < b.N; r++ {
		maxUsed := 0
		maxLatest := 0
		for _, pType := range [5]uint8{1, 2, 3, 4, 8} {
			for _, ctx := range allCtx[pType] {
				nU, nL := runNextPoints(&ctx, nbRound)
				if nU > maxUsed {
					maxUsed = nU
				}
				if nL > maxLatest {
					maxLatest = nL
				}
			}
		}
		Log.Infof("Max size for all context %d, %d with %d runs", maxUsed, maxLatest, nbRound)
	}
}

func runNextPoints(ctx *GrowthContext, nbRound int) (int, int) {
	usedPoints := make(map[Point]bool, 10*nbRound*nbRound)
	totalUsedPoints := 1
	latestPoints := make([]Point, 1)
	latestPoints[0] = Origin
	usedPoints[Origin] = true
	for d := 0; d < nbRound; d++ {
		nbLatestPoints := len(latestPoints)
		// Send all orig new points
		origNewPoints := make(chan Point, 4*SPLIT)
		wg := sync.WaitGroup{}
		if nbLatestPoints < 4*SPLIT {
			// too small for split send all
			wg.Add(1)
			go nextPointsSplit(&latestPoints, 0, nbLatestPoints, ctx, origNewPoints, &wg)
		} else {
			sizePerSplit := int(nbLatestPoints / SPLIT)
			for currentPos := 0; currentPos < nbLatestPoints; currentPos += sizePerSplit {
				wg.Add(1)
				go nextPointsSplit(&latestPoints, currentPos, sizePerSplit, ctx, origNewPoints, &wg)
			}
		}
		go func(step int) {
			wg.Wait()
			close(origNewPoints)
		}(d)

		finalPoints := make([]Point, 0, int(1.7*float32(nbLatestPoints)))
		for p := range origNewPoints {
			_, ok := usedPoints[p]
			if !ok {
				finalPoints = append(finalPoints, p)
				usedPoints[p] = true
			}
		}

		totalUsedPoints += len(finalPoints)
		latestPoints = finalPoints
	}
	return totalUsedPoints, len(latestPoints)
}

func runNextPointsAsync(ctx *GrowthContext, nbRound int) (int, int) {
	//usedPoints := make(map[Point]bool, 10*nbRound*nbRound)
	usedPoints := new(sync.Map)
	totalUsedPoints := 1
	latestPoints := make([]Point, 1)
	latestPoints[0] = Origin
	usedPoints.Store(Origin, true)
	o := make(chan Point, 100)
	for d := 0; d < nbRound; d++ {
		finalPoints := make([]Point, 0, int(1.2*float32(len(latestPoints))))
		for _, p := range latestPoints {
			go asyncNextPoints(p, ctx, o, nil)
		}
		// I'll always get 3 tines the amount of latest points
		newPoints := 3 * len(latestPoints)
		for i := 0; i < newPoints; i++ {
			p, ok := <-o
			if !ok {
				break
			} else {
				_, ok := usedPoints.LoadOrStore(p, true)
				if !ok {
					finalPoints = append(finalPoints, p)
				}
			}
		}
		latestPoints = finalPoints
		totalUsedPoints += len(latestPoints)
	}
	return totalUsedPoints, len(latestPoints)
}

func nextPointsSplit(lps *[]Point, currentPos, nb int, ctx *GrowthContext, o chan Point, wg *sync.WaitGroup) {
	c := 0
	for i := currentPos; i < len(*lps); i++ {
		p := (*lps)[i]
		for _, np := range p.GetNextPoints(ctx) {
			o <- np
		}
		c++
		if c == nb {
			break
		}
	}
	wg.Done()
}

func asyncNextPoints(p Point, ctx *GrowthContext, o chan Point, wg *sync.WaitGroup) {
	for _, np := range p.GetNextPoints(ctx) {
		o <- np
	}
	wg.Done()
}

var allContexts map[uint8][]GrowthContext

func getAllContexts() map[uint8][]GrowthContext {
	if allContexts != nil {
		return allContexts
	}
	res := make(map[uint8][]GrowthContext)
	res[1] = make([]GrowthContext, 0, 8)
	res[3] = make([]GrowthContext, 0, 8*4)
	res[2] = make([]GrowthContext, 0, 12*2*2)
	res[4] = make([]GrowthContext, 0, 12*4*2)
	res[8] = make([]GrowthContext, 0, 12*8*2)

	for pIdx := 0; pIdx < 8; pIdx++ {
		res[1] = append(res[1], GrowthContext{Origin, 1, pIdx, false, 0,})
		for offset := 0; offset < 3; offset++ {
			res[3] = append(res[3], GrowthContext{Origin, 3, pIdx, false, offset,})
		}
	}

	for _, pType := range [3]uint8{2, 4, 8} {
		for pIdx := 0; pIdx < 12; pIdx++ {
			for offset := 0; offset < int(pType); offset++ {
				res[pType] = append(res[pType],
					GrowthContext{Origin, pType, pIdx, false, offset,},
					GrowthContext{Origin, pType, pIdx, true, offset,})
			}
		}
	}

	allContexts = res
	return res
}

func TestDivByThree(t *testing.T) {
	runDivByThree(t)
}

func runDivByThree(t assert.TestingT) {
	Log.Level = m3util.DEBUG
	someCenter1 := Point{3, -6, 9}
	ctx := GrowthContext{someCenter1, 1, 1, false, 0,}
	assert.Equal(t, someCenter1, ctx.center)
	assert.Equal(t, uint8(1), ctx.permutationType)
	assert.Equal(t, 1, ctx.permutationIndex)
	assert.Equal(t, false, ctx.permutationNegFlow)
	assert.Equal(t, 0, ctx.permutationOffset)

	assert.Equal(t, uint64(1), ctx.GetDivByThree(Point{0, -6, 9}))
	assert.Equal(t, uint64(1), ctx.GetDivByThree(Point{6, -6, 9}))
	assert.Equal(t, uint64(1), ctx.GetDivByThree(Point{3, -3, 9}))
	assert.Equal(t, uint64(1), ctx.GetDivByThree(Point{3, -9, 9}))
	assert.Equal(t, uint64(1), ctx.GetDivByThree(Point{3, -6, 12}))
	assert.Equal(t, uint64(1), ctx.GetDivByThree(Point{3, -6, 6}))

	assert.Equal(t, uint64(6), ctx.GetDivByThree(Point{0, 0, 0}))

	// Verify trio index unaffected
	for d := uint64(0); d < 30; d++ {
		assert.Equal(t, 1, ctx.GetTrioIndex(d), "failed trio index for ctx %v and divByThree=%d", ctx, d)
	}

}

func TestGrowthContext1(t *testing.T) {
	Log.Level = m3util.DEBUG
	ctx := GrowthContext{Origin, 1, 3, false, 0,}
	assert.Equal(t, uint8(1), ctx.permutationType)
	assert.Equal(t, 3, ctx.permutationIndex)
	assert.Equal(t, false, ctx.permutationNegFlow)
	assert.Equal(t, 0, ctx.permutationOffset)
	for d := uint64(0); d < 30; d++ {
		assert.Equal(t, 3, ctx.GetTrioIndex(d), "failed trio index for ctx %v and divByThree=%d", ctx, d)
	}
	ctx.permutationIndex = 4
	ctx.permutationNegFlow = true
	ctx.permutationOffset = 2
	assert.Equal(t, uint8(1), ctx.permutationType)
	assert.Equal(t, 4, ctx.permutationIndex)
	assert.Equal(t, true, ctx.permutationNegFlow)
	assert.Equal(t, 2, ctx.permutationOffset)
	for d := uint64(0); d < 30; d++ {
		assert.Equal(t, 4, ctx.GetTrioIndex(d), "failed trio index for ctx %v and divByThree=%d", ctx, d)
	}
}

func TestGrowthContext3(t *testing.T) {
	Log.Level = m3util.DEBUG

	for idx := 0; idx < 4; idx++ {
		ctx := GrowthContext{Origin, 3, idx, false, 0,}
		assert.Equal(t, uint8(3), ctx.permutationType)
		assert.Equal(t, idx, ctx.permutationIndex)
		assert.Equal(t, false, ctx.permutationNegFlow)
		assert.Equal(t, 0, ctx.permutationOffset)
		for d := uint64(0); d < 9; d++ {
			if d%2 == 0 {
				assert.Equal(t, idx, ctx.GetTrioIndex(d), "failed trio index for ctx %v step %d", ctx, d)
			} else {
				expected := 4 + (int(d/2) % 3)
				if expected >= idx+4 {
					expected++
				}
				assert.Equal(t, expected, ctx.GetTrioIndex(d), "failed trio index for ctx %v step %d", ctx, d)
			}
		}
	}
}

func TestGrowthContextsExpectType3(t *testing.T) {
	runGrowthContextsExpectType3(t)
}

func runGrowthContextsExpectType3(t assert.TestingT) {
	Log.Level = m3util.DEBUG

	growthContexts := getAllContexts()
	for _, ctx := range growthContexts[1] {
		assert.Equal(t, uint8(1), ctx.permutationType)
		for d := uint64(0); d < 30; d++ {
			assert.Equal(t, ctx.permutationIndex, ctx.GetTrioIndex(d), "failed trio index for ctx %v and divByThree=%d", ctx, d)
		}
	}

	for _, ctx := range growthContexts[2] {
		assert.Equal(t, uint8(2), ctx.permutationType)
		oneTwo := ValidNextTrio[ctx.permutationIndex]
		twoIdx := ctx.permutationOffset
		if ctx.permutationNegFlow {
			twoIdx = reverse2Map[twoIdx]
		}
		for d := uint64(0); d < 30; d++ {
			assert.Equal(t, oneTwo[twoIdx], ctx.GetTrioIndex(d), "failed trio index for ctx %v and divByThree=%d twoIdx=%d", ctx, d, twoIdx)
			// Positive flow
			if ctx.permutationNegFlow {
				twoIdx--
				if twoIdx == -1 {
					twoIdx = 1
				}
			} else {
				twoIdx++
				if twoIdx == 2 {
					twoIdx = 0
				}
			}
		}
	}

	for _, ctx := range growthContexts[4] {
		assert.Equal(t, uint8(4), ctx.permutationType)
		oneToFour := AllMod4Permutations[ctx.permutationIndex]
		fourIdx := ctx.permutationOffset
		if ctx.permutationNegFlow {
			fourIdx = reverse4Map[fourIdx]
		}
		for d := uint64(0); d < 30; d++ {
			assert.Equal(t, oneToFour[fourIdx], ctx.GetTrioIndex(d), "failed trio index for ctx %v and divByThree=%d fourIdx=%d", ctx, d, fourIdx)
			// Positive flow
			if ctx.permutationNegFlow {
				fourIdx--
				if fourIdx == -1 {
					fourIdx = 3
				}
			} else {
				fourIdx++
				if fourIdx == 4 {
					fourIdx = 0
				}
			}
		}
	}

	for _, ctx := range growthContexts[8] {
		assert.Equal(t, uint8(8), ctx.permutationType)
		oneToEight := AllMod8Permutations[ctx.permutationIndex]
		eightIdx := ctx.permutationOffset
		if ctx.permutationNegFlow {
			eightIdx = reverse8Map[eightIdx]
		}
		for d := uint64(0); d < 30; d++ {
			assert.Equal(t, oneToEight[eightIdx], ctx.GetTrioIndex(d), "failed trio index for ctx %v and divByThree=%d eightIdx=%d", ctx, d, eightIdx)
			// Positive flow
			if ctx.permutationNegFlow {
				eightIdx--
				if eightIdx == -1 {
					eightIdx = 7
				}
			} else {
				eightIdx++
				if eightIdx == 8 {
					eightIdx = 0
				}
			}
		}
	}

}
