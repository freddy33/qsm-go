package m3space

type Point [3]int64

func (p Point) X() int64 {
	return p[0]
}

func (p Point) Y() int64 {
	return p[1]
}

func (p Point) Z() int64 {
	return p[2]
}

type Node struct {
	P *Point
	C [3]*Node
}

type TickTime uint64

type Distance uint64

type EventID uint64

type Event struct {
	ID EventID
	N  *Node
	T  TickTime
}

type EventOutgrowth struct {
	Evt   EventID
	D     Distance
	Nodes []*Node
}

type Space struct {
	events  map[EventID]*Event
	current TickTime
}

const THREE = 3

var Origin = Point{0, 0, 0}
var BasePoints = [3]Point{{1, 1, 0}, {0, -1, 1}, {-1, 0, -1}}

func (p Point) Mul(m int64) Point {
	return Point{p[0] * m, p[1] * m, p[2] * m}
}

func (p1 Point) Add(p2 Point) Point {
	return Point{p1[0] + p2[0], p1[1] + p2[1], p1[2] + p2[2]}
}

func (p1 Point) Sub(p2 Point) Point {
	return Point{p1[0] - p2[0], p1[1] - p2[1], p1[2] - p2[2]}
}

func DS(p1, p2 Point) int64 {
	return (p2[0] - p1[0]) ^ 2 + (p2[1] - p1[1]) ^ 2 + (p2[2] - p1[2]) ^ 2
}
