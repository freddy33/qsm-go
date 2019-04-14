package m3point

import (
	"github.com/freddy33/qsm-go/m3util"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

func TestDS(t *testing.T) {
	Log.Level = m3util.DEBUG
	assert.Equal(t, int64(0), DS(Point{1,2,3}, Point{1,2,3}))
	assert.Equal(t, int64(1), DS(Point{1,2,3}, Point{0,2,3}))
	assert.Equal(t, int64(1), DS(Point{1,2,3}, Point{2,2,3}))
	assert.Equal(t, int64(1), DS(Point{1,2,3}, Point{1,3,3}))
	assert.Equal(t, int64(1), DS(Point{1,2,3}, Point{1,1,3}))
	assert.Equal(t, int64(1), DS(Point{1,2,3}, Point{1,2,4}))
	assert.Equal(t, int64(1), DS(Point{1,2,3}, Point{1,2,2}))

	assert.Equal(t, int64(3), Point{1,1,1}.DistanceSquared())
	assert.Equal(t, int64(3), Point{-1,1,1}.DistanceSquared())
	assert.Equal(t, int64(3), Point{1,-1,1}.DistanceSquared())
	assert.Equal(t, int64(3), Point{1,1,-1}.DistanceSquared())
	assert.Equal(t, int64(3), Point{-1,-1,-1}.DistanceSquared())

	assert.Equal(t, int64(14), Point{1,2,3}.DistanceSquared())

	assert.Equal(t, int64(0), DS(Point{-3,-2,-1}, Point{-3,-2,-1}))
	assert.Equal(t, int64(3), DS(Point{-3,-2,-1}, Point{-2,-1,0}))
}

func TestNbPosCoord(t *testing.T) {
	Log.Level = m3util.DEBUG
	assert.Equal(t, int64(0), Origin.SumOfPositiveCoord())
	assert.Equal(t, int64(0), Point{-1,0,0}.SumOfPositiveCoord())
	assert.Equal(t, int64(0), Point{0,-1,0}.SumOfPositiveCoord())
	assert.Equal(t, int64(0), Point{0,0,-1}.SumOfPositiveCoord())
	assert.Equal(t, int64(0), Point{-34,-45,-14}.SumOfPositiveCoord())
	assert.Equal(t, int64(34), Point{34,-45,-14}.SumOfPositiveCoord())
	assert.Equal(t, int64(45), Point{-34,45,-14}.SumOfPositiveCoord())
	assert.Equal(t, int64(14), Point{-34,-45,14}.SumOfPositiveCoord())
	assert.Equal(t, int64(6), Point{1,2,3}.SumOfPositiveCoord())
}

func TestPoint(t *testing.T) {
	Log.Level = m3util.DEBUG

	Orig := Point{0, 0, 0}
	OneTwoThree := Point{1, 2, 3}
	P := Point{17, 11, 13}

	// Test equal
	assert.Equal(t, Orig, Origin)
	assert.Equal(t, Point{1, 2, 3}, OneTwoThree)
	assert.Equal(t, Point{17, 11, 13}, P)

	// Test DS
	assert.Equal(t, int64(3), DS(OneTwoThree, Point{0, 1, 2}))
	// Make sure OneTwoThree did not change
	assert.Equal(t, Point{1, 2, 3}, OneTwoThree)

	assert.Equal(t, int64(4), DS(OneTwoThree, Point{-1, 2, 3}))
	assert.Equal(t, int64(16), DS(OneTwoThree, Point{1, -2, 3}))
	assert.Equal(t, int64(36), DS(OneTwoThree, Point{1, 2, -3}))

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

func randomInt64(max int64) int64 {
	r := rand.Int63n(max)
	if rand.Float32() < 0.5 {
		return -r
	}
	return r
}
