package m3space

import "fmt"

type Trio [3]Point

var AllBaseTrio [8]Trio

var AllConnectionsPossible map[Point]ConnectionDetails

type ConnectionDetails struct {
	Vector     Point
	ConnNumber uint8
	ConnNeg    bool
}

var EmptyConnDetails = ConnectionDetails{Origin, 0, false}

func init() {
	// Initial Trio 0
	AllBaseTrio[0] = MakeBaseConnectingVectorsTrio([3]Point{{1, 1, 0}, {-1, 0, -1}, {0, -1, 1}})
	// Initial Trio 0 prime
	AllBaseTrio[4] = MakeBaseConnectingVectorsTrio([3]Point{{1, 1, 0}, {-1, 0, 1}, {0, -1, -1}})
	for i := 1; i < 4; i++ {
		AllBaseTrio[i] = AllBaseTrio[i-1].PlusX()
		AllBaseTrio[i+4] = AllBaseTrio[i+4-1].PlusX()
	}
}

func (t Trio) PlusX() Trio {
	return MakeBaseConnectingVectorsTrio([3]Point{t[0].PlusX(),t[1].PlusX(),t[2].PlusX()})
}

func MakeBaseConnectingVectorsTrio(points [3]Point) Trio {
	res := Trio{}
	// All points should be a connecting vector
	for _, p := range points {
		if !p.IsBaseConnectingVector() {
			fmt.Println("Trying to create a base trio out of non base vector!",p)
			return res
		}
	}
	for _, p := range points {
		for i:=0;i<3;i++ {
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
	fmt.Println("Impossible! For all trio there should be a +X vector")
	return Origin
}

func (t Trio) getMinusXVector() Point {
	for _, vec := range t {
		if vec.X() == -1 {
			return vec
		}
	}
	fmt.Println("Impossible! For all trio there should be a -X vector")
	return Origin
}

func (t Trio) getPlusYVector() Point {
	for _, vec := range t {
		if vec.Y() == 1 {
			return vec
		}
	}
	fmt.Println("Impossible! For all trio there should be a +Y vector")
	return Origin
}

func (t Trio) getMinusYVector() Point {
	for _, vec := range t {
		if vec.Y() == -1 {
			return vec
		}
	}
	fmt.Println("Impossible! For all trio there should be a -Y vector")
	return Origin
}

func (t Trio) getPlusZVector() Point {
	for _, vec := range t {
		if vec.Z() == 1 {
			return vec
		}
	}
	fmt.Println("Impossible! For all trio there should be a +Z vector")
	return Origin
}

func (t Trio) getMinusZVector() Point {
	for _, vec := range t {
		if vec.Z() == -1 {
			return vec
		}
	}
	fmt.Println("Impossible! For all trio there should be a -Z vector")
	return Origin
}

func InitConnectionDetails() uint8 {
	connNumber := uint8(0)
	AllConnectionsPossible = make(map[Point]ConnectionDetails)

	// All Trio and all combi of Trio
	for _, tr := range AllBaseTrio {
		for _, vec := range tr {
			addConnDetail(&connNumber, vec)
		}
		for _, tB := range AllBaseTrio {
			conns := GetNonBaseConnections(tr,tB)
			for _, conn := range conns {
				addConnDetail(&connNumber, conn)
			}
		}
	}
	fmt.Println("Number of connection details created ", connNumber)
	return connNumber
}

func addConnDetail(connNumber *uint8, connVector Point) {
	cd, ok := AllConnectionsPossible[connVector]
	if !ok {
		cd, ok = AllConnectionsPossible[connVector.Neg()]
		if !ok {
			cd = ConnectionDetails{connVector, *connNumber, false,}
			*connNumber++
		} else {
			cd = ConnectionDetails{connVector, cd.ConnNumber, true,}
		}
		AllConnectionsPossible[connVector] = cd
	}
}

func GetConnectionDetails(p1, p2 Point) ConnectionDetails {
	vector := p2.Sub(p1)
	cd, ok := AllConnectionsPossible[vector]
	if !ok {
		fmt.Println("Trying to connect to Pos",p1,p2,"that cannot be connected with any known connection details")
		return EmptyConnDetails
	}
	return cd
}
