package m3path

type OpenNodeBuilder struct {
	pathCtx                        *PathContextDb
	d                              int
	openNodes                      []*PathNodeDb
	expectedSize                   int
	openNodesMap                   PathNodeMap
	selectConflict, insertConflict int
}

func createNewNodeBuilder(previous *OpenNodeBuilder) *OpenNodeBuilder {
	res := new(OpenNodeBuilder)
	if previous == nil {
		res.d = 0
		res.expectedSize = 1
		res.openNodes = make([]*PathNodeDb, 0, 1)
		res.openNodesMap = MakeSimplePathNodeMap(1)
	} else {
		res.pathCtx = previous.pathCtx
		res.d = previous.d + 1
		res.expectedSize = previous.nextOpenNodesLen()
		res.openNodes = make([]*PathNodeDb, 0, res.expectedSize)
		if res.expectedSize > 32 {
			res.openNodesMap = MakeHashPathNodeMap(res.expectedSize)
		} else {
			res.openNodesMap = MakeSimplePathNodeMap(res.expectedSize)
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
		pn, err := readRowOnlyIds(rows)
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

func (onb *OpenNodeBuilder) addPathNode(pn *PathNodeDb) int {
	_, inserted := onb.openNodesMap.AddPathNode(pn)
	if inserted {
		onb.openNodes = append(onb.openNodes, pn)
	}
	return len(onb.openNodes)
}

func (onb *OpenNodeBuilder) nextOpenNodesLen() int {
	return calculatePredictedSize(onb.d, onb.openNodesMap.Size())
}

func (onb *OpenNodeBuilder) clear() {
	// Do not release the first three steps
	if onb.d > 3 {
		for _, on := range onb.openNodes {
			on.release()
		}
	}
	// Clear the map
	onb.openNodesMap.Clear()
}
