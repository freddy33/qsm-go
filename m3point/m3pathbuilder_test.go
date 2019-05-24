package m3point

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAllPathBuilders(t *testing.T) {
	Log.SetAssert(true)
	nb := createAllPathBuilders()
	assert.Equal(t, 1664, nb)
	for _, ctxType := range GetAllContextTypes() {
		nbIndexes := ctxType.GetNbIndexes()
		for pIdx := 0; pIdx < nbIndexes; pIdx++ {
			trCtx := GetTrioIndexContext(ctxType, pIdx)
			maxOffset := MaxOffsetPerType[ctxType]
			for offset := 0; offset < maxOffset; offset++ {
				for div := uint64(0); div < 12; div++ {
					pnb := GetPathNodeBuilder(trCtx, offset, div)
					assert.NotNil(t, pnb, "did not find builder for %v %v %v", *trCtx, offset, div)
					rpnb, tok := pnb.(*RootPathNodeBuilder)
					assert.True(t, tok, "%s is not a root builder", pnb.String())
					trioIdx := pnb.GetTrioIndex()
					assert.NotEqual(t, NilTrioIndex, trioIdx, "no trio index for builder %s", pnb.String())
					assert.Equal(t, rpnb.trIdx, trioIdx, "trio index mismatch for builder %s", pnb.String())
					switch ctxType {
					case 1:
						assert.Equal(t, TrioIndex(pIdx), trioIdx, "wrong trio index for %s", pnb.String())
					case 3:
						if PosMod2(div) == 0 {
							assert.Equal(t, TrioIndex(pIdx), trioIdx, "wrong trio index for %s", pnb.String())
						} else {
							assert.NotEqual(t, TrioIndex(pIdx), trioIdx, "wrong trio index for %s", pnb.String())
						}
					case 2:
						idx := int(PosMod2(div + uint64(offset)))
						assert.Equal(t, validNextTrio[pIdx][idx], trioIdx, "wrong trio index for %s", pnb.String())
					case 4:
						idx := int(PosMod4(div + uint64(offset)))
						assert.Equal(t, AllMod4Permutations[pIdx][idx], trioIdx, "wrong trio index for %s", pnb.String())
					case 8:
						idx := int(PosMod8(div + uint64(offset)))
						assert.Equal(t, AllMod8Permutations[pIdx][idx], trioIdx, "wrong trio index for %s", pnb.String())
					}
					td := GetTrioDetails(trioIdx)
					assert.NotNil(t, td, "did not find trio index %s for path builder %s", trioIdx.String(), pnb.String())
					for i, cd := range td.conns {
						assert.Equal(t, cd.Id, rpnb.pathLinks[i].connId, "connId mismatch at %d for %s", i, pnb.String())
						npnb := pnb.GetNextPathNodeBuilder(cd.Id)
						assert.NotNil(t, npnb, "nil next path node at %d for connId %s and pnb %s", i, cd.Id.String(), pnb.String())
						assert.NotEqual(t, NilTrioIndex, npnb.GetTrioIndex(), "no trio index for next path node at %d for connId %s and pnb %s", i, cd.Id.String(), pnb.String())
						ntd := GetTrioDetails(npnb.GetTrioIndex())
						ipnb, iok := npnb.(*IntermediatePathNodeBuilder)
						assert.True(t, iok, "next path node not an intermediate at %d for connId %s and pnb %s", i, cd.Id.String(), pnb.String())
						for _, ncd := range ntd.conns {
							if ncd.GetNegId() != cd.GetId() {
								found := false
								for _, ipl := range ipnb.pathLinks {
									if ipl.connId == ncd.GetId() {
										found = true
										lipnb, liok := ipl.pathNode.(*LastIntermediatePathNodeBuilder)
										assert.True(t, liok, "next path node not an intermediate at %d for connId %s and pnb %s", i, cd.Id.String(), pnb.String())
										assert.NotEqual(t, NilTrioIndex, lipnb.GetTrioIndex(), "no trio index for next path node at %d for connId %s and pnb %s", i, cd.Id.String(), pnb.String())

									}
								}
								assert.True(t, found, "not found inter cid %s for connId %s and pnb %s", ncd.GetId(), i, cd.Id.String(), pnb.String())
							}
						}
					}
				}
			}
		}
	}
}
