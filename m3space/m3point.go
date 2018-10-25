package m3space

import (
	"fmt"
)

type Point [3]int64

var Origin = Point{0, 0, 0}
var XFirst = Point{THREE, 0, 0}
var YFirst = Point{0, THREE, 0}
var ZFirst = Point{0, 0, THREE}
var BasePoints = [4][3]Point{{{1, 1, 0}, {0, -1, 1}, {-1, 0, -1}},}
var NextMapping = [3][4]int{
	{0,1,2,3},
	{0,2,1,3},
	{0,1,3,2},
}
var AllMod4Possible = make(map[Point]int)

func init() {
	for i := 1; i < 4; i++ {
		for j := 0; j < 3; j++ {
			BasePoints[i][j] = BasePoints[i-1][j].PlusX()
		}
	}
	for x := int64(0); x < 4; x++ {
		for y := int64(0); y < 4; y++ {
			for z := int64(0); z < 4; z++ {
				p := Point{3*x, 3*y, 3*z}
				AllMod4Possible[p.GetMod4Point()] = p.CalculateMod4Value()
			}
		}
	}
}

func PosMod4(i int64) int64 {
	return i&0x0000000000000003
}

func (p Point) GetMod4Point() Point {
	if !p.IsMainPoint() {
		panic(fmt.Sprintf("cannot ask for Mod4 on non main point %v!",p))
	}
	return Point{PosMod4(p[0] / 3), PosMod4(p[1] / 3), PosMod4(p[2] / 3)}
}

func (p Point) GetMod4Value() int {
	return AllMod4Possible[p.GetMod4Point()]
}

func (p Point) CalculateMod4Value() int {
	pMod4 := p.GetMod4Point()
	xMod4 := NextMapping[0][int(pMod4[0])]
	// Find in Y line the k matching xMod4
	inYK := -1
	for k := 0; k<4;k++ {
		if NextMapping[1][k] == xMod4 {
			inYK = k
			break
		}
	}
	if inYK == -1 {
		fmt.Println("Something went really wrong trying to find x mod 4",xMod4,"from",p,"pMod4",pMod4)
		return -1
	}
	yMod4 := NextMapping[1][int((pMod4[1]+int64(inYK))%4)]
	// Find in Z line the k matching yMod4
	inZK := -1
	for k := 0; k<4;k++ {
		if NextMapping[2][k] == yMod4 {
			inZK = k
			break
		}
	}
	if inZK == -1 {
		fmt.Println("Something went really wrong trying to find y mod 4",yMod4,"from",p,"pMod4",pMod4)
		return -1
	}
	finalMod4 := NextMapping[2][int((pMod4[2]+int64(inZK))%4)]
	return finalMod4
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

func (p Point) IsMainPoint() bool {
	allDivByThree := true
	for _, c := range p {
		if c%THREE != 0 {
			allDivByThree = false
		}
	}
	return allDivByThree
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
