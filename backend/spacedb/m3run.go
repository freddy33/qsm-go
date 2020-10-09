package spacedb

import (
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/freddy33/qsm-go/model/m3space"
	"log"
	"math"
	"sort"
)

var LogRun = m3util.NewDataLogger("m3run", m3util.DEBUG)

func GetPyramidSize(points [4]m3point.Point) m3point.DInt {
	// Sum all the edges
	totalSize := m3point.DInt(0)
	totalSize += m3point.MakeVector(points[0], points[1]).DistanceSquared()
	totalSize += m3point.MakeVector(points[0], points[2]).DistanceSquared()
	totalSize += m3point.MakeVector(points[0], points[3]).DistanceSquared()
	totalSize += m3point.MakeVector(points[1], points[2]).DistanceSquared()
	totalSize += m3point.MakeVector(points[1], points[3]).DistanceSquared()
	totalSize += m3point.MakeVector(points[2], points[3]).DistanceSquared()
	return totalSize
}

type Pyramid [4]m3point.Point

func (pyramid Pyramid) ordered() Pyramid {
	slice := make([]m3point.Point, 4)
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

func createPyramidWithParams(space *SpaceDb, pyramidSize m3point.CInt, ctxTypes [4]m3point.GrowthType, indexes [4]int, offsets [4]int) {
	_, err := space.CreateEvent(ctxTypes[0], indexes[0], offsets[0], m3space.ZeroDistAndTime, m3point.Point{3, 0, 3}.Mul(pyramidSize), m3space.RedEvent)
	if err != nil {
		Log.Error(err)
		return
	}
	_, err = space.CreateEvent(ctxTypes[1], indexes[1], offsets[1], m3space.ZeroDistAndTime, m3point.Point{-3, 3, 3}.Mul(pyramidSize), m3space.GreenEvent)
	if err != nil {
		Log.Error(err)
		return
	}
	_, err = space.CreateEvent(ctxTypes[2], indexes[2], offsets[2], m3space.ZeroDistAndTime, m3point.Point{-3, -3, 3}.Mul(pyramidSize), m3space.BlueEvent)
	if err != nil {
		Log.Error(err)
		return
	}
	_, err = space.CreateEvent(ctxTypes[3], indexes[3], offsets[3], m3space.ZeroDistAndTime, m3point.Point{0, 0, -3}.Mul(pyramidSize), m3space.YellowEvent)
	if err != nil {
		Log.Error(err)
		return
	}
}

func RunSpacePyramidWithParams(space *SpaceDb, pSize m3point.CInt, ctxTypes [4]m3point.GrowthType, indexes [4]int, offsets [4]int) (bool, Pyramid, m3space.DistAndTime, Pyramid, int) {
	createPyramidWithParams(space, pSize, ctxTypes, indexes, offsets)

	originalPyramid := Pyramid{}
	idx := 0
	for _, evt := range space.GetActiveEventsAt(0) {
		if evt != nil {
			pa, err := evt.GetCenterNode().GetPoint()
			if err != nil {
				Log.Error(err)
				originalPyramid[idx] = m3point.Origin
			} else {
				originalPyramid[idx] = *pa
			}
			idx++
		}
	}
	originalPyramid = originalPyramid.ordered()
	LogRun.Infof("Starting with pyramid %v : %d", originalPyramid, GetPyramidSize(originalPyramid))

	expectedTime := m3space.ZeroDistAndTime
	finalTime := m3space.DistAndTime(9)
	//if finalTime < DistAndTime(25) {
	//	finalTime = DistAndTime(25)
	//}
	found := false
	var bestPyramid Pyramid
	var bestSize m3point.DInt
	var nbPossibilities int
	var spaceTime *SpaceTime

	for expectedTime < finalTime {
		expectedTime++
		spaceTime = space.GetSpaceTimeAt(expectedTime).(*SpaceTime)
		frwdRes := spaceTime.GetRuleAnalyzer()
		// This collection contains all the blocks of three events that have points activated at the same time
		pointsPer3Ids := frwdRes.PointsPerThreeIds
		nbThreeIdsActive := len(pointsPer3Ids)
		if nbThreeIdsActive >= 3 {
			LogRun.Debugf("Found a 3 match with %d elements", nbThreeIdsActive)
			if nbThreeIdsActive >= 4 {
				LogRun.Debug("Found a 4 match")
				builder := PyramidBuilder{make(map[Pyramid]m3point.DInt, 1)}
				builder.createPyramids(pointsPer3Ids, &Pyramid{}, 0, nbThreeIdsActive-4)
				allPyramids := builder.allPyramids
				nbPossibilities = len(allPyramids)
				LogRun.Debugf("AllPyramids %d", nbPossibilities)
				if len(allPyramids) > 0 {
					bestSize = m3point.DInt(0)
					for pyramid, size := range allPyramids {
						LogRun.Debugf("%v : %d", pyramid, size)
						if size > bestSize {
							bestSize = size
							bestPyramid = pyramid
						}
					}
					found = true
					LogRun.Infof("We have a winner out of %d possible %v at size %d", nbPossibilities, bestPyramid, bestSize)
					break
				}
			}
		}
	}
	if spaceTime == nil {
		return false, originalPyramid, m3space.ZeroDistAndTime, Pyramid{}, 0
	}
	return found, originalPyramid, spaceTime.GetCurrentTime(), bestPyramid, nbPossibilities
}

// Builder to extract possible pyramids out of a list of ThreeIds that have common points
type PyramidBuilder struct {
	// All the possible pyramids built out
	allPyramids map[Pyramid]m3point.DInt
}

func (b *PyramidBuilder) createPyramids(currentPointsPer3Ids map[ThreeIds][]m3point.Point, currentPyramid *Pyramid, currentPos int, possibleSkip int) {
	// Recursive Algorithm:
	// Find threeIds with smallest list of points (small3Ids),
	// Iterate though each point in the list of points for this small3Ids -> pickedPoint,
	//   Stop Condition: If currentPos is 3:
	//     - Create all the pyramids with the currentPos point being pickedPoint
	//   Logic for next call:
	//     - Recreate the map of PointsPerThreeIds removing the small3Ids and the pickedPoint from all the lists
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
			for _, pickedPoint := range points {
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
		l := len(points)
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
		newCurrentPointsPer3Ids := make(map[ThreeIds][]m3point.Point, curLength-1)
		for tIds, points := range currentPointsPer3Ids {
			if tIds != small3Ids {
				newCurrentPointsPer3Ids[tIds] = points
			}
		}
		b.createPyramids(newCurrentPointsPer3Ids, currentPyramid, currentPos, possibleSkip-1)
	}

	// Do the full logic
	for _, pickedPoint := range currentPointsPer3Ids[small3Ids] {
		// Dereference creates a copy
		newPyramid := *(currentPyramid)
		newPyramid[currentPos] = pickedPoint
		newCurrentPointsPer3Ids := make(map[ThreeIds][]m3point.Point, curLength-1)
		for tIds, points := range currentPointsPer3Ids {
			if tIds != small3Ids {
				newList := make([]m3point.Point, 0, len(points))
				for _, p := range points {
					if p != pickedPoint {
						newList = append(newList, p)
					}
				}
				newCurrentPointsPer3Ids[tIds] = newList
			}
		}
		b.createPyramids(newCurrentPointsPer3Ids, &newPyramid, currentPos+1, possibleSkip)
	}
}
