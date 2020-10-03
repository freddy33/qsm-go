package pointdb

import (
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAllPathBuilders(t *testing.T) {
	Log.SetAssert(true)
	Log.SetDebug()
	m3util.SetToTestMode()

	env := GetPointDbFullEnv(m3util.PointTestEnv)
	ppd := GetServerPointPackData(env)

	assert.Equal(t, TotalNumberOfCubes+1, ppd.GetNbPathBuilders())
	for _, ctxType := range m3point.GetAllGrowthTypes() {
		nbIndexes := ctxType.GetNbIndexes()
		for pIdx := 0; pIdx < nbIndexes; pIdx++ {
			growthCtx := ppd.GetGrowthContextByTypeAndIndex(ctxType, pIdx)
			maxOffset := ctxType.GetMaxOffset()
			for offset := 0; offset < maxOffset; offset++ {
				centerPoint := m3point.Origin
				for div := uint64(0); div < 8*3; div++ {
					nbRoot := 0
					nbIntemediate := 0
					nbLastInter := 0
					if div != 0 {
						switch div % 3 {
						case 0:
							centerPoint = centerPoint.Add(m3point.XFirst)
						case 1:
							centerPoint = centerPoint.Add(m3point.YFirst)
						case 2:
							centerPoint = centerPoint.Add(m3point.ZFirst)
						}
					}
					assert.Equal(t, div, growthCtx.GetBaseDivByThree(centerPoint), "something wrong with div by three for %s", growthCtx.String())
					pnb := ppd.GetPathNodeBuilder(growthCtx, offset, centerPoint)
					assert.NotNil(t, pnb, "did not find builder for %s %v %v", growthCtx.String(), offset, div)
					rpnb, tok := pnb.(*RootPathNodeBuilder)
					assert.True(t, tok, "%s is not a root builder", pnb.String())
					nbRoot++
					trioIdx := pnb.GetTrioIndex()
					assert.NotEqual(t, m3point.NilTrioIndex, trioIdx, "no trio index for builder %s", pnb.String())
					assert.Equal(t, rpnb.TrIdx, trioIdx, "trio index mismatch for builder %s", pnb.String())
					assert.True(t, trioIdx.IsBaseTrio(), "trio index is not a base one for builder %s", pnb.String())
					switch ctxType {
					case 1:
						assert.Equal(t, m3point.TrioIndex(pIdx), trioIdx, "wrong trio index for %s", pnb.String())
					case 3:
						if m3util.PosMod2(div) == 0 {
							assert.Equal(t, m3point.TrioIndex(pIdx), trioIdx, "wrong trio index for %s", pnb.String())
						} else {
							assert.NotEqual(t, m3point.TrioIndex(pIdx), trioIdx, "wrong trio index for %s", pnb.String())
						}
					case 2:
						idx := int(m3util.PosMod2(div + uint64(offset)))
						assert.Equal(t, validNextTrio[pIdx][idx], trioIdx, "wrong trio index for %s", pnb.String())
					case 4:
						idx := int(m3util.PosMod4(div + uint64(offset)))
						assert.Equal(t, allMod4Permutations[pIdx][idx], trioIdx, "wrong trio index for %s", pnb.String())
					case 8:
						idx := int(m3util.PosMod8(div + uint64(offset)))
						assert.Equal(t, allMod8Permutations[pIdx][idx], trioIdx, "wrong trio index for %s", pnb.String())
					}
					td := ppd.GetTrioDetails(trioIdx)
					assert.NotNil(t, td, "did not find trio index %s for path builder %s", trioIdx.String(), pnb.String())
					for i, cd := range td.Conns {
						assert.Equal(t, cd.Id, rpnb.PathLinks[i].ConnId, "connId mismatch at %d for %s", i, pnb.String())
						npnb, np := pnb.GetNextPathNodeBuilder(centerPoint, cd.Id, offset)
						assert.NotNil(t, npnb, "nil next path node at %d for connId %s and pnb %s", i, cd.Id.String(), pnb.String())
						assert.NotEqual(t, m3point.NilTrioIndex, npnb.GetTrioIndex(), "no trio index for next path node at %d for connId %s and pnb %s", i, cd.Id.String(), pnb.String())
						assert.False(t, npnb.GetTrioIndex().IsBaseTrio(), "trio index should not a base one for builder %s", npnb.String())
						assert.Equal(t, centerPoint.Add(cd.Vector), np, "failed next point for builder %s", npnb.String())
						ntd := ppd.GetTrioDetails(npnb.GetTrioIndex())
						ipnb, iok := npnb.(*IntermediatePathNodeBuilder)
						nbIntemediate++
						assert.True(t, iok, "next path node not an intermediate at %d for connId %s and pnb %s", i, cd.Id.String(), pnb.String())
						for _, ncd := range ntd.Conns {
							if ncd.GetNegId() != cd.GetId() {
								found := false
								var lipnb *LastPathNodeBuilder
								for _, ipl := range ipnb.PathLinks {
									if ipl.ConnId == ncd.GetId() {
										found = true
										liok := false
										lipnb, liok = ipl.PathNode.(*LastPathNodeBuilder)
										assert.True(t, liok, "next path node not an intermediate at %d for connId %s and pnb %s", i, cd.Id.String(), ipnb.String())
										assert.NotEqual(t, m3point.NilTrioIndex, lipnb.GetTrioIndex(), "no trio index for next path node at %d for connId %s and pnb %s", i, cd.Id.String(), lipnb.String())
										assert.False(t, lipnb.GetTrioIndex().IsBaseTrio(), "trio index should not a base one for builder %s", lipnb.String())
									}
								}
								assert.True(t, found, "not found inter cid %s for connId %s and pnb %s", ncd.GetId(), i, cd.Id.String(), pnb.String())
								lastIpnb, lip := ipnb.GetNextPathNodeBuilder(np, ncd.GetId(), offset)
								olipnb, liok := lastIpnb.(*LastPathNodeBuilder)
								nbLastInter++
								assert.True(t, liok, "next path node not an intermediate at %d for connId %s and pnb %s", i, cd.Id.String(), pnb.String())
								assert.Equal(t, lipnb, olipnb, "next path node not an intermediate at %d for connId %s and pnb %s", i, cd.Id.String(), pnb.String())
								assert.NotEqual(t, m3point.NilTrioIndex, lastIpnb.GetTrioIndex(), "no trio index for next path node at %d for connId %s and pnb %s", i, cd.Id.String(), pnb.String())
								assert.Equal(t, np.Add(ncd.Vector), lip, "next path node failed points at %d for connId %s and pnb %s", i, cd.Id.String(), pnb.String())

								nextMainPB, nmp := lastIpnb.GetNextPathNodeBuilder(lip, olipnb.NextMainConnId, offset)
								_, tok := nextMainPB.(*RootPathNodeBuilder)
								assert.True(t, tok, "%s is not a root builder", nextMainPB.String())
								assert.True(t, nmp.IsMainPoint(), "last node builder main does not give main for builder %s", nextMainPB.String())
								assert.True(t, nextMainPB.GetTrioIndex().IsBaseTrio(), "trio index is not a base one for builder %s", nextMainPB.String())
								assert.Equal(t, lip.GetNearMainPoint(), nmp, "next path node failed points at %d for connId %s and pnb %s", i, cd.Id.String(), pnb.String())

								// Make sure the way back get same trio
								if Log.IsTrace() {
									Log.Tracef("get back from %s %s %v", nextMainPB.String(), GetBaseTrioDetails(growthCtx, nmp, offset).String(), nmp)
								}
								backIpnb, oLip := nextMainPB.GetNextPathNodeBuilder(nmp, olipnb.NextMainConnId.GetNegId(), offset)
								assert.NotNil(t, backIpnb, "%s next root builder is nil", nextMainPB.String())
								assert.Equal(t, lip, oLip, "%s next root builder does not point back to same point", nextMainPB.String())
								assert.Equal(t, lastIpnb.GetTrioIndex(), backIpnb.GetTrioIndex(), "%s next root builder does not point back to same trio index", nextMainPB.String())

								nextInterPB, nlip := lastIpnb.GetNextPathNodeBuilder(lip, olipnb.NextInterConnId, offset)
								_, tiok := nextInterPB.(*LastPathNodeBuilder)
								assert.True(t, tiok, "%s is not a last inter builder", nextInterPB.String())
								assert.NotEqual(t, nlip.GetNearMainPoint(), nmp, "next path node failed points at %d for connId %s and pnb %s", i, cd.Id.String(), pnb.String())
							}
						}
					}
					assert.Equal(t, 1, nbRoot, "for builder %s", pnb.String())
					assert.Equal(t, 3, nbIntemediate, "for builder %s", pnb.String())
					assert.Equal(t, 6, nbLastInter, "for builder %s", pnb.String())
				}
			}
		}
	}
}
