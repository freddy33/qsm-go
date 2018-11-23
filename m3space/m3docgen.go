package m3space

import (
	"fmt"
	"os"
	"log"
	"encoding/csv"
	"sort"
)

func WriteAllTables() {
	changeToDocsDir()
	InitConnectionDetails()
	writeAllTrioTable()
	writeTrioConnectionsTable()
	writeAllConnectionDetails()
}

func changeToDocsDir() {
	if _, err := os.Stat("docs"); !os.IsNotExist(err) {
		os.Chdir("docs")
		if _, err := os.Stat("generated"); os.IsNotExist(err) {
			os.Mkdir("generated", os.ModePerm)
		}
		os.Chdir("generated")
	}
}

type Int2 struct {
	a, b int
}

// Return the kind of connection between 2 trios depending of the distance square values
// A3 => All connections have a DS of 3
// A5 => All connections have a DS of 5
// X135 => All DS present 1, 3 and 5
// G13 => 1 and 3 are present but no DS 5 (The type we use)
func GetTrioConnType(conns [6]Point) string {
	has1 := false
	has3 := false
	has5 := false
	for _, conn := range conns {
		ds := conn.DistanceSquared()
		switch ds {
		case 1:
			has1 = true
		case 3:
			has3 = true
		case 5:
			has5 = true
		}
	}
	if !has1 && !has3 && has5 {
		// All 5
		return "A5  "
	}
	if !has1 && has3 && !has5 {
		// All 3
		return "A3  "
	}
	if has1 && has3 && has5 {
		// 1, 3 and 5
		return "X135"
	}
	if has1 && has3 && !has5 {
		// Good ones with 1 and 3
		return "G13 "
	}
	log.Fatalf("Trio connection list inconsistent got 1=%t, 3=%t, 5=%t", has1, has3, has5)
	return "WRONG"
}

func GetTrioTransitionTableTxt() map[Int2][7]string {
	result := make(map[Int2][7]string, 8*8)
	for a, tA := range AllBaseTrio {
		for b, tB := range AllBaseTrio {
			txtOut := [7]string{}
			conns := GetNonBaseConnections(tA, tB)
			txtOut[0] = GetTrioConnType(conns)
			for i, conn := range conns {
				cd := AllConnectionsPossible[conn]
				// Total size 18
				txtOut[i+1] = fmt.Sprintf("%v %s", conn, cd.GetName())
			}
			result[Int2{a, b}] = txtOut
		}
	}
	return result
}

func GetTrioTransitionTableCsv() [][]string {
	csvOutput := make([][]string, 8*8)
	for a, tA := range AllBaseTrio {
		for b, tB := range AllBaseTrio {
			lineNb := a * 8
			if b == 0 {
				csvOutput[lineNb] = make([]string, 7*8)
			}
			baseColumn := b * 7
			columnNb := baseColumn
			csvOutput[lineNb][columnNb] = fmt.Sprintf("%d", a)
			columnNb++
			csvOutput[lineNb][columnNb] = fmt.Sprintf("%d", b)
			columnNb++
			for ; columnNb < 7; columnNb++ {
				csvOutput[lineNb][columnNb] = ""
			}

			conns := GetNonBaseConnections(tA, tB)
			for _, conn := range conns {
				ds := conn.DistanceSquared()

				lineNb++
				if b == 0 {
					csvOutput[lineNb] = make([]string, 7*8)
				}
				// Empty to first column
				for columnNb = baseColumn; columnNb < baseColumn+2; columnNb++ {
					csvOutput[lineNb][columnNb] = ""
				}
				csvOutput[lineNb][columnNb] = fmt.Sprintf("%d", conn[0])
				columnNb++
				csvOutput[lineNb][columnNb] = fmt.Sprintf("%d", conn[1])
				columnNb++
				csvOutput[lineNb][columnNb] = fmt.Sprintf("%d", conn[2])
				columnNb++
				csvOutput[lineNb][columnNb] = fmt.Sprintf("%d", ds)
			}
		}
	}
	return csvOutput
}

func GetTrioTableCsv() [][]string {
	nbColumns := 5
	nbRowsPerTrio := 4
	csvOutput := make([][]string, len(AllBaseTrio)*nbColumns)
	for a, trio := range AllBaseTrio {
		lineNb := a * nbRowsPerTrio
		csvOutput[lineNb] = make([]string, nbColumns)
		columnNb := 0
		csvOutput[lineNb][columnNb] = fmt.Sprintf("%d", a)
		columnNb++
		for ; columnNb < nbColumns; columnNb++ {
			csvOutput[lineNb][columnNb] = ""
		}
		for _, bv := range trio {
			lineNb++
			csvOutput[lineNb] = make([]string, nbColumns)
			columnNb := 0
			csvOutput[lineNb][columnNb] = ""
			columnNb++
			csvOutput[lineNb][columnNb] = fmt.Sprintf("%d", bv[0])
			columnNb++
			csvOutput[lineNb][columnNb] = fmt.Sprintf("%d", bv[1])
			columnNb++
			csvOutput[lineNb][columnNb] = fmt.Sprintf("%d", bv[2])
			columnNb++
			csvOutput[lineNb][columnNb] = AllConnectionsPossible[bv].GetName()
		}
	}
	return csvOutput
}

// Write all the 8x8 connections possible for all trio in text and CSV files, and classify the connections size DS
func writeTrioConnectionsTable() {
	txtFile, err := os.Create("TrioConnectionsTable.txt")
	if err != nil {
		log.Fatal("Cannot create text file", err)
	}
	csvFile, err := os.Create("TrioConnectionsTable.csv")
	if err != nil {
		log.Fatal("Cannot create csv file", err)
	}
	defer txtFile.Close()
	defer csvFile.Close()

	csvWriter := csv.NewWriter(csvFile)
	csvWriter.WriteAll(GetTrioTransitionTableCsv())

	txtOutputs := GetTrioTransitionTableTxt()
	for a := 0; a < 8; a++ {
		for b := 0; b < 8; b++ {
			out := txtOutputs[Int2{a, b}]
			if b == 7 {
				txtFile.WriteString(fmt.Sprintf("%d, %d %s", a, b, out[0]))
			} else {
				txtFile.WriteString(fmt.Sprintf("%d, %d %s            ", a, b, out[0]))
			}
		}
		txtFile.WriteString("\n")
		for i := 0; i < 6; i++ {
			for b := 0; b < 8; b++ {
				out := txtOutputs[Int2{a, b}]
				// this is 18 chars
				txtFile.WriteString(out[i+1])
				if b != 7 {
					txtFile.WriteString("  ")
				}
			}
			txtFile.WriteString("\n")
		}
		txtFile.WriteString("\n")
	}
}

// Write all the 8 base vectors trio in text and CSV files
func writeAllTrioTable() {
	txtFile, err := os.Create("AllTrioTable.txt")
	if err != nil {
		log.Fatal("Cannot create text file", err)
	}
	csvFile, err := os.Create("AllTrioTable.csv")
	if err != nil {
		log.Fatal("Cannot create csv file", err)
	}
	defer txtFile.Close()
	defer csvFile.Close()

	csvWriter := csv.NewWriter(csvFile)
	csvWriter.WriteAll(GetTrioTableCsv())
	for a, trio := range AllBaseTrio {
		txtFile.WriteString(fmt.Sprintf("T%d:\t%v\t%s\n", a, trio[0], AllConnectionsPossible[trio[0]].GetName()))
		txtFile.WriteString(fmt.Sprintf("\t%v\t%s\n", trio[1], AllConnectionsPossible[trio[1]].GetName()))
		txtFile.WriteString(fmt.Sprintf("\t%v\t%s\n", trio[2], AllConnectionsPossible[trio[2]].GetName()))
		txtFile.WriteString("\n")
	}
}

// Write all the connection details in text and CSV files
func writeAllConnectionDetails() {
	txtFile, err := os.Create("AllConnectionDetails.txt")
	if err != nil {
		log.Fatal("Cannot create text file", err)
	}
	csvFile, err := os.Create("AllConnectionDetails.csv")
	if err != nil {
		log.Fatal("Cannot create csv file", err)
	}
	defer txtFile.Close()
	defer csvFile.Close()

	nbConnDetails := uint8(len(AllConnectionsPossible) / 2)
	csvWriter := csv.NewWriter(csvFile)
	for cdNb := uint8(0); cdNb < nbConnDetails; cdNb++ {
		for _, v := range AllConnectionsPossible {
			if v.ConnNumber == cdNb && !v.ConnNeg {
				ds := v.ConnDS
				posVec := v.Vector
				negVec := v.Vector.Neg()
				csvWriter.Write([]string{
					fmt.Sprintf(" %d", cdNb),
					fmt.Sprintf("% d", posVec[0]),
					fmt.Sprintf("% d", posVec[1]),
					fmt.Sprintf("% d", posVec[2]),
					fmt.Sprintf("% d", ds),
				})
				csvWriter.Write([]string{
					fmt.Sprintf("-%d", cdNb),
					fmt.Sprintf("% d", negVec[0]),
					fmt.Sprintf("% d", negVec[1]),
					fmt.Sprintf("% d", negVec[2]),
					fmt.Sprintf("% d", ds),
				})
				txtFile.WriteString(fmt.Sprintf("%s: %v = %d\n", v.GetName(), posVec, ds))
				negCD := AllConnectionsPossible[negVec]
				txtFile.WriteString(fmt.Sprintf("%s: %v = %d\n", negCD.GetName(), negVec, ds))
				break
			}
		}
	}
}

type PointFrom struct {
	time  TickTime
	index int
}

type PointState struct {
	globalIdx    int
	creationTime TickTime
	main         bool
	trioIndex    int
	from         []PointFrom
}

func (ps PointState) GetFromString() string {
	if ps.from == nil || len(ps.from) == 0 {
		return fmt.Sprintf("%3s %3s", " ", " ")
	}
	if len(ps.from) == 1 {
		return fmt.Sprintf("%3d %3s", ps.from[0].index, " ")
	}
	if len(ps.from) == 2 {
		return fmt.Sprintf("%3d %3d", ps.from[0].index, ps.from[1].index)
	}
	return fmt.Sprintf("%d:%3d %3d", len(ps.from), ps.from[0].index, ps.from[1].index)
}

// Write all the points, base vector used, DS and connection details used from1 T=0 to T=X when transitioning from1 Trio Index 0 to 4 back and forth
func Write0To4TimeFlow() {
	changeToDocsDir()
	InitConnectionDetails()
	// Start from1 origin with growth context type 2 index 0
	ctx := &GrowthContext{&Origin, 2, 0, false, 0}

	untilTime := TickTime(8)
	txtFile, err := os.Create(fmt.Sprintf("Center_%03d_%03d_%03d_Growth_%d_%d_Time_%03d.txt",
		ctx.center[0], ctx.center[1], ctx.center[2],
		ctx.permutationType, ctx.permutationIndex, untilTime))
	if err != nil {
		log.Fatal("Cannot create text file", err)
	}

	globalPointIdx := 0
	time := TickTime(0)
	allPoints := make(map[Point]*PointState, 100)
	allPoints[Origin] = &PointState{globalPointIdx, time, true, ctx.GetTrioIndex(ctx.GetDivByThree(Origin)), nil,}
	globalPointIdx++
	WriteCurrentPointsToFile(txtFile, time, &allPoints)

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
					from := make([]PointFrom, 1, 2)
					from[0] = PointFrom{currentState.creationTime, currentState.globalIdx}
					isMainPoint := np.IsMainPoint()
					if isMainPoint {
						allPoints[np] = &PointState{globalPointIdx, time + 1, isMainPoint, ctx.GetTrioIndex(ctx.GetDivByThree(np)), from,}
					} else {
						allPoints[np] = &PointState{globalPointIdx, time + 1, isMainPoint, -1, from,}
					}
					globalPointIdx++
				} else {
					npState.from = append(npState.from, PointFrom{currentState.creationTime, currentState.globalIdx})
				}
			}
		}
		currentPoints = newPoints
		time++
		WriteCurrentPointsToFile(txtFile, time, &allPoints)
	}
}

func WriteCurrentPointsToFile(txtFile *os.File, time TickTime, allPoints *map[Point]*PointState) {
	mainPoints := make([]Point, 0, len(*allPoints)/3)
	currentPoints := make([]Point, 0, len(*allPoints)-cap(mainPoints))
	for k, v := range *allPoints {
		if v.creationTime == time {
			if k.IsMainPoint() {
				mainPoints = append(mainPoints, k)
			} else {
				currentPoints = append(currentPoints, k)
			}
		}
	}
	sort.Slice(mainPoints, func (i, j int) bool { return len((*allPoints)[mainPoints[i]].from) > len((*allPoints)[mainPoints[j]].from) })
	sort.Slice(currentPoints, func (i, j int) bool { return len((*allPoints)[currentPoints[i]].from) > len((*allPoints)[currentPoints[j]].from) })
	txtFile.WriteString("\n**************************************************\n")
	txtFile.WriteString(fmt.Sprintf("Time: %4d       %4d       %4d\n######  MAIN POINTS: %4d #######", time, time, time, len(mainPoints)))
	for i, p := range mainPoints {
		if i%4 == 0 {
			txtFile.WriteString("\n")
		}
		pState := (*allPoints)[p]
		txtFile.WriteString(fmt.Sprintf("%3d - %d: %2d, %2d, %2d <= %s | ", pState.globalIdx, pState.trioIndex, p[0], p[1], p[2], pState.GetFromString()))
	}
	txtFile.WriteString(fmt.Sprintf("\n###### OTHER POINTS: %4d #######", len(currentPoints)))
	for i, p := range currentPoints {
		if i%6 == 0 {
			txtFile.WriteString("\n")
		}
		pState := (*allPoints)[p]
		txtFile.WriteString(fmt.Sprintf("%3d: %2d, %2d, %2d <= %s | ", pState.globalIdx, p[0], p[1], p[2], pState.GetFromString()))
	}
}
