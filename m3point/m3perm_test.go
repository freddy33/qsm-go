package m3point

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSamePermutation(t *testing.T) {
	assert.True(t, samePermutation([]int{1, 2, 3, 4}, []int{1, 2, 3, 4}))
	assert.True(t, samePermutation([]int{1, 2, 3, 4}, []int{4, 1, 2, 3}))
	assert.True(t, samePermutation([]int{1, 2, 3, 4}, []int{3, 4, 1, 2}))
	assert.True(t, samePermutation([]int{1, 2, 3, 4}, []int{2, 3, 4, 1}))

	assert.False(t, samePermutation([]int{1, 2, 3, 4}, []int{1, 2, 4, 3}))
	assert.False(t, samePermutation([]int{1, 2, 3, 4}, []int{3, 1, 2, 4}))
}

func TestPermBuilder(t *testing.T) {
	Log.Level = m3util.DEBUG
	p := PermBuilder{4, 0, make([][]int, 12)}
	p.fill(0, make([]int, p.size))
	fmt.Println(p.collector)
	assert.Equal(t, 12, len(p.collector))
	for i, c := range p.collector {
		assert.Equal(t, 4, len(c), "population failed for %d %v", i, c)
	}
}

