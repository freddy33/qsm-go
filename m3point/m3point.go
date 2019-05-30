package m3point

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3util"
)

var Log = m3util.NewLogger("m3point", m3util.INFO)

const (
	// Where the number matters and appear. Remember that 3 is the number!
	THREE = 3
)

type Point [3]int64

var Origin = Point{0, 0, 0}
var XFirst = Point{THREE, 0, 0}
var YFirst = Point{0, THREE, 0}
var ZFirst = Point{0, 0, THREE}

/***************************************************************/
// Util Functions
/***************************************************************/

func Abs64(i int64) int64 {
	if i < 0 {
		return -i
	}
	return i
}

func Abs8(i int8) int8 {
	if i < 0 {
		return -i
	}
	return i
}

func DS(p1, p2 Point) int64 {
	x := p2.X() - p1.X()
	y := p2.Y() - p1.Y()
	z := p2.Z() - p1.Z()
	return x*x + y*y + z*z
}

/***************************************************************/
// Point Functions for ALL points not only nextMainPoint
// TODO: Make MainPoint a type
/***************************************************************/

func MakeVector(p1,p2 Point) Point {
	return p2.Sub(p1)
}

func (p Point) String() string {
	return fmt.Sprintf("[ % d, % d, % d ]", p[0], p[1], p[2])
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

func (p Point) Add(p2 Point) Point {
	return Point{p[0] + p2[0], p[1] + p2[1], p[2] + p2[2]}
}

func (p Point) Sub(p2 Point) Point {
	return Point{p[0] - p2[0], p[1] - p2[1], p[2] - p2[2]}
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

func (p Point) SumOfPositiveCoord() int64 {
	res := int64(0)
	for _, c := range p {
		if c > 0 {
			res += c
		}
	}
	return res
}

func (p Point) IsOnlyOneAndZero() bool {
	for _, c := range p {
		if c != 0 && c != 1 && c != -1 {
			return false
		}
	}
	return true
}

func (p Point) IsOnlyTwoOneAndZero() bool {
	for _, c := range p {
		if c != 0 && c != 1 && c != -1 && c != 2 && c != -2 {
			return false
		}
	}
	return true
}

func (p Point) GetNearMainPoint() Point {
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
