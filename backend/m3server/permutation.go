package m3server

import "github.com/freddy33/qsm-go/model/m3point"

type TrioIndexPermBuilder struct {
	size      int
	colIdx    int
	collector [][]m3point.TrioIndex
}

func samePermutation(p1, p2 []m3point.TrioIndex) bool {
	if len(p1) != len(p2) {
		m3point.Log.Fatalf("cannot test 2 permutation of different sizes %v %v", p1, p2)
	}
	permSize := len(p1)
	// Index in p2 of first entry in p1
	idx0 := -1
	for idx := 0; idx < permSize; idx++ {
		if p2[idx] == p1[0] {
			idx0 = idx
			break
		}
	}
	if idx0 == -1 {
		// did not find p1[0] so not same permutation
		return false
	}
	// Now they are same permutation if translation index of idx0 get same values
	for idx := 0; idx < permSize; idx++ {
		if p2[(idx0+idx)%permSize] != p1[idx] {
			// just one failure means doom
			return false
		}
	}
	return true
}

func (p *TrioIndexPermBuilder) fill(pos int, current []m3point.TrioIndex) {
	if pos == p.size {
		exists := false
		for i := 0; i < p.colIdx; i++ {
			if samePermutation(p.collector[i], current) {
				exists = true
				break
			}
		}
		if !exists {
			p.collector[p.colIdx] = current
			p.colIdx++
		}
		return
	}
	for i := 0; i < 4; i++ {
		// non prime index
		newIndex := m3point.TrioIndex(i)
		if pos%2 == 1 {
			// prime index
			newIndex = m3point.TrioIndex(i + 4)
		}
		usable := true
		if pos-1 >= 0 {
			// any index only once
			for j := 0; j < pos-1; j++ {
				if current[j] == newIndex {
					usable = false
				}
			}
			// Cannot have prime before
			if isPrime(newIndex, current[pos-1]) {
				usable = false
			}
		}
		// If last cannot be prime with first
		if pos+1 == p.size {
			if isPrime(newIndex, current[0]) {
				usable = false
			}
		}
		if usable {
			perm := make([]m3point.TrioIndex, p.size)
			copy(perm, current)
			perm[pos] = newIndex
			p.fill(pos+1, perm)
		}
	}
}
