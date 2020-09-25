package m3space

import (
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3point"
	"sort"
	"sync"
)

var LogStat = m3util.NewStatLogger("m3stat", m3util.INFO)

type ThreeIds [3]EventId

var NilThreeIds = ThreeIds{NilEvent, NilEvent, NilEvent}

type ForwardResult struct {
	pointsPerThreeIds map[ThreeIds][]m3point.Point
}

func MakeForwardResult() *ForwardResult {
	res := ForwardResult{make(map[ThreeIds][]m3point.Point, 16)}
	return &res
}

func (fr *ForwardResult) addPoint(tIds []ThreeIds, p m3point.Point) {
	for _, tid := range tIds {
		pList, ok := fr.pointsPerThreeIds[tid]
		if !ok {
			pList = make([]m3point.Point, 1)
			pList[0] = p
		} else {
			pList = append(pList, p)
		}
		fr.pointsPerThreeIds[tid] = pList
	}

}

func (space *Space) ForwardTime() *ForwardResult {
	nbLatest := 0
	expectedLatestNodes := 0
	for _, evt := range space.events {
		if evt != nil {
			nbLatest += evt.PathContext.GetNumberOfOpenNodes()
			expectedLatestNodes += evt.PathContext.PredictedNextOpenNodesLen()
		}
	}
	space.latestNodes = make([]Node, 0, expectedLatestNodes)
	if Log.IsInfo() {
		Log.Infof("Stepping up to %d: %d events, %d actNodes, %d actConn, %d latestOpen, %d expectedOpen",
			space.CurrentTime+1, space.GetNbEvents(), len(space.ActiveNodes), len(space.ActiveLinks), nbLatest, expectedLatestNodes)
	}
	LogStat.Infof("%4d: %d: %d: %d: %d: %d",
		space.CurrentTime, space.GetNbEvents(), len(space.ActiveNodes), len(space.ActiveLinks), nbLatest, expectedLatestNodes)

	wg := sync.WaitGroup{}
	for _, evt := range space.events {
		if evt != nil {
			wg.Add(1)
			go evt.moveToNext(&wg)
		}
	}
	wg.Wait()

	space.CurrentTime++

	for _, evt := range space.events {
		if evt != nil {
			for _, opn := range evt.PathContext.GetAllOpenPathNodes() {
				// TODO: Remove PathNodeMap need. Use DB
				evt.pathNodeMap.AddPathNode(opn)
			}
		}
	}

	newActiveNodes := NodeList(make([]Node, 0, expectedLatestNodes))
	newActiveLinks := NodeLinkList(make([]NodeLink, 0, expectedLatestNodes))
	res := MakeForwardResult()
	for _, n := range space.latestNodes {
		space.populateActiveNodesAndLinks(n, res, &newActiveNodes, &newActiveLinks)
	}
	for _, n := range space.ActiveNodes {
		space.populateActiveNodesAndLinks(n, res, &newActiveNodes, &newActiveLinks)
	}
	space.ActiveNodes = newActiveNodes
	space.ActiveLinks = newActiveLinks

	return res
}

func (space *Space) populateActiveNodesAndLinks(n Node, res *ForwardResult, nodes *NodeList, links *NodeLinkList) {
	nbActive := n.GetNbActiveEvents(space)
	point := n.GetPoint()
	if point != nil && point.IsMainPoint() && nbActive >= m3point.THREE {
		tIds := MakeThreeIds(n.GetActiveEventIds(space))
		res.addPoint(tIds, *point)
	}
	if nbActive > 0 {
		nodes.addNode(n)
		links.addAll(n.GetActiveLinks(space))
	}
}

func (evt *Event) moveToNext(wg *sync.WaitGroup) {
	evt.PathContext.MoveToNextNodes()
	wg.Done()
}

func SortEventIDs(ids *[]EventId) {
	sort.Slice(*ids, func(i, j int) bool {
		return (*ids)[i] < (*ids)[j]
	})
}

func MakeThreeIds(ids []EventId) []ThreeIds {
	SortEventIDs(&ids)
	if len(ids) == 3 {
		return []ThreeIds{{ids[0], ids[1], ids[2]}}
	} else if len(ids) == 4 {
		return []ThreeIds{
			{ids[0], ids[1], ids[2]},
			{ids[0], ids[2], ids[3]},
			{ids[0], ids[1], ids[3]},
			{ids[1], ids[2], ids[3]},
		}
	}
	Log.Fatal("WHAT!")
	return nil
}

func (tIds ThreeIds) contains(id EventId) bool {
	for _, tid := range tIds {
		if tid == id {
			return true
		}
	}
	return false
}
