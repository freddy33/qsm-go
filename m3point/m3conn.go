package m3point

import (
	"fmt"
	"sort"
)

var allConnectionsByVector map[Point]*ConnectionDetails
var allConnections []*ConnectionDetails

type ConnectionDetails struct {
	Id     int8
	Vector Point
	ConnDS int64
}

var EmptyConnDetails = ConnectionDetails{0, Origin, 0,}

type ByConnVector []*ConnectionDetails
type ByConnId []*ConnectionDetails

func (cds ByConnVector) Len() int      { return len(cds) }
func (cds ByConnVector) Swap(i, j int) { cds[i], cds[j] = cds[j], cds[i] }
func (cds ByConnVector) Less(i, j int) bool {
	cd1 := cds[i]
	cd2 := cds[j]
	dsDiff := cd1.ConnDS - cd2.ConnDS
	if dsDiff == 0 {
		// X < Y < Z
		for c := 0; c < 3; c++ {
			d := Abs64(cd1.Vector[c]) - Abs64(cd2.Vector[c])
			if d != 0 {
				return d > 0
			}
		}
		// All abs value equal the first coord that is positive is less
		for c := 0; c < 3; c++ {
			d := cd1.Vector[c] - cd2.Vector[c]
			if d != 0 {
				return d > 0
			}
		}
	}
	return dsDiff < 0
}

func (cds ByConnId) Len() int      { return len(cds) }
func (cds ByConnId) Swap(i, j int) { cds[i], cds[j] = cds[j], cds[i] }
func (cds ByConnId) Less(i, j int) bool {
	return IsLessConnId(cds[i], cds[j])
}

func IsLessConnId(cd1, cd2 *ConnectionDetails) bool {
	absDiff := cd1.GetPosIntId() - cd2.GetPosIntId()
	if absDiff < 0 {
		return true
	} else if absDiff > 0 {
		return false
	} else {
		return cd1.Id > 0
	}
}

/***************************************************************/
// ConnectionDetails Functions
/***************************************************************/

func GetMaxConnId() int8 {
	// The pos conn Id of the last one
	return allConnections[len(allConnections)-1].GetPosIntId()
}

func (cd *ConnectionDetails) IsValid() bool {
	return cd.Id != 0
}

func (cd *ConnectionDetails) GetIntId() int8 {
	return cd.Id
}

func (cd *ConnectionDetails) GetPosIntId() int8 {
	return Abs8(cd.Id)
}

func (cd *ConnectionDetails) DistanceSquared() int64 {
	absId := Abs8(cd.Id)
	if absId <= 3 {
		return int64(1)
	} else if absId <= 9 {
		return int64(2)
	} else if absId <= 13 {
		return int64(3)
	} else if absId <= 25 {
		return int64(5)
	}
	Log.Fatalf("abs Id of connection details %v invalid", cd)
	return int64(0)
}

func (cd *ConnectionDetails) String() string {
	return GetConnectionName(cd.Id)
}

func GetConnectionName(connId int8) string {
	if connId < 0 {
		return fmt.Sprintf("CN%02d", -connId)
	} else {
		return fmt.Sprintf("CP%02d", connId)
	}
}

func initConnectionDetails() uint8 {
	connMap := make(map[Point]*ConnectionDetails)
	// Going through all Trio and all combination of Trio, to aggregate connection details
	for _, tr := range allBaseTrio {
		for _, vec := range tr {
			addConnDetail(&connMap, vec)
		}
		for _, tB := range allBaseTrio {
			connectingVectors := GetNonBaseConnections(tr, tB)
			for _, conn := range connectingVectors {
				addConnDetail(&connMap, conn)
			}
		}
	}
	Log.Debug("Number of connection details created", len(connMap))
	nbConnDetails := int8(len(connMap) / 2)

	// Reordering connection details number by size, and x, y, z
	allConnections = make([]*ConnectionDetails, len(connMap))
	idx := 0
	for _, cd := range connMap {
		allConnections[idx] = cd
		idx++
	}
	sort.Sort(ByConnVector(allConnections))

	currentConnNumber := int8(1)
	for _, cd := range allConnections {
		if cd.Id == 0 {
			vec1 := cd.Vector
			vec2 := vec1.Neg()
			var posVec, negVec Point
			// first one with non 0 pos coord
			for _, c := range vec1 {
				if c > 0 {
					posVec = vec1
					negVec = vec2
					break
				} else if c < 0 {
					posVec = vec2
					negVec = vec1
					break
				}
			}
			posCD := connMap[posVec]
			posCD.Id = currentConnNumber
			negCD := connMap[negVec]
			negCD.Id = -currentConnNumber
			currentConnNumber++
		}
	}
	sort.Sort(ByConnId(allConnections))
	allConnectionsByVector = connMap

	return uint8(nbConnDetails)
}

func addConnDetail(connMap *map[Point]*ConnectionDetails, connVector Point) {
	ds := connVector.DistanceSquared()
	if ds == 0 {
		panic("zero vector cannot be a connection")
	}
	if !(ds == 1 || ds == 2 || ds == 3 || ds == 5) {
		panic(fmt.Sprintf("vector %v of ds=%d cannot be a connection", connVector, ds))
	}
	_, ok := (*connMap)[connVector]
	if !ok {
		// Add both pos and neg
		posVec := connVector
		negVec := connVector.Neg()
		posConnDetails := ConnectionDetails{0, posVec, ds,}
		negConnDetails := ConnectionDetails{0, negVec, ds,}
		(*connMap)[posVec] = &posConnDetails
		(*connMap)[negVec] = &negConnDetails
	}
}

func GetConnDetailsById(id int8) *ConnectionDetails {
	if id > 0 {
		return allConnections[2*id-2]
	} else {
		return allConnections[-2*id-1]
	}
}

func GetConnDetailsByPoints(p1, p2 Point) *ConnectionDetails {
	vector := MakeVector(p1, p2)
	cd, ok := allConnectionsByVector[vector]
	if !ok {
		Log.Error("Trying to connect to Pos", p1, p2, "that cannot be connected with any known connection details")
		return &EmptyConnDetails
	}
	return cd
}

func GetConnDetailsByVector(vector Point) *ConnectionDetails {
	cd, ok := allConnectionsByVector[vector]
	if !ok {
		Log.Error("Vector", vector, "is not a known connection details")
		return &EmptyConnDetails
	}
	return cd
}
