package m3point

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAllTrioBuilder(t *testing.T) {
	Log.Level = m3util.DEBUG

	// array of vec DS are in the possible list only: [2,2,2] [1,2,3], [2,3,3], [2,5,5]
	PossibleDSArray := [4][3]int64{{2, 2, 2}, {1, 2, 3}, {2, 3, 3}, {2, 5, 5}}

	assert.Equal(t, 92, len(AllTrioDetails))
	indexInPossDS := make([]int, len(AllTrioDetails))
	for i, tr := range AllTrioDetails {
		// All vec should have conn details
		cds := tr.Conns
		// Conn ID increase always
		assert.True(t, cds[0].GetPosIntId() <= cds[1].GetPosIntId(), "Mess in %v for trio %d = %v", cds, i, tr)
		assert.True(t, cds[1].GetPosIntId() <= cds[2].GetPosIntId(), "Mess in %v for trio %d = %v", cds, i, tr)

		dsArray := [3]int64{cds[0].ConnDS, cds[1].ConnDS, cds[2].ConnDS}
		found := false
		for k, posDsArray := range PossibleDSArray {
			if posDsArray == dsArray {
				found = true
				indexInPossDS[i] = k
			}
		}
		assert.True(t, found, "DS array %v not correct for trio %d = %v", dsArray, i, tr)
		assert.Equal(t, indexInPossDS[i], tr.GetDSIndex(), "DS array %v not correct for trio %d = %v", dsArray, i, tr)
	}

	// Check that All trio is ordered correctly
	countPerIndex := [4]int{}
	countPerIndexPerPosId := [4][10]int{}
	for i, tr := range AllTrioDetails {
		if i > 0 {
			assert.True(t, indexInPossDS[i-1] <= indexInPossDS[i], "Wrong order for trios %d = %v and %d = %v", i-1, AllTrioDetails[i-1], i, tr)
		}

		dsIndex := tr.GetDSIndex()
		countPerIndex[dsIndex]++
		countPerIndexPerPosId[dsIndex][tr.Conns[0].GetPosIntId()]++
	}
	assert.Equal(t, 8, countPerIndex[0])
	assert.Equal(t, 3*8*2, countPerIndex[1])
	assert.Equal(t, 3*4*2, countPerIndex[2])
	assert.Equal(t, 3*2*2, countPerIndex[3])
	for i, v := range countPerIndexPerPosId[0] {
		if i == 4 || i == 5 {
			assert.Equal(t, 4, v, "Index 0 wrong for %d", i)
		} else {
			assert.Equal(t, 0, v, "Index 0 wrong for %d", i)
		}
	}
	for i, v := range countPerIndexPerPosId[1] {
		if i == 1 || i == 2 || i == 3 {
			assert.Equal(t, 16, v, "Index 1 wrong for %d", i)
		} else {
			assert.Equal(t, 0, v, "Index 1 wrong for %d", i)
		}
	}
	for i, v := range countPerIndexPerPosId[2] {
		if i >= 4 && i <= 9 {
			assert.Equal(t, 4, v, "Index 2 wrong for %d", i)
		} else {
			assert.Equal(t, 0, v, "Index 2 wrong for %d", i)
		}
	}
	for i, v := range countPerIndexPerPosId[3] {
		if i >= 4 && i <= 9 {
			assert.Equal(t, 2, v, "Index 3 wrong for %d", i)
		} else {
			assert.Equal(t, 0, v, "Index 3 wrong for %d", i)
		}
	}
}

func TestInitialTrioConnectingVectors(t *testing.T) {
	Log.Level = m3util.DEBUG
	assert.Equal(t, AllBaseTrio[0][0], Point{1, 1, 0})
	assert.Equal(t, AllBaseTrio[0][1], Point{-1, 0, -1})
	assert.Equal(t, AllBaseTrio[0][1], AllBaseTrio[0][0].PlusX().PlusY().PlusY())
	assert.Equal(t, AllBaseTrio[0][2], Point{0, -1, 1})
	assert.Equal(t, AllBaseTrio[0][2], AllBaseTrio[0][0].PlusY().PlusX().PlusX())

	assert.Equal(t, AllBaseTrio[4][0], Point{-1, -1, 0})
	assert.Equal(t, AllBaseTrio[4][1], Point{1, 0, 1})
	assert.Equal(t, AllBaseTrio[4][2], Point{0, 1, -1})
}

func TestAllTrio(t *testing.T) {
	Log.Level = m3util.DEBUG
	for i, tr := range AllBaseTrio {
		assert.Equal(t, int64(0), tr[0][2], "Failed on Trio %d", i)
		assert.Equal(t, int64(0), tr[1][1], "Failed on Trio %d", i)
		assert.Equal(t, int64(0), tr[2][0], "Failed on Trio %d", i)
		BackToOrig := Origin
		for j, vec := range tr {
			for c := 0; c < 3; c++ {
				abs := Abs64(vec[c])
				assert.True(t, int64(1) == abs || int64(0) == abs, "Something wrong with coordinate of connecting vector %d %d %d = %v", i, j, c, vec)
			}
			assert.Equal(t, int64(2), vec.DistanceSquared(), "Failed vec at %d %d", i, j)
			assert.True(t, vec.IsBaseConnectingVector(), "Failed vec at %d %d", i, j)
			BackToOrig = BackToOrig.Add(vec)
		}
		assert.Equal(t, Origin, BackToOrig, "Something wrong with sum of Trio %d %v", i, tr)
		for j, tB := range AllBaseTrio {
			assertIsGenericNonBaseConnectingVector(t, GetNonBaseConnections(tr, tB), i, j)
		}
	}
}

func TestAllFull5Trio(t *testing.T) {
	Log.Level = m3util.DEBUG
	idxMap := createAll8IndexMap()
	// All trio with prime (neg of all vec) will have a full 5 connection length
	for i := 0; i < 4; i++ {
		nextTrio := [2]int{i, i + 4}
		assertValidNextTrio(t, nextTrio, i)

		// All conns are only 5
		assertIsFull5NonBaseConnectingVector(t, GetNonBaseConnections(AllBaseTrio[nextTrio[0]], AllBaseTrio[nextTrio[1]]), i, -1)
		idxMap[nextTrio[0]]++
		idxMap[nextTrio[1]]++
	}
	assertAllIndexUsed(t, idxMap, 1, "full 5 trios")
}

func TestAllValidTrio(t *testing.T) {
	idxMap := createAll8IndexMap()
	for i, nextTrio := range ValidNextTrio {
		assertValidNextTrio(t, nextTrio, i)

		// All conns are 3 or 1, no more 5
		assertIsThreeOr1NonBaseConnectingVector(t, GetNonBaseConnections(AllBaseTrio[nextTrio[0]], AllBaseTrio[nextTrio[1]]), i, -1)
		idxMap[nextTrio[0]]++
		idxMap[nextTrio[1]]++
	}
	assertAllIndexUsed(t, idxMap, 3, "valid trios")
}

func TestAllMod4Permutations(t *testing.T) {
	initMod4Permutations()
	idxMap := createAll8IndexMap()
	for i, permutMap := range AllMod4Permutations {
		for j := 0; j < 4; j++ {
			startIdx := permutMap[j]
			endIdx := permutMap[(j+1)%4]
			assertExistsInValidNextTrio(t, startIdx, endIdx, fmt.Sprint("in mod4 permutation[", i, "]=", permutMap, "idx", j))
			idxMap[permutMap[j]]++
		}
	}
	assertAllIndexUsed(t, idxMap, 6, "all mod4 permutations")
}

func TestAllMod8Permutations(t *testing.T) {
	for i, permutMap := range AllMod8Permutations {
		idxMap := createAll8IndexMap()
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

func assertIsGenericNonBaseConnectingVector(t *testing.T, conns [6]Point, i, j int) {
	for _, conn := range conns {
		assert.True(t, conn.IsConnectionVector(), "Found wrong connection %v at %d %d", conn, i, j)
		assert.False(t, conn.IsBaseConnectingVector(), "Found wrong connection %v at %d %d", conn, i, j)
		ds := conn.DistanceSquared()
		assert.True(t, ds == 1 || ds == 3 || ds == 5, "Found wrong connection %v at %d %d", conn, i, j)
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
