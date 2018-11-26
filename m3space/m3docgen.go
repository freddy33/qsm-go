package m3space

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
)

func WriteAllTables() {
	changeToDocsGeneratedDir()
	InitConnectionDetails()
	writeAllTrioTable()
	writeTrioConnectionsTable()
	writeAllConnectionDetails()
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
	defer closeFile(txtFile)
	defer closeFile(csvFile)

	csvWriter := csv.NewWriter(csvFile)
	writeAll(csvWriter, GetTrioTransitionTableCsv())
	csvWriter.Flush()

	txtOutputs := GetTrioTransitionTableTxt()
	for a := 0; a < 8; a++ {
		for b := 0; b < 8; b++ {
			out := txtOutputs[Int2{a, b}]
			if b == 7 {
				writeNextString(txtFile, fmt.Sprintf("%d, %d %s", a, b, out[0]))
			} else {
				writeNextString(txtFile, fmt.Sprintf("%d, %d %s            ", a, b, out[0]))
			}
		}
		writeNextString(txtFile, "\n")
		for i := 0; i < 6; i++ {
			for b := 0; b < 8; b++ {
				out := txtOutputs[Int2{a, b}]
				// this is 18 chars
				writeNextString(txtFile, out[i+1])
				if b != 7 {
					writeNextString(txtFile, "  ")
				}
			}
			writeNextString(txtFile, "\n")
		}
		writeNextString(txtFile, "\n")
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
	defer closeFile(txtFile)
	defer closeFile(csvFile)

	csvWriter := csv.NewWriter(csvFile)
	writeAll(csvWriter, GetTrioTableCsv())
	for a, trio := range AllBaseTrio {
		writeNextString(txtFile, fmt.Sprintf("T%d:\t%v\t%s\n", a, trio[0], AllConnectionsPossible[trio[0]].GetName()))
		writeNextString(txtFile, fmt.Sprintf("\t%v\t%s\n", trio[1], AllConnectionsPossible[trio[1]].GetName()))
		writeNextString(txtFile, fmt.Sprintf("\t%v\t%s\n", trio[2], AllConnectionsPossible[trio[2]].GetName()))
		writeNextString(txtFile, "\n")
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
	defer closeFile(txtFile)
	defer closeFile(csvFile)

	nbConnDetails := uint8(len(AllConnectionsPossible) / 2)
	csvWriter := csv.NewWriter(csvFile)
	for cdNb := uint8(0); cdNb < nbConnDetails; cdNb++ {
		for _, v := range AllConnectionsPossible {
			if v.ConnNumber == cdNb && !v.ConnNeg {
				ds := v.ConnDS
				posVec := v.Vector
				negVec := v.Vector.Neg()
				write(csvWriter, []string{
					fmt.Sprintf(" %d", cdNb),
					fmt.Sprintf("% d", posVec[0]),
					fmt.Sprintf("% d", posVec[1]),
					fmt.Sprintf("% d", posVec[2]),
					fmt.Sprintf("% d", ds),
				})
				write(csvWriter, []string{
					fmt.Sprintf("-%d", cdNb),
					fmt.Sprintf("% d", negVec[0]),
					fmt.Sprintf("% d", negVec[1]),
					fmt.Sprintf("% d", negVec[2]),
					fmt.Sprintf("% d", ds),
				})
				writeNextString(txtFile, fmt.Sprintf("%s: %v = %d\n", v.GetName(), posVec, ds))
				negCD := AllConnectionsPossible[negVec]
				writeNextString(txtFile, fmt.Sprintf("%s: %v = %d\n", negCD.GetName(), negVec, ds))
				break
			}
		}
	}
}

// Write all the points, base vector used, DS and connection details used from1 T=0 to T=X when transitioning from Trio Index 0 to 4 back and forth
func Write0To4TimeFlow() {
	changeToDocsGeneratedDir()

	// Start from origin with growth context type 2 index 0
	ctx := &GrowthContext{&Origin, 2, 0, false, 0}
	untilTime := TickTime(8)

	txtFile, err := os.Create(fmt.Sprintf("Center_%03d_%03d_%03d_Growth_%d_%d_Time_%03d.txt",
		ctx.center[0], ctx.center[1], ctx.center[2],
		ctx.permutationType, ctx.permutationIndex, untilTime))
	if err != nil {
		log.Fatal("Cannot create text file", err)
	}

	collectFlow(ctx, untilTime, func(pointMap *map[Point]*PointState, time TickTime) {
		WriteCurrentPointsToFile(txtFile, time, pointMap)
	})
}

func WriteCurrentPointsToFile(txtFile *os.File, time TickTime, allPoints *map[Point]*PointState) {
	mainPoints, currentPoints := extractMainAndOtherPoints(allPoints, time)

	writeNextString(txtFile, "\n**************************************************\n")
	writeNextString(txtFile, fmt.Sprintf("Time: %4d       %4d       %4d\n######  MAIN POINTS: %4d #######", time, time, time, len(mainPoints)))
	for i, p := range mainPoints {
		if i%4 == 0 {
			writeNextString(txtFile, "\n")
		}
		pState := (*allPoints)[p]
		writeNextString(txtFile, fmt.Sprintf("%3d - %d: %2d, %2d, %2d <= %s | ", pState.globalIdx, pState.trioIndex, p[0], p[1], p[2], pState.GetFromString()))
	}
	writeNextString(txtFile, fmt.Sprintf("\n###### OTHER POINTS: %4d #######", len(currentPoints)))
	for i, p := range currentPoints {
		if i%6 == 0 {
			writeNextString(txtFile, "\n")
		}
		pState := (*allPoints)[p]
		writeNextString(txtFile, fmt.Sprintf("%3d: %2d, %2d, %2d <= %s | ", pState.globalIdx, p[0], p[1], p[2], pState.GetFromString()))
	}
}

func GenerateDataTimeFlow0() {
	changeToDocsDataDir()

	// Start from origin with growth context type 2 index 0
	ctx := &GrowthContext{&Origin, 2, 0, false, 0}
	untilTime := TickTime(30)

	binFile, err := os.Create(fmt.Sprintf("Growth_%d_%d_Time_%03d.data",
		ctx.permutationType, ctx.permutationIndex, untilTime))
	if err != nil {
		log.Fatal("Cannot create bin data file", err)
	}

	collectFlow(ctx, untilTime, func(pointMap *map[Point]*PointState, time TickTime) {
		WriteCurrentPointsDataToFile(binFile, time, pointMap)
	})
}

func WriteCurrentPointsDataToFile(file *os.File, time TickTime, allPoints *map[Point]*PointState) {
	for _, ps := range *allPoints {
		if ps.creationTime == time {
			writeNextString(file, ps.ToDataString())
			writeNextString(file, "\n")
		}
	}
}
