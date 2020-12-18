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
	ppd.InitializeAll()

	if !assert.Equal(t, TotalNumberOfCubes+1, ppd.GetNbPathBuilders()) {
		return
	}
	nbTest := 0
	for _, ctxType := range m3point.GetAllGrowthTypes() {
		nbIndexes := ctxType.GetNbIndexes()
		for pIdx := 0; pIdx < nbIndexes; pIdx++ {
			growthCtx := ppd.GetGrowthContextByTypeAndIndex(ctxType, pIdx)
			maxOffset := ctxType.GetMaxOffset()
			for offset := 0; offset < maxOffset; offset++ {
				centerPoint := m3point.Origin
				for div := uint64(0); div < 8*3; div++ {
					nbTest++
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
					good := assert.Equal(t, div, growthCtx.GetBaseDivByThree(centerPoint), "something wrong with div by three for %s", growthCtx.String())
					if !good {
						return
					}
					pnb := ppd.GetPathNodeBuilder(growthCtx, offset, centerPoint)
					good = assert.NotNil(t, pnb, "did not find builder for %s %v %v", growthCtx.String(), offset, div)
					if !good {
						return
					}
					rpnb, tok := pnb.(*RootPathNodeBuilder)
					good = assert.True(t, tok, "%s is not a root builder", pnb.String())
					if !good {
						return
					}
					nbRoot++
					trioIdx := pnb.GetTrioIndex()
					good = assert.NotEqual(t, m3point.NilTrioIndex, trioIdx, "no trio index for builder %s", pnb.String()) &&
						assert.Equal(t, rpnb.TrIdx, trioIdx, "trio index mismatch for builder %s", pnb.String()) &&
						assert.True(t, trioIdx.IsBaseTrio(), "trio index is not a base one for builder %s", pnb.String())
					if !good {
						return
					}
					switch ctxType {
					case 1:
						good = assert.Equal(t, m3point.TrioIndex(pIdx), trioIdx, "wrong trio index for %s", pnb.String())
					case 3:
						if m3util.PosMod2(div) == 0 {
							good = assert.Equal(t, m3point.TrioIndex(pIdx), trioIdx, "wrong trio index for %s", pnb.String())
						} else {
							good = assert.NotEqual(t, m3point.TrioIndex(pIdx), trioIdx, "wrong trio index for %s", pnb.String())
						}
					case 2:
						idx := int(m3util.PosMod2(div + uint64(offset)))
						good = assert.Equal(t, validNextTrio[pIdx][idx], trioIdx, "wrong trio index for %s", pnb.String())
					case 4:
						idx := int(m3util.PosMod4(div + uint64(offset)))
						good = assert.Equal(t, allMod4Permutations[pIdx][idx], trioIdx, "wrong trio index for %s", pnb.String())
					case 8:
						idx := int(m3util.PosMod8(div + uint64(offset)))
						good = assert.Equal(t, allMod8Permutations[pIdx][idx], trioIdx, "wrong trio index for %s", pnb.String())
					}
					if !good {
						return
					}
					td := ppd.GetTrioDetails(trioIdx)
					good = assert.NotNil(t, td, "did not find trio index %s for path builder %s", trioIdx.String(), pnb.String())
					if !good {
						return
					}
					for i, cd := range td.Conns {
						good = assert.Equal(t, cd.Id, rpnb.PathLinks[i].ConnId, "connId mismatch at %d for %s", i, pnb.String())
						if !good {
							return
						}
						npnb, np, err := pnb.GetNextPathNodeBuilder(centerPoint, cd.Id, offset)
						good = assert.NoError(t, err, "err next path node at %d for connId %s and pnb %s", i, cd.Id.String(), pnb.String()) &&
							assert.NotNil(t, npnb, "nil next path node at %d for connId %s and pnb %s", i, cd.Id.String(), pnb.String()) &&
							assert.NotEqual(t, m3point.NilTrioIndex, npnb.GetTrioIndex(), "no trio index for next path node at %d for connId %s and pnb %s", i, cd.Id.String(), pnb.String()) &&
							assert.False(t, npnb.GetTrioIndex().IsBaseTrio(), "trio index should not a base one for builder %s", npnb.String()) &&
							assert.Equal(t, centerPoint.Add(cd.Vector), np, "failed next point for builder %s", npnb.String())
						if !good {
							return
						}
						ntd := ppd.GetTrioDetails(npnb.GetTrioIndex())
						ipnb, iok := npnb.(*IntermediatePathNodeBuilder)
						nbIntemediate++
						good = assert.True(t, iok, "next path node not an intermediate at %d for connId %s and pnb %s", i, cd.Id.String(), pnb.String())
						if !good {
							return
						}
						for _, ncd := range ntd.Conns {
							if ncd.GetNegId() != cd.GetId() {
								found := false
								var lipnb *LastPathNodeBuilder
								for _, ipl := range ipnb.PathLinks {
									if ipl.ConnId == ncd.GetId() {
										found = true
										liok := false
										lipnb, liok = ipl.PathNode.(*LastPathNodeBuilder)
										good = assert.True(t, liok, "next path node not an intermediate at %d for connId %s and pnb %s", i, cd.Id.String(), ipnb.String()) &&
											assert.NotEqual(t, m3point.NilTrioIndex, lipnb.GetTrioIndex(), "no trio index for next path node at %d for connId %s and pnb %s", i, cd.Id.String(), lipnb.String()) &&
											assert.False(t, lipnb.GetTrioIndex().IsBaseTrio(), "trio index should not a base one for builder %s", lipnb.String())
										if !good {
											return
										}
									}
								}
								good = assert.True(t, found, "not found inter cid %s for connId %s and pnb %s", ncd.GetId(), i, cd.Id.String(), pnb.String())
								if !good {
									return
								}
								lastIpnb, lip, err := ipnb.GetNextPathNodeBuilder(np, ncd.GetId(), offset)
								olipnb, liok := lastIpnb.(*LastPathNodeBuilder)
								nbLastInter++
								good = assert.NoError(t, err, "next path node not an intermediate at %d for connId %s and pnb %s", i, cd.Id.String(), pnb.String()) &&
									assert.True(t, liok, "next path node not an intermediate at %d for connId %s and pnb %s", i, cd.Id.String(), pnb.String()) &&
									assert.Equal(t, lipnb, olipnb, "next path node not an intermediate at %d for connId %s and pnb %s", i, cd.Id.String(), pnb.String()) &&
									assert.NotEqual(t, m3point.NilTrioIndex, lastIpnb.GetTrioIndex(), "no trio index for next path node at %d for connId %s and pnb %s", i, cd.Id.String(), pnb.String()) &&
									assert.Equal(t, np.Add(ncd.Vector), lip, "next path node failed points at %d for connId %s and pnb %s", i, cd.Id.String(), pnb.String())
								if !good {
									return
								}

								nextMainPB, nmp, err := lastIpnb.GetNextPathNodeBuilder(lip, olipnb.NextMainConnId, offset)
								_, tok := nextMainPB.(*RootPathNodeBuilder)
								good = assert.NoError(t, err, "last node builder main failed for builder %s", nextMainPB.String()) &&
									assert.True(t, tok, "%s is not a root builder", nextMainPB.String()) &&
									assert.True(t, nmp.IsMainPoint(), "last node builder main does not give main for builder %s", nextMainPB.String()) &&
									assert.True(t, nextMainPB.GetTrioIndex().IsBaseTrio(), "trio index is not a base one for builder %s", nextMainPB.String()) &&
									assert.Equal(t, lip.GetNearMainPoint(), nmp, "next path node failed points at %d for connId %s and pnb %s", i, cd.Id.String(), pnb.String())
								if !good {
									return
								}

								// Make sure the way back get same trio
								if Log.IsTrace() {
									Log.Tracef("get back from %s %s %v", nextMainPB.String(), GetBaseTrioDetails(growthCtx, nmp, offset).String(), nmp)
								}
								backIpnb, oLip, err := nextMainPB.GetNextPathNodeBuilder(nmp, olipnb.NextMainConnId.GetNegId(), offset)
								good = assert.NoError(t, err, "%s next root builder failed", nextMainPB.String()) &&
									assert.NotNil(t, backIpnb, "%s next root builder is nil", nextMainPB.String()) &&
									assert.Equal(t, lip, oLip, "%s next root builder does not point back to same point", nextMainPB.String()) &&
									assert.Equal(t, lastIpnb.GetTrioIndex(), backIpnb.GetTrioIndex(), "%s next root builder does not point back to same trio index", nextMainPB.String())
								if !good {
									return
								}

								nextInterPB, nlip, err := lastIpnb.GetNextPathNodeBuilder(lip, olipnb.NextInterConnId, offset)
								_, tiok := nextInterPB.(*LastPathNodeBuilder)
								good = assert.NoError(t, err, "%s failed to get a last inter builder", nextInterPB.String()) &&
									assert.True(t, tiok, "%s is not a last inter builder", nextInterPB.String()) &&
									assert.NotEqual(t, nlip.GetNearMainPoint(), nmp, "next path node failed points at %d for connId %s and pnb %s", i, cd.Id.String(), pnb.String())
								if !good {
									return
								}
							}
						}
					}
					good = assert.Equal(t, 1, nbRoot, "for builder %s", pnb.String()) &&
						assert.Equal(t, 3, nbIntemediate, "for builder %s", pnb.String()) &&
						assert.Equal(t, 6, nbLastInter, "for builder %s", pnb.String())
					if !good {
						return
					}
				}
			}
		}
	}
	assert.Equal(t, 4800, nbTest)
}
