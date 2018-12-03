package m3space

import "fmt"

type Trio [3]Point

var AllBaseTrio [8]Trio

var ValidNextTrio = [12][2]int{
	{0,4},{0,6},{0,7},
	{1,4},{1,5},{1,7},
	{2,4},{2,5},{2,6},
	{3,5},{3,6},{3,7},
}

var AllMod4Permutations = [12][4]int{
	{0,4,1,7},
	{0,4,2,6},
	{0,6,2,4},
	{0,6,3,7},
	{0,7,1,4},
	{0,7,3,6},
	{1,4,2,5},
	{1,5,2,4},
	{1,5,3,7},
	{1,7,3,5},
	{2,5,3,6},
	{2,6,3,5},
}

var AllMod8Permutations = [12][8]int{
	{0,4,1,5,2,6,3,7},
	{0,4,1,7,3,5,2,6},
	{0,4,2,5,1,7,3,6},
	{0,4,2,6,3,5,1,7},
	{0,6,2,4,1,5,3,7},
	{0,6,2,5,3,7,1,4},
	{0,6,3,5,2,4,1,7},
	{0,6,3,7,1,5,2,4},
	{0,7,1,4,2,5,3,6},
	{0,7,1,5,3,6,2,4},
	{0,7,3,5,1,4,2,6},
	{0,7,3,6,2,5,1,4},
}

var AllConnectionsPossible map[Point]ConnectionDetails
var AllConnectionsIds map[int8]ConnectionDetails

type ConnectionDetails struct {
	Id     int8
	Vector Point
	ConnDS int64
}

var EmptyConnDetails = ConnectionDetails{0, Origin, 0,}

func init() {
	// Initial Trio 0
	AllBaseTrio[0] = MakeBaseConnectingVectorsTrio([3]Point{{1, 1, 0}, {-1, 0, -1}, {0, -1, 1}})
	// Initial Trio 0 prime
	AllBaseTrio[4] = MakeBaseConnectingVectorsTrio([3]Point{{1, 1, 0}, {-1, 0, 1}, {0, -1, -1}})
	for i := 1; i < 4; i++ {
		AllBaseTrio[i] = AllBaseTrio[i-1].PlusX()
		AllBaseTrio[i+4] = AllBaseTrio[i+4-1].PlusX()
	}
	initConnectionDetails()
}

func (cd ConnectionDetails) GetIntId() int8 {
	return cd.Id
}

func (cd ConnectionDetails) GetName() string {
	if cd.Id < 0 {
		return fmt.Sprintf("CN%02d", -cd.Id)
	} else {
		return fmt.Sprintf("CP%02d", cd.Id)
	}
}

func (t Trio) PlusX() Trio {
	return MakeBaseConnectingVectorsTrio([3]Point{t[0].PlusX(), t[1].PlusX(), t[2].PlusX()})
}

func MakeBaseConnectingVectorsTrio(points [3]Point) Trio {
	res := Trio{}
	// All points should be a connecting vector
	for _, p := range points {
		if !p.IsBaseConnectingVector() {
			Log.Error("Trying to create a base trio out of non base vector!", p)
			return res
		}
	}
	for _, p := range points {
		for i := 0; i < 3; i++ {
			if p[i] == 0 {
				res[reverse3Map[i]] = p
			}
		}
	}
	return res
}

// Return the 6 connections possible +X, -X, +Y, -Y, +Z, -Z vectors between 2 Trio
func GetNonBaseConnections(tA, tB Trio) [6]Point {
	res := [6]Point{}
	for _, aVec := range tA {
		switch aVec.X() {
		case 0:
			// Nothing connect
		case 1:
			// This is +X, find the -X on the other side
			bVec := tB.getMinusXVector()
			res[0] = XFirst.Add(bVec).Sub(aVec)
		case -1:
			// This is -X, find the +X on the other side
			bVec := tB.getPlusXVector()
			res[1] = XFirst.Neg().Add(bVec).Sub(aVec)
		}
		switch aVec.Y() {
		case 0:
			// Nothing connect
		case 1:
			// This is +Y, find the -Y on the other side
			bVec := tB.getMinusYVector()
			res[2] = YFirst.Add(bVec).Sub(aVec)
		case -1:
			// This is -Y, find the +Y on the other side
			bVec := tB.getPlusYVector()
			res[3] = YFirst.Neg().Add(bVec).Sub(aVec)
		}
		switch aVec.Z() {
		case 0:
			// Nothing connect
		case 1:
			// This is +Z, find the -Z on the other side
			bVec := tB.getMinusZVector()
			res[4] = ZFirst.Add(bVec).Sub(aVec)
		case -1:
			// This is -Z, find the +Z on the other side
			bVec := tB.getPlusZVector()
			res[5] = ZFirst.Neg().Add(bVec).Sub(aVec)
		}
	}
	return res
}

func (t Trio) getPlusXVector() Point {
	for _, vec := range t {
		if vec.X() == 1 {
			return vec
		}
	}
	Log.Error("Impossible! For all trio there should be a +X vector")
	return Origin
}

func (t Trio) getMinusXVector() Point {
	for _, vec := range t {
		if vec.X() == -1 {
			return vec
		}
	}
	Log.Error("Impossible! For all trio there should be a -X vector")
	return Origin
}

func (t Trio) getPlusYVector() Point {
	for _, vec := range t {
		if vec.Y() == 1 {
			return vec
		}
	}
	Log.Error("Impossible! For all trio there should be a +Y vector")
	return Origin
}

func (t Trio) getMinusYVector() Point {
	for _, vec := range t {
		if vec.Y() == -1 {
			return vec
		}
	}
	Log.Error("Impossible! For all trio there should be a -Y vector")
	return Origin
}

func (t Trio) getPlusZVector() Point {
	for _, vec := range t {
		if vec.Z() == 1 {
			return vec
		}
	}
	Log.Error("Impossible! For all trio there should be a +Z vector")
	return Origin
}

func (t Trio) getMinusZVector() Point {
	for _, vec := range t {
		if vec.Z() == -1 {
			return vec
		}
	}
	Log.Error("Impossible! For all trio there should be a -Z vector")
	return Origin
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
		smallestCD := ConnectionDetails{0,Origin, 0}
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
		smallestCD.Id = currentConnNumber
		connMap[smallestCD.Vector] = smallestCD
		negVec := smallestCD.Vector.Neg()
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

func GetConnectionDetails(p1, p2 Point) ConnectionDetails {
	vector := MakeVector(p1, p2)
	cd, ok := AllConnectionsPossible[vector]
	if !ok {
		Log.Error("Trying to connect to Pos", p1, p2, "that cannot be connected with any known connection details")
		return EmptyConnDetails
	}
	return cd
}
