package m3space

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/gonum/stat"
	"github.com/stretchr/testify/assert"
	"log"
	"math"
	"sort"
	"testing"
)

var LogTest = m3util.NewDataLogger("m3test", m3util.DEBUG)

func GetPyramidSize(points [4]Point) int64 {
	// Sum all the edges
	totalSize := int64(0)
	totalSize += MakeVector(points[0], points[1]).DistanceSquared()
	totalSize += MakeVector(points[0], points[2]).DistanceSquared()
	totalSize += MakeVector(points[0], points[3]).DistanceSquared()
	totalSize += MakeVector(points[1], points[2]).DistanceSquared()
	totalSize += MakeVector(points[1], points[3]).DistanceSquared()
	totalSize += MakeVector(points[2], points[3]).DistanceSquared()
	return totalSize
}

type Pyramid [4]Point

func (pyramid Pyramid) ordered() Pyramid {
	slice := make([]Point, 4)
	for i, p := range pyramid {
		slice[i] = p
	}
	sort.Slice(slice, func(i, j int) bool {
		iP := slice[i]
		jP := slice[j]
		if iP.X() < jP.X() {
			return true
		}
		if iP.X() > jP.X() {
			return false
		}
		if iP.Y() < jP.Y() {
			return true
		}
		if iP.Y() > jP.Y() {
			return false
		}
		if iP.Z() < jP.Z() {
			return true
		}
		if iP.Z() > jP.Z() {
			return false
		}
		return false
	})
	return Pyramid{slice[0], slice[1], slice[2], slice[3]}
}

func TestStatPack(t *testing.T) {
	Log.Level = m3util.WARN
	LogStat.Level = m3util.INFO
	fmt.Println(stat.StdDev([]float64{1.3, 1.5, 1.7, 1.1}, nil))
	space := MakeSpace(3 * 30)
	space.MaxConnections = 3
	space.blockOnSameEvent = 3
	space.SetEventOutgrowthThreshold(Distance(0))
	space.CreatePyramid(50)
	/*
		i:=0
		for _, evt := range space.events {
			evt.growthContext.permutationType = 1
			evt.growthContext.permutationIndex = i*2
			evt.growthContext.permutationOffset = 0
			evt.growthContext.permutationNegFlow = false
			i++
		}
	*/

	pyramidPoints := Pyramid{}
	idx := 0
	for p := range space.activeNodesMap {
		pyramidPoints[idx] = p
		idx++
	}
	LogTest.Infof("Starting with pyramid %v : %d", pyramidPoints, GetPyramidSize(pyramidPoints))

	expectedTime := TickTime(0)
	for expectedTime < 250 {
		assert.Equal(t, expectedTime, space.currentTime)
		col := space.ForwardTime()
		expectedTime++
		// This collection contains all the blocks of three events that have points activated at the same time
		pointsPer3Ids := col.multiEvents.pointsPerThreeIds
		nbThreeIdsActive := len(pointsPer3Ids)
		if nbThreeIdsActive >= 3 {
			LogTest.Infof("Found a 3 match with %d elements", nbThreeIdsActive)
			if nbThreeIdsActive >= 4 {
				LogTest.Info("Found a 4 match")
				builder := PyramidBuilder{make(map[Pyramid]int64, 1)}
				builder.createPyramids(pointsPer3Ids, &Pyramid{}, 0, nbThreeIdsActive-4)
				allPyramids := builder.allPyramids
				LogTest.Infof("AllPyramids %d", len(allPyramids))
				assert.True(t, len(allPyramids) > 0)
				if len(allPyramids) > 1 {
					bestSize := int64(0)
					var bestPyramid [4]Point
					for pyramid, size := range allPyramids {
						LogTest.Debugf("%v : %d", pyramid, size)
						if size > bestSize {
							bestSize = size
							bestPyramid = pyramid
						}
					}
					LogTest.Infof("We have a winner %v at size %d", bestPyramid, bestSize)
					break
				}
			}
		}
	}
}

// Builder to extract possible pyramids out of a list of ThreeIds that have common points
type PyramidBuilder struct {
	// All the possible pyramids built out
	allPyramids map[Pyramid]int64
}

func (b *PyramidBuilder) createPyramids(currentPointsPer3Ids map[ThreeIds]*[]Point, currentPyramid *Pyramid, currentPos int, possibleSkip int) {
	// Recursive Algorithm:
	// Find threeIds with smallest list of points (small3Ids),
	// Iterate though each point in the list of points for this small3Ids -> pickedPoint,
	//   Stop Condition: If currentPos is 3:
	//     - Create all the pyramids with the currentPos point being pickedPoint
	//   Logic for next call:
	//     - Recreate the map of pointsPerThreeIds removing the small3Ids and the pickedPoint from all the lists
	//     - Recurse to createPyramids with params:
	//       - the new maps filtered above
	//       - new pyramid with the currentPos point being pickedPoint
	//       - currentPos + 1
	curLength := len(currentPointsPer3Ids)
	if curLength == 0 {
		log.Fatal("Should never reach here with an empty map")
	}
	if curLength == 1 && currentPos != 3 {
		log.Fatal("Reached the end of the map but not the ned of the pyramid building for:", currentPos, currentPointsPer3Ids)
	}

	// Last points in pyramid
	// TODO: May be allowed bigger structure
	if currentPos == 3 {
		for _, points := range currentPointsPer3Ids {
			for _, pickedPoint := range *(points) {
				// Dereference creates a copy
				newPyramid := *(currentPyramid)
				newPyramid[currentPos] = pickedPoint
				b.allPyramids[newPyramid.ordered()] = GetPyramidSize(newPyramid)
			}
		}
		return
	}

	small3Ids := NilThreeIds
	minLength := int(math.MaxInt32)
	for tIds, points := range currentPointsPer3Ids {
		l := len(*points)
		if l < minLength {
			minLength = l
			small3Ids = tIds
		}
	}
	if small3Ids == NilThreeIds {
		log.Fatal("Did not find any smallest in", currentPointsPer3Ids)
	}

	// If there are some possible skips do a skip of this ThreeIds
	if possibleSkip > 0 {
		newCurrentPointsPer3Ids := make(map[ThreeIds]*[]Point, curLength-1)
		for tIds, points := range currentPointsPer3Ids {
			if tIds != small3Ids {
				newCurrentPointsPer3Ids[tIds] = points
			}
		}
		b.createPyramids(newCurrentPointsPer3Ids, currentPyramid, currentPos, possibleSkip-1)
	}

	// Do the full logic
	for _, pickedPoint := range *(currentPointsPer3Ids[small3Ids]) {
		// Dereference creates a copy
		newPyramid := *(currentPyramid)
		newPyramid[currentPos] = pickedPoint
		newCurrentPointsPer3Ids := make(map[ThreeIds]*[]Point, curLength-1)
		for tIds, points := range currentPointsPer3Ids {
			if tIds != small3Ids {
				newList := make([]Point, 0, len(*points))
				for _, p := range *points {
					if p != pickedPoint {
						newList = append(newList, p)
					}
				}
				newCurrentPointsPer3Ids[tIds] = &newList
			}
		}
		b.createPyramids(newCurrentPointsPer3Ids, &newPyramid, currentPos+1, possibleSkip)
	}
}
