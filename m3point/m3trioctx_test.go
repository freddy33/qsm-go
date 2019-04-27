package m3point

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConnectionDetailsInGrowthContext(t *testing.T) {
	for _, ctxType := range GetAllContextTypes() {
		nbIndexes := ctxType.GetNbIndexes()
		for pIdx := 0; pIdx < nbIndexes; pIdx++ {
			trCtx := GetTrioIndexContext(ctxType, pIdx)
			runConnectionDetailsCheck(t, trCtx)
		}
	}
}

func runConnectionDetailsCheck(t *testing.T, trCtx *TrioIndexContext) {
	// For all trioIndex rotations, any 2 close nextMainPoint points there should be a connection details
	min := int64(-5)
	max := int64(5)
	for x := min; x < max; x++ {
		for y := min; y < max; y++ {
			for z := min; z < max; z++ {
				mainPoint := Point{x, y, z}.Mul(3)
				connectingVectors := trCtx.GetBaseTrio(mainPoint, 0)
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
						nextConnectingVectors := trCtx.GetBaseTrio(nextMain, 0)
						for _, nbp := range nextConnectingVectors {
							if nbp.X() == -cVec.X() {
								assertValidConnDetails(t, mainPoint.Add(cVec), nextMain.Add(nbp), fmt.Sprint("Main Pos=", mainPoint,
									"next Pos=", nextMain, "trio index=", trCtx.GetBaseTrioIndex(trCtx.GetBaseDivByThree(mainPoint), 0),
									"nextMainPoint base vector", cVec, "next base vector", nbp))
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
						nextConnectingVectors := trCtx.GetBaseTrio(nextMain, 0)
						for _, nbp := range nextConnectingVectors {
							if nbp.Y() == -cVec.Y() {
								assertValidConnDetails(t, mainPoint.Add(cVec), nextMain.Add(nbp), fmt.Sprint("Main Pos=", mainPoint,
									"next Pos=", nextMain, "trio index=", trCtx.GetBaseTrioIndex(trCtx.GetBaseDivByThree(mainPoint), 0),
									"nextMainPoint base vector", cVec, "next base vector", nbp))
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
						nextConnectingVectors := trCtx.GetBaseTrio(nextMain, 0)
						for _, nbp := range nextConnectingVectors {
							if nbp.Z() == -cVec.Z() {
								assertValidConnDetails(t, mainPoint.Add(cVec), nextMain.Add(nbp), fmt.Sprint("Main Pos=", mainPoint,
									"next Pos=", nextMain, "trio index=", trCtx.GetBaseTrioIndex(trCtx.GetBaseDivByThree(mainPoint), 0),
									"nextMainPoint base vector", cVec, "next base vector", nbp))
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
