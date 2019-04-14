package m3point

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConnectionDetails(t *testing.T) {
	Log.Level = m3util.DEBUG
	for k, v := range AllConnectionsPossible {
		assert.Equal(t, k, v.Vector)
		assert.Equal(t, k.DistanceSquared(), v.DistanceSquared())
		currentNumber := v.GetPosIntId()
		sameNumber := 0
		for _, nv := range AllConnectionsPossible {
			if nv.GetPosIntId() == currentNumber {
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
	Log.Info("ConnId usage:", countConnId)

}

func TestConnectionDetailsInGrowthContext(t *testing.T) {
	allCtx := getAllTestContexts()
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
	connDetails1 := GetConnDetailsByPoints(p1, p2)
	assert.NotEqual(t, EmptyConnDetails, connDetails1, msg)
	assert.Equal(t, MakeVector(p1, p2), connDetails1.Vector, msg)

	connDetails2 := GetConnDetailsByPoints(p2, p1)
	assert.NotEqual(t, EmptyConnDetails, connDetails2, msg)
	assert.Equal(t, MakeVector(p2, p1), connDetails2.Vector, msg)
}
