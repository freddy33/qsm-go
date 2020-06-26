package m3point

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSamePermutation(t *testing.T) {
	assert.True(t, samePermutation([]TrioIndex{1, 2, 3, 4}, []TrioIndex{1, 2, 3, 4}))
	assert.True(t, samePermutation([]TrioIndex{1, 2, 3, 4}, []TrioIndex{4, 1, 2, 3}))
	assert.True(t, samePermutation([]TrioIndex{1, 2, 3, 4}, []TrioIndex{3, 4, 1, 2}))
	assert.True(t, samePermutation([]TrioIndex{1, 2, 3, 4}, []TrioIndex{2, 3, 4, 1}))

	assert.False(t, samePermutation([]TrioIndex{1, 2, 3, 4}, []TrioIndex{1, 2, 4, 3}))
	assert.False(t, samePermutation([]TrioIndex{1, 2, 3, 4}, []TrioIndex{3, 1, 2, 4}))
}

func TestPermBuilder(t *testing.T) {
	Log.SetDebug()
	p := TrioIndexPermBuilder{4, 0, make([][]TrioIndex, 12)}
	p.fill(0, make([]TrioIndex, p.size))
	//fmt.Println(p.collector)
	assert.Equal(t, 12, len(p.collector))
	for i, c := range p.collector {
		assert.Equal(t, 4, len(c), "population failed for %d %v in %v", i, c, p.collector)
	}
}

