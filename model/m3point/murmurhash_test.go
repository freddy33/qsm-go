package m3point

import (
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
	"time"
)

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

	mHashes map[uint32]int
	hashes map[int]*[]*Point
	mHistogram []int
	histogram []int
}

func TestHashCodeConflicts(t *testing.T) {
	hConf := createHashConf(CInt(100), 2.0, 2.0, 10)
	runHashCodeFromConf(t, hConf)
	hConf = createHashConf(CInt(200), 7.0, 1.7, 10)
	runHashCodeFromConf(t, hConf)
	hConf = createHashConf(CInt(500), 7000.0, 2.5, 10)
	runHashCodeFromConf(t, hConf)
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
	hEnv.mHistogram = make([]int, hConf.maxFails)
	hEnv.histogram = make([]int, hConf.maxFails)
	hEnv.mHashes = make(map[uint32]int, hConf.nbRun)
	hEnv.hashes = make(map[int]*[]*Point, hConf.nbRun)
	return hEnv
}

func (hConf *HashTestConf) dumpInfo() {
	Log.Infof("Conf rdMax=%d | hashSizeRatio=%f | runRatio=%f | maxElements=%f | hashSize = %d | nbRun=%d",
		hConf.rdMax, hConf.hashSizeRatio, hConf.runRatio, hConf.maxElements, hConf.hashSize, hConf.nbRun)
}

func (hEnv *HashTestEnv) dumpInfo() {
	hEnv.conf.dumpInfo()
	Log.Infof("\t%d entries with %d foundSame, %d conflicts and %f conflict ratio and\nMurmur Hash Histo: %v\nHash Index Histo: %v",
		hEnv.mapSize, hEnv.foundSames, hEnv.conflicts, float64(100*hEnv.conflicts)/float64(hEnv.conf.nbRun), hEnv.mHistogram, hEnv.histogram)
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
		mHash := randomPoint.MurmurHash()
		hEnv.mHashes[mHash]++
		hash := murmurHashToInt(mHash, hEnv.conf.hashSize)
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
	for _, f := range hEnv.mHashes {
		hEnv.mHistogram[f-1]++
	}
	for _, f := range hEnv.hashes {
		hEnv.histogram[len(*f)-1]++
	}
	assert.Equal(t, 0, hEnv.noMoreSpace)
}

