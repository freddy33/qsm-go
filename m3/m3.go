package m3

type Point struct {
	X, Y, Z int64
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

var BasePoints = [3]Point{{1, 1, 0}, {0, -1, 1}, {-1, 0, -1}}

func (p Point) Mul(m int64) Point {
	return Point{p.X * m, p.Y * m, p.Z * m}
}

func (p1 Point) Add(p2 Point) Point {
	return Point{p1.X + p2.X, p1.Y + p2.Y, p1.Z + p2.Z}
}

func DS(p1, p2 Point) int64 {
	return (p1.X-p2.X)^2 + (p1.Y-p2.Y)^2 + (p1.Z-p2.Z)^2
}
