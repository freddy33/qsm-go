package m3point

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3util"
	"math/rand"
	"strconv"
	"time"
)

var Log = m3util.NewLogger("m3point", m3util.INFO)

const (
	// Where the number matters and appear. Remember that 3 is the number!
	THREE = 3
)

type DInt int64

func (D *DInt) String() string {
	return strconv.FormatInt(int64(*D), 10)
}

func (D DInt) MurmurHash() uint32 {
	c64 := uint64(D)
	return MurmurHashUint32([]uint32{uint32(c64 & low), uint32(c64 & high >> 32)})
}

type CInt int32
type Point [3]CInt

var Origin = Point{0, 0, 0}
var XFirst = Point{THREE, 0, 0}
var YFirst = Point{0, THREE, 0}
var ZFirst = Point{0, 0, THREE}

func init() {
	rand.Seed(time.Now().UnixNano())
}

/***************************************************************/
// Util Functions
/***************************************************************/

func AbsCInt(i CInt) CInt {
	if i < 0 {
		return -i
	}
	return i
}

func AbsDInt(i DInt) DInt {
	if i < 0 {
		return -i
	}
	return i
}

func AbsDIntFromC(i CInt) DInt {
	if i < 0 {
		return DInt(-i)
	}
	return DInt(i)
}

func DS(p1, p2 Point) DInt {
	x := p2.X() - p1.X()
	y := p2.Y() - p1.Y()
	z := p2.Z() - p1.Z()
	return DInt(x)*DInt(x) + DInt(y)*DInt(y) + DInt(z)*DInt(z)
}

/***************************************************************/
// Point Functions for ALL points not only nextMainPoint
// TODO: Make MainPoint a type
/***************************************************************/

func MakeVector(p1, p2 Point) Point {
	return p2.Sub(p1)
}

func (p Point) MurmurHash() uint32 {
	return MurmurHashUint32([]uint32{uint32(p[0]), uint32(p[1]), uint32(p[2])})
}

func (p Point) String() string {
	return fmt.Sprintf("[ % d, % d, % d ]", p[0], p[1], p[2])
}

func (p Point) X() CInt {
	return p[0]
}

func (p Point) Y() CInt {
	return p[1]
}

func (p Point) Z() CInt {
	return p[2]
}

func (p Point) Mul(m CInt) Point {
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
func (p Point) RotPlusX() Point {
	return Point{p[0], -p[2], p[1]}
}

// Negative PI/2 rotation on X
func (p Point) RotNegX() Point {
	return Point{p[0], p[2], -p[1]}
}

// Positive PI/2 rotation on Y
func (p Point) RotPlusY() Point {
	return Point{p[2], p[1], -p[0]}
}

// Negative PI/2 rotation on Y
func (p Point) RotNegY() Point {
	return Point{-p[2], p[1], p[0]}
}

// Positive PI/2 rotation on Z
func (p Point) RotPlusZ() Point {
	return Point{-p[1], p[0], p[2]}
}

// Negative PI/2 rotation on X
func (p Point) RotNegZ() Point {
	return Point{p[1], -p[0], p[2]}
}

func (p Point) DistanceSquared() DInt {
	return DInt(p[0])*DInt(p[0]) + DInt(p[1])*DInt(p[1]) + DInt(p[2])*DInt(p[2])
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

func (p Point) SumOfPositiveCoord() DInt {
	res := DInt(0)
	for _, c := range p {
		if c > 0 {
			res += DInt(c)
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

/***************************************************************/
// Utility methods for test
/***************************************************************/

func CreateRandomPoint(max CInt) Point {
	return Point{GetRandomCInt(max), GetRandomCInt(max), GetRandomCInt(max)}
}

func GetRandomCInt(max CInt) CInt {
	r := CInt(rand.Int31n(int32(max)))
	if rand.Float32() < 0.5 {
		return -r
	}
	return r
}
