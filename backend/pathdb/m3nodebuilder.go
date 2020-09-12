package pathdb

import (
	"github.com/freddy33/qsm-go/model/m3path"
	"github.com/freddy33/qsm-go/model/m3point"
)

type OpenNodeBuilder struct {
	pathCtx                        *PathContextDb
	d                              int
	expectedSize                   int
	openNodesMap                   m3path.PathNodeMap
	selectConflict, insertConflict int
}

func createNewNodeBuilder(previous *OpenNodeBuilder) *OpenNodeBuilder {
	res := new(OpenNodeBuilder)
	if previous == nil {
		res.d = 0
		res.expectedSize = 1
		res.openNodesMap = m3path.MakeSimplePathNodeMap(1)
	} else {
		res.pathCtx = previous.pathCtx
		res.d = previous.d + 1
		res.expectedSize = previous.nextOpenNodesLen()
		if res.expectedSize > 32 {
			res.openNodesMap = m3path.MakeHashPathNodeMap(res.expectedSize)
		} else {
			res.openNodesMap = m3path.MakeSimplePathNodeMap(res.expectedSize)
		}
	}
	return res
}

func (onb *OpenNodeBuilder) fillOpenPathNodes() {
	pathCtx := onb.pathCtx
	rows, err := pathCtx.pathNodesTe().Query(SelectPathNodesByCtxAndDistance, pathCtx.id, onb.d)
	if err != nil {
		Log.Fatal(err)
	}
	for rows.Next() {
		pn, err := fetchDbRow(rows)
		if err != nil {
			Log.Fatalf("Could not read row of %s due to %v", PathNodesTable, err)
		} else {
			if pn.pathCtxId != pathCtx.id {
				Log.Fatalf("While retrieving all path nodes got a node with context id %d instead of %d",
					pn.pathCtxId, pathCtx.id)
				return
			}
			pn.pathCtx = pathCtx
			onb.addPathNode(pn)
		}
	}
}

func (onb *OpenNodeBuilder) addPathNode(pn *PathNodeDb) *PathNodeDb {
	res, _ := onb.openNodesMap.AddPathNode(pn)
	return res.(*PathNodeDb)
}

func (onb *OpenNodeBuilder) openNodesSize() int {
	return onb.openNodesMap.Size()
}

func (onb *OpenNodeBuilder) nextOpenNodesLen() int {
	return m3path.CalculatePredictedSize(onb.d, onb.openNodesMap.Size())
}

func (onb *OpenNodeBuilder) clear() {
	// Do not release the first three steps
	if onb.d > 3 {
		onb.openNodesMap.Range(func(point m3point.Point, pn m3path.PathNode) bool {
			// do not release root nodes
			if !pn.IsRoot() {
				pn.(*PathNodeDb).release()
			}
			return false
		}, nbParallelProcesses)
	}
	// Clear the map
	onb.openNodesMap.Clear()
}
