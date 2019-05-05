package m3path

import (
	"github.com/freddy33/qsm-go/m3point"
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

const (
	TestNbRound = 8*3+1
)

var allTestContexts map[m3point.ContextType][]GrowthContext

func getAllTestContexts() map[m3point.ContextType][]GrowthContext {
	if allTestContexts != nil {
		return allTestContexts
	}
	res := make(map[m3point.ContextType][]GrowthContext)

	for _, ctxType := range m3point.GetAllContextTypes() {
		nbIndexes := ctxType.GetNbIndexes()
		maxOffset := maxOffsetPerType[ctxType]
		res[ctxType] = make([]GrowthContext, nbIndexes*maxOffset)
		idx := 0
		for pIdx := 0; pIdx < nbIndexes; pIdx++ {
			rootCtx := m3point.GetTrioIndexContext(ctxType, pIdx)
			for offset := 0; offset < maxOffset; offset++ {
				res[ctxType][idx] = *CreateFromRoot(rootCtx, m3point.Origin, offset)
				idx++
			}
		}
	}

	allTestContexts = res
	return res
}

func TestCtx2(t *testing.T) {
	Log.SetInfo()
	runForCtxType(1, TestNbRound, 2)
}

func TestCtxPerType(t *testing.T) {
	Log.SetInfo()
	for _, pType := range m3point.GetAllContextTypes() {
		runForCtxType(1, TestNbRound, pType)
	}
}

func TestAllGrowth(t *testing.T) {
	Log.SetInfo()
	Log.SetAssert(true)
	nbRound := TestNbRound
	allCtx := getAllTestContexts()
	maxUsed := 0
	maxLatest := 0
	for _, pType := range m3point.GetAllContextTypes() {
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

func runForCtxType(N, nbRound int, pType m3point.ContextType) {
	allCtx := getAllTestContexts()
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
		Log.Debugf("Max size for all context of type %d: %d, %d with %d runs", pType, maxUsed, maxLatest, nbRound)
	}
}

func runNextPoints(ctx *GrowthContext, nbRound int) (int, int) {
	usedPoints := make(map[m3point.Point]bool, 5*nbRound*nbRound)
	latestPoints := make([]m3point.Point, 1)
	latestPoints[0] = m3point.Origin
	usedPoints[m3point.Origin] = true
	for d := 0; d < nbRound; d++ {
		nbLatestPoints := len(latestPoints)
		finalPoints := make([]m3point.Point, 0, int(1.7*float32(nbLatestPoints)))
		for _, lp := range latestPoints {
			for _, np := range ctx.GetNextPoints(lp) {
				_, ok := usedPoints[np]
				if !ok {
					finalPoints = append(finalPoints, np)
					usedPoints[np] = true
				}
			}
		}
		latestPoints = finalPoints
	}
	return len(usedPoints), len(latestPoints)
}

func TestDivByThree(t *testing.T) {
	runDivByThree(t)
}

func runDivByThree(t assert.TestingT) {
	Log.SetDebug()
	someCenter1 := m3point.Point{3, -6, 9}
	ctx := CreateGrowthContext(someCenter1, 1, 1, 0)
	assert.Equal(t, someCenter1, ctx.center)
	assert.Equal(t, m3point.ContextType(1), ctx.GetType())
	assert.Equal(t, 1, ctx.GetIndex())
	assert.Equal(t, 0, ctx.offset)

	assert.Equal(t, uint64(1), ctx.GetDivByThree(m3point.Point{0, -6, 9}))
	assert.Equal(t, uint64(1), ctx.GetDivByThree(m3point.Point{6, -6, 9}))
	assert.Equal(t, uint64(1), ctx.GetDivByThree(m3point.Point{3, -3, 9}))
	assert.Equal(t, uint64(1), ctx.GetDivByThree(m3point.Point{3, -9, 9}))
	assert.Equal(t, uint64(1), ctx.GetDivByThree(m3point.Point{3, -6, 12}))
	assert.Equal(t, uint64(1), ctx.GetDivByThree(m3point.Point{3, -6, 6}))

	assert.Equal(t, uint64(6), ctx.GetDivByThree(m3point.Point{0, 0, 0}))

	// Verify trio index unaffected
	for d := uint64(0); d < 30; d++ {
		assert.Equal(t, m3point.TrioIndex(1), ctx.GetTrioIndex(d), "failed trio index for ctx %v and divByThree=%d", ctx, d)
	}

}

func TestGrowthContext1(t *testing.T) {
	Log.SetDebug()
	ctx := CreateGrowthContext(m3point.Origin, 1, 3, 0)
	assert.Equal(t, m3point.ContextType(1), ctx.GetType())
	assert.Equal(t, 3, ctx.GetIndex())
	assert.Equal(t, 0, ctx.offset)
	for d := uint64(0); d < 30; d++ {
		assert.Equal(t, m3point.TrioIndex(3), ctx.GetTrioIndex(d), "failed trio index for ctx %v and divByThree=%d", ctx, d)
	}
	ctx.SetIndex(4)
	ctx.offset = 2
	assert.Equal(t, m3point.ContextType(1), ctx.GetType())
	assert.Equal(t, 4, ctx.GetIndex())
	assert.Equal(t, 2, ctx.offset)
	for d := uint64(0); d < 30; d++ {
		assert.Equal(t, m3point.TrioIndex(4), ctx.GetTrioIndex(d), "failed trio index for ctx %v and divByThree=%d", ctx, d)
	}
}

func TestGrowthContext3(t *testing.T) {
	Log.SetDebug()

	for idx := m3point.TrioIndex(0); idx < 4; idx++ {
		ctx := CreateGrowthContext(m3point.Origin, 3, int(idx), 0)
		assert.Equal(t, m3point.ContextType(3), ctx.GetType())
		assert.Equal(t, int(idx), ctx.GetIndex())
		assert.Equal(t, 0, ctx.offset)
		for d := uint64(0); d < 9; d++ {
			if d%2 == 0 {
				assert.Equal(t, idx, ctx.GetTrioIndex(d), "failed trio index for ctx %v step %d", ctx, d)
			} else {
				expected := m3point.TrioIndex(4 + (int(d/2) % 3))
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
	Log.SetDebug()

	growthContexts := getAllTestContexts()
	for _, ctx := range growthContexts[1] {
		assert.Equal(t, m3point.ContextType(1), ctx.GetType())
		for d := uint64(0); d < 30; d++ {
			assert.Equal(t, m3point.TrioIndex(ctx.GetIndex()), ctx.GetTrioIndex(d), "failed trio index for ctx %v and divByThree=%d", ctx.String(), d)
		}
	}

	for _, ctx := range growthContexts[2] {
		assert.Equal(t, m3point.ContextType(2), ctx.GetType())
		oneTwo := m3point.GetValidNextTrioPair(m3point.TrioIndex(ctx.GetIndex()))
		twoIdx := ctx.offset
		for d := uint64(0); d < 30; d++ {
			assert.Equal(t, oneTwo[twoIdx], ctx.GetTrioIndex(d), "failed trio index for ctx %v and divByThree=%d twoIdx=%d in %v", ctx.String(), d, twoIdx, oneTwo)
			twoIdx++
			if twoIdx == 2 {
				twoIdx = 0
			}
		}
	}

	for _, ctx := range growthContexts[4] {
		assert.Equal(t, m3point.ContextType(4), ctx.GetType())
		oneToFour := m3point.AllMod4Permutations[ctx.GetIndex()]
		fourIdx := ctx.offset
		for d := uint64(0); d < 30; d++ {
			assert.Equal(t, oneToFour[fourIdx], ctx.GetTrioIndex(d), "failed trio index for ctx %v and divByThree=%d fourIdx=%d in %v", ctx.String(), d, fourIdx, oneToFour)
			fourIdx++
			if fourIdx == 4 {
				fourIdx = 0
			}
		}
	}

	for _, ctx := range growthContexts[8] {
		assert.Equal(t, m3point.ContextType(8), ctx.GetType())
		oneToEight := m3point.AllMod8Permutations[ctx.GetIndex()]
		eightIdx := ctx.offset
		for d := uint64(0); d < 30; d++ {
			assert.Equal(t, oneToEight[eightIdx], ctx.GetTrioIndex(d), "failed trio index for ctx %v and divByThree=%d eightIdx=%d in %v", ctx.String(), d, eightIdx, oneToEight)
			eightIdx++
			if eightIdx == 8 {
				eightIdx = 0
			}
		}
	}

}

func TestTrioListPerContext(t *testing.T) {
	Log.SetInfo()
	contexts := getAllTestContexts()
	for _, ctxs := range contexts {
		stableStep := -1
		indexList := make(map[int][]int)
		for _, ctx := range ctxs {
			s, l := runAllTrioList(t, &ctx)
			if stableStep == -1 {
				stableStep = s
			} else {
				assert.Equal(t, stableStep, s, "failed same stable step for %s", ctx.String())
			}
			curList, ok := indexList[ctx.GetIndex()]
			if !ok {
				indexList[ctx.GetIndex()] = l
			} else {
				assert.True(t, EqualIntSlice(curList, l), "failed same index list for %s %v != %v", ctx.String(), curList, l)
			}
		}
	}
}

// Return the ordered list of trio index used
func runAllTrioList(t *testing.T, ctx *GrowthContext) (stableStep int, indexList []int) {
	// The result list of trio index used
	var currentIndexList []int

	countUsedIdx := make(map[m3point.TrioIndex]int)
	usedPoints := make(map[m3point.Point]m3point.TrioIndex)
	latestPoints := make([]m3point.Point, 1)
	latestPoints[0] = ctx.center

	// If currentIndexList is stable for "verifyStable" iterations we should stop
	verifyStable := 8
	stableIndexList := 0
	stepStable := 0

	inError := make(map[m3point.TrioIndex]bool)
	possibleTrios := ctx.GetPossibleTrioList()
	possIds := possibleTrios.IdList()
	assert.True(t, possibleTrios.Len() > 2, "wrong possible trio for %v", ctx.String())

	for d := uint64(0); d < 30; d++ {
		stepStable = int(d)
		newPoints := make([]m3point.Point, 0, 2*len(latestPoints))
		stepCountIdx := make(map[m3point.TrioIndex]int)
		stepConflictCount := make(map[m3point.Point]int)

		for _, p := range latestPoints {
			nextPoints := ctx.GetNextPoints(p)
			var trCtx *m3point.TrioIndexContext
			trCtx = m3point.GetTrioIndexContext(ctx.GetType(), ctx.GetIndex())
			tdIdx, link := m3point.FindTrioIndex(p, nextPoints, trCtx, ctx.offset)
			assert.True(t, tdIdx < m3point.GetNumberOfTrioDetails(), "wrong trio detail index=%d for %v, %v, %s", tdIdx, p, nextPoints, ctx.String())
			td := m3point.GetTrioDetails(tdIdx)

			idExists := possibleTrios.ExistsById(td)
			if !idExists && !inError[td.GetId()] {
				assert.True(t, idExists, "did not find for ctx %s trio %s in %v", ctx.String(), td.String(), possIds)
				inError[td.GetId()] = true
			}

			assert.True(t, td.Links.Exists(&link), "did not find for ctx %s link in idx=%d %v in %v", ctx.String(), tdIdx, link, td.Links)

			countUsedIdx[tdIdx]++
			stepCountIdx[tdIdx]++

			existingIdx, ok := usedPoints[p]
			if !ok {
				usedPoints[p] = tdIdx
				for _, np := range nextPoints {
					_, present := usedPoints[np]
					if !present {
						newPoints = append(newPoints, np)
					}
				}
			} else {
				stepConflictCount[p]++
				assert.Equal(t, existingIdx, tdIdx, "conflict on %v step %d ctx %s", p, d, ctx.String())
			}
		}
		stepConflictSummary := make(map[int]int)
		for _, v := range stepConflictCount {
			stepConflictSummary[v]++
		}
		_, impossible := stepConflictSummary[3]
		assert.False(t, impossible)

		Log.Tracef("Run: %2d %4d : %4d %2d %v", d, len(latestPoints), stepConflictSummary[1], stepConflictSummary[2], stepCountIdx)

		newIndexList := make([]int, 0, len(countUsedIdx))
		for trIdx := range countUsedIdx {
			newIndexList = append(newIndexList, int(trIdx))
		}
		sort.Ints(newIndexList)

		if EqualIntSlice(currentIndexList, newIndexList) {
			stableIndexList++
		} else {
			stableIndexList = 0
			currentIndexList = newIndexList
		}

		if stableIndexList >= verifyStable {
			break
		}

		latestPoints = newPoints
	}
	Log.Debug(ctx.String(), stepStable-verifyStable, currentIndexList)

	return stepStable - verifyStable, currentIndexList
}

// Equal tells whether a and b contain the same elements.
// A nil argument is equivalent to an empty slice.
func EqualIntSlice(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
