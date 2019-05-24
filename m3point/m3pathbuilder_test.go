package m3point

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAllPathBuilders(t *testing.T) {
	nb := createAllPathBuilders()
	assert.Equal(t, 1664, nb)
	for _, ctxType := range GetAllContextTypes() {
		nbIndexes := ctxType.GetNbIndexes()
		for pIdx := 0; pIdx < nbIndexes; pIdx++ {
			trCtx := GetTrioIndexContext(ctxType, pIdx)
			maxOffset := MaxOffsetPerType[ctxType]
			for offset := 0; offset < maxOffset; offset++ {
				for div := uint64(0); div < 12; div++ {
					pathNodeBuilder := GetPathNodeBuilder(trCtx, offset, div)
					assert.NotNil(t, pathNodeBuilder, "did not find builder for %v %v %v", *trCtx, offset, div)
				}
			}
		}
	}
}

