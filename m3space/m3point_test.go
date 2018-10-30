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

func TestInitialTrioConnectingVectors(t *testing.T) {
	DEBUG = true
	assert.Equal(t, AllBaseTrio[0][0], Point{1, 1, 0})
	assert.Equal(t, AllBaseTrio[0][1], Point{-1, 0, -1})
	assert.Equal(t, AllBaseTrio[0][1], AllBaseTrio[0][0].PlusX().PlusY().PlusY())
	assert.Equal(t, AllBaseTrio[0][2], Point{0, -1, 1})
	assert.Equal(t, AllBaseTrio[0][2], AllBaseTrio[0][0].PlusY().PlusX().PlusX())

	assert.Equal(t, AllBaseTrio[4][0], Point{1, 1, 0})
	assert.Equal(t, AllBaseTrio[4][1], Point{-1, 0, 1})
	assert.Equal(t, AllBaseTrio[4][2], Point{0, -1, -1})
}

var Full5NextTrio = [4][2]int{
	{0, 5}, {1, 6}, {2, 7}, {3, 4},
}

func TestAllTrio(t *testing.T) {
	DEBUG = true
	outTable := false
	for i, tr := range AllBaseTrio {
		assert.Equal(t, int64(0), tr[0][2], "Failed on Trio %d", i)
		assert.Equal(t, int64(0), tr[1][1], "Failed on Trio %d", i)
		assert.Equal(t, int64(0), tr[2][0], "Failed on Trio %d", i)
		BackToOrig := Origin
		for j, vec := range tr {
			for c := 0; c < 3; c++ {
				abs := Abs(vec[c])
				assert.True(t, int64(1) == abs || int64(0) == abs, "Something wrong with coordinate of connecting vector %d %d %d = %v", i, j, c, vec)
			}
			assert.Equal(t, int64(2), vec.DistanceSquared(), "Failed vec at %d %d", i, j)
			assert.True(t, vec.IsBaseConnectingVector(), "Failed vec at %d %d", i, j)
			BackToOrig = BackToOrig.Add(vec)
		}
		assert.Equal(t, Origin, BackToOrig, "Something wrong with sum of Trio %d %v", i, tr)
		for j, tB := range AllBaseTrio {
			if outTable {
				fmt.Println(i, ",", j)
			}
			assertIsGenericNonBaseConnectingVector(t, GetNonBaseConnections(tr, tB), i, j, outTable)
		}
	}

	idxMap := createAll8IndexMap()
	for i, nextTrio := range Full5NextTrio {
		assertValidNextTrio(t, nextTrio, i)

		// All conns are only 5
		assertIsFull5NonBaseConnectingVector(t, GetNonBaseConnections(AllBaseTrio[nextTrio[0]], AllBaseTrio[nextTrio[1]]), i, -1)
		idxMap[nextTrio[0]]++
		idxMap[nextTrio[1]]++
	}
	assertAllIndexUsed(t, idxMap, 1, "full 5 trios")

	idxMap = createAll8IndexMap()
	for i, nextTrio := range ValidNextTrio {
		assertValidNextTrio(t, nextTrio, i)

		// All conns are 3 or 1, no more 5
		assertIsThreeOr1NonBaseConnectingVector(t, GetNonBaseConnections(AllBaseTrio[nextTrio[0]], AllBaseTrio[nextTrio[1]]), i, -1)
		idxMap[nextTrio[0]]++
		idxMap[nextTrio[1]]++
	}
	assertAllIndexUsed(t, idxMap, 3, "valid trios")

	idxMap = createAll8IndexMap()
	for i, permutMap := range AllMod4Permutations {
		for j := 0; j < 4; j++ {
			startIdx := permutMap[j]
			endIdx := permutMap[(j+1)%4]
			assertExistsInValidNextTrio(t, startIdx, endIdx, fmt.Sprint("in mod4 permutation[", i, "]=", permutMap, "idx", j))
			idxMap[permutMap[j]]++
		}
	}
	assertAllIndexUsed(t, idxMap, 6, "all mod4 permutations")

	for i, permutMap := range AllMod8Permutations {
		idxMap = createAll8IndexMap()
		for j := 0; j < 8; j++ {
			startIdx := permutMap[j]
			endIdx := permutMap[(j+1)%8]
			assertExistsInValidNextTrio(t, startIdx, endIdx, fmt.Sprint("in mod8 permutation[", i, "]=", permutMap, "idx", j))
			idxMap[permutMap[j]]++
		}
		assertAllIndexUsed(t, idxMap, 1, fmt.Sprint("in mod8 permutation[", i, "]=", permutMap))
	}

}

func assertExistsInValidNextTrio(t *testing.T, startIdx int, endIdx int, msg string) {
	assert.NotEqual(t, startIdx, endIdx, "start and end index cannot be equal for %s", msg)
	// Order the indexes
	trioToFind := [2]int{-1, -1}
	if startIdx >= 4 {
		trioToFind[1] = startIdx
	} else {
		trioToFind[0] = startIdx
	}
	if endIdx >= 4 {
		trioToFind[1] = endIdx
	} else {
		trioToFind[0] = endIdx
	}

	assert.True(t, trioToFind[0] >= 0 && trioToFind[0] <= 3, "Something wrong with trioToFind first value for %s", msg)
	assert.True(t, trioToFind[1] >= 4 && trioToFind[1] <= 7, "Something wrong with trioToFind second value for %s", msg)

	foundNextTrio := false
	for _, nextTrio := range ValidNextTrio {
		if trioToFind == nextTrio {
			foundNextTrio = true
		}
	}
	assert.True(t, foundNextTrio, "Did not find trio %v in list of valid trio for %s", trioToFind, msg)
}

func assertValidNextTrio(t *testing.T, nextTrio [2]int, i int) {
	assert.NotEqual(t, nextTrio[0], nextTrio[1], "Something wrong with nextTrio index %d %v", i, nextTrio)
	assert.True(t, nextTrio[0] >= 0 && nextTrio[0] <= 3, "Something wrong with nextTrio first value index %d %v", i, nextTrio)
	assert.True(t, nextTrio[1] >= 4 && nextTrio[1] <= 7, "Something wrong with nextTrio second value index %d %v", i, nextTrio)
}

func createAll8IndexMap() map[int]int {
	res := make(map[int]int)
	for i := 0; i < 8; i++ {
		res[i] = 0
	}
	return res
}

func assertAllIndexUsed(t *testing.T, idxMap map[int]int, expectedTimes int, msg string) {
	assert.Equal(t, 8, len(idxMap))
	for i := 0; i < 8; i++ {
		v, ok := idxMap[i]
		assert.True(t, ok, "did not find index %d in %v for %s", i, idxMap, msg)
		assert.Equal(t, expectedTimes, v, "failed nb times at index %d in %v for %s", i, idxMap, msg)
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
		assert.True(t, conn.IsOnlyOneAndZero(), "Found wrong connection %v at %d %d", conn, i, j)
		assert.True(t, conn.IsConnectionVector(), "Found wrong connection %v at %d %d", conn, i, j)
		assert.False(t, conn.IsBaseConnectingVector(), "Found wrong connection %v at %d %d", conn, i, j)
		ds := conn.DistanceSquared()
		assert.True(t, ds == 1 || ds == 3, "Found wrong connection %v at %d %d", conn, i, j)
	}
}

func assertIsFull5NonBaseConnectingVector(t *testing.T, conns [6]Point, i, j int) {
	for _, conn := range conns {
		assert.True(t, conn.IsOnlyTwoOneAndZero(), "Found wrong connection %v at %d %d", conn, i, j)
		assert.True(t, conn.IsConnectionVector(), "Found wrong connection %v at %d %d", conn, i, j)
		assert.False(t, conn.IsBaseConnectingVector(), "Found wrong connection %v at %d %d", conn, i, j)
		ds := conn.DistanceSquared()
		assert.True(t, ds == 5, "Found wrong connection %v at %d %d", conn, i, j)
	}
}

func getAllContexes() []GrowthContext {
	res := make([]GrowthContext, 0, 8+3*12*2*(2+4+8))

	for pIdx := 0; pIdx < 8; pIdx++ {
		res = append(res, GrowthContext{1, pIdx, false, 0,})
	}

	for _, pType := range [4]uint8{2, 4, 8} {
		for pIdx := 0; pIdx < 12; pIdx++ {
			for offset := 0; offset < int(pType); offset++ {
				res = append(res, GrowthContext{pType, pIdx, false, offset,})
				res = append(res, GrowthContext{pType, pIdx, true, offset,})
			}
		}
	}
	return res
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

	allCtx := getAllContexes()
	fmt.Println("Created", len(allCtx), "contexes")
	// For all trioIndex rotations, any 2 close main points there should be a connection details
	min := int64(-2) // -5
	max := int64(2)  // 5
	for _, ctx := range allCtx {
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

func randomInt64(max int64) int64 {
	r := rand.Int63n(max)
	if rand.Float32() < 0.5 {
		return -r
	}
	return r
}
