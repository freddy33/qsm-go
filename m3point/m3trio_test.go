package m3point

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

func TestPosMod2(t *testing.T) {
	Log.SetDebug()
	assert.Equal(t, uint64(1), PosMod2(5))
	assert.Equal(t, uint64(0), PosMod2(4))
	assert.Equal(t, uint64(1), PosMod2(3))
	assert.Equal(t, uint64(0), PosMod2(2))
	assert.Equal(t, uint64(1), PosMod2(1))
	assert.Equal(t, uint64(0), PosMod2(0))
}

func TestPosMod4(t *testing.T) {
	Log.SetDebug()
	assert.Equal(t, uint64(1), PosMod4(5))
	assert.Equal(t, uint64(0), PosMod4(4))
	assert.Equal(t, uint64(3), PosMod4(3))
	assert.Equal(t, uint64(2), PosMod4(2))
	assert.Equal(t, uint64(1), PosMod4(1))
	assert.Equal(t, uint64(0), PosMod4(0))
}

func TestPosMod8(t *testing.T) {
	Log.SetDebug()
	assert.Equal(t, uint64(1), PosMod8(9))
	assert.Equal(t, uint64(0), PosMod8(8))
	assert.Equal(t, uint64(7), PosMod8(7))
	assert.Equal(t, uint64(6), PosMod8(6))
	assert.Equal(t, uint64(5), PosMod8(5))
	assert.Equal(t, uint64(4), PosMod8(4))
	assert.Equal(t, uint64(3), PosMod8(3))
	assert.Equal(t, uint64(2), PosMod8(2))
	assert.Equal(t, uint64(1), PosMod8(1))
	assert.Equal(t, uint64(0), PosMod8(0))
}

func TestAllTrioLinks(t *testing.T) {
	Log.SetDebug()
	assert.Equal(t, 8*8*(1+8)/2, len(allTrioLinks), "%v", allTrioLinks)
	for a := TrioIndex(0); a < 8; a++ {
		for b := TrioIndex(0); b < 8; b++ {
			for c := TrioIndex(0); c < 8; c++ {
				count := 0
				for _, tl := range allTrioLinks {
					if a == tl.a && b == tl.b && c == tl.c {
						count++
					}
				}
				if b <= c {
					assert.Equal(t, 1, count, "sould have found one instance of %d %d %d", a, b, c)
				} else {
					assert.Equal(t, 0, count, "sould have found no instances of %d %d %d", a, b, c)
				}
			}
		}
	}
}

func TestAllTrioDetails(t *testing.T) {
	Log.SetDebug()
	Log.SetAssert(true)

	assert.Equal(t, 200, len(allTrioDetails))
	for i, td := range allTrioDetails {
		// All vec should have conn details
		cds := td.GetConnections()
		// Conn ID increase always
		assert.True(t, cds[0].GetPosId() <= cds[1].GetPosId(), "Mess in order for %v for trio %d = %s", cds, i, td.String())
		assert.True(t, cds[1].GetPosId() <= cds[2].GetPosId(), "Mess in order for %v for trio %d = %s", cds, i, td.String())
		// For base trio verify we have all good
		assert.Equal(t, td.IsBaseTrio(), i < 8, "trio %d = %s should be or not base", i, td.String())
		if i < 8 {
			// All connections are base connection
			for _, c := range cds {
				assert.True(t, c.IsBaseConnection(), "found non base connection %s in %d = %s", c.String(), i, td.String())
			}
			// all find connection with assertion are good
			assert.NotNil(t, td.getPlusXConn(), "trio %d = %s did not find conn", i, td.String())
			assert.NotNil(t, td.getMinusXConn(), "trio %d = %s did not find conn", i, td.String())
			assert.NotNil(t, td.getPlusYConn(), "trio %d = %s did not find conn", i, td.String())
			assert.NotNil(t, td.getMinusYConn(), "trio %d = %s did not find conn", i, td.String())
			assert.NotNil(t, td.getPlusZConn(), "trio %d = %s did not find conn", i, td.String())
			assert.NotNil(t, td.getMinusZConn(), "trio %d = %s did not find conn", i, td.String())
			// all find connection without assertion are good
			Log.SetAssert(false)
			assert.NotNil(t, td.getPlusXConn(), "trio %d = %s did not find conn", i, td.String())
			assert.NotNil(t, td.getMinusXConn(), "trio %d = %s did not find conn", i, td.String())
			assert.NotNil(t, td.getPlusYConn(), "trio %d = %s did not find conn", i, td.String())
			assert.NotNil(t, td.getMinusYConn(), "trio %d = %s did not find conn", i, td.String())
			assert.NotNil(t, td.getPlusZConn(), "trio %d = %s did not find conn", i, td.String())
			assert.NotNil(t, td.getMinusZConn(), "trio %d = %s did not find conn", i, td.String())
			Log.SetAssert(true)
		}
	}

	// Check that All trio is ordered correctly
	for i, tr := range allTrioDetails {
		if i > 0 {
			assert.True(t, allTrioDetails[i-1].GetDSIndex() <= tr.GetDSIndex(), "Wrong order for trios %d = %v and %d = %v", i-1, allTrioDetails[i-1], i, tr)
		}
	}
}

func TestTrioDetailsPerDSIndex(t *testing.T) {
	Log.SetDebug()

	// array of vec DS are in the possible list only: [2,2,2] [1,2,3], [2,3,3], [2,5,5]
	PossibleDSArray := [NbTrioDsIndex][3]int64{{2, 2, 2}, {1, 1, 2}, {1, 2, 3}, {1, 2, 5}, {2, 3, 3}, {2, 3, 5}, {2, 5, 5}}

	indexInPossDS := make([]int, len(allTrioDetails))
	for i, td := range allTrioDetails {
		cds := td.conns
		dsArray := [3]int64{cds[0].ConnDS, cds[1].ConnDS, cds[2].ConnDS}
		found := false
		for k, posDsArray := range PossibleDSArray {
			if posDsArray == dsArray {
				found = true
				indexInPossDS[i] = k
			}
		}
		assert.True(t, found, "DS array %v not correct for trio %d = %v", dsArray, i, td)
		assert.Equal(t, indexInPossDS[i], td.GetDSIndex(), "DS array %v not correct for trio %d = %v", dsArray, i, td)
	}

	// Check that All trio is ordered correctly
	countPerIndex := [NbTrioDsIndex]int{}
	countPerIndexPerFirstConnPosId := [NbTrioDsIndex][10]int{}
	for i, td := range allTrioDetails {
		if i > 0 {
			assert.True(t, indexInPossDS[i-1] <= indexInPossDS[i], "Wrong order for trios %d = %v and %d = %v", i-1, allTrioDetails[i-1], i, td)
		}
		dsIndex := td.GetDSIndex()
		countPerIndex[dsIndex]++
		countPerIndexPerFirstConnPosId[dsIndex][td.conns[0].GetPosId()]++
	}
	assert.Equal(t, 8, countPerIndex[0])
	assert.Equal(t, 3*2*2, countPerIndex[1])
	assert.Equal(t, 3*8*2, countPerIndex[2])
	assert.Equal(t, 3*4*2, countPerIndex[3])
	assert.Equal(t, 3*8*2, countPerIndex[4])
	assert.Equal(t, 3*8*2, countPerIndex[5])
	assert.Equal(t, 3*2*2, countPerIndex[6])
	for i, v := range countPerIndexPerFirstConnPosId[0] {
		if i == 4 || i == 5 {
			assert.Equal(t, 4, v, "Index 0 wrong for %d", i)
		} else {
			assert.Equal(t, 0, v, "Index 0 wrong for %d", i)
		}
	}
	for i, v := range countPerIndexPerFirstConnPosId[1] {
		if i == 1 {
			assert.Equal(t, 8, v, "Index 1 wrong for %d", i)
		} else if i == 2 {
			assert.Equal(t, 4, v, "Index 1 wrong for %d", i)
		} else {
			assert.Equal(t, 0, v, "Index 1 wrong for %d", i)
		}
	}
	for i, v := range countPerIndexPerFirstConnPosId[2] {
		if i == 1 || i == 2 || i == 3 {
			assert.Equal(t, 16, v, "Index 2 wrong for %d", i)
		} else {
			assert.Equal(t, 0, v, "Index 2 wrong for %d", i)
		}
	}
	for i, v := range countPerIndexPerFirstConnPosId[3] {
		if i == 1 || i == 2 || i == 3 {
			assert.Equal(t, 8, v, "Index 3 wrong for %d", i)
		} else {
			assert.Equal(t, 0, v, "Index 3 wrong for %d", i)
		}
	}
	for i, v := range countPerIndexPerFirstConnPosId[4] {
		if i >= 4 && i <= 9 {
			assert.Equal(t, 8, v, "Index 4 wrong for %d", i)
		} else {
			assert.Equal(t, 0, v, "Index 4 wrong for %d", i)
		}
	}
	for i, v := range countPerIndexPerFirstConnPosId[5] {
		if i >= 4 && i <= 9 {
			assert.Equal(t, 8, v, "Index 5 wrong for %d", i)
		} else {
			assert.Equal(t, 0, v, "Index 5 wrong for %d", i)
		}
	}
	for i, v := range countPerIndexPerFirstConnPosId[6] {
		if i >= 4 && i <= 9 {
			assert.Equal(t, 2, v, "Index 6 wrong for %d", i)
		} else {
			assert.Equal(t, 0, v, "Index 6 wrong for %d", i)
		}
	}
}

func TestTrioDetailsConnectionsMethods(t *testing.T) {
	td0 := GetTrioDetails(0)
	assert.True(t, td0.HasConnection(4))
	assert.False(t, td0.HasConnection(-4))
	assert.True(t, td0.HasConnection(-6))
	assert.False(t, td0.HasConnection(6))
	assert.True(t, td0.HasConnection(-9))
	assert.False(t, td0.HasConnection(9))
	failedOc := td0.OtherConnectionsFrom(4)
	assert.Equal(t, (*ConnectionDetails)(nil), failedOc[0])
	assert.Equal(t, (*ConnectionDetails)(nil), failedOc[1])

	oc := td0.OtherConnectionsFrom(-4)
	assert.Equal(t, *GetConnDetailsById(-6), *oc[0])
	assert.Equal(t, *GetConnDetailsById(-9), *oc[1])

	td92 := GetTrioDetails(92)
	assert.True(t, td92.HasConnection(4))
	assert.False(t, td92.HasConnection(-4))
	assert.True(t, td92.HasConnection(12))
	assert.True(t, td92.HasConnection(-12))

	oc = td92.OtherConnectionsFrom(-4)
	assert.Equal(t, *GetConnDetailsById(12), *oc[0])
	assert.Equal(t, *GetConnDetailsById(-12), *oc[1])
}

func TestTrioDetailsLinks(t *testing.T) {
	countPerTrioLinks := make(map[TrioLink]int)

	for _, td := range allTrioDetails {
		switch td.GetDSIndex() {
		case 0:
			assert.Equal(t, 92, len(td.Links), "Nb links wrong for %v", td.String())
		case 6:
			assert.Equal(t, 6, len(td.Links), "Nb links wrong for %v", td.String())
		default:
			assert.Equal(t, 8, len(td.Links), "Nb links wrong for %v", td.String())
		}
		for _, tl := range td.Links {
			countPerTrioLinks[*tl]++
		}
	}

	collPerCount := make(map[int]*TrioLinkList)
	for tl, c := range countPerTrioLinks {
		coll, ok := collPerCount[c]
		if !ok {
			coll = &TrioLinkList{}
			collPerCount[c] = coll
		}
		copyTl := makeTrioLink(tl.a, tl.b, tl.c)
		coll.addUnique(&copyTl)
	}
	for _, tll := range collPerCount {
		sort.Sort(tll)
	}
	//fmt.Println(collPerCount)
	assert.Equal(t, 8, collPerCount[4].Len(), "wrong number of 4 times in nextMainTd %v", *collPerCount[4])
	assert.Equal(t, 8*7, collPerCount[5].Len(), "wrong number of 5 times in nextMainTd %v", *collPerCount[5])
	assert.Equal(t, 8*10, collPerCount[8].Len(), "wrong number of 8 times in nextMainTd %v", *collPerCount[8])
	assert.Equal(t, 8*18, collPerCount[9].Len(), "wrong number of 9 times in nextMainTd %v", *collPerCount[9])

	// The size 4 are when a,b,c are equals
	for _, td := range *collPerCount[4] {
		assert.True(t, td.a == td.b && td.a == td.c, "not all equal on 4 %v", td.String())
	}

	// The size 5 are when b and c are equals
	for _, td := range *collPerCount[5] {
		assert.True(t, td.b == td.c, "not b = c equal on 5 %v", td.String())
	}

	// Size on 8 are either a is equal to b or c, or b or c is prime of a and the other on same side
	for _, td := range *collPerCount[8] {
		if td.a != td.b && td.a != td.c {
			if td.a < 4 {
				assert.True(t, isPrime(td.a, td.c), "not prime on 8 %v", td.String())
				assert.True(t, td.b < 4 && td.b != td.a, "wrong side on 8 %v", td.String())
			} else {
				assert.True(t, isPrime(td.a, td.b), "not prime on 8 %v", td.String())
				assert.True(t, td.c >= 4 && td.c != td.a, "wrong side on 8 %v", td.String())
			}
		} else {
			// a is equal to b or c but not all equal
			assert.False(t, td.a == td.b && td.a == td.c, "all equal on 8 %v", td.String())
		}
	}
}

func TestInitialTrioConnectingVectors(t *testing.T) {
	Log.SetDebug()
	assert.Equal(t, allBaseTrio[0][0], Point{1, 1, 0})
	assert.Equal(t, allBaseTrio[0][1], Point{-1, 0, -1})
	assert.Equal(t, allBaseTrio[0][1], allBaseTrio[0][0].RotPlusX().RotPlusY().RotPlusY())
	assert.Equal(t, allBaseTrio[0][2], Point{0, -1, 1})
	assert.Equal(t, allBaseTrio[0][2], allBaseTrio[0][0].RotPlusY().RotPlusX().RotPlusX())

	assert.Equal(t, allBaseTrio[4][0], Point{-1, -1, 0})
	assert.Equal(t, allBaseTrio[4][1], Point{1, 0, 1})
	assert.Equal(t, allBaseTrio[4][2], Point{0, 1, -1})
}

func TestAllBaseTrio(t *testing.T) {
	Log.SetDebug()
	for i, tr := range allBaseTrio {
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
		for j, tB := range allBaseTrio {
			assertIsGenericNonBaseConnectingVector(t, GetNonBaseConnections(tr, tB), i, j)
		}
	}
}

func TestAllFull5Trio(t *testing.T) {
	Log.SetDebug()
	idxMap := createAll8IndexMap()
	// All trio with prime (neg of all vec) will have a full 5 connection length
	for i := 0; i < 4; i++ {
		nextTrio := [2]TrioIndex{TrioIndex(i), TrioIndex(i + 4)}
		assertValidNextTrio(t, nextTrio, i)

		// All conns are only 5
		assertIsFull5NonBaseConnectingVector(t, GetNonBaseConnections(allBaseTrio[nextTrio[0]], allBaseTrio[nextTrio[1]]), i, -1)
		idxMap[nextTrio[0]]++
		idxMap[nextTrio[1]]++
	}
	assertAllIndexUsed(t, idxMap, 1, "full 5 trios")
}

func TestAllValidTrio(t *testing.T) {
	idxMap := createAll8IndexMap()
	for i, nextTrio := range validNextTrio {
		assertValidNextTrio(t, nextTrio, i)

		// All conns are 3 or 1, no more 5
		assertIsThreeOr1NonBaseConnectingVector(t, GetNonBaseConnections(allBaseTrio[nextTrio[0]], allBaseTrio[nextTrio[1]]), i, -1)
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

func assertExistsInValidNextTrio(t *testing.T, startIdx TrioIndex, endIdx TrioIndex, msg string) {
	assert.NotEqual(t, startIdx, endIdx, "start and end index cannot be equal for %s", msg)
	// Order the indexes
	trioToFind := [2]TrioIndex{NilTrioIndex, NilTrioIndex}
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

	assert.True(t, trioToFind[0] != NilTrioIndex && trioToFind[0] <= 3, "Something wrong with trioToFind first value for %s", msg)
	assert.True(t, trioToFind[1] != NilTrioIndex && trioToFind[1] >= 4 && trioToFind[1] <= 7, "Something wrong with trioToFind second value for %s", msg)

	foundNextTrio := false
	for _, nextTrio := range validNextTrio {
		if trioToFind == nextTrio {
			foundNextTrio = true
		}
	}
	assert.True(t, foundNextTrio, "Did not find trio %v in list of valid trio for %s", trioToFind, msg)
}

func assertValidNextTrio(t *testing.T, nextTrio [2]TrioIndex, i int) {
	assert.NotEqual(t, nextTrio[0], nextTrio[1], "Something wrong with nextTrio index %d %v", i, nextTrio)
	assert.True(t, nextTrio[0] != NilTrioIndex && nextTrio[0] <= 3, "Something wrong with nextTrio first value index %d %v", i, nextTrio)
	assert.True(t, nextTrio[1] != NilTrioIndex && nextTrio[1] >= 4 && nextTrio[1] <= 7, "Something wrong with nextTrio second value index %d %v", i, nextTrio)
}

func createAll8IndexMap() map[TrioIndex]int {
	res := make(map[TrioIndex]int)
	for i := TrioIndex(0); i < 8; i++ {
		res[i] = 0
	}
	return res
}

func assertAllIndexUsed(t *testing.T, idxMap map[TrioIndex]int, expectedTimes int, msg string) {
	assert.Equal(t, 8, len(idxMap))
	for i := TrioIndex(0); i < 8; i++ {
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
