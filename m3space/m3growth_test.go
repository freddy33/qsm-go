package m3space

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"fmt"
)

func getAllContexes() map[uint8][]GrowthContext {
	res := make(map[uint8][]GrowthContext)
	res[1] = make([]GrowthContext, 0, 8)
	res[3] = make([]GrowthContext, 0, 8*4)
	res[2] = make([]GrowthContext, 0, 12*2*2)
	res[4] = make([]GrowthContext, 0, 12*4*2)
	res[8] = make([]GrowthContext, 0, 12*8*2)

	for pIdx := 0; pIdx < 8; pIdx++ {
		res[1] = append(res[1], GrowthContext{&Origin, 1, pIdx, false, 0,})
		for offset := 0; offset < 3; offset++ {
			res[3] = append(res[3], GrowthContext{&Origin, 3, pIdx, false, offset,})
		}
	}

	for _, pType := range [4]uint8{2, 4, 8} {
		for pIdx := 0; pIdx < 12; pIdx++ {
			for offset := 0; offset < int(pType); offset++ {
				res[pType] = append(res[pType], GrowthContext{&Origin, pType, pIdx, false, offset,})
				res[pType] = append(res[pType], GrowthContext{&Origin, pType, pIdx, true, offset,})
			}
		}
	}
	return res
}

func TestDivByThree(t *testing.T) {
	DEBUG = true
	someCenter1 := Point{3,-6,9}
	ctx := GrowthContext{&someCenter1, 1, 1, false, 0,}
	assert.Equal(t, someCenter1, *(ctx.center))
	assert.Equal(t, uint8(1), ctx.permutationType)
	assert.Equal(t, 1, ctx.permutationIndex)
	assert.Equal(t, false, ctx.permutationNegFlow)
	assert.Equal(t, 0, ctx.permutationOffset)

	assert.Equal(t, int64(1), Point{0,-6,9}.GetDivByThree(&ctx))
	assert.Equal(t, int64(1), Point{6,-6,9}.GetDivByThree(&ctx))
	assert.Equal(t, int64(1), Point{3,-3,9}.GetDivByThree(&ctx))
	assert.Equal(t, int64(1), Point{3,-9,9}.GetDivByThree(&ctx))
	assert.Equal(t, int64(1), Point{3,-6,12}.GetDivByThree(&ctx))
	assert.Equal(t, int64(1), Point{3,-6,6}.GetDivByThree(&ctx))

	assert.Equal(t, int64(6), Point{0,0,0}.GetDivByThree(&ctx))

	// Verify trio index unaffected
	for d := int64(-10); d < 10; d++ {
		assert.Equal(t, 1, ctx.GetTrioIndex(d), "failed trio index for ctx %v and divByThree=%d", ctx, d)
	}

}

func TestGrowthContext1(t *testing.T) {
	DEBUG = true
	ctx := GrowthContext{&Origin, 1, 3, false, 0,}
	assert.Equal(t, uint8(1), ctx.permutationType)
	assert.Equal(t, 3, ctx.permutationIndex)
	assert.Equal(t, false, ctx.permutationNegFlow)
	assert.Equal(t, 0, ctx.permutationOffset)
	for d := int64(-10); d < 10; d++ {
		assert.Equal(t, 3, ctx.GetTrioIndex(d), "failed trio index for ctx %v and divByThree=%d", ctx, d)
	}
	ctx.permutationIndex = 4
	ctx.permutationNegFlow = true
	ctx.permutationOffset = 2
	assert.Equal(t, uint8(1), ctx.permutationType)
	assert.Equal(t, 4, ctx.permutationIndex)
	assert.Equal(t, true, ctx.permutationNegFlow)
	assert.Equal(t, 2, ctx.permutationOffset)
	for d := int64(-10); d < 10; d++ {
		assert.Equal(t, 4, ctx.GetTrioIndex(d), "failed trio index for ctx %v and divByThree=%d", ctx, d)
	}
}

func TestGrowthContext3(t *testing.T) {
	DEBUG = true
	ctx := GrowthContext{&Origin, 3, 0, false, 0,}
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

func TestGrowthContexes(t *testing.T) {
	DEBUG = true

	growthContexts := getAllContexes()
	for _, ctx := range growthContexts[1] {
		assert.Equal(t, uint8(1), ctx.permutationType)
		for d := int64(-10); d < 10; d++ {
			assert.Equal(t, ctx.permutationIndex, ctx.GetTrioIndex(d), "failed trio index for ctx %v and divByThree=%d", ctx, d)
		}
	}
	/*
	for _, ctx := range growthContexts[2] {
		assert.Equal(t, 2, ctx.permutationType)
		if !ctx.permutationNegFlow {
			// Positive flow
			for d := int64(-10); d < 10; d++ {
				assert.Equal(t, ctx.permutationIndex, ctx.GetTrioIndex(d), "failed trio index for ctx %v and divByThree=%d", ctx, d)
			}
		} else {
			// Neg flow
		}
	}
	*/
}

func TestConnectionDetails(t *testing.T) {
	DEBUG = true
	InitConnectionDetails()
	for k, v := range AllConnectionsPossible {
		assert.Equal(t, k, v.Vector)
		currentNumber := v.ConnNumber
		sameNumber := 0
		for _, nv := range AllConnectionsPossible {
			if nv.ConnNumber == currentNumber {
				sameNumber++
				if nv.Vector != v.Vector {
					assert.Equal(t, nv.Vector.Neg(), v.Vector, "Should have neg vector")
					assert.Equal(t, !nv.ConnNeg, v.ConnNeg, "Should have opposite connNeg flag")
				}
			}
		}
		assert.Equal(t, 2, sameNumber, "Should have 2 with same conn number for %d", currentNumber)
	}

	countPosConnNumbers := make(map[uint8]int)
	countNegConnNumbers := make(map[uint8]int)
	for i, tA := range AllBaseTrio {
		for j, tB := range AllBaseTrio {
			connVectors := GetNonBaseConnections(tA, tB)
			for k, connVector := range connVectors {
				connDetails, ok := AllConnectionsPossible[connVector]
				assert.True(t, ok, "Connection between 2 trio (%d,%d) number %k is not in conn details", i, j, k)
				assert.Equal(t, connVector, connDetails.Vector, "Connection between 2 trio (%d,%d) number %k is not in conn details", i, j, k)
				if connDetails.ConnNeg {
					countNegConnNumbers[connDetails.ConnNumber]++
				} else {
					countPosConnNumbers[connDetails.ConnNumber]++
				}
			}
		}
	}

	allCtx := getAllContexes()
	assert.Equal(t, 5, len(allCtx))

	nbCtx := 0
	for _, contextList := range allCtx {
		nbCtx += len(contextList)
	}
	fmt.Println("Created", nbCtx, "contexes")
	fmt.Println("Using", len(allCtx[8]), " contexes from the 8 context")
	// For all trioIndex rotations, any 2 close main points there should be a connection details
	min := int64(-2) // -5
	max := int64(2)  // 5
	for _, ctx := range allCtx[8] {
		for x := min; x < max; x++ {
			for y := min; y < max; y++ {
				for z := min; z < max; z++ {
					mainPoint := Point{x, y, z}.Mul(3)
					connectingVectors := mainPoint.GetTrio(&ctx)
					for _, cVec := range connectingVectors {

						assertValidConnDetails(t, mainPoint, mainPoint.Add(cVec), fmt.Sprint("Main Point", mainPoint, "base vector", cVec))

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
							nextConnectingVectors := nextMain.GetTrio(&ctx)
							for _, nbp := range nextConnectingVectors {
								if nbp.X() == -cVec.X() {
									assertValidConnDetails(t, mainPoint.Add(cVec), nextMain.Add(nbp), fmt.Sprint("Main Point=", mainPoint,
										"next point=", nextMain, "trio index=", mainPoint.GetTrioIndex(&ctx),
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
							nextConnectingVectors := nextMain.GetTrio(&ctx)
							for _, nbp := range nextConnectingVectors {
								if nbp.Y() == -cVec.Y() {
									assertValidConnDetails(t, mainPoint.Add(cVec), nextMain.Add(nbp), fmt.Sprint("Main Point=", mainPoint,
										"next point=", nextMain, "trio index=", mainPoint.GetTrioIndex(&ctx),
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
							nextConnectingVectors := nextMain.GetTrio(&ctx)
							for _, nbp := range nextConnectingVectors {
								if nbp.Z() == -cVec.Z() {
									assertValidConnDetails(t, mainPoint.Add(cVec), nextMain.Add(nbp), fmt.Sprint("Main Point=", mainPoint,
										"next point=", nextMain, "trio index=", mainPoint.GetTrioIndex(&ctx),
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
	assert.Equal(t, p2.Sub(p1), connDetails1.Vector, msg)

	connDetails2 := GetConnectionDetails(p2, p1)
	assert.NotEqual(t, EmptyConnDetails, connDetails2, msg)
	assert.Equal(t, p1.Sub(p2), connDetails2.Vector, msg)
}
