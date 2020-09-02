package pointdb

import (
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSamePermutation(t *testing.T) {
	assert.True(t, samePermutation([]m3point.TrioIndex{1, 2, 3, 4}, []m3point.TrioIndex{1, 2, 3, 4}))
	assert.True(t, samePermutation([]m3point.TrioIndex{1, 2, 3, 4}, []m3point.TrioIndex{4, 1, 2, 3}))
	assert.True(t, samePermutation([]m3point.TrioIndex{1, 2, 3, 4}, []m3point.TrioIndex{3, 4, 1, 2}))
	assert.True(t, samePermutation([]m3point.TrioIndex{1, 2, 3, 4}, []m3point.TrioIndex{2, 3, 4, 1}))

	assert.False(t, samePermutation([]m3point.TrioIndex{1, 2, 3, 4}, []m3point.TrioIndex{1, 2, 4, 3}))
	assert.False(t, samePermutation([]m3point.TrioIndex{1, 2, 3, 4}, []m3point.TrioIndex{3, 1, 2, 4}))
}

func TestPermBuilder(t *testing.T) {
	Log.SetDebug()
	p := TrioIndexPermBuilder{4, 0, make([][]m3point.TrioIndex, 12)}
	p.fill(0, make([]m3point.TrioIndex, p.size))
	//fmt.Println(p.collector)
	assert.Equal(t, 12, len(p.collector))
	for i, c := range p.collector {
		assert.Equal(t, 4, len(c), "population failed for %d %v in %v", i, c, p.collector)
	}
}
