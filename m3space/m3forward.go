package m3space

import (
	"github.com/freddy33/qsm-go/m3point"
	"github.com/freddy33/qsm-go/m3util"
	"sort"
	"sync"
)

var LogStat = m3util.NewStatLogger("m3stat", m3util.INFO)

type ThreeIds [3]EventID

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
	expectedActiveNodes := 0
	for _, evt := range space.events {
		if evt != nil {
			nbLatest += evt.pathContext.GetNumberOfOpenNodes()
			expectedActiveNodes += evt.pathContext.GetNextOpenNodesLen()
		}
	}
	if Log.IsInfo() {
		Log.Infof("Stepping up to %d: %d events, %d actNodes, %d actConn, %d latestOpen, %d expectedOpen",
			space.currentTime+1, space.GetNbEvents(), len(space.activeNodes), len(space.activeLinks), nbLatest, expectedActiveNodes)
	}
	LogStat.Infof("%4d: %d: %d: %d: %d: %d",
		space.currentTime, space.GetNbEvents(), len(space.activeNodes), len(space.activeLinks), nbLatest, expectedActiveNodes)

	wg := sync.WaitGroup{}
	for _, evt := range space.events {
		if evt != nil {
			wg.Add(1)
			go evt.moveToNext(&wg)
		}
	}
	wg.Wait()

	res := MakeForwardResult()
	for _, n := range space.activeNodes {
		if n.GetNbActiveEvents() > 3 {
			tIds := MakeThreeIds(n.GetActiveEventIds())
			res.addPoint(tIds, *n.GetPoint())
		}
	}

	space.currentTime++

	return res
}

func (evt *Event) moveToNext(wg *sync.WaitGroup) {
	evt.pathContext.MoveToNextNodes()
	wg.Done()
}

func SortEventIDs(ids *[]EventID) {
	sort.Slice(*ids, func(i, j int) bool {
		return (*ids)[i] < (*ids)[j]
	})
}

func MakeThreeIds(ids []EventID) []ThreeIds {
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

func (tIds ThreeIds) contains(id EventID) bool {
	for _, tid := range tIds {
		if tid == id {
			return true
		}
	}
	return false
}
