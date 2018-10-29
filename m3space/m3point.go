package m3space

import (
	"fmt"
)

type Point [3]int64

var Origin = Point{0, 0, 0}
var XFirst = Point{THREE, 0, 0}
var YFirst = Point{0, THREE, 0}
var ZFirst = Point{0, 0, THREE}

func PosMod4(i int64) int64 {
	return i & 0x0000000000000003
}

func PosMod8(i int64) int64 {
	return i & 0x0000000000000007
}

func (p Point) GetMod4Point() Point {
	if !p.IsMainPoint() {
		panic(fmt.Sprintf("cannot ask for Mod4 on non main point %v!", p))
	}
	return Point{PosMod4(p[0] / 3), PosMod4(p[1] / 3), PosMod4(p[2] / 3)}
}

var Mod4Function = "sum" // or "sum"

func (p Point) GetTrio() Trio {
	return AllBaseTrio[AllMod8Rotations[0][int(PosMod8(p[0]/3 + p[1]/3 + p[2]/3))]]
}

func (p Point) GetMod4Value() int {
	switch Mod4Function {
	case "sum":
		return p.calculateMod4ValueBySum()
	case "map":
		return p.calculateMod4ValueByNextMapping()
	default:
		fmt.Println("Mod4 function", Mod4Function, "not supported")
		panic("unsupported Mod4 function")
	}
	return -1
}

func (p Point) calculateMod4ValueBySum() int {
	return int(PosMod4(p[0]/3 + p[1]/3 + p[2]/3))
}

func (p Point) calculateMod4ValueByNextMapping() int {
	pMod4 := p.GetMod4Point()
	xMod4 := NextMapping[0][int(pMod4[0])]
	// Find in Y line the k matching xMod4
	inYK := -1
	for k := 0; k < 4; k++ {
		if NextMapping[1][k] == xMod4 {
			inYK = k
			break
		}
	}
	if inYK == -1 {
		fmt.Println("Something went really wrong trying to find x mod 4", xMod4, "from", p, "pMod4", pMod4)
		return -1
	}
	yMod4 := NextMapping[1][int((pMod4[1]+int64(inYK))%4)]
	// Find in Z line the k matching yMod4
	inZK := -1
	for k := 0; k < 4; k++ {
		if NextMapping[2][k] == yMod4 {
			inZK = k
			break
		}
	}
	if inZK == -1 {
		fmt.Println("Something went really wrong trying to find y mod 4", yMod4, "from", p, "pMod4", pMod4)
		return -1
	}
	finalMod4 := NextMapping[2][int((pMod4[2]+int64(inZK))%4)]
	return finalMod4

}

func (p Point) getNearMainPoint() Point {
	res := Point{}
	for i, c := range p {
		switch c % THREE {
		case 0:
			res[i] = c
		case 1:
			res[i] = c - 1
		case 2:
			res[i] = c + 1
		case -1:
			res[i] = c + 1
		case -2:
			res[i] = c - 1
		}
	}
	return res
}

func getNextPoints(mainPoint Point, cVec Point) [2]Point {
	offset := 0
	result := [2]Point{}

	nextMain := mainPoint
	switch cVec.X() {
	case 0:
		// Nothing out
	case 1:
		nextMain = mainPoint.Add(XFirst)
	case -1:
		nextMain = mainPoint.Sub(XFirst)
	default:
		fmt.Printf("There should not be a connecting vector with x value %d\n", cVec.X())
		return result
	}
	if nextMain != mainPoint {
		// Find the base point on the other side ( the opposite 1 or -1 on X() )
		nextConnectingVectors := nextMain.GetTrio()
		for _, nbp := range nextConnectingVectors {
			if nbp.X() == -cVec.X() {
				result[offset] = nextMain.Add(nbp)
				offset++
			}
		}
	}

	nextMain = mainPoint
	switch cVec.Y() {
	case 0:
		// Nothing out
	case 1:
		nextMain = mainPoint.Add(YFirst)
	case -1:
		nextMain = mainPoint.Sub(YFirst)
	default:
		fmt.Printf("There should not be a connecting vector with y value %d\n", cVec.Y())
	}
	if nextMain != mainPoint {
		// Find the base point on the other side ( the opposite 1 or -1 on Y() )
		nextConnectingVectors := nextMain.GetTrio()
		for _, nbp := range nextConnectingVectors {
			if nbp.Y() == -cVec.Y() {
				result[offset] = nextMain.Add(nbp)
				offset++
			}
		}
	}

	nextMain = mainPoint
	switch cVec.Z() {
	case 0:
		// Nothing out
	case 1:
		nextMain = mainPoint.Add(ZFirst)
	case -1:
		nextMain = mainPoint.Sub(ZFirst)
	default:
		fmt.Printf("There should not be a connecting vector with z value %d\n", cVec.Z())
	}
	if nextMain != mainPoint {
		// Find the base point on the other side ( the opposite 1 or -1 on Z() )
		nextConnectingVectors := nextMain.GetTrio()
		for _, nbp := range nextConnectingVectors {
			if nbp.Z() == -cVec.Z() {
				result[offset] = nextMain.Add(nbp)
				offset++
			}
		}
	}
	return result
}

func Abs(i int64) int64 {
	if i < 0 {
		return -i
	}
	return i
}

func (p Point) X() int64 {
	return p[0]
}

func (p Point) Y() int64 {
	return p[1]
}

func (p Point) Z() int64 {
	return p[2]
}

func (p Point) Mul(m int64) Point {
	return Point{p[0] * m, p[1] * m, p[2] * m}
}

func (p1 Point) Add(p2 Point) Point {
	return Point{p1[0] + p2[0], p1[1] + p2[1], p1[2] + p2[2]}
}

func (p1 Point) Sub(p2 Point) Point {
	return Point{p1[0] - p2[0], p1[1] - p2[1], p1[2] - p2[2]}
}

func (p Point) Neg() Point {
	return Point{-p[0], -p[1], -p[2]}
}

// Positive PI/2 rotation on X
func (p Point) PlusX() Point {
	return Point{p[0], -p[2], p[1]}
}

// Negative PI/2 rotation on X
func (p Point) NegX() Point {
	return Point{p[0], p[2], -p[1]}
}

// Positive PI/2 rotation on Y
func (p Point) PlusY() Point {
	return Point{p[2], p[1], -p[0]}
}

// Negative PI/2 rotation on Y
func (p Point) NegY() Point {
	return Point{-p[2], p[1], p[0]}
}

// Positive PI/2 rotation on Z
func (p Point) PlusZ() Point {
	return Point{-p[1], p[0], p[2]}
}

// Negative PI/2 rotation on X
func (p Point) NegZ() Point {
	return Point{p[1], -p[0], p[2]}
}

func DS(p1, p2 *Point) int64 {
	x := p2.X() - p1.X()
	y := p2.Y() - p1.Y()
	z := p2.Z() - p1.Z()
	return x*x + y*y + z*z
}

func (p Point) DistanceSquared() int64 {
	return p[0]*p[0] + p[1]*p[1] + p[2]*p[2]
}

func (p Point) IsMainPoint() bool {
	allDivByThree := true
	for _, c := range p {
		if c%THREE != 0 {
			allDivByThree = false
		}
	}
	return allDivByThree
}

func (p Point) IsBaseConnectingVector() bool {
	if p.IsOnlyOneAndZero() {
		return p.DistanceSquared() == 2
	}
	return false
}

func (p Point) IsConnectionVector() bool {
	if p.IsOnlyTwoOneAndZero() {
		sizeSquared := p.DistanceSquared()
		return sizeSquared == 1 || sizeSquared == 2 || sizeSquared == 3 || sizeSquared == 5
	}
	return false
}

func (p Point) IsOnlyOneAndZero() bool {
	for _, c := range p {
		if c != 0 && c != 1 && c!= -1 {
			return false
		}
	}
	return true
}

func (p Point) IsOnlyTwoOneAndZero() bool {
	for _, c := range p {
		if c != 0 && c != 1 && c!= -1 && c != 2 && c!= -2 {
			return false
		}
	}
	return true
}

func (p Point) IsBorder(max int64) bool {
	for _, c := range p {
		if c > 0 && c >= max-1 {
			return true
		}
		if c < 0 && c <= -max+1 {
			return true
		}
	}
	return false
}

func (p Point) IsOutBorder(max int64) bool {
	for _, c := range p {
		if c > 0 && c > max {
			return true
		}
		if c < 0 && c < -max {
			return true
		}
	}
	return false
}
