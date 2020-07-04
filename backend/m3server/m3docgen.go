package m3server

import (
	"encoding/csv"
	"fmt"
	"github.com/freddy33/qsm-go/backend/m3db"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/freddy33/qsm-go/utils/m3util"
	"log"
)

func GenerateTextFilesEnv(env m3util.QsmEnvironment) {
	InitializePointDBEnv(env.(*m3db.QsmDbEnvironment), false)
	genDoc := m3util.GetGenDocDir()

	ppd, _ := getServerPointPackData(env)
	ppd.writeAllTrioDetailsTable(genDoc)
	ppd.writeAllTrioPermutationsTable(genDoc)
	ppd.writeTrioConnectionsTable(genDoc)
	ppd.writeAllConnectionDetails(genDoc)
}

type Int2 struct {
	a, b int
}

// Return the kind of connection between 2 trios depending of the distance square values
// A3 => All connections have a DS of 3
// A5 => All connections have a DS of 5
// X135 => All DS present 1, 3 and 5
// G13 => 1 and 3 are present but no DS 5 (The type we use)
func GetTrioConnType(conns [6]m3point.Point) string {
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
	log.Fatalf("trio connection list inconsistent got 1=%t, 3=%t, 5=%t", has1, has3, has5)
	return "WRONG"
}

func (ppd *PointPackData) GetTrioTransitionTableTxt() map[Int2][7]string {
	result := make(map[Int2][7]string, 8*8)
	for a, tA := range allBaseTrio {
		for b, tB := range allBaseTrio {
			txtOut := [7]string{}
			conns := GetNonBaseConnections(tA, tB)
			txtOut[0] = GetTrioConnType(conns)
			for i, conn := range conns {
				cd := ppd.GetConnDetailsByVector(conn)
				// Total size 18
				txtOut[i+1] = fmt.Sprintf("%v %s", conn, cd.String())
			}
			result[Int2{a, b}] = txtOut
		}
	}
	return result
}

func GetTrioTransitionTableCsv() [][]string {
	csvOutput := make([][]string, 8*8)
	for a, tA := range allBaseTrio {
		for b, tB := range allBaseTrio {
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

func (ppd *PointPackData) GetTrioTableCsv() [][]string {
	nbColumns := 5
	nbRowsPerTrio := 4
	csvOutput := make([][]string, len(allBaseTrio)*nbColumns)
	for a, trio := range allBaseTrio {
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
			csvOutput[lineNb][columnNb] = ppd.GetConnDetailsByVector(bv).String()
		}
	}
	return csvOutput
}

// Write all the 8x8 connections possible for all trio in text and CSV files, and classify the connections size DS
func (ppd *PointPackData) writeTrioConnectionsTable(dir string) {
	txtFile := m3util.CreateFile(dir,"TrioConnectionsTable.txt")
	csvFile := m3util.CreateFile(dir,"TrioConnectionsTable.csv")
	defer m3util.CloseFile(txtFile)
	defer m3util.CloseFile(csvFile)

	csvWriter := csv.NewWriter(csvFile)
	m3util.WriteAll(csvWriter, GetTrioTransitionTableCsv())
	csvWriter.Flush()

	txtOutputs := ppd.GetTrioTransitionTableTxt()
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

func (ppd *PointPackData) writeAllTrioDetailsTable(dir string) {
	txtFile := m3util.CreateFile(dir, "AllTrioTable.txt")
	csvFile := m3util.CreateFile(dir,"AllTrioTable.csv")
	defer m3util.CloseFile(txtFile)
	defer m3util.CloseFile(csvFile)

	csvWriter := csv.NewWriter(csvFile)
	m3util.WriteAll(csvWriter, ppd.GetTrioTableCsv())
	for _, td := range ppd.AllTrioDetails {
		m3util.WriteNextString(txtFile, fmt.Sprintf("%s: %v %s\n", td.GetId().String(), td.Conns[0].Vector, td.Conns[0].String()))
		m3util.WriteNextString(txtFile, fmt.Sprintf("      %v %s\n", td.Conns[1].Vector, td.Conns[1].String()))
		m3util.WriteNextString(txtFile, fmt.Sprintf("      %v %s\n", td.Conns[2].Vector, td.Conns[2].String()))
		m3util.WriteNextString(txtFile, "\n")
	}
}

func (ppd *PointPackData) writeAllTrioPermutationsTable(dir string) {
	txtFile := m3util.CreateFile(dir,"AllTrioPermTable.txt")
	defer m3util.CloseFile(txtFile)

	m3util.WriteNextString(txtFile, "Valid next trio Index permutation 2\n")
	for i, perm := range validNextTrio {
		m3util.WriteNextString(txtFile, fmt.Sprintf("%2d: %v\n", i, perm))
	}
	m3util.WriteNextString(txtFile, "\nAll trio Index permutation 4\n")
	for i, perm := range AllMod4Permutations {
		m3util.WriteNextString(txtFile, fmt.Sprintf("%2d: %v\n", i, perm))
	}
	m3util.WriteNextString(txtFile, "\nAll trio Index permutation 8\n")
	for i, perm := range AllMod8Permutations {
		m3util.WriteNextString(txtFile, fmt.Sprintf("%2d: %v\n", i, perm))
	}
}

func (ppd *PointPackData) getAllConnDetailsByVector() map[m3point.Point]*m3point.ConnectionDetails {
	ppd.CheckConnInitialized()
	return ppd.AllConnectionsByVector
}

// Write all the connection details in text and CSV files
func (ppd *PointPackData) writeAllConnectionDetails(dir string) {
	txtFile := m3util.CreateFile(dir, "AllConnectionDetails.txt")
	csvFile := m3util.CreateFile(dir, "AllConnectionDetails.csv")
	defer m3util.CloseFile(txtFile)
	defer m3util.CloseFile(csvFile)

	allCons := ppd.getAllConnDetailsByVector()
	nbConnDetails := m3point.ConnectionId(len(allCons) / 2)
	csvWriter := csv.NewWriter(csvFile)
	for cdNb := m3point.ConnectionId(1); cdNb <= nbConnDetails; cdNb++ {
		for _, v := range allCons {
			if v.GetId() == cdNb {
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
				m3util.WriteNextString(txtFile, fmt.Sprintf("%s: %v = %d\n", v.String(), posVec, ds))
				negCD := ppd.GetConnDetailsByVector(negVec)
				m3util.WriteNextString(txtFile, fmt.Sprintf("%s: %v = %d\n", negCD.String(), negVec, ds))
				break
			}
		}
	}
}

