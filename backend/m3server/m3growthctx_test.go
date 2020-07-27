package m3server

import (
	"fmt"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConnectionDetailsInGrowthContext(t *testing.T) {
	m3util.SetToTestMode()

	env := getServerFullTestDb(m3util.PointTestEnv)
	ppd, _ := getServerPointPackData(env)
	for _, ctxType := range m3point.GetAllContextTypes() {
		nbIndexes := ctxType.GetNbIndexes()
		for pIdx := 0; pIdx < nbIndexes; pIdx++ {
			runConnectionDetailsCheck(t, ppd.GetGrowthContextByTypeAndIndex(ctxType, pIdx))
		}
	}
}

func GetBaseTrioDetails(growthCtx m3point.GrowthContext, mainPoint m3point.Point, offset int) *m3point.TrioDetails {
	ppd := growthCtx.GetEnv().GetData(m3util.PointIdx).(m3point.PointPackDataIfc)
	return ppd.GetTrioDetails(growthCtx.GetBaseTrioIndex(ppd, growthCtx.GetBaseDivByThree(mainPoint), offset))
}

func runConnectionDetailsCheck(t *testing.T, growthCtx m3point.GrowthContext) {
	ppd, _ := getServerPointPackData(growthCtx.GetEnv())
	// For all trioIndex rotations, any 2 close nextMainPoint points there should be a connection details
	min := m3point.CInt(-5)
	max := m3point.CInt(5)
	for x := min; x < max; x++ {
		for y := min; y < max; y++ {
			for z := min; z < max; z++ {
				mainPoint := m3point.Point{x, y, z}.Mul(3)
				connectingVectors := GetBaseTrioDetails(growthCtx, mainPoint, 0).GetConnections()
				for _, conn := range connectingVectors {
					cVec := conn.Vector

					assertValidConnDetails(t, ppd, mainPoint, mainPoint.Add(cVec), fmt.Sprint("Main Pos", mainPoint, "base vector", cVec))

					nextMain := m3point.Origin
					switch cVec.X() {
					case 0:
						// Nothing out
					case 1:
						nextMain = mainPoint.Add(m3point.XFirst)
					case -1:
						nextMain = mainPoint.Sub(m3point.XFirst)
					default:
						assert.Fail(t, "There should not be a connecting vector with x value %d", cVec.X())
					}
					if nextMain != m3point.Origin {
						// Find the connecting vector on the other side ( the opposite 1 or -1 on X() )
						nextConnectingVectors := GetBaseTrioDetails(growthCtx, nextMain, 0).GetConnections()
						for _, conn := range nextConnectingVectors {
							nbp := conn.Vector
							if nbp.X() == -cVec.X() {
								assertValidConnDetails(t, ppd, mainPoint.Add(cVec), nextMain.Add(nbp), fmt.Sprint("Main Pos=", mainPoint,
									"next Pos=", nextMain, "trio index=", growthCtx.GetBaseTrioIndex(ppd, growthCtx.GetBaseDivByThree(mainPoint), 0),
									"nextMainPoint base vector", cVec, "next base vector", nbp))
							}
						}
					}

					nextMain = m3point.Origin
					switch cVec.Y() {
					case 0:
						// Nothing out
					case 1:
						nextMain = mainPoint.Add(m3point.YFirst)
					case -1:
						nextMain = mainPoint.Sub(m3point.YFirst)
					default:
						assert.Fail(t, "There should not be a connecting vector with y value %d", cVec.Y())
					}
					if nextMain != m3point.Origin {
						// Find the connecting vector on the other side ( the opposite 1 or -1 on Y() )
						nextConnectingVectors := GetBaseTrioDetails(growthCtx, nextMain, 0).GetConnections()
						for _, conn := range nextConnectingVectors {
							nbp := conn.Vector
							if nbp.Y() == -cVec.Y() {
								assertValidConnDetails(t, ppd, mainPoint.Add(cVec), nextMain.Add(nbp), fmt.Sprint("Main Pos=", mainPoint,
									"next Pos=", nextMain, "trio index=", growthCtx.GetBaseTrioIndex(ppd, growthCtx.GetBaseDivByThree(mainPoint), 0),
									"nextMainPoint base vector", cVec, "next base vector", nbp))
							}
						}
					}

					nextMain = m3point.Origin
					switch cVec.Z() {
					case 0:
						// Nothing out
					case 1:
						nextMain = mainPoint.Add(m3point.ZFirst)
					case -1:
						nextMain = mainPoint.Sub(m3point.ZFirst)
					default:
						assert.Fail(t, "There should not be a connecting vector with Z value %d", cVec.Z())
					}
					if nextMain != m3point.Origin {
						// Find the connecting vector on the other side ( the opposite 1 or -1 on Z() )
						nextConnectingVectors := GetBaseTrioDetails(growthCtx, nextMain, 0).GetConnections()
						for _, conn := range nextConnectingVectors {
							nbp := conn.Vector
							if nbp.Z() == -cVec.Z() {
								assertValidConnDetails(t, ppd, mainPoint.Add(cVec), nextMain.Add(nbp), fmt.Sprint("Main Pos=", mainPoint,
									"next Pos=", nextMain, "trio index=", growthCtx.GetBaseTrioIndex(ppd, growthCtx.GetBaseDivByThree(mainPoint), 0),
									"nextMainPoint base vector", cVec, "next base vector", nbp))
							}
						}
					}
				}
			}
		}
	}

}

func assertValidConnDetails(t *testing.T, ppd *PointPackData, p1, p2 m3point.Point, msg string) {
	connDetails1 := ppd.GetConnDetailsByPoints(p1, p2)
	assert.NotEqual(t, m3point.EmptyConnDetails, connDetails1, msg)
	assert.Equal(t, m3point.MakeVector(p1, p2), connDetails1.Vector, msg)

	connDetails2 := ppd.GetConnDetailsByPoints(p2, p1)
	assert.NotEqual(t, m3point.EmptyConnDetails, connDetails2, msg)
	assert.Equal(t, m3point.MakeVector(p2, p1), connDetails2.Vector, msg)
}
