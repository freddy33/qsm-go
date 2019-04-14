package m3space

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3point"
	"github.com/freddy33/qsm-go/m3util"
	"log"
	"os"
	"sort"
)

type PointState struct {
	creationTime TickTime
	globalIdx    int
	pos          m3point.Point
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

func extractMainAndOtherPoints(pointMap *map[m3point.Point]*PointState, time TickTime) (mainPoints, otherPoints []m3point.Point) {
	mainPoints = make([]m3point.Point, 0, len(*pointMap)/3)
	otherPoints = make([]m3point.Point, 0, len(*pointMap)-cap(mainPoints))
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

func collectFlow(ctx *m3point.GrowthContext, untilTime TickTime, writeAllPoints func(pointMap *map[m3point.Point]*PointState, time TickTime)) {
	globalPointIdx := 0
	time := TickTime(0)
	allPoints := make(map[m3point.Point]*PointState, 100)
	allPoints[m3point.Origin] = &PointState{time, globalPointIdx, m3point.Origin, ctx.GetTrioIndex(ctx.GetDivByThree(m3point.Origin)), -1, -1,}
	globalPointIdx++
	writeAllPoints(&allPoints, time)

	nbPoints := 3
	for ; time < untilTime; {
		currentPoints := make([]m3point.Point, 0, nbPoints)
		for k, v := range allPoints {
			if v.creationTime == time {
				currentPoints = append(currentPoints, k)
			}
		}
		nbPoints = len(currentPoints) * 2
		newPoints := make([]m3point.Point, 0, nbPoints)
		for _, p := range currentPoints {
			currentState := allPoints[p]
			nps := p.GetNextPoints(ctx)
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

// Write all the points, base vector used, DS and connection details used from1 T=0 to T=X when transitioning from Trio Index 0 to 4 back and forth
func Write0To4TimeFlow() {
	m3util.ChangeToDocsGeneratedDir()

	// Start from origin with growth context type 2 index 0
	ctx := m3point.CreateGrowthContext(m3point.Origin, 2, 0, 0)
	untilTime := TickTime(8)

	txtFile, err := os.Create(fmt.Sprintf("%s_Time_%03d.txt", ctx.GetFileName(), untilTime))
	if err != nil {
		log.Fatal("Cannot create text file", err)
	}

	collectFlow(ctx, untilTime, func(pointMap *map[m3point.Point]*PointState, time TickTime) {
		WriteCurrentPointsToFile(txtFile, time, pointMap)
	})
}


func WriteCurrentPointsToFile(txtFile *os.File, time TickTime, allPoints *map[m3point.Point]*PointState) {
	mainPoints, currentPoints := extractMainAndOtherPoints(allPoints, time)

	m3util.WriteNextString(txtFile, "\n**************************************************\n")
	m3util.WriteNextString(txtFile, fmt.Sprintf("Time: %4d       %4d       %4d\n######  MAIN POINTS: %4d #######", time, time, time, len(mainPoints)))
	for i, p := range mainPoints {
		if i%4 == 0 {
			m3util.WriteNextString(txtFile, "\n")
		}
		pState := (*allPoints)[p]
		m3util.WriteNextString(txtFile, fmt.Sprintf("%3d - %d: %2d, %2d, %2d <= %s | ", pState.globalIdx, pState.trioIndex, p[0], p[1], p[2], pState.GetFromString()))
	}
	m3util.WriteNextString(txtFile, fmt.Sprintf("\n###### OTHER POINTS: %4d #######", len(currentPoints)))
	for i, p := range currentPoints {
		if i%6 == 0 {
			m3util.WriteNextString(txtFile, "\n")
		}
		pState := (*allPoints)[p]
		m3util.WriteNextString(txtFile, fmt.Sprintf("%3d: %2d, %2d, %2d <= %s | ", pState.globalIdx, p[0], p[1], p[2], pState.GetFromString()))
	}
}

func GenerateDataTimeFlow0() {
	m3util.ChangeToDocsDataDir()

	// Start from origin with growth context type 2 index 0
	ctx := m3point.CreateGrowthContext(m3point.Origin, 2, 0, 0)
	untilTime := TickTime(30)

	binFile, err := os.Create(fmt.Sprintf("%s_Time_%03d.data", ctx.GetFileName(), untilTime))
	if err != nil {
		log.Fatal("Cannot create bin data file", err)
	}

	collectFlow(ctx, untilTime, func(pointMap *map[m3point.Point]*PointState, time TickTime) {
		WriteCurrentPointsDataToFile(binFile, time, pointMap)
	})
}

func WriteCurrentPointsDataToFile(file *os.File, time TickTime, allPoints *map[m3point.Point]*PointState) {
	for _, ps := range *allPoints {
		if ps.creationTime == time {
			m3util.WriteNextString(file, ps.ToDataString())
			m3util.WriteNextString(file, "\n")
		}
	}
}
