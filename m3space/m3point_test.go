package m3space

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"fmt"
	"math/rand"
)

func TestPosMod4(t *testing.T) {
	DEBUG = true
	assert.Equal(t, int64(1), PosMod4(5))
	assert.Equal(t, int64(0), PosMod4(4))
	assert.Equal(t, int64(3), PosMod4(3))
	assert.Equal(t, int64(2), PosMod4(2))
	assert.Equal(t, int64(1), PosMod4(1))
	assert.Equal(t, int64(0), PosMod4(0))
	assert.Equal(t, int64(3), PosMod4(-1))
	assert.Equal(t, int64(2), PosMod4(-2))
	assert.Equal(t, int64(1), PosMod4(-3))
	assert.Equal(t, int64(0), PosMod4(-4))
	assert.Equal(t, int64(3), PosMod4(-5))
}

func TestPoint(t *testing.T) {
	DEBUG = true

	Orig := Point{0, 0, 0}
	OneTwoThree := Point{1, 2, 3}
	P := Point{17, 11, 13}

	// Test equal
	assert.Equal(t, Orig, Origin)
	assert.Equal(t, Point{1, 2, 3}, OneTwoThree)
	assert.Equal(t, Point{17, 11, 13}, P)

	// Test DS
	assert.Equal(t, int64(3), DS(&OneTwoThree, &Point{0, 1, 2}))
	// Make sure OneTwoThree did not change
	assert.Equal(t, Point{1, 2, 3}, OneTwoThree)

	assert.Equal(t, int64(4), DS(&OneTwoThree, &Point{-1, 2, 3}))
	assert.Equal(t, int64(16), DS(&OneTwoThree, &Point{1, -2, 3}))
	assert.Equal(t, int64(36), DS(&OneTwoThree, &Point{1, 2, -3}))

	// Test Add
	assert.Equal(t, Point{3, 0, 0}, Orig.Add(XFirst))
	// Make sure orig did not change
	assert.Equal(t, Orig, Origin)
	assert.Equal(t, Point{0, 3, 0}, Orig.Add(YFirst))
	assert.Equal(t, Point{0, 0, 3}, Orig.Add(ZFirst))
	assert.Equal(t, Point{18, 13, 16}, P.Add(OneTwoThree))
	// Make sure P and OneTwoThree did not change
	assert.Equal(t, Point{17, 11, 13}, P)
	assert.Equal(t, Point{1, 2, 3}, OneTwoThree)

	// Test Sub
	assert.Equal(t, Point{-3, 0, 0}, Orig.Sub(XFirst))
	// Make sure orig did not change
	assert.Equal(t, Orig, Origin)

	assert.Equal(t, Point{0, -3, 0}, Orig.Sub(YFirst))
	assert.Equal(t, Point{0, 0, -3}, Orig.Sub(ZFirst))
	assert.Equal(t, Point{16, 9, 10}, P.Sub(OneTwoThree))
	// Make sure P and OneTwoThree did not change
	assert.Equal(t, Point{17, 11, 13}, P)
	assert.Equal(t, Point{1, 2, 3}, OneTwoThree)

	// Test Neg
	assert.Equal(t, Point{-1, -2, -3}, OneTwoThree.Neg())
	// Make sure OneTwoThree did not change
	assert.Equal(t, Point{1, 2, 3}, OneTwoThree)

	// Test Mul
	assert.Equal(t, OneTwoThree.Mul(2), Point{2, 4, 6})
	// Make sure OneTwoThree did not change
	assert.Equal(t, OneTwoThree, Point{1, 2, 3})
	assert.Equal(t, OneTwoThree.Mul(-3), Point{-3, -6, -9})

	// Test PlusX, NegX, PlusY, NegY, PlusZ, NegZ
	assert.Equal(t, OneTwoThree.PlusX(), Point{1, -3, 2})
	assert.Equal(t, OneTwoThree.NegX(), Point{1, 3, -2})
	assert.Equal(t, OneTwoThree.PlusY(), Point{3, 2, -1})
	assert.Equal(t, OneTwoThree.NegY(), Point{-3, 2, 1})
	assert.Equal(t, OneTwoThree.PlusZ(), Point{-2, 1, 3})
	assert.Equal(t, OneTwoThree.NegZ(), Point{2, -1, 3})

	// Test bunch of equations using random points
	nbRun := 100
	rdMax := int64(100000000)
	for i := 0; i < nbRun; i++ {
		randomPoint := Point{randomInt64(rdMax), randomInt64(rdMax), randomInt64(rdMax)}
		assert.Equal(t, Orig.Sub(randomPoint), randomPoint.Neg())
		assert.Equal(t, randomPoint.Sub(randomPoint.Add(OneTwoThree)), OneTwoThree.Neg())
		assert.Equal(t, randomPoint.Sub(randomPoint.Add(OneTwoThree)), OneTwoThree.Mul(-1))
		assert.Equal(t, randomPoint.Add(randomPoint.Neg()), Orig)
		assert.Equal(t, randomPoint.Add(randomPoint.Mul(-1)), Orig)

		assert.Equal(t, randomPoint.PlusX().NegX(), randomPoint)
		assert.Equal(t, randomPoint.NegX().PlusX(), randomPoint)
		assert.Equal(t, randomPoint.PlusY().NegY(), randomPoint)
		assert.Equal(t, randomPoint.NegY().PlusY(), randomPoint)
		assert.Equal(t, randomPoint.PlusZ().NegZ(), randomPoint)
		assert.Equal(t, randomPoint.NegZ().PlusZ(), randomPoint)

		assert.Equal(t, randomPoint.PlusX().PlusX().PlusX().PlusX(), randomPoint)
		assert.Equal(t, randomPoint.PlusY().PlusY().PlusY().PlusY(), randomPoint)
		assert.Equal(t, randomPoint.PlusZ().PlusZ().PlusZ().PlusZ(), randomPoint)
		assert.Equal(t, randomPoint.NegX().NegX().NegX().NegX(), randomPoint)
		assert.Equal(t, randomPoint.NegY().NegY().NegY().NegY(), randomPoint)
		assert.Equal(t, randomPoint.NegZ().NegZ().NegZ().NegZ(), randomPoint)

		assert.Equal(t, randomPoint.NegX().NegX(), randomPoint.PlusX().PlusX())
		assert.Equal(t, randomPoint.NegY().NegY(), randomPoint.PlusY().PlusY())
		assert.Equal(t, randomPoint.NegZ().NegZ(), randomPoint.PlusZ().PlusZ())
	}
}

func TestConnectingVectors(t *testing.T) {
	DEBUG = true
	assert.Equal(t, AllBaseTrio[0][0], Point{1, 1, 0})
	assert.Equal(t, AllBaseTrio[0][1], Point{-1, 0, -1})
	assert.Equal(t, AllBaseTrio[0][1], AllBaseTrio[0][0].PlusX().PlusY().PlusY())
	assert.Equal(t, AllBaseTrio[0][2], Point{0, -1, 1})
	assert.Equal(t, AllBaseTrio[0][2], AllBaseTrio[0][0].PlusY().PlusX().PlusX())

	for i := 0; i < 4; i++ {
		BackToOrig := Origin
		BackToOrig2 := Origin
		for j := 0; j < 3; j++ {
			assert.Equal(t, int64(2), AllBaseTrio[i][j].DistanceSquared(), "Something wrong with size of connecting vector %d %d", i, j)
			assert.Equal(t, int64(2), AllBaseTrio[i+4][j].DistanceSquared(), "Something wrong with size of connecting vector 2 %d %d", i, j)
			for c := 0; c < 3; c++ {
				abs := Abs(AllBaseTrio[i][j][c])
				assert.True(t, int64(1) == abs || int64(0) == abs, "Something wrong with coordinate of connecting vector %d %d %d = %d", i, j, c, AllBaseTrio[i][j][c])
				abs = Abs(AllBaseTrio[i+4][j][c])
				assert.True(t, int64(1) == abs || int64(0) == abs, "Something wrong with coordinate of connecting vector 2 %d %d %d = %d", i, j, c, AllBaseTrio[i][j][c])
			}
			BackToOrig = BackToOrig.Add(AllBaseTrio[i][j])
			BackToOrig2 = BackToOrig2.Add(AllBaseTrio[i+4][j])
		}
		assert.Equal(t, Origin, BackToOrig, "Something wrong with sum of connecting vectors %d", i)
		assert.Equal(t, Origin, BackToOrig2, "Something wrong with sum of connecting vectors 2 %d", i)
	}
}

func TestConnectingVectorsRotation(t *testing.T) {
	// For each axe (first index), the three connecting vectors evolves with PlusX, plusY and plusZ
	currentConnectingVectors := [3][3]Point{}
	currentConnectingVectors2 := [3][3]Point{}
	for axe := 0; axe < 3; axe++ {
		currentConnectingVectors[axe][0] = AllBaseTrio[0][0]
		currentConnectingVectors[axe][1] = AllBaseTrio[0][1]
		currentConnectingVectors[axe][2] = AllBaseTrio[0][2]

		currentConnectingVectors2[axe][0] = AllBaseTrio[4][0]
		currentConnectingVectors2[axe][1] = AllBaseTrio[4][1]
		currentConnectingVectors2[axe][2] = AllBaseTrio[4][2]
	}

	for k := -4; k < 6; k++ {
		mapColumn := int(PosMod4(int64(k)))
		fmt.Println("Checking map column", mapColumn, "from", k)
		assertSameTrio(t, AllBaseTrio[NextMapping[0][mapColumn]], currentConnectingVectors[0])
		assertSameTrio(t, AllBaseTrio[NextMapping[1][mapColumn]], currentConnectingVectors[1])
		assertSameTrio(t, AllBaseTrio[NextMapping[2][mapColumn]], currentConnectingVectors[2])

		assertSameTrio(t, AllBaseTrio[NextMapping2[0][mapColumn]+4], currentConnectingVectors2[0])
		assertSameTrio(t, AllBaseTrio[NextMapping2[1][mapColumn]+4], currentConnectingVectors2[1])
		assertSameTrio(t, AllBaseTrio[NextMapping2[2][mapColumn]+4], currentConnectingVectors2[2])

		for i := 0; i < 3; i++ {
			currentConnectingVectors[0][i] = currentConnectingVectors[0][i].PlusX()
			currentConnectingVectors[1][i] = currentConnectingVectors[1][i].PlusY()
			currentConnectingVectors[2][i] = currentConnectingVectors[2][i].PlusZ()

			currentConnectingVectors2[0][i] = currentConnectingVectors2[0][i].PlusX()
			currentConnectingVectors2[1][i] = currentConnectingVectors2[1][i].PlusY()
			currentConnectingVectors2[2][i] = currentConnectingVectors2[2][i].PlusZ()
		}
	}

	nbRun := 100
	rdMax := int64(100000000000)
	for i := 0; i < nbRun; i++ {
		randomPoint := Point{randomInt64(rdMax) * 3, randomInt64(rdMax) * 3, randomInt64(rdMax) * 3}
		assert.True(t, randomPoint.IsMainPoint(), "point %v should be main", randomPoint)
		mod4Point := randomPoint.GetMod4Point()
		mod4Val := randomPoint.GetMod4Value()
		assert.True(t, mod4Val >= 0 && mod4Val < 4, "Mod4 does not exists for %v mod4 point %v", randomPoint, mod4Point)
	}
}

func TestAllTrio(t *testing.T) {
	DEBUG = true
	print := false
	for i, tr := range AllBaseTrio {
		assert.Equal(t, int64(0), tr[0][2], "Failed on Trio %d", i)
		assert.Equal(t, int64(0), tr[1][1], "Failed on Trio %d", i)
		assert.Equal(t, int64(0), tr[2][0], "Failed on Trio %d", i)
		for j, tB := range AllBaseTrio {
			if print {
				fmt.Println(i, ",", j)
			}
			assertIsGenericNonBaseConnectingVector(t, GetNonBaseConnections(tr, tB), i, j, print)
		}
	}
	for i, mr := range AllMod4Rotations {
		for j := 0; j < 4; j++ {
			// All conns are 3 or 1, no more 5
			assertIsThreeOr1NonBaseConnectingVector(t, GetNonBaseConnections(AllBaseTrio[mr[j]], AllBaseTrio[mr[(j+1)%4]]), i, j)
		}
	}
}

func assertIsGenericNonBaseConnectingVector(t *testing.T, conns [6]Point, i, j int, print bool) {
	for _, conn := range conns {
		assert.True(t, conn.IsConnectionVector(), "Found wrong connection %v at %d %d", conn, i, j)
		assert.False(t, conn.IsBaseConnectingVector(), "Found wrong connection %v at %d %d", conn, i, j)
		ds := conn.DistanceSquared()
		assert.True(t, ds == 1 || ds == 3 || ds == 5, "Found wrong connection %v at %d %d", conn, i, j)
		if print {
			fmt.Println("\t", conn, "\t", ds)
		}
	}
}

func assertIsThreeOr1NonBaseConnectingVector(t *testing.T, conns [6]Point, i, j int) {
	for _, conn := range conns {
		assert.True(t, conn.IsConnectionVector(), "Found wrong connection %v at %d %d", conn, i, j)
		assert.False(t, conn.IsBaseConnectingVector(), "Found wrong connection %v at %d %d", conn, i, j)
		ds := conn.DistanceSquared()
		assert.True(t, ds == 1 || ds == 3, "Found wrong connection %v at %d %d", conn, i, j)
	}
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

	// For all mod4 rotations, any 2 close main points there should be a connection details
	min := int64(-10) // -5
	max := int64(10)  // 5
	//for _, mod4Rot := range AllMod4Rotations {
		for x := min; x < max; x++ {
			for y := min; y < max; y++ {
				for z := min; z < max; z++ {
					mainPoint := Point{x, y, z}.Mul(3)
					connectingVectors := mainPoint.GetTrio()
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
							nextConnectingVectors := nextMain.GetTrio()
							for _, nbp := range nextConnectingVectors {
								if nbp.X() == -cVec.X() {
									assertValidConnDetails(t, mainPoint.Add(cVec), nextMain.Add(nbp), fmt.Sprint("Main Point", mainPoint, "mod4", mainPoint.GetMod4Value(),
										"next point", nextMain, "mod4", mainPoint.GetMod4Value(),
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
							nextConnectingVectors := nextMain.GetTrio()
							for _, nbp := range nextConnectingVectors {
								if nbp.Y() == -cVec.Y() {
									assertValidConnDetails(t, mainPoint.Add(cVec), nextMain.Add(nbp), fmt.Sprint("Main Point", mainPoint, "mod4", mainPoint.GetMod4Value(),
										"next point", nextMain, "mod4", mainPoint.GetMod4Value(),
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
							nextConnectingVectors := nextMain.GetTrio()
							for _, nbp := range nextConnectingVectors {
								if nbp.Z() == -cVec.Z() {
									assertValidConnDetails(t, mainPoint.Add(cVec), nextMain.Add(nbp), fmt.Sprint("Main Point", mainPoint, "mod4", mainPoint.GetMod4Value(),
										"next point", nextMain, "mod4", mainPoint.GetMod4Value(),
										"main base vector", cVec, "next base vector", nbp))
								}
							}
						}
					}
				}
			}
		//}
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

func randomInt64(max int64) int64 {
	r := rand.Int63n(max)
	if rand.Float32() < 0.5 {
		return -r
	}
	return r
}

// Verify the 2 arrays are actually identical just not particular order
func assertSameTrio(t *testing.T, t1 [3]Point, t2 [3]Point) {
	for _, p1 := range t1 {
		found := false
		for _, p2 := range t2 {
			if p1 == p2 {
				found = true
			}
		}
		assert.True(t, found, "Did not find point %v in Trio %v", p1, t2)
	}
	for _, p2 := range t2 {
		found := false
		for _, p1 := range t1 {
			if p1 == p2 {
				found = true
			}
		}
		assert.True(t, found, "Did not find point %v in Trio %v", p2, t1)
	}
}
