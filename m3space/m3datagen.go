package m3space

import (
	"fmt"
	"sort"
)

type PointState struct {
	creationTime TickTime
	globalIdx    int
	pos          Point
	trioIndex    int
	from1, from2 int
}

func (ps *PointState) ToDataString() string {
	return fmt.Sprintf("%d:%d:[%d,%d,%d]:%d:%d:%d",
		ps.creationTime, ps.globalIdx, ps.pos[0], ps.pos[1], ps.pos[2], ps.trioIndex, ps.from1, ps.from2)
}

func FromDataString(line string) *PointState {
	res := PointState{}
	nbRead, err := fmt.Sscanf(line, "%d:%d:[%d,%d,%d]:%d:%d:%d",
		&(res.creationTime), &(res.globalIdx), &(res.pos[0]), &(res.pos[1]), &(res.pos[2]), &(res.trioIndex), &(res.from1), &(res.from2))
	if err != nil {
		Log.Warnf("parsing line %s failed with %v", line, err)
		return nil
	}
	if nbRead != 8 {
		Log.Warnf("parsing line %s dis not return 8 field but %d", line, nbRead)
		return nil
	}
	return &res
}

func (ps PointState) HasFrom1() bool {
	return ps.from1 >= 0
}

func (ps PointState) HasFrom2() bool {
	return ps.from2 >= 0
}

func (ps PointState) GetFromString() string {
	if !ps.HasFrom1() {
		return fmt.Sprintf("%3s %3s", " ", " ")
	}
	if ps.HasFrom2() {
		return fmt.Sprintf("%3d %3d", ps.from1, ps.from2)
	} else {
		return fmt.Sprintf("%3d %3s", ps.from1, " ")
	}
}

func extractMainAndOtherPoints(pointMap *map[Point]*PointState, time TickTime) (mainPoints, otherPoints []Point) {
	mainPoints = make([]Point, 0, len(*pointMap)/3)
	otherPoints = make([]Point, 0, len(*pointMap)-cap(mainPoints))
	for k, v := range *pointMap {
		if v.creationTime == time {
			if k.IsMainPoint() {
				mainPoints = append(mainPoints, k)
			} else {
				otherPoints = append(otherPoints, k)
			}
		}
	}
	sort.Slice(mainPoints, func(i, j int) bool { return (*pointMap)[mainPoints[i]].HasFrom2() && !(*pointMap)[mainPoints[j]].HasFrom2() })
	sort.Slice(otherPoints, func(i, j int) bool { return (*pointMap)[otherPoints[i]].HasFrom2() && !(*pointMap)[otherPoints[j]].HasFrom2() })

	return
}

func collectFlow(ctx *GrowthContext, untilTime TickTime, writeAllPoints func(pointMap *map[Point]*PointState, time TickTime)) {
	InitConnectionDetails()

	globalPointIdx := 0
	time := TickTime(0)
	allPoints := make(map[Point]*PointState, 100)
	allPoints[Origin] = &PointState{time, globalPointIdx, Origin, ctx.GetTrioIndex(ctx.GetDivByThree(Origin)), -1, -1,}
	globalPointIdx++
	writeAllPoints(&allPoints, time)

	nbPoints := 3
	for ; time < untilTime; {
		currentPoints := make([]Point, 0, nbPoints)
		for k, v := range allPoints {
			if v.creationTime == time {
				currentPoints = append(currentPoints, k)
			}
		}
		nbPoints = len(currentPoints) * 2
		newPoints := make([]Point, 0, nbPoints)
		for _, p := range currentPoints {
			currentState := allPoints[p]
			nps := p.getNextPoints(ctx)
			for _, np := range nps {
				npState, ok := allPoints[np]
				if !ok {
					if np.IsMainPoint() {
						allPoints[np] = &PointState{time + 1, globalPointIdx, np, ctx.GetTrioIndex(ctx.GetDivByThree(np)), currentState.globalIdx, -1}
					} else {
						allPoints[np] = &PointState{time + 1, globalPointIdx, np, -1, currentState.globalIdx, -1}
					}
					globalPointIdx++
				} else if npState.creationTime == time+1 {
					// Created now but already populated
					if npState.HasFrom2() {
						Log.Error("Got 3 overlap for", npState)
					} else {
						npState.from2 = currentState.globalIdx
					}
				}
			}
		}
		currentPoints = newPoints
		time++
		writeAllPoints(&allPoints, time)
	}
}
