package spacedb

import (
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/freddy33/qsm-go/model/m3space"
	"sort"
)

var LogStat = m3util.NewStatLogger("m3stat", m3util.INFO)

type ThreeIds [3]m3space.EventId

var NilThreeIds = ThreeIds{m3space.NilEvent, m3space.NilEvent, m3space.NilEvent}

type SpaceTimeRuleAnalyzer struct {
	spaceTime         *SpaceTime
	PointsPerThreeIds map[ThreeIds][]m3point.Point
}

func MakeRuleAnalyzer(st *SpaceTime) *SpaceTimeRuleAnalyzer {
	return &SpaceTimeRuleAnalyzer{spaceTime: st,
		PointsPerThreeIds: make(map[ThreeIds][]m3point.Point, 16)}
}

func (fr *SpaceTimeRuleAnalyzer) addPoint(tIds []ThreeIds, p m3point.Point) {
	for _, tid := range tIds {
		pList, ok := fr.PointsPerThreeIds[tid]
		if !ok {
			pList = make([]m3point.Point, 1)
			pList[0] = p
		} else {
			pList = append(pList, p)
		}
		fr.PointsPerThreeIds[tid] = pList
	}

}

func (fr *SpaceTimeRuleAnalyzer) VisitNode(node m3space.SpaceTimeNodeIfc) {
	eventIds := node.GetEventIds()
	if len(eventIds) >= m3point.THREE {
		point, err := node.GetPoint()
		if err != nil {
			Log.Error(err)
			return
		}
		if point.IsMainPoint() {
			tIds := MakeThreeIds(eventIds)
			fr.addPoint(tIds, *point)
		}
	}
}

func SortEventIDs(ids *[]m3space.EventId) {
	sort.Slice(*ids, func(i, j int) bool {
		return (*ids)[i] < (*ids)[j]
	})
}

func MakeThreeIds(ids []m3space.EventId) []ThreeIds {
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

func (tIds ThreeIds) contains(id m3space.EventId) bool {
	for _, tid := range tIds {
		if tid == id {
			return true
		}
	}
	return false
}
