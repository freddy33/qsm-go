package m3space

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/stretchr/testify/assert"
	"gonum.org/v1/gonum/stat"
	"testing"
)

type ThreeIds [3]EventID

var LogDatagen = m3util.NewLogger("datagen ", m3util.DEBUG)

func MakeThreeIds(ids []EventID) []ThreeIds {
	SortEventIDs(&ids)
	if len(ids) == 3 {
		return []ThreeIds{{ids[0], ids[1], ids[2]},}
	} else if len(ids) == 4 {
		return []ThreeIds{
			{ids[0], ids[1], ids[2]},
			{ids[0], ids[2], ids[3]},
			{ids[0], ids[1], ids[3]},
			{ids[1], ids[2], ids[3]},
		}
	}
	LogDatagen.Fatal("WHAT!")
	return nil
}

func (tIds ThreeIds) contains(id EventID) bool {
	for _, tid := range tIds {
		if tid == id {
			return true
		}
	}
	return false
}

func GetPyramidSize(points [4]Point) int64 {
	// Sum all the edges
	totalSize := int64(0)
	totalSize += points[0].Sub(points[1]).DistanceSquared()
	totalSize += points[0].Sub(points[2]).DistanceSquared()
	totalSize += points[0].Sub(points[3]).DistanceSquared()
	totalSize += points[1].Sub(points[2]).DistanceSquared()
	totalSize += points[1].Sub(points[3]).DistanceSquared()
	totalSize += points[2].Sub(points[3]).DistanceSquared()
	return totalSize
}

func TestStatPack(t *testing.T) {
	Log.Level = m3util.WARN
	LogStat.Level = m3util.INFO
	fmt.Println(stat.StdDev([]float64{1.3, 1.5, 1.7, 1.1}, nil))
	space := MakeSpace(3 * 30)
	InitConnectionDetails()
	space.SetEventOutgrowthThreshold(Distance(0))
	space.CreatePyramid(18)

	pyramidPoints := [4]Point{}
	idx := 0
	for p := range space.activeNodesMap {
		pyramidPoints[idx] = p
		idx++
	}
	LogDatagen.Info("Starting with pyramid size of ", GetPyramidSize(pyramidPoints))

	expectedTime := TickTime(0)
	for expectedTime < 200 {
		assert.Equal(t, expectedTime, space.currentTime)
		col := space.ForwardTime()
		expectedTime++
		// This collection contains all the points activated by 3 separate events at the same time
		if len(col.multiEvents.moreThan3EventsPerPoint) >= 3 {
			LogDatagen.Debugf("Analyzing match with %d elements in %v", len(col.multiEvents.moreThan3EventsPerPoint), col.multiEvents.moreThan3EventsPerPoint)
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
			LogDatagen.Debugf("Reorganization of map size %d elements in %v", len(eventsPerPoints), eventsPerPoints)
			LogDatagen.Debugf("All three ids size %d in %v", len(allThreeIds), allThreeIds)
			LogDatagen.Debugf("Points per events size is %d in %v filtering less than three points", len(pointsPerEvent), pointsPerEvent)
			validEventIds := make([]EventID, 0, 3)
			for id, points := range pointsPerEvent {
				if len(points) < 3 {
					LogDatagen.Debug("Event id", id, "does not have enough points. Removing it!")
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
				}
			}
			SortEventIDs(&validEventIds)
			for p, threeIds := range eventsPerPoints {
				if len(threeIds) == 0 {
					delete(eventsPerPoints, p)
				}
			}
			LogDatagen.Debugf("After filter: validIds=%d events=%d points=%d", len(validEventIds), len(pointsPerEvent), len(eventsPerPoints))

			if len(pointsPerEvent) >= 3 && len(validEventIds) >= 3 && len(eventsPerPoints) >= 3 {
				LogDatagen.Info("Found a 3 match")
				if len(pointsPerEvent) >= 4 && len(validEventIds) >= 4 && len(eventsPerPoints) >= 4 {
					idx = 0
					for p := range eventsPerPoints {
						pyramidPoints[idx] = p
						idx++
						if idx == 4 {
							break
						}
					}
					LogDatagen.Info("We have a winner at size", GetPyramidSize(pyramidPoints))
					break
				}
			}
		}
	}
}

type ThreeEventsCollector struct {
	points []Point
}
