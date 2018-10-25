package m3space

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"fmt"
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

func TestBasePoints(t *testing.T) {
	DEBUG = true
	assert.Equal(t, BasePoints[0][0], Point{1, 1, 0})
	assert.Equal(t, BasePoints[0][1], Point{0, -1, 1})
	assert.Equal(t, BasePoints[0][1], BasePoints[0][0].PlusY().PlusX().PlusX())
	assert.Equal(t, BasePoints[0][2], Point{-1, 0, -1})
	assert.Equal(t, BasePoints[0][2], BasePoints[0][0].PlusX().PlusY().PlusY())

	for i := 1; i < 4; i++ {
		for j := 0; j < 3; j++ {
			assert.Equal(t, BasePoints[i][j], BasePoints[i-1][j].PlusX(), "Something wrong with base points %d %d", i, j)
			assert.Equal(t, BasePoints2[i][j], BasePoints2[i-1][j].PlusX(), "Something wrong with base points 2 %d %d", i, j)
		}
	}

	for i := 0; i < 4; i++ {
		BackToOrig := Origin
		BackToOrig2 := Origin
		for j := 0; j < 3; j++ {
			assert.Equal(t, int64(2), DS(&Origin, &BasePoints[i][j]), "Something wrong with size of base point %d %d", i, j)
			assert.Equal(t, int64(2), DS(&Origin, &BasePoints2[i][j]), "Something wrong with size of base point 2 %d %d", i, j)
			for c := 0; c < 3; c++ {
				abs := Abs(BasePoints[i][j][c])
				assert.True(t, int64(1) == abs || int64(0) == abs, "Something wrong with coordinate of base point %d %d %d = %d", i, j, c, BasePoints[i][j][c])
				abs = Abs(BasePoints2[i][j][c])
				assert.True(t, int64(1) == abs || int64(0) == abs, "Something wrong with coordinate of base point 2 %d %d %d = %d", i, j, c, BasePoints[i][j][c])
			}
			BackToOrig = BackToOrig.Add(BasePoints[i][j])
			BackToOrig2 = BackToOrig2.Add(BasePoints2[i][j])
		}
		assert.Equal(t, Origin, BackToOrig, "Something wrong with sum of base points %d", i)
		assert.Equal(t, Origin, BackToOrig2, "Something wrong with sum of base points 2 %d", i)
	}
}

func TestBasePointsRotation(t *testing.T) {
	// For each axe (first index), the three base point evolves with plusX, plusY and plusZ
	currentBasePoints := [3][3]Point{}
	currentBasePoints2 := [3][3]Point{}
	for axe := 0; axe < 3; axe++ {
		currentBasePoints[axe][0] = BasePoints[0][0]
		currentBasePoints[axe][1] = BasePoints[0][1]
		currentBasePoints[axe][2] = BasePoints[0][2]

		currentBasePoints2[axe][0] = BasePoints2[0][0]
		currentBasePoints2[axe][1] = BasePoints2[0][1]
		currentBasePoints2[axe][2] = BasePoints2[0][2]
	}

	for k := -4; k < 6; k++ {
		mapColumn := int(PosMod4(int64(k)))
		fmt.Println("Checking map column", mapColumn, "from", k)
		assertSameTrio(t, BasePoints[NextMapping[0][mapColumn]], currentBasePoints[0])
		assertSameTrio(t, BasePoints[NextMapping[1][mapColumn]], currentBasePoints[1])
		assertSameTrio(t, BasePoints[NextMapping[2][mapColumn]], currentBasePoints[2])

		assertSameTrio(t, BasePoints2[NextMapping2[0][mapColumn]], currentBasePoints2[0])
		assertSameTrio(t, BasePoints2[NextMapping2[1][mapColumn]], currentBasePoints2[1])
		assertSameTrio(t, BasePoints2[NextMapping2[2][mapColumn]], currentBasePoints2[2])

		for i := 0; i < 3; i++ {
			currentBasePoints[0][i] = currentBasePoints[0][i].PlusX()
			currentBasePoints[1][i] = currentBasePoints[1][i].PlusY()
			currentBasePoints[2][i] = currentBasePoints[2][i].PlusZ()

			currentBasePoints2[0][i] = currentBasePoints2[0][i].PlusX()
			currentBasePoints2[1][i] = currentBasePoints2[1][i].PlusY()
			currentBasePoints2[2][i] = currentBasePoints2[2][i].PlusZ()
		}
	}

	nbRun := 100
	rdMax := int64(100000000000)
	for i := 0; i < nbRun; i++ {
		randomPoint := Point{randomInt64(rdMax) * 3, randomInt64(rdMax) * 3, randomInt64(rdMax) * 3}
		assert.True(t, randomPoint.IsMainPoint(), "point %v should be main", randomPoint)
		mod4Point := randomPoint.GetMod4Point()
		mod4Val, ok := AllMod4Possible[mod4Point]
		assert.True(t, ok, "Mod4 does not exists for %v mod4 point %v", randomPoint, mod4Point)
		assert.Equal(t, randomPoint.CalculateMod4Value(), mod4Val, "Wrong Mod4 value for %v mod4 point %v", randomPoint, mod4Point)
	}
}

func TestConnectionDetails(t *testing.T) {
	DEBUG = true
	for k, v := range AllConnectionsPossible {
		assert.Equal(t, k, v.vector)
		currentNumber := v.connNumber
		sameNumber := 0
		for _, nv := range AllConnectionsPossible {
			if nv.connNumber == currentNumber {
				sameNumber++
				if nv.vector != v.vector {
					assert.Equal(t, nv.vector.Neg(), v.vector, "Should have neg vector")
					assert.Equal(t, !nv.connNeg, v.connNeg, "Should have opposite connNeg flag")
				}
			}
		}
		assert.Equal(t, 2, sameNumber, "Should have 2 with same conn number for %d", currentNumber)
	}

	// For any 2 close main points there should be a connection details if DS <= 3
	min := int64(-5) // -5
	max := int64(5)  // 5
	for x := min; x < max; x++ {
		for y := min; y < max; y++ {
			for z := min; z < max; z++ {
				mainPoint := Point{x, y, z}.Mul(3)
				basePoints := BasePoints[mainPoint.GetMod4Value()]
				for _, bp := range basePoints {

					assertValidConnDetails(t, mainPoint, mainPoint.Add(bp), fmt.Sprint("Main Point", mainPoint, "base vector", bp))

					nextMain := Origin
					switch bp.X() {
					case 0:
						// Nothing out
					case 1:
						nextMain = mainPoint.Add(XFirst)
					case -1:
						nextMain = mainPoint.Sub(XFirst)
					default:
						assert.Fail(t, "There should not be a base point with x value %d", bp.X())
					}
					if nextMain != Origin {
						// Find the base point on the other side ( the opposite 1 or -1 on X() )
						nextBasePoints := BasePoints[nextMain.GetMod4Value()]
						for _, nbp := range nextBasePoints {
							if nbp.X() == -bp.X() {
								assertValidConnDetails(t, mainPoint.Add(bp), nextMain.Add(nbp), fmt.Sprint("Main Point", mainPoint, "mod4", mainPoint.GetMod4Value(),
									"next point", nextMain, "mod4", mainPoint.GetMod4Value(),
									"main base vector", bp, "next base vector", nbp))
							}
						}
					}

					nextMain = Origin
					switch bp.Y() {
					case 0:
						// Nothing out
					case 1:
						nextMain = mainPoint.Add(YFirst)
					case -1:
						nextMain = mainPoint.Sub(YFirst)
					default:
						assert.Fail(t, "There should not be a base point with y value %d", bp.X())
					}
					if nextMain != Origin {
						// Find the base point on the other side ( the opposite 1 or -1 on Y() )
						nextBasePoints := BasePoints[nextMain.GetMod4Value()]
						for _, nbp := range nextBasePoints {
							if nbp.Y() == -bp.Y() {
								assertValidConnDetails(t, mainPoint.Add(bp), nextMain.Add(nbp), fmt.Sprint("Main Point", mainPoint, "mod4", mainPoint.GetMod4Value(),
									"next point", nextMain, "mod4", mainPoint.GetMod4Value(),
									"main base vector", bp, "next base vector", nbp))
							}
						}
					}

					nextMain = Origin
					switch bp.Z() {
					case 0:
						// Nothing out
					case 1:
						nextMain = mainPoint.Add(ZFirst)
					case -1:
						nextMain = mainPoint.Sub(ZFirst)
					default:
						assert.Fail(t, "There should not be a base point with Z value %d", bp.X())
					}
					if nextMain != Origin {
						// Find the base point on the other side ( the opposite 1 or -1 on Z() )
						nextBasePoints := BasePoints[nextMain.GetMod4Value()]
						for _, nbp := range nextBasePoints {
							if nbp.Z() == -bp.Z() {
								assertValidConnDetails(t, mainPoint.Add(bp), nextMain.Add(nbp), fmt.Sprint("Main Point", mainPoint, "mod4", mainPoint.GetMod4Value(),
									"next point", nextMain, "mod4", mainPoint.GetMod4Value(),
									"main base vector", bp, "next base vector", nbp))
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
	assert.Equal(t, p2.Sub(p1), connDetails1.vector, msg)

	connDetails2 := GetConnectionDetails(p2, p1)
	assert.NotEqual(t, EmptyConnDetails, connDetails2, msg)
	assert.Equal(t, p1.Sub(p2), connDetails2.vector, msg)
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

func TestSpace(t *testing.T) {
	DEBUG = true

	assert.Equal(t, int64(9), SpaceObj.max)
	assert.Equal(t, 0, len(SpaceObj.nodes))
	assert.Equal(t, 0, len(SpaceObj.connections))
	assert.Equal(t, 0, len(SpaceObj.events))
	assert.Equal(t, 0, len(SpaceObj.Elements))

	SpaceObj.CreateStuff(3, 1)
	assert.Equal(t, int64(3), SpaceObj.max)
	// Big nodes = (center + center face + middle edge + corner) * (main + 3)
	nbNodes := (1 + 6 + 12 + 8) * 4

	/*******************  STEP 0 ******************/
	nbOutgrowthsStep0 := 4
	assertSpaceMax3(t, nbNodes)

	assert.Equal(t, TickTime(0), SpaceObj.currentTime)

	assertOutgrowth(t, 4)
	assertOutgrowthDistance(t, map[EventID]int{0: 1, 1: 1, 2: 1, 3: 1})
	assertOutgrowthColors(t, 0, map[uint8]int{0: int(nbNodes - 4), 1: 4, 10: 4})
	assertOutgrowthColors(t, 20, map[uint8]int{0: int(nbNodes - 4), 1: 4, 10: 4})

	/*******************  STEP 1 ******************/
	SpaceObj.ForwardTime()
	// Same elements just color changes
	assertSpaceMax3(t, nbNodes)

	assert.Equal(t, TickTime(1), SpaceObj.currentTime)
	newOutgrowthsStep1 := 4 * 3
	nbOutgrowthsStep1 := nbOutgrowthsStep0 + newOutgrowthsStep1

	assertOutgrowth(t, nbOutgrowthsStep1)
	assertOutgrowthDistance(t, map[EventID]int{0: 3, 1: 3, 2: 3, 3: 3})
	assertOutgrowthColors(t, 0, map[uint8]int{0: int(nbNodes - newOutgrowthsStep1 - 4), 1: newOutgrowthsStep1 + 4, 10: 4})
	assertOutgrowthColors(t, 1, map[uint8]int{0: int(nbNodes - nbOutgrowthsStep1), 1: nbOutgrowthsStep1, 10: 4})
	assertOutgrowthColors(t, 20, map[uint8]int{0: int(nbNodes - nbOutgrowthsStep1), 1: nbOutgrowthsStep1, 10: 4})

	/*******************  STEP 2 ******************/
	SpaceObj.ForwardTime()
	assertSpaceMax3(t, nbNodes)

	assert.Equal(t, TickTime(2), SpaceObj.currentTime)
	newOutgrowthsStep2 := (4 * 3) + 2
	nbOutgrowthsStep2 := nbOutgrowthsStep1 + newOutgrowthsStep2

	assertOutgrowth(t, nbOutgrowthsStep1+newOutgrowthsStep2)
	assertOutgrowthDistance(t, map[EventID]int{0: 3, 1: 3, 2: 3, 3: 5})
	assertOutgrowthColors(t, 0, map[uint8]int{0: int(nbNodes - newOutgrowthsStep2 - 4), 1: newOutgrowthsStep2 + 4, 10: 4})
	assertOutgrowthColors(t, 1, map[uint8]int{0: int(nbNodes - (newOutgrowthsStep1 + newOutgrowthsStep2) - 4), 1: newOutgrowthsStep1 + newOutgrowthsStep2 + 4, 10: 4})
	assertOutgrowthColors(t, 2, map[uint8]int{0: int(nbNodes - nbOutgrowthsStep2), 1: nbOutgrowthsStep2, 10: 4})
	assertOutgrowthColors(t, 20, map[uint8]int{0: int(nbNodes - nbOutgrowthsStep2), 1: nbOutgrowthsStep2, 10: 4})

	/*******************  STEP 3 ******************/
	SpaceObj.ForwardTime()
	assertSpaceMax3(t, nbNodes)

	assert.Equal(t, TickTime(3), SpaceObj.currentTime)
	newOutgrowthsStep3 := (4*3)*2 - 1
	nbOutgrowthsStep3 := nbOutgrowthsStep2 + newOutgrowthsStep3
	nb2colorsStep3 := 2

	assertOutgrowth(t, nbOutgrowthsStep3)
	assertOutgrowthDistance(t, map[EventID]int{0: 4, 1: 6, 2: 4, 3: 9})
	assertOutgrowthColors(t, 0, map[uint8]int{0: int(nbNodes - newOutgrowthsStep3 + nb2colorsStep3 - 4), 1: newOutgrowthsStep3 + 4 - 2*nb2colorsStep3, 2: nb2colorsStep3, 10: 4})
	assertOutgrowthColors(t, 3, map[uint8]int{0: int(nbNodes - nbOutgrowthsStep3 + nb2colorsStep3), 1: nbOutgrowthsStep3 - 2*nb2colorsStep3, 2: nb2colorsStep3, 10: 4})

	/*******************  STEP 4 ******************/
	SpaceObj.ForwardTime()
	assertSpaceMax3(t, nbNodes)

	assert.Equal(t, TickTime(4), SpaceObj.currentTime)
	newOutgrowthsStep4 := (4*3)*4 - 5
	nbOutgrowthsStep4 := nbOutgrowthsStep3 + newOutgrowthsStep4
	nb2colorsStep4 := 5 + nb2colorsStep3

	assertOutgrowth(t, nbOutgrowthsStep4)
	assertOutgrowthDistance(t, map[EventID]int{0: 7, 1: 12, 2: 7, 3: 17})
	assertOutgrowthColors(t, 0, map[uint8]int{0: int(nbNodes - newOutgrowthsStep4 + nb2colorsStep4 - 4), 1: newOutgrowthsStep4 - 2*nb2colorsStep4 + 4, 2: nb2colorsStep4, 10: 4})

	/*******************  STEP 5 ******************/
	SpaceObj.ForwardTime()
	assertSpaceMax3(t, nbNodes)

	assert.Equal(t, TickTime(5), SpaceObj.currentTime)
	newOutgrowthsStep5 := (4*3)*4 - 4
	nbOutgrowthsStep5 := nbOutgrowthsStep4 + newOutgrowthsStep5
	nb2colorsStep5 := 6

	assertOutgrowth(t, nbOutgrowthsStep5)
	assertOutgrowthDistance(t, map[EventID]int{0: 6, 1: 13, 2: 8, 3: 17})
	assertOutgrowthColors(t, 0, map[uint8]int{0: int(nbNodes - newOutgrowthsStep5 + nb2colorsStep5 + 2*2 - 4), 1: newOutgrowthsStep5 - 2*nb2colorsStep5 - 3*2 + 4, 2: nb2colorsStep5, 3: 2, 10: 4})
}

func assertSpaceMax3(t *testing.T, nbNodes int) {
	assert.Equal(t, nbNodes, len(SpaceObj.nodesMap))
	assert.Equal(t, nbNodes, len(SpaceObj.nodes))
	assert.Equal(t, nbNodes+35, len(SpaceObj.connections))
	assert.Equal(t, 4, len(SpaceObj.events))
	assert.Equal(t, 2*nbNodes+35+6, len(SpaceObj.Elements))
}

func assertOutgrowth(t *testing.T, expect int) {
	nbOutgrowth := 0
	for _, evt := range SpaceObj.events {
		nbOutgrowth += len(evt.outgrowths)
	}
	assert.Equal(t, expect, nbOutgrowth)
	nbOutgrowth = 0
	for _, node := range SpaceObj.nodes {
		nbOutgrowth += len(node.outgrowths)
	}
	assert.Equal(t, expect, nbOutgrowth)
}

func assertOutgrowthDistance(t *testing.T, topOnes map[EventID]int) {
	for _, evt := range SpaceObj.events {
		nbTopOnes := 0
		for _, eo := range evt.outgrowths {
			if eo.distance == Distance(SpaceObj.currentTime-evt.created) {
				assert.Equal(t, eo.state, EventOutgrowthLatest, "Event outgrowth state test failed for evtID=%d node=%v . Should be latest", evt.id, *(eo.node))
				nbTopOnes++
			} else {
				assert.Equal(t, eo.state, EventOutgrowthOld, "Event outgrowth state test failed for evtID=%d node=%v . Should be old", evt.id, *(eo.node))
			}
		}
		assert.Equal(t, topOnes[evt.id], nbTopOnes, "NB top ones expected failed for evtID=%d", evt.id)
	}
}

func assertOutgrowthColors(t *testing.T, threshold Distance, multiOutgrowths map[uint8]int) {
	// map of how many nodes have 0, 1, 2, 3, 4 event outgrowth, the key 10 is for the amount of root
	count := make(map[uint8]int)
	for _, node := range SpaceObj.nodes {
		if node.IsRoot() {
			count[10]++
		}
		count[node.HowManyColors(threshold)]++
	}
	for k, v := range count {
		assert.Equal(t, multiOutgrowths[k], v, "color outgrowth count failed for k=%d and th=%d", k, threshold)
	}
}
