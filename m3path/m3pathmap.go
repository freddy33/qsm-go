package m3path

import "github.com/freddy33/qsm-go/m3point"

type PathNodeMap interface {
	GetSize() int
	GetPathNode(p m3point.Point) (PathNode, bool)
	AddPathNode(pathNode PathNode)
}

type SimplePathNodeMap map[m3point.Point]PathNode

func MakeSimplePathNodeMap(initSize int) PathNodeMap {
	res := SimplePathNodeMap(make(map[m3point.Point]PathNode, initSize))
	return &res
}

func (pnm *SimplePathNodeMap) GetSize() int {
	return len(*pnm)
}

func (pnm *SimplePathNodeMap) GetPathNode(p m3point.Point) (PathNode, bool) {
	res, ok := (*pnm)[p]
	return res, ok
}

func (pnm *SimplePathNodeMap) AddPathNode(pathNode PathNode) {
	(*pnm)[pathNode.P()] = pathNode
}
