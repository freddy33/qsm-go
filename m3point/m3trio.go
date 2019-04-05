package m3point

import (
	"fmt"
)

type Trio [3]Point

var AllBaseTrio [8]Trio
var AllTrio []Trio

var ValidNextTrio [12][2]int

var AllMod4Permutations [12][4]int

var AllMod8Permutations [12][8]int

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
	for i := 1; i < 4; i++ {
		AllBaseTrio[i] = AllBaseTrio[i-1].PlusX()
	}
	// Initial Trio 0 prime
	for i := 0; i < 4; i++ {
		AllBaseTrio[i+4] = AllBaseTrio[i].Neg()
	}

	initValidTrios()
	initMod4Permutations()
	initMod8Permutations()
	initConnectionDetails()
}

func initValidTrios() {
	// Valid next trio are all but prime
	idx := 0
	for i := 0; i < 4; i++ {
		for j := 4; j < 8; j++ {
			// j index cannot be the prime (neg) trio
			if !isPrime(i, j) {
				ValidNextTrio[idx] = [2]int{i, j}
				idx++
			}
		}
	}
}

type PermBuilder struct {
	size      int
	colIdx    int
	collector [][]int
}

func samePermutation(p1, p2 []int) bool {
	if len(p1) != len(p2) {
		Log.Fatalf("cannot test 2 permutation of different sizes %v %v", p1, p2)
	}
	permSize := len(p1)
	// Index in p2 of first entry in p1
	idx0 := -1
	for idx := 0; idx < permSize; idx++ {
		if p2[idx] == p1[0] {
			idx0 = idx
			break
		}
	}
	if idx0 == -1 {
		// did not find p1[0] so not same permutation
		return false
	}
	// Now they are same permutation if translation index of idx0 get same values
	for idx := 0; idx < permSize; idx++ {
		if p2[(idx0+idx)%permSize] != p1[idx] {
			// just one failure means doom
			return false
		}
	}
	return true
}

func (p *PermBuilder) fill(pos int, current []int) {
	if pos == p.size {
		exists := false
		for i := 0; i < p.colIdx; i++ {
			if samePermutation(p.collector[i], current) {
				exists = true
				break
			}
		}
		if !exists {
			p.collector[p.colIdx] = current
			p.colIdx++
		}
		return
	}
	for i := 0; i < 4; i++ {
		// non prime index
		newIndex := i
		if pos%2 == 1 {
			// prime index
			newIndex = i + 4
		}
		usable := true
		if pos-1 >= 0 {
			// any index only once
			for j := 0; j < pos-1; j++ {
				if current[j] == newIndex {
					usable = false
				}
			}
			// Cannot have prime before
			if isPrime(newIndex, current[pos-1]) {
				usable = false
			}
		}
		// If last cannot be prime with first
		if pos+1 == p.size {
			if isPrime(newIndex, current[0]) {
				usable = false
			}
		}
		if usable {
			perm := make([]int, p.size)
			copy(perm, current)
			perm[pos] = newIndex
			p.fill(pos+1, perm)
		}
	}
}

func initMod4Permutations() {
	p := PermBuilder{4, 0, make([][]int, 12)}
	p.fill(0, make([]int, p.size))
	for pIdx := 0; pIdx < len(AllMod4Permutations); pIdx++ {
		for i := 0; i < 4; i++ {
			AllMod4Permutations[pIdx][i] = p.collector[pIdx][i]
		}
	}
}

func initMod8Permutations() {
	p := PermBuilder{8, 0, make([][]int, 12)}
	// In 8 size permutation the first index always 0 since we use all the indexes
	first := make([]int, p.size)
	first[0] = 0
	p.fill(1, first)
	for pIdx := 0; pIdx < len(AllMod8Permutations); pIdx++ {
		for i := 0; i < 8; i++ {
			AllMod8Permutations[pIdx][i] = p.collector[pIdx][i]
		}
	}
}

func isPrime(i1, i2 int) bool {
	return i2-i1 == 4 || i2-i1 == -4
}

/***************************************************************/
// Trio Functions
/***************************************************************/

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

func (t Trio) PlusX() Trio {
	return MakeBaseConnectingVectorsTrio([3]Point{t[0].PlusX(), t[1].PlusX(), t[2].PlusX()})
}

func (t Trio) Neg() Trio {
	return MakeBaseConnectingVectorsTrio([3]Point{t[0].Neg(), t[1].Neg(), t[2].Neg()})
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

// Return the 3 new Trio out of Origin + tA
func GetNextTrios(tA, tB Trio) [3]Trio {
	// 0 z=0 for first element, x connector, y connector
	// 1 y=0 for first element, x connector, z connector
	// 2 x=0 for first element, y connector, z connector
	res := [3]Trio{}

	noZ := tA[0]
	var xConn, yConn, zConn Point
	if noZ.X() > 0 {
		xConn = XFirst.Add(tB.getMinusXVector()).Sub(noZ)
	} else {
		xConn = XFirst.Neg().Add(tB.getPlusXVector()).Sub(noZ)
	}
	if noZ.Y() > 0 {
		yConn = YFirst.Add(tB.getMinusYVector()).Sub(noZ)
	} else {
		yConn = YFirst.Neg().Add(tB.getPlusYVector()).Sub(noZ)
	}
	res[0] = Trio{noZ.Neg(), xConn, yConn}

	noY := tA[1]
	if noY.X() > 0 {
		xConn = XFirst.Add(tB.getMinusXVector()).Sub(noY)
	} else {
		xConn = XFirst.Neg().Add(tB.getPlusXVector()).Sub(noY)
	}
	if noY.Z() > 0 {
		zConn = ZFirst.Add(tB.getMinusZVector()).Sub(noY)
	} else {
		zConn = ZFirst.Neg().Add(tB.getPlusZVector()).Sub(noY)
	}
	res[1] = Trio{noY.Neg(), xConn, zConn}

	noX := tA[2]
	if noX.Y() > 0 {
		yConn = YFirst.Add(tB.getMinusYVector()).Sub(noX)
	} else {
		yConn = YFirst.Neg().Add(tB.getPlusYVector()).Sub(noX)
	}
	if noX.Z() > 0 {
		zConn = ZFirst.Add(tB.getMinusZVector()).Sub(noX)
	} else {
		zConn = ZFirst.Neg().Add(tB.getPlusZVector()).Sub(noX)
	}
	res[2] = Trio{noX.Neg(), yConn, zConn}

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

/***************************************************************/
// ConnectionDetails Functions
/***************************************************************/

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

func GetConnectionDetails(p1, p2 Point) ConnectionDetails {
	vector := MakeVector(p1, p2)
	cd, ok := AllConnectionsPossible[vector]
	if !ok {
		Log.Error("Trying to connect to Pos", p1, p2, "that cannot be connected with any known connection details")
		return EmptyConnDetails
	}
	return cd
}
