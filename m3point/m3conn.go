package m3point

import "fmt"

var AllConnectionsPossible map[Point]ConnectionDetails
var AllConnectionsIds map[int8]ConnectionDetails

type ConnectionDetails struct {
	Id     int8
	Vector Point
	ConnDS int64
}

var EmptyConnDetails = ConnectionDetails{0, Origin, 0,}

/***************************************************************/
// ConnectionDetails Functions
/***************************************************************/

func (cd ConnectionDetails) GetIntId() int8 {
	return cd.Id
}

func (cd ConnectionDetails) GetPosIntId() int8 {
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

func (cd ConnectionDetails) GetName() string {
	if cd.Id < 0 {
		return fmt.Sprintf("CN%02d", -cd.Id)
	} else {
		return fmt.Sprintf("CP%02d", cd.Id)
	}
}

func initConnectionDetails() uint8 {
	connMap := make(map[Point]ConnectionDetails)
	// Going through all Trio and all combination of Trio, to aggregate connection details
	for _, tr := range AllBaseTrio {
		for _, vec := range tr {
			addConnDetail(&connMap, vec)
		}
		for _, tB := range AllBaseTrio {
			connectingVectors := GetNonBaseConnections(tr, tB)
			for _, conn := range connectingVectors {
				addConnDetail(&connMap, conn)
			}
		}
	}
	Log.Info("Number of connection details created", len(connMap))
	nbConnDetails := int8(len(connMap) / 2)

	// Reordering connection details number by size, and x, y, z
	AllConnectionsIds = make(map[int8]ConnectionDetails)
	for currentConnNumber := int8(1); currentConnNumber <= nbConnDetails; currentConnNumber++ {
		smallestCD := ConnectionDetails{0, Origin, 0}
		for _, cd := range connMap {
			if cd.Id == int8(0) {
				if smallestCD.Vector == Origin {
					smallestCD = cd
				} else if smallestCD.ConnDS > cd.ConnDS {
					smallestCD = cd
				} else if smallestCD.ConnDS == cd.ConnDS {
					if Abs64(cd.Vector.X()) > Abs64(smallestCD.Vector.X()) {
						smallestCD = cd
					} else if Abs64(cd.Vector.X()) == Abs64(smallestCD.Vector.X()) && Abs64(cd.Vector.Y()) > Abs64(smallestCD.Vector.Y()) {
						smallestCD = cd
					} else if Abs64(cd.Vector.X()) == Abs64(smallestCD.Vector.X()) && Abs64(cd.Vector.Y()) == Abs64(smallestCD.Vector.Y()) && Abs64(cd.Vector.Z()) > Abs64(smallestCD.Vector.Z()) {
						smallestCD = cd
					}
				}
			}
		}
		vec1 := smallestCD.Vector
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

		smallestCD = connMap[posVec]
		smallestCD.Id = currentConnNumber
		connMap[smallestCD.Vector] = smallestCD

		negSmallestCD := connMap[negVec]
		negSmallestCD.Id = -currentConnNumber
		connMap[negVec] = negSmallestCD

		AllConnectionsIds[smallestCD.GetIntId()] = smallestCD
		AllConnectionsIds[negSmallestCD.GetIntId()] = negSmallestCD
	}
	AllConnectionsPossible = connMap

	return uint8(nbConnDetails)
}

func addConnDetail(connMap *map[Point]ConnectionDetails, connVector Point) {
	ds := connVector.DistanceSquared()
	if ds == 0 {
		panic("zero vector cannot be a connection")
	}
	if !(ds == 1 || ds == 2 || ds == 3 || ds == 5) {
		panic(fmt.Sprintf("vector %v of ds=%d cannot be a connection", connVector, ds))
	}
	_, ok := (*connMap)[connVector]
	if !ok {
		// Consider negative if X, then Y, then Z is neg
		// If vector negative need to flip
		posVec := connVector
		negVec := connVector.Neg()
		if connVector.X() < 0 {
			// flip
			posVec = negVec
			negVec = connVector
		} else if connVector.X() == 0 {
			if connVector.Y() < 0 {
				// flip
				posVec = negVec
				negVec = connVector
			} else if connVector.Y() == 0 {
				if connVector.Z() < 0 {
					// flip
					posVec = negVec
					negVec = connVector
				}
			}
		}
		posConnDetails := ConnectionDetails{0, posVec, ds,}
		negConnDetails := ConnectionDetails{0, negVec, ds,}
		(*connMap)[posVec] = posConnDetails
		(*connMap)[negVec] = negConnDetails
	}
}

func GetConnDetailsById(id int8) ConnectionDetails {
	return AllConnectionsIds[id]
}

func GetConnDetailsByPoints(p1, p2 Point) ConnectionDetails {
	vector := MakeVector(p1, p2)
	cd, ok := AllConnectionsPossible[vector]
	if !ok {
		Log.Error("Trying to connect to Pos", p1, p2, "that cannot be connected with any known connection details")
		return EmptyConnDetails
	}
	return cd
}

