package m3space

import (
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3point"
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

func CreateAllIndexes(nbIndexes int) ([][4]int, [12]int) {
	// TODO: Equivalence not evident at all finally dues to relation between trioIdx and the axis
	// the points of the pyramid are equivalent. So, any reorder that end up with same array of 4 is equivalent.
	// so creating the arrays starting from previous index will create the combinations of indexes

	// This are not true combinations as each index can be duplicated
	// So, its the sum of t4 (all the same) + t3 (3 identical) + t2a (2 identical) + t2b ( 2 x 2 identical) + t1 (all different)
	// Formula | for nb = 8 | for nb = 12
	// t4 = nbIndexes | 8 | 12
	// t3 = nbIndexes * (nbIndexes-1) | 56 | 132
	// t2a = nbIndexes * (nbIndexes-1) / 2 | 28 | 66
	// t2b = nbIndexes * ( (nbIndexes-1)! / ( 2! (nbIndexes-1-2)! ) ) | 168 | 660
	// with n is number of indexes and k=4 we have number of combinations: t1 = n! / ( k! (n-k)! ) | 70 | 495
	var t4, t3, t2a, t2b, t1 int
	if nbIndexes == 8 {
		t4 = 8
		t3 = 56
		t2a = 28
		t2b = 168
		t1 = 70
	} else if nbIndexes == 12 {
		t4 = 12
		t3 = 132
		t2a = 66
		t2b = 660
		t1 = 495
	} else {
		Log.Fatalf("Nb indexes %d is not supported", nbIndexes)
		return nil, [12]int{}
	}
	nbConbinations := t4 + t3 + t2a + t2b + t1
	res := make([][4]int, nbConbinations)
	idx := 0
	var nbT4, nbT3, nbT2a, nbT2b, nbT1 int
	for i1 := 0; i1 < nbIndexes; i1++ {
		res[idx] = [4]int{i1, i1, i1, i1}
		idx++
		nbT4++
		for iT3 := 0; iT3 < nbIndexes; iT3++ {
			if iT3 != i1 {
				res[idx] = [4]int{i1, i1, i1, iT3}
				idx++
				nbT3++
			}
		}
		for i2 := i1 + 1; i2 < nbIndexes; i2++ {
			res[idx] = [4]int{i1, i1, i2, i2}
			idx++
			nbT2a++
			for iT2b := 0; iT2b < nbIndexes; iT2b++ {
				if iT2b != i1 && iT2b != i2 {
					res[idx] = [4]int{i1, i2, iT2b, iT2b}
					idx++
					nbT2b++
				}
			}
			for i3 := i2 + 1; i3 < nbIndexes; i3++ {
				for i4 := i3 + 1; i4 < nbIndexes; i4++ {
					res[idx] = [4]int{i1, i2, i3, i4}
					idx++
					nbT1++
				}
			}
		}
	}
	return res, [12]int{nbConbinations, idx, t1, nbT1, t2a, nbT2a, t2b, nbT2b, t3, nbT3, t4, nbT4}
}

func createPyramidWithParams(space *Space, pyramidSize m3point.CInt, ctxTypes [4]m3point.GrowthType, indexes [4]int, offsets [4]int) {
	space.CreateEventAtZeroTime(ctxTypes[0], indexes[0], offsets[0], m3point.Point{3, 0, 3}.Mul(pyramidSize), RedEvent)
	space.CreateEventAtZeroTime(ctxTypes[1], indexes[1], offsets[1], m3point.Point{-3, 3, 3}.Mul(pyramidSize), GreenEvent)
	space.CreateEventAtZeroTime(ctxTypes[2], indexes[2], offsets[2], m3point.Point{-3, -3, 3}.Mul(pyramidSize), BlueEvent)
	space.CreateEventAtZeroTime(ctxTypes[3], indexes[3], offsets[3], m3point.Point{0, 0, -3}.Mul(pyramidSize), YellowEvent)
}

func RunSpacePyramidWithParams(env m3util.QsmEnvironment, pSize m3point.CInt, ctxTypes [4]m3point.GrowthType, indexes [4]int, offsets [4]int) (bool, Pyramid, DistAndTime, Pyramid, int) {
	space := MakeSpace(env, 3 * 30)
	space.MaxConnections = 3
	space.BlockOnSameEvent = 3
	space.SetEventOutgrowthThreshold(DistAndTime(0))
	createPyramidWithParams(&space, pSize, ctxTypes, indexes, offsets)

	originalPyramid := Pyramid{}
	idx := 0
	for _, evt := range space.events {
		if evt != nil {
			pa, err := evt.node.GetPoint()
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

	expectedTime := DistAndTime(0)
	finalTime := DistAndTime(3)
	//if finalTime < DistAndTime(25) {
	//	finalTime = DistAndTime(25)
	//}
	found := false
	var bestPyramid Pyramid
	var bestSize m3point.DInt
	var nbPossibilities int

	for expectedTime < finalTime {
		frwdRes := space.ForwardTime()
		expectedTime++
		// This collection contains all the blocks of three events that have points activated at the same time
		pointsPer3Ids := frwdRes.pointsPerThreeIds
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
	return found, originalPyramid, space.CurrentTime, bestPyramid, nbPossibilities
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
