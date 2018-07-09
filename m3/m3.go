package m3

type Point struct {
	X, Y, Z int64
}

type Node struct {
	Point
	C [3]*Node
}

type TickTime uint64

type Distance uint64

type EventID uint64

type Event struct {
	ID EventID
	N Node
	T TickTime
}

type EventOutgrowth struct {
	Evt EventID
	D Distance
	Nodes []Node
}

type Space struct {
	events []Event
	current TickTime
}
