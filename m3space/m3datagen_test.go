package m3space

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/stretchr/testify/assert"
	"gonum.org/v1/gonum/stat"
	"sort"
	"testing"
)

var LogTest = m3util.NewDataLogger("m3test", m3util.DEBUG)

func GetPyramidSize(points [4]Point) int64 {
	// Sum all the edges
	totalSize := int64(0)
	totalSize += MakeVector(points[0],points[1]).DistanceSquared()
	totalSize += MakeVector(points[0],points[2]).DistanceSquared()
	totalSize += MakeVector(points[0],points[3]).DistanceSquared()
	totalSize += MakeVector(points[1],points[2]).DistanceSquared()
	totalSize += MakeVector(points[1],points[3]).DistanceSquared()
	totalSize += MakeVector(points[2],points[3]).DistanceSquared()
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
	return Pyramid{slice[0], slice[1], slice[2], slice[3],}
}

func TestStatPack(t *testing.T) {
	Log.Level = m3util.WARN
	LogStat.Level = m3util.INFO
	fmt.Println(stat.StdDev([]float64{1.3, 1.5, 1.7, 1.1}, nil))
	space := MakeSpace(3 * 30)
	space.MaxConnections = 3
	space.blockOnSameEvent = 3
	space.SetEventOutgrowthThreshold(Distance(0))
	space.CreatePyramid(20)
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
	for expectedTime < 50 {
		assert.Equal(t, expectedTime, space.currentTime)
		col := space.ForwardTime()
		expectedTime++
		// This collection contains all the blocks of three events that have points activated at the same time
		nbThreeIdsActive := len(col.multiEvents.pointsPerThreeIds)
		if nbThreeIdsActive >= 3 {
			LogTest.Infof("Found a 3 match with %d elements", nbThreeIdsActive)
			if nbThreeIdsActive >= 4 {
				LogTest.Info("Found a 4 match")
				builder := PyramidBuilder{pointsPerEvent, validEventIds, make(map[Pyramid]int64, 1),}
				builder.processEventId(validEventIds[0], Pyramid{})
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

			// Reorganizing the map into maps of block of three ids
			eventsPerPoints := make(map[Point]map[ThreeIds]int, 4)
			allThreeIds := make(map[ThreeIds]int, 4)
			// Let's collect for every event involved in the collection all the ones which have 3 separate points in it
			pointsPerEvent := make(map[EventID][]Point, 4)
			for p, ids := range col.multiEvents.moreThan3EventsPerPoint {
				for _, id := range ids {
					points, ok := pointsPerEvent[id]
					if !ok {
						points = make([]Point, 1)
						points[0] = p
					} else {
						points = append(points, p)
					}
					pointsPerEvent[id] = points
				}
				currentThreeIds := MakeThreeIds(ids)
				threeIds, ok := eventsPerPoints[p]
				if !ok {
					threeIds = make(map[ThreeIds]int, 1)
				}
				for _, tid := range currentThreeIds {
					threeIds[tid]++
					allThreeIds[tid]++
				}
				eventsPerPoints[p] = threeIds
			}
			LogTest.Debugf("Reorganization of maps points=%d, events=%d, threeIds=%d, full=%v", len(eventsPerPoints), len(pointsPerEvent), len(allThreeIds), allThreeIds)
			validEventIds := make([]EventID, 0, 3)
			maxPoints := 0
			for id, points := range pointsPerEvent {
				nbPoints := len(points)
				if nbPoints < 3 {
					LogTest.Debug("Event id", id, "does not have enough points. Removing it!")
					delete(pointsPerEvent, id)
					for p, threeIds := range eventsPerPoints {
						for tIds := range threeIds {
							if tIds.contains(id) {
								delete(threeIds, tIds)
								eventsPerPoints[p] = threeIds
							}
						}
					}
					for tIds := range allThreeIds {
						if tIds.contains(id) {
							delete(allThreeIds, tIds)
						}
					}
				} else {
					validEventIds = append(validEventIds, id)
					if nbPoints > maxPoints {
						maxPoints = nbPoints
					}
				}
			}
			SortEventIDs(&validEventIds)
			for p, threeIds := range eventsPerPoints {
				if len(threeIds) == 0 {
					delete(eventsPerPoints, p)
				}
			}
			LogTest.Debugf("After filter: validIds=%d events=%d points=%d vslidIds=%v", len(validEventIds), len(pointsPerEvent), len(eventsPerPoints), validEventIds)

			if len(pointsPerEvent) >= 3 && len(validEventIds) >= 3 && len(eventsPerPoints) >= 3 {
				if len(pointsPerEvent) >= 4 && len(validEventIds) >= 4 && len(eventsPerPoints) >= 4 {
					LogTest.Info("Found a 4 match")
					builder := PyramidBuilder{pointsPerEvent, validEventIds, make(map[Pyramid]int64, 1),}
					builder.processEventId(validEventIds[0], Pyramid{})
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
}

type PyramidBuilder struct {
	pointsPerEvent map[EventID][]Point
	validEventIds  []EventID
	allPyramids    map[Pyramid]int64
}

func (b *PyramidBuilder) processEventId(evtId EventID, pyramid Pyramid) {
	evtIdIdx := -1
	for i, validId := range b.validEventIds {
		if validId == evtId {
			evtIdIdx = i
			break
		}
	}
	if evtIdIdx < 0 {
		LogTest.Fatalf("did not find event id %d in valid list %v", evtId, b.validEventIds)
	}
	for _, p := range b.pointsPerEvent[evtId] {
		pointAlreadyThere := false
		if evtIdIdx > 0 {
			for j := 0; j < evtIdIdx; j++ {
				if pyramid[j] == p {
					pointAlreadyThere = true
				}
			}
		}
		if !pointAlreadyThere {
			pyramid[evtIdIdx] = p
			if evtIdIdx+1 < len(b.validEventIds) {
				b.processEventId(b.validEventIds[evtIdIdx+1], pyramid)
			} else {
				orderedPyramid := pyramid.ordered()
				b.allPyramids[orderedPyramid] = GetPyramidSize(orderedPyramid)
			}
		}
	}
}
