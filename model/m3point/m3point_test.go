package m3point

import (
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
	"time"
)

func TestDS(t *testing.T) {
	Log.SetDebug()
	assert.Equal(t, DInt(0), DS(Point{1, 2, 3}, Point{1, 2, 3}))
	assert.Equal(t, DInt(1), DS(Point{1, 2, 3}, Point{0, 2, 3}))
	assert.Equal(t, DInt(1), DS(Point{1, 2, 3}, Point{2, 2, 3}))
	assert.Equal(t, DInt(1), DS(Point{1, 2, 3}, Point{1, 3, 3}))
	assert.Equal(t, DInt(1), DS(Point{1, 2, 3}, Point{1, 1, 3}))
	assert.Equal(t, DInt(1), DS(Point{1, 2, 3}, Point{1, 2, 4}))
	assert.Equal(t, DInt(1), DS(Point{1, 2, 3}, Point{1, 2, 2}))

	assert.Equal(t, DInt(3), Point{1, 1, 1}.DistanceSquared())
	assert.Equal(t, DInt(3), Point{-1, 1, 1}.DistanceSquared())
	assert.Equal(t, DInt(3), Point{1, -1, 1}.DistanceSquared())
	assert.Equal(t, DInt(3), Point{1, 1, -1}.DistanceSquared())
	assert.Equal(t, DInt(3), Point{-1, -1, -1}.DistanceSquared())

	assert.Equal(t, DInt(14), Point{1, 2, 3}.DistanceSquared())

	assert.Equal(t, DInt(0), DS(Point{-3, -2, -1}, Point{-3, -2, -1}))
	assert.Equal(t, DInt(3), DS(Point{-3, -2, -1}, Point{-2, -1, 0}))
}

func TestNbPosCoord(t *testing.T) {
	Log.SetDebug()
	assert.Equal(t, DInt(0), Origin.SumOfPositiveCoord())
	assert.Equal(t, DInt(0), Point{-1, 0, 0}.SumOfPositiveCoord())
	assert.Equal(t, DInt(0), Point{0, -1, 0}.SumOfPositiveCoord())
	assert.Equal(t, DInt(0), Point{0, 0, -1}.SumOfPositiveCoord())
	assert.Equal(t, DInt(0), Point{-34, -45, -14}.SumOfPositiveCoord())
	assert.Equal(t, DInt(34), Point{34, -45, -14}.SumOfPositiveCoord())
	assert.Equal(t, DInt(45), Point{-34, 45, -14}.SumOfPositiveCoord())
	assert.Equal(t, DInt(14), Point{-34, -45, 14}.SumOfPositiveCoord())
	assert.Equal(t, DInt(6), Point{1, 2, 3}.SumOfPositiveCoord())
}

func TestPointEqualAndDS(t *testing.T) {
	Log.SetDebug()

	Orig := Point{0, 0, 0}
	OneTwoThree := Point{1, 2, 3}
	P := Point{17, 11, 13}

	// Test equal
	assert.Equal(t, Orig, Origin)
	assert.Equal(t, Point{3, 0, 0}, XFirst)
	assert.Equal(t, Point{0, 3, 0}, YFirst)
	assert.Equal(t, Point{0, 0, 3}, ZFirst)
	assert.Equal(t, Point{1, 2, 3}, OneTwoThree)
	assert.Equal(t, Point{17, 11, 13}, P)

	// Test DS
	assert.Equal(t, DInt(3), DS(OneTwoThree, Point{0, 1, 2}))
	// Make sure OneTwoThree did not change
	assert.Equal(t, Point{1, 2, 3}, OneTwoThree)

	assert.Equal(t, DInt(4), DS(OneTwoThree, Point{-1, 2, 3}))
	assert.Equal(t, DInt(16), DS(OneTwoThree, Point{1, -2, 3}))
	assert.Equal(t, DInt(36), DS(OneTwoThree, Point{1, 2, -3}))
}

func TestIsMainPoint(t *testing.T) {
	assert.True(t, Origin.IsMainPoint())
	assert.True(t, XFirst.IsMainPoint())
	assert.True(t, XFirst.Neg().IsMainPoint())
	assert.True(t, YFirst.IsMainPoint())
	assert.True(t, YFirst.Neg().IsMainPoint())
	assert.True(t, ZFirst.IsMainPoint())
	assert.True(t, ZFirst.Neg().IsMainPoint())
	assert.True(t, Point{6, 3, 0}.IsMainPoint())
	assert.True(t, Point{-6, -3, 0}.IsMainPoint())
	assert.True(t, Point{-9, 12, 3}.IsMainPoint())
	assert.True(t, Point{21, -12, -3}.IsMainPoint())

	assert.False(t, Point{1, 0, 0}.IsMainPoint())
	assert.False(t, Point{0, 1, 0}.IsMainPoint())
	assert.False(t, Point{0, 0, 1}.IsMainPoint())
	assert.False(t, Point{1, 0, 0}.Neg().IsMainPoint())
	assert.False(t, Point{0, 1, 0}.Neg().IsMainPoint())
	assert.False(t, Point{0, 0, 1}.Neg().IsMainPoint())
	assert.False(t, Point{1, 0, 0}.Mul(2).IsMainPoint())
	assert.False(t, Point{0, 1, 0}.Mul(2).IsMainPoint())
	assert.False(t, Point{0, 0, 1}.Mul(2).IsMainPoint())
	assert.False(t, Point{1, 0, 0}.Mul(-2).IsMainPoint())
	assert.False(t, Point{0, 1, 0}.Mul(-2).IsMainPoint())
	assert.False(t, Point{0, 0, 1}.Mul(-2).IsMainPoint())
}

func TestIsConnectionVector(t *testing.T) {
	// Test IsBaseConnectingVector
	assert.True(t, Point{1, 1, 0}.IsBaseConnectingVector())
	assert.True(t, Point{0, 1, 1}.IsBaseConnectingVector())
	assert.True(t, Point{1, 0, 1}.IsBaseConnectingVector())
	assert.True(t, Point{1, -1, 0}.IsBaseConnectingVector())
	assert.True(t, Point{0, 1, -1}.IsBaseConnectingVector())
	assert.True(t, Point{-1, 0, 1}.IsBaseConnectingVector())

	assert.False(t, Origin.IsBaseConnectingVector())
	assert.False(t, Point{1, 0, 0}.IsBaseConnectingVector())
	assert.False(t, Point{-1, 0, 0}.IsBaseConnectingVector())
	assert.False(t, Point{1, 1, 1}.IsBaseConnectingVector())
	assert.False(t, Point{1, -1, 1}.IsBaseConnectingVector())

	assert.False(t, Point{0, 0, 2}.IsBaseConnectingVector())
	assert.False(t, Point{0, -2, 0}.IsBaseConnectingVector())
	assert.False(t, Point{2, 0, 0}.IsBaseConnectingVector())

	// Test IsConnectionVector
	assert.True(t, Point{1, 1, 0}.IsConnectionVector())
	assert.True(t, Point{0, 1, 1}.IsConnectionVector())
	assert.True(t, Point{1, 0, 1}.IsConnectionVector())
	assert.True(t, Point{1, -1, 0}.IsConnectionVector())
	assert.True(t, Point{0, 1, -1}.IsConnectionVector())
	assert.True(t, Point{-1, 0, 1}.IsConnectionVector())

	assert.True(t, Point{1, 0, 0}.IsConnectionVector())
	assert.True(t, Point{-1, 0, 0}.IsConnectionVector())
	assert.True(t, Point{1, 1, 1}.IsConnectionVector())
	assert.True(t, Point{1, -1, 1}.IsConnectionVector())
	assert.True(t, Point{1, -2, 0}.IsConnectionVector())
	assert.True(t, Point{0, -2, 1}.IsConnectionVector())

	assert.False(t, Origin.IsConnectionVector())
	assert.False(t, Point{2, 0, 0}.IsConnectionVector())
	assert.False(t, Point{0, 2, 0}.IsConnectionVector())
	assert.False(t, Point{0, 0, 2}.IsConnectionVector())
	assert.False(t, Point{-2, 0, 0}.IsConnectionVector())
	assert.False(t, Point{0, -2, 0}.IsConnectionVector())
	assert.False(t, Point{0, 0, -2}.IsConnectionVector())
	assert.False(t, Point{2, 1, 1}.IsConnectionVector())
	assert.False(t, Point{2, 2, 1}.IsConnectionVector())

	assert.False(t, Point{3, 0, 1}.IsConnectionVector())
	assert.False(t, Point{2, 3, 0}.IsConnectionVector())
	assert.False(t, Point{1, 0, 3}.IsConnectionVector())
}

func TestGetNearMainPoint(t *testing.T) {
	assert.Equal(t, Origin, Origin.GetNearMainPoint())
	assert.Equal(t, XFirst, XFirst.GetNearMainPoint())
	assert.Equal(t, XFirst.Neg(), XFirst.Neg().GetNearMainPoint())
	assert.Equal(t, YFirst, YFirst.GetNearMainPoint())
	assert.Equal(t, YFirst.Neg(), YFirst.Neg().GetNearMainPoint())
	assert.Equal(t, ZFirst, ZFirst.GetNearMainPoint())
	assert.Equal(t, ZFirst.Neg(), ZFirst.Neg().GetNearMainPoint())

	assert.Equal(t, Origin, Point{1, 0, 0}.GetNearMainPoint())
	assert.Equal(t, Origin, Point{-1, 0, 0}.GetNearMainPoint())
	assert.Equal(t, Origin, Point{0, 1, 0}.GetNearMainPoint())
	assert.Equal(t, Origin, Point{0, -1, 0}.GetNearMainPoint())
	assert.Equal(t, Origin, Point{0, 0, 1}.GetNearMainPoint())
	assert.Equal(t, Origin, Point{0, 0, -1}.GetNearMainPoint())

	assert.Equal(t, Origin, Point{1, 1, 0}.GetNearMainPoint())
	assert.Equal(t, Origin, Point{-1, 1, 0}.GetNearMainPoint())
	assert.Equal(t, Origin, Point{-1, 1, 0}.GetNearMainPoint())
	assert.Equal(t, Origin, Point{-1, -1, 0}.GetNearMainPoint())
	assert.Equal(t, Origin, Point{1, 1, 1}.GetNearMainPoint())
	assert.Equal(t, Origin, Point{-1, -1, -1}.GetNearMainPoint())

	assert.Equal(t, XFirst, Point{2, 0, 0}.GetNearMainPoint())
	assert.Equal(t, XFirst, Point{2, 1, 0}.GetNearMainPoint())
	assert.Equal(t, XFirst, Point{2, 1, 1}.GetNearMainPoint())
	assert.Equal(t, XFirst, Point{2, 1, -1}.GetNearMainPoint())
	assert.Equal(t, XFirst, Point{2, -1, -1}.GetNearMainPoint())
}

func TestPointAddSubMulNeg(t *testing.T) {
	Log.SetDebug()

	Orig := Point{0, 0, 0}
	OneTwoThree := Point{1, 2, 3}
	P := Point{17, 11, 13}

	// Test Add
	assert.Equal(t, Point{3, 0, 0}, Orig.Add(XFirst))
	// Make sure orig did not change
	assert.Equal(t, Orig, Origin)
	assert.Equal(t, Point{0, 3, 0}, Orig.Add(YFirst))
	assert.Equal(t, Point{0, 0, 3}, Orig.Add(ZFirst))
	assert.Equal(t, Point{3, 3, 3}, XFirst.Add(YFirst).Add(ZFirst))
	assert.Equal(t, Point{-3, -3, -3}, XFirst.Add(YFirst).Add(ZFirst).Neg())
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
}

func TestPointRotations(t *testing.T) {
	Log.SetDebug()

	Orig := Point{0, 0, 0}
	OneTwoThree := Point{1, 2, 3}

	// Test RotPlusX, RotNegX, RotPlusY, RotNegY, RotPlusZ, RotNegZ
	assert.Equal(t, OneTwoThree.RotPlusX(), Point{1, -3, 2})
	assert.Equal(t, OneTwoThree.RotNegX(), Point{1, 3, -2})
	assert.Equal(t, OneTwoThree.RotPlusY(), Point{3, 2, -1})
	assert.Equal(t, OneTwoThree.RotNegY(), Point{-3, 2, 1})
	assert.Equal(t, OneTwoThree.RotPlusZ(), Point{-2, 1, 3})
	assert.Equal(t, OneTwoThree.RotNegZ(), Point{2, -1, 3})

	// Test bunch of equations using random points
	nbRun := 100
	rdMax := CInt(100000000)
	for i := 0; i < nbRun; i++ {
		randomPoint := CreateRandomPoint(rdMax)
		assert.Equal(t, Orig.Sub(randomPoint), randomPoint.Neg())
		assert.Equal(t, randomPoint.Sub(randomPoint.Add(OneTwoThree)), OneTwoThree.Neg())
		assert.Equal(t, randomPoint.Sub(randomPoint.Add(OneTwoThree)), OneTwoThree.Mul(-1))
		assert.Equal(t, randomPoint.Add(randomPoint.Neg()), Orig)
		assert.Equal(t, randomPoint.Add(randomPoint.Mul(-1)), Orig)

		assert.Equal(t, randomPoint.RotPlusX().RotNegX(), randomPoint)
		assert.Equal(t, randomPoint.RotNegX().RotPlusX(), randomPoint)
		assert.Equal(t, randomPoint.RotPlusY().RotNegY(), randomPoint)
		assert.Equal(t, randomPoint.RotNegY().RotPlusY(), randomPoint)
		assert.Equal(t, randomPoint.RotPlusZ().RotNegZ(), randomPoint)
		assert.Equal(t, randomPoint.RotNegZ().RotPlusZ(), randomPoint)

		assert.Equal(t, randomPoint.RotPlusX().RotPlusX().RotPlusX().RotPlusX(), randomPoint)
		assert.Equal(t, randomPoint.RotPlusY().RotPlusY().RotPlusY().RotPlusY(), randomPoint)
		assert.Equal(t, randomPoint.RotPlusZ().RotPlusZ().RotPlusZ().RotPlusZ(), randomPoint)
		assert.Equal(t, randomPoint.RotNegX().RotNegX().RotNegX().RotNegX(), randomPoint)
		assert.Equal(t, randomPoint.RotNegY().RotNegY().RotNegY().RotNegY(), randomPoint)
		assert.Equal(t, randomPoint.RotNegZ().RotNegZ().RotNegZ().RotNegZ(), randomPoint)

		assert.Equal(t, randomPoint.RotNegX().RotNegX(), randomPoint.RotPlusX().RotPlusX())
		assert.Equal(t, randomPoint.RotNegY().RotNegY(), randomPoint.RotPlusY().RotPlusY())
		assert.Equal(t, randomPoint.RotNegZ().RotNegZ(), randomPoint.RotPlusZ().RotPlusZ())
	}
}

type HashTestConf struct {
	rdMax                   CInt
	runRatio, hashSizeRatio float64
	maxFails                int

	maxElements float64
	hashSize, nbRun int
	dataSet                      []*Point
}

type HashTestEnv struct {
	conf *HashTestConf

	foundSames, conflicts, noMoreSpace int
	mapSize                            int

	hashes map[int]*[]*Point
	histogram []int
}

func createHashConf(rdMax CInt, runRatio, hashSizeRatio float64, maxFails int) *HashTestConf {
	hConf := new(HashTestConf)
	hConf.rdMax = rdMax
	hConf.runRatio = runRatio
	hConf.hashSizeRatio = hashSizeRatio

	hConf.maxElements = math.Pow(float64(rdMax), 3)
	hConf.nbRun = int(hConf.maxElements / runRatio)
	hConf.hashSize = int(float64(hConf.nbRun) * hashSizeRatio)

	hConf.maxFails = maxFails

	hConf.dataSet = make([]*Point, hConf.nbRun)
	for i := 0; i < hConf.nbRun; i++ {
		p := CreateRandomPoint(hConf.rdMax)
		hConf.dataSet[i] = &p
	}

	return hConf
}

func (hConf *HashTestConf) createHashEnv() *HashTestEnv {
	hEnv := new(HashTestEnv)
	hEnv.conf = hConf
	hEnv.histogram = make([]int, hConf.maxFails)
	hEnv.hashes = make(map[int]*[]*Point, hConf.nbRun)
	return hEnv
}

func (hConf *HashTestConf) dumpInfo() {
	Log.Infof("Conf rdMax=%d | hashSizeRatio=%f | runRatio=%f | maxElements=%f | hashSize = %d | nbRun=%d",
		hConf.rdMax, hConf.hashSizeRatio, hConf.runRatio, hConf.maxElements, hConf.hashSize, hConf.nbRun)
}

func (hEnv *HashTestEnv) dumpInfo() {
	hEnv.conf.dumpInfo()
	Log.Infof("\t%d entries with %d foundSame, %d conflicts and %f conflict ratio and %v",
		hEnv.mapSize, hEnv.foundSames, hEnv.conflicts, float64(100*hEnv.conflicts)/float64(hEnv.conf.nbRun), hEnv.histogram)
}

func TestHashCodeConflicts(t *testing.T) {
	hConf := createHashConf(CInt(100), 2.0, 2.0, 10)
	runHashCodeFromConf(t, hConf)
	hConf = createHashConf(CInt(200), 7.0, 1.7, 10)
	runHashCodeFromConf(t, hConf)
	hConf = createHashConf(CInt(500), 7000.0, 2.5, 10)
	runHashCodeFromConf(t, hConf)
}

func runHashCodeFromConf(t *testing.T, hConf *HashTestConf) {
	env := hConf.createHashEnv()
	start := time.Now()
	runHashCode(t, env)
	Log.Infof("Took %v", time.Now().Sub(start))
	env.dumpInfo()
}

func runHashCode(t *testing.T, hEnv *HashTestEnv) {
	for _, randomPoint := range hEnv.conf.dataSet {
		hash := randomPoint.Hash(hEnv.conf.hashSize)
		assert.True(t, hash >= 0 && hash < hEnv.conf.hashSize, "hash %d not correct for %d", hash, hEnv.conf.hashSize)
		f, ok := hEnv.hashes[hash]
		if ok {
			points := *f
			foundSame := false
			for _, op := range points {
				if *op == *randomPoint {
					foundSame = true
				}
			}
			if foundSame {
				hEnv.foundSames++
			} else {
				hEnv.conflicts++
				points = append(points, randomPoint)
				if len(points) > hEnv.conf.maxFails {
					assert.FailNow(t, "no space", "did not find space for %v in %v hash %d", randomPoint, *f, hash)
					hEnv.noMoreSpace++
				}
				hEnv.hashes[hash] = &points
			}
		} else {
			newF := make([]*Point, 1)
			newF[0] = randomPoint
			hEnv.hashes[hash] = &newF
		}
	}
	hEnv.mapSize = len(hEnv.hashes)
	for _, f := range hEnv.hashes {
		hEnv.histogram[len(*f)-1]++
	}
	assert.Equal(t, 0, hEnv.noMoreSpace)
}
