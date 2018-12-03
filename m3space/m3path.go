package m3space

import "github.com/freddy33/qsm-go/m3util"

// An element in the path from event base node to latest outgrowth
// Forward is from event to outgrowth
// Backwards is from latest outgrowth to event
type PathElement interface {
	IsEnd() bool
	NbForwardElements() int
	GetForwardConnId(idx int) int8
	GetForwardElement(idx int) PathElement
	Copy() PathElement
	SetLastNext(path PathElement)
	GetLength() int
}

// End of path marker
type EndPathElement int8

// The int8 here is the forward connection Id
type SimplePathElement struct {
	forwardConnId int8
	next          PathElement
}

// We count only forward fork
type ForkPathElement struct {
	simplePaths []*SimplePathElement
}

var TheEnd = EndPathElement(0)

/***************************************************************/
// Simple Path Functions
/***************************************************************/

func (spe EndPathElement) IsEnd() bool {
	return true
}

func (spe EndPathElement) NbForwardElements() int {
	return 0
}

func (spe EndPathElement) GetForwardConnId(idx int) int8 {
	return int8(spe)
}

func (spe EndPathElement) GetForwardElement(idx int) PathElement {
	return nil
}

func (spe EndPathElement) Copy() PathElement {
	return spe
}

func (spe EndPathElement) SetLastNext(path PathElement) {
	Log.Fatalf("cannot set last on end element")
}

func (spe EndPathElement) GetLength() int {
	return 0
}

/***************************************************************/
// Simple Path Functions
/***************************************************************/

func (spe *SimplePathElement) IsEnd() bool {
	return false
}

func (spe *SimplePathElement) NbForwardElements() int {
	return 1
}

func (spe *SimplePathElement) GetForwardConnId(idx int) int8 {
	if idx != 0 {
		Log.Fatalf("index out of bound for %d", idx)
	}
	return spe.forwardConnId
}

func (spe *SimplePathElement) GetForwardElement(idx int) PathElement {
	if idx != 0 {
		Log.Fatalf("index out of bound for %d", idx)
	}
	return spe.next
}

func (spe *SimplePathElement) Copy() PathElement {
	return spe.internalCopy()
}

func (spe *SimplePathElement) internalCopy() *SimplePathElement {
	if spe.next == nil {
		return &SimplePathElement{spe.forwardConnId, nil}
	}
	return &SimplePathElement{spe.forwardConnId, spe.next.Copy()}
}

func (spe *SimplePathElement) SetLastNext(path PathElement) {
	if spe.next == nil {
		spe.next = path
	} else {
		spe.next.SetLastNext(path)
	}
}

func (spe *SimplePathElement) GetLength() int {
	if spe.next == nil {
		return 1
	} else {
		return 1 + spe.next.GetLength()
	}
}

/***************************************************************/
// Forked Path Functions
/***************************************************************/

func (fpe *ForkPathElement) IsEnd() bool {
	return false
}

func (fpe *ForkPathElement) NbForwardElements() int {
	return len(fpe.simplePaths)
}

func (fpe *ForkPathElement) GetForwardConnId(idx int) int8 {
	return fpe.simplePaths[idx].GetForwardConnId(0)
}

func (fpe *ForkPathElement) GetForwardElement(idx int) PathElement {
	return fpe.simplePaths[idx].GetForwardElement(0)
}

func (fpe *ForkPathElement) Copy() PathElement {
	res := ForkPathElement{make([]*SimplePathElement, len(fpe.simplePaths))}
	for i, spe := range fpe.simplePaths {
		res.simplePaths[i] = spe.internalCopy()
	}
	return &res
}

func (fpe *ForkPathElement) SetLastNext(path PathElement) {
	for _, spe := range fpe.simplePaths {
		spe.SetLastNext(path)
	}
}

func (fpe *ForkPathElement) GetLength() int {
	length := fpe.simplePaths[0].GetLength()
	if Log.Level <= m3util.DEBUG {
		// All length should be identical
		for i := 1; i < len(fpe.simplePaths); i++ {
			otherLength := fpe.simplePaths[i].GetLength()
			if otherLength != length {
				Log.Errorf("fork points to 2 path with diff length %d != %d", length, otherLength)
			}
		}
	}
	return length
}

/***************************************************************/
// Merge Path Functions
/***************************************************************/

func MergePath(path1, path2 PathElement) PathElement {
	if path1 == nil && path2 == nil {
		return nil
	}
	if (path1 != nil && path2 == nil) || (path1 == nil && path2 != nil) {
		Log.Errorf("cannot merge path if one nil and not the other")
		return nil
	}
	if path1.GetLength() != path2.GetLength() {
		Log.Errorf("cannot merge path of different length")
		return nil
	}
	nb1 := path1.NbForwardElements()
	nb2 := path2.NbForwardElements()
	if nb1 == 1 && nb2 == 1 {
		p1ConnId := path1.GetForwardConnId(0)
		p2ConnId := path2.GetForwardConnId(0)
		p1Next := path1.GetForwardElement(0)
		p2Next := path2.GetForwardElement(0)
		if p1ConnId == p2ConnId {
			return &SimplePathElement{p1ConnId, MergePath(p1Next, p2Next)}
		}
		if p1Next != nil {
			p1Next = p1Next.Copy()
		}
		if p2Next != nil {
			p2Next = p2Next.Copy()
		}
		fpe := ForkPathElement{make([]*SimplePathElement, 2)}
		fpe.simplePaths[0] = &SimplePathElement{p1ConnId, p1Next}
		fpe.simplePaths[1] = &SimplePathElement{p2ConnId, p2Next}
		return &fpe
	}
	pathsPerConnId := make(map[int8][]*SimplePathElement)
	for i := 0; i < nb1; i++ {
		addCopyToMap(path1, i, &pathsPerConnId)
	}
	for i := 0; i < nb2; i++ {
		addCopyToMap(path2, i, &pathsPerConnId)
	}
	i := 0
	res := ForkPathElement{make([]*SimplePathElement, len(pathsPerConnId))}
	for connId, paths := range pathsPerConnId {
		if len(paths) == 1 {
			res.simplePaths[i] = paths[0]
			i++
		} else if len(paths) == 2 {
			res.simplePaths[i] = &SimplePathElement{connId, MergePath(paths[0].GetForwardElement(0), paths[1].GetForwardElement(0))}
			i++
		} else {
			Log.Errorf("Cannot have paths in merge for same connection ids not 1 or 2 for %d %d", connId, len(paths))
		}
	}
	return &res
}

func addCopyToMap(path PathElement, idx int, pathsPerConnId *map[int8][]*SimplePathElement) {
	connId := path.GetForwardConnId(idx)
	next := path.GetForwardElement(idx)
	if next != nil {
		next = next.Copy()
	}
	paths, ok := (*pathsPerConnId)[connId]
	newPath := &SimplePathElement{connId, next}
	if !ok {
		paths = make([]*SimplePathElement, 1)
		paths[0] = newPath
	} else {
		paths = append(paths, newPath)
	}
	(*pathsPerConnId)[connId] = paths
}

/***************************************************************/
// Path building Functions
/***************************************************************/

func (eo *EventOutgrowth) GetRootPathElement(evt *Event) PathElement {
	if eo.rootPath == nil {
		eo.rootPath = eo.BuildPath(TheEnd)
	}
	return eo.rootPath
}

func (seo *SavedEventOutgrowth) GetRootPathElement(evt *Event) PathElement {
	if seo.rootPath == nil {
		seo.rootPath = seo.BuildPath(TheEnd)
	}
	return seo.rootPath
}

func (eo *EventOutgrowth) BuildPath(path PathElement) PathElement {
	if eo.IsRoot() {
		return path
	}
	return path
	/* TODO
	fromConnIds := eo.GetFromConnIds()
	firstPath := eo.fromList[0].BuildPath(&SimplePathElement{-fromConnIds[0], path,})
	if len(eo.fromList) == 1 {
		return firstPath
	}
	for i := 1; i < len(eo.fromList); i++ {
		newPath := eo.fromList[i].BuildPath(&SimplePathElement{-fromConnIds[i], path,})
		firstPath = MergePath(firstPath, newPath)
	}
	return firstPath
	*/
}

func (seo *SavedEventOutgrowth) BuildPath(path PathElement) PathElement {
	if seo.IsRoot() {
		return path
	}
	newPath := seo.rootPath.Copy()
	newPath.SetLastNext(path)
	return newPath
}
