package m3space

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

func init() {
	for i := 1; i < 4; i++ {
		for j := 0; j < 3; j++ {
			BasePoints[i][j] = BasePoints[i-1][j].PlusX()
		}
	}
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
