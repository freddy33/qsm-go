package m3path

import "math/rand"

type OpenNodeBuilder struct {
	pathCtx   *PathContextDb
	d         int
	openNodes []*PathNodeDb
	expectedSize int
	openNodesMap [][3]*PathNodeDb
	mapSize int
	selectConflict, insertConflict int
}

func createNewNodeBuilder(previous *OpenNodeBuilder) *OpenNodeBuilder {
	res := new(OpenNodeBuilder)
	if previous == nil {
		res.d = 0
		res.expectedSize = 1
		res.openNodes = make([]*PathNodeDb, 0, 1)
	} else {
		res.pathCtx = previous.pathCtx
		res.d = previous.d + 1
		res.expectedSize = previous.nextOpenNodesLen()
		res.openNodes = make([]*PathNodeDb, 0, res.expectedSize)
	}
	return res
}

func (onb *OpenNodeBuilder) fillOpenPathNodes() []*PathNodeDb {
	pathCtx := onb.pathCtx
	rows, err := pathCtx.pathNodesTe().Query(SelectPathNodesByCtxAndDistance, pathCtx.id, onb.d)
	if err != nil {
		Log.Fatal(err)
	}
	res := make([]*PathNodeDb, 0, 100)
	for rows.Next() {
		pn, err := readRowOnlyIds(rows)
		if err != nil {
			Log.Fatalf("Could not read row of %s due to %v", PathNodesTable, err)
		} else {
			if pn.pathCtxId != pathCtx.id {
				Log.Fatalf("While retrieving all path nodes got a node with context id %d instead of %d",
					pn.pathCtxId, pathCtx.id)
				return nil
			}
			pn.pathCtx = pathCtx
			res = append(res, pn)
		}
	}
	return res
}

func (onb *OpenNodeBuilder) addPathNode(pn *PathNodeDb) int {
	if onb.expectedSize == 1 {
		onb.openNodes = append(onb.openNodes, pn)
	} else {
		onb.openNodes = append(onb.openNodes, pn)
	}
	return len(onb.openNodes)
}

func (onb *OpenNodeBuilder) nextOpenNodesLen() int {
	return calculatePredictedSize(onb.d, len(onb.openNodes))
}

func (onb *OpenNodeBuilder) shuffle() {
	rand.Shuffle(len(onb.openNodes), func(i, j int) { onb.openNodes[i], onb.openNodes[j] = onb.openNodes[j], onb.openNodes[i] })
}

func (onb *OpenNodeBuilder) clear() {
	// Do not release the first three steps
	if onb.d > 3 {
		for _, on := range onb.openNodes {
			on.release()
		}
	}
}
