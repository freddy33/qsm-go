package m3point

import (
	"encoding/csv"
	"fmt"
	"github.com/freddy33/qsm-go/m3util"
	"log"
	"os"
)

func WriteAllTables() {
	m3util.ChangeToDocsGeneratedDir()
	writeAllTrioDetailsTable()
	writeAllTrioPermutationsTable()
	writeTrioConnectionsTable()
	writeAllConnectionDetails()
	writeAllTrioDetailsLinks()
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
	defer m3util.CloseFile(txtFile)
	defer m3util.CloseFile(csvFile)

	csvWriter := csv.NewWriter(csvFile)
	m3util.WriteAll(csvWriter, GetTrioTransitionTableCsv())
	csvWriter.Flush()

	txtOutputs := GetTrioTransitionTableTxt()
	for a := 0; a < 8; a++ {
		for b := 0; b < 8; b++ {
			out := txtOutputs[Int2{a, b}]
			if b == 7 {
				m3util.WriteNextString(txtFile, fmt.Sprintf("%d, %d %s", a, b, out[0]))
			} else {
				m3util.WriteNextString(txtFile, fmt.Sprintf("%d, %d %s            ", a, b, out[0]))
			}
		}
		m3util.WriteNextString(txtFile, "\n")
		for i := 0; i < 6; i++ {
			for b := 0; b < 8; b++ {
				out := txtOutputs[Int2{a, b}]
				// this is 18 chars
				m3util.WriteNextString(txtFile, out[i+1])
				if b != 7 {
					m3util.WriteNextString(txtFile, "  ")
				}
			}
			m3util.WriteNextString(txtFile, "\n")
		}
		m3util.WriteNextString(txtFile, "\n")
	}
}

func writeAllTrioDetailsTable() {
	txtFile, err := os.Create("AllTrioTable.txt")
	if err != nil {
		log.Fatal("Cannot create text file", err)
	}
	csvFile, err := os.Create("AllTrioTable.csv")
	if err != nil {
		log.Fatal("Cannot create csv file", err)
	}
	defer m3util.CloseFile(txtFile)
	defer m3util.CloseFile(csvFile)

	csvWriter := csv.NewWriter(csvFile)
	m3util.WriteAll(csvWriter, GetTrioTableCsv())
	for a, td := range AllTrioDetails {
		m3util.WriteNextString(txtFile, fmt.Sprintf("T%03d: %v %s\n", a, td.conns[0].Vector, td.conns[0].GetName()))
		m3util.WriteNextString(txtFile, fmt.Sprintf("      %v %s\n", td.conns[1].Vector, td.conns[1].GetName()))
		m3util.WriteNextString(txtFile, fmt.Sprintf("      %v %s\n", td.conns[2].Vector, td.conns[2].GetName()))
		m3util.WriteNextString(txtFile, "\n")
	}
}

func writeAllTrioDetailsLinks() {
	txtFile, err := os.Create("AllTrioLinks.txt")
	if err != nil {
		log.Fatal("Cannot create text file", err)
	}
	defer m3util.CloseFile(txtFile)

	for a, td := range AllTrioDetails {
		m3util.WriteNextString(txtFile, fmt.Sprintf("T%03d: %s\n", a, td.Links.String()))
	}
}

func writeAllTrioPermutationsTable() {
	txtFile, err := os.Create("AllTrioPermTable.txt")
	if err != nil {
		log.Fatal("Cannot create text file", err)
	}
	defer m3util.CloseFile(txtFile)

	m3util.WriteNextString(txtFile, "Valid next Trio Index permutation 2\n")
	for i, perm := range ValidNextTrio {
		m3util.WriteNextString(txtFile, fmt.Sprintf("%2d: %v\n", i, perm))
	}
	m3util.WriteNextString(txtFile, "\nAll Trio Index permutation 4\n")
	for i, perm := range AllMod4Permutations {
		m3util.WriteNextString(txtFile, fmt.Sprintf("%2d: %v\n", i, perm))
	}
	m3util.WriteNextString(txtFile, "\nAll Trio Index permutation 8\n")
	for i, perm := range AllMod8Permutations {
		m3util.WriteNextString(txtFile, fmt.Sprintf("%2d: %v\n", i, perm))
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
	defer m3util.CloseFile(txtFile)
	defer m3util.CloseFile(csvFile)

	nbConnDetails := int8(len(AllConnectionsPossible) / 2)
	csvWriter := csv.NewWriter(csvFile)
	for cdNb := int8(1); cdNb <= nbConnDetails; cdNb++ {
		for _, v := range AllConnectionsPossible {
			if v.GetIntId() == cdNb {
				ds := v.ConnDS
				posVec := v.Vector
				negVec := v.Vector.Neg()
				m3util.Write(csvWriter, []string{
					fmt.Sprintf(" %d", cdNb),
					fmt.Sprintf("% d", posVec[0]),
					fmt.Sprintf("% d", posVec[1]),
					fmt.Sprintf("% d", posVec[2]),
					fmt.Sprintf("% d", ds),
				})
				m3util.Write(csvWriter, []string{
					fmt.Sprintf("-%d", cdNb),
					fmt.Sprintf("% d", negVec[0]),
					fmt.Sprintf("% d", negVec[1]),
					fmt.Sprintf("% d", negVec[2]),
					fmt.Sprintf("% d", ds),
				})
				m3util.WriteNextString(txtFile, fmt.Sprintf("%s: %v = %d\n", v.GetName(), posVec, ds))
				negCD := AllConnectionsPossible[negVec]
				m3util.WriteNextString(txtFile, fmt.Sprintf("%s: %v = %d\n", negCD.GetName(), negVec, ds))
				break
			}
		}
	}
}

