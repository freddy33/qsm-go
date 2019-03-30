package m3point

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func BenchmarkAllGrowth(b *testing.B) {
	allCtx := getAllContexts()
	nbRound := 25
	for r:=0;r<b.N;r++ {
		for _, ctx := range allCtx[1] {
			runNextPoints(&ctx, nbRound)
		}
		for _, ctx := range allCtx[2] {
			runNextPoints(&ctx, nbRound)
		}
		for _, ctx := range allCtx[4] {
			runNextPoints(&ctx, nbRound)
		}
		for _, ctx := range allCtx[8] {
			runNextPoints(&ctx, nbRound)
		}
	}
}

func runNextPoints(ctx *GrowthContext, nbRound int) {
	usedPoints := make(map[Point]bool, 15000)
	latestPoints := make([]Point, 1)
	latestPoints[0] = Origin
	usedPoints[Origin] = true
	for d := 0; d < nbRound; d++ {
		finalPoints := latestPoints[:0]
		for _, p := range latestPoints {
			newPoints := p.GetNextPoints(ctx)
			for _, np := range newPoints {
				if !usedPoints[np] {
					finalPoints = append(finalPoints, np)
					usedPoints[np] = true
				}
			}
		}
		latestPoints = finalPoints
	}
}

func getAllContexts() map[uint8][]GrowthContext {
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

	for _, pType := range [4]uint8{2, 4, 8} {
		for pIdx := 0; pIdx < 12; pIdx++ {
			for offset := 0; offset < int(pType); offset++ {
				res[pType] = append(res[pType], GrowthContext{Origin, pType, pIdx, false, offset,})
				res[pType] = append(res[pType], GrowthContext{Origin, pType, pIdx, true, offset,})
			}
		}
	}
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
	ctx := GrowthContext{Origin, 3, 0, false, 0,}
	assert.Equal(t, uint8(3), ctx.permutationType)
	assert.Equal(t, 0, ctx.permutationIndex)
	assert.Equal(t, false, ctx.permutationNegFlow)
	assert.Equal(t, 0, ctx.permutationOffset)
	assert.Equal(t, 0, ctx.GetTrioIndex(0), "failed trio index for ctx %v", ctx)
	assert.Equal(t, 4, ctx.GetTrioIndex(1), "failed trio index for ctx %v", ctx)
	assert.Equal(t, 0, ctx.GetTrioIndex(2), "failed trio index for ctx %v", ctx)
	assert.Equal(t, 6, ctx.GetTrioIndex(3), "failed trio index for ctx %v", ctx)
	assert.Equal(t, 0, ctx.GetTrioIndex(4), "failed trio index for ctx %v", ctx)
	assert.Equal(t, 7, ctx.GetTrioIndex(5), "failed trio index for ctx %v", ctx)
	assert.Equal(t, 0, ctx.GetTrioIndex(6), "failed trio index for ctx %v", ctx)
	assert.Equal(t, 4, ctx.GetTrioIndex(7), "failed trio index for ctx %v", ctx)
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

func TestConnectionDetails(t *testing.T) {
	Log.Level = m3util.DEBUG
	for k, v := range AllConnectionsPossible {
		assert.Equal(t, k, v.Vector)
		currentNumber := Abs8(v.GetIntId())
		sameNumber := 0
		for _, nv := range AllConnectionsPossible {
			if Abs8(nv.GetIntId()) == currentNumber {
				sameNumber++
				if nv.Vector != v.Vector {
					assert.Equal(t, nv.GetIntId(), -v.GetIntId(), "Should have opposite id")
					assert.Equal(t, nv.Vector.Neg(), v.Vector, "Should have neg vector")
				}
			}
		}
		assert.Equal(t, 2, sameNumber, "Should have 2 with same conn number for %d", currentNumber)
	}

	countConnId := make(map[int8]int)
	for i, tA := range AllBaseTrio {
		for j, tB := range AllBaseTrio {
			connVectors := GetNonBaseConnections(tA, tB)
			for k, connVector := range connVectors {
				connDetails, ok := AllConnectionsPossible[connVector]
				assert.True(t, ok, "Connection between 2 trio (%d,%d) number %k is not in conn details", i, j, k)
				assert.Equal(t, connVector, connDetails.Vector, "Connection between 2 trio (%d,%d) number %k is not in conn details", i, j, k)
				countConnId[connDetails.GetIntId()]++
			}
		}
	}
	Log.Info("ConnId usage:",countConnId)

	allCtx := getAllContexts()
	assert.Equal(t, 5, len(allCtx))

	nbCtx := 0
	for _, contextList := range allCtx {
		nbCtx += len(contextList)
	}
	Log.Info("Created", nbCtx, "contexts")
	Log.Info("Using", len(allCtx[8]), " contexts from the 8 context")
	// For all trioIndex rotations, any 2 close main points there should be a connection details
	min := int64(-2) // -5
	max := int64(2)  // 5
	for _, ctx := range allCtx[8] {
		for x := min; x < max; x++ {
			for y := min; y < max; y++ {
				for z := min; z < max; z++ {
					mainPoint := Point{x, y, z}.Mul(3)
					connectingVectors := ctx.GetTrio(mainPoint)
					for _, cVec := range connectingVectors {

						assertValidConnDetails(t, mainPoint, mainPoint.Add(cVec), fmt.Sprint("Main Pos", mainPoint, "base vector", cVec))

						nextMain := Origin
						switch cVec.X() {
						case 0:
							// Nothing out
						case 1:
							nextMain = mainPoint.Add(XFirst)
						case -1:
							nextMain = mainPoint.Sub(XFirst)
						default:
							assert.Fail(t, "There should not be a connecting vector with x value %d", cVec.X())
						}
						if nextMain != Origin {
							// Find the connecting vector on the other side ( the opposite 1 or -1 on X() )
							nextConnectingVectors := ctx.GetTrio(nextMain)
							for _, nbp := range nextConnectingVectors {
								if nbp.X() == -cVec.X() {
									assertValidConnDetails(t, mainPoint.Add(cVec), nextMain.Add(nbp), fmt.Sprint("Main Pos=", mainPoint,
										"next Pos=", nextMain, "trio index=", ctx.GetTrioIndex(ctx.GetDivByThree(mainPoint)),
										"main base vector", cVec, "next base vector", nbp))
								}
							}
						}

						nextMain = Origin
						switch cVec.Y() {
						case 0:
							// Nothing out
						case 1:
							nextMain = mainPoint.Add(YFirst)
						case -1:
							nextMain = mainPoint.Sub(YFirst)
						default:
							assert.Fail(t, "There should not be a connecting vector with y value %d", cVec.Y())
						}
						if nextMain != Origin {
							// Find the connecting vector on the other side ( the opposite 1 or -1 on Y() )
							nextConnectingVectors := ctx.GetTrio(nextMain)
							for _, nbp := range nextConnectingVectors {
								if nbp.Y() == -cVec.Y() {
									assertValidConnDetails(t, mainPoint.Add(cVec), nextMain.Add(nbp), fmt.Sprint("Main Pos=", mainPoint,
										"next Pos=", nextMain, "trio index=", ctx.GetTrioIndex(ctx.GetDivByThree(mainPoint)),
										"main base vector", cVec, "next base vector", nbp))
								}
							}
						}

						nextMain = Origin
						switch cVec.Z() {
						case 0:
							// Nothing out
						case 1:
							nextMain = mainPoint.Add(ZFirst)
						case -1:
							nextMain = mainPoint.Sub(ZFirst)
						default:
							assert.Fail(t, "There should not be a connecting vector with Z value %d", cVec.Z())
						}
						if nextMain != Origin {
							// Find the connecting vector on the other side ( the opposite 1 or -1 on Z() )
							nextConnectingVectors := ctx.GetTrio(nextMain)
							for _, nbp := range nextConnectingVectors {
								if nbp.Z() == -cVec.Z() {
									assertValidConnDetails(t, mainPoint.Add(cVec), nextMain.Add(nbp), fmt.Sprint("Main Pos=", mainPoint,
										"next Pos=", nextMain, "trio index=", ctx.GetTrioIndex(ctx.GetDivByThree(mainPoint)),
										"main base vector", cVec, "next base vector", nbp))
								}
							}
						}
					}
				}
			}
		}
	}

}

func assertValidConnDetails(t *testing.T, p1, p2 Point, msg string) {
	connDetails1 := GetConnectionDetails(p1, p2)
	assert.NotEqual(t, EmptyConnDetails, connDetails1, msg)
	assert.Equal(t, MakeVector(p1, p2), connDetails1.Vector, msg)

	connDetails2 := GetConnectionDetails(p2, p1)
	assert.NotEqual(t, EmptyConnDetails, connDetails2, msg)
	assert.Equal(t, MakeVector(p2, p1), connDetails2.Vector, msg)
}
