package m3point

import (
	"github.com/freddy33/qsm-go/utils/m3db"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDisplayPathBuilders(t *testing.T) {
	Log.SetAssert(true)
	m3db.SetToTestMode()

	env := GetFullTestDb(m3db.PointTestEnv)
	InitializeDBEnv(env, false)
	ppd := GetPointPackData(env)
	assert.Equal(t, TotalNumberOfCubes+1, len(ppd.PathBuilders))
	growthCtx := ppd.GetGrowthContextByTypeAndIndex(GrowthType(8), 0)
	pnb := ppd.GetPathNodeBuilder(growthCtx, 0, Origin)
	assert.NotNil(t, pnb, "did not find builder for %s", growthCtx.String())
	rpnb, tok := pnb.(*RootPathNodeBuilder)
	assert.True(t, tok, "%s is not a root builder", pnb.String())
	Log.Debug(rpnb.dumpInfo())
}

func TestAllPathBuilders(t *testing.T) {
	Log.SetAssert(true)
	Log.SetDebug()
	m3db.SetToTestMode()

	env := GetFullTestDb(m3db.PointTestEnv)
	InitializeDBEnv(env, true)
	ppd := GetPointPackData(env)

	assert.Equal(t, TotalNumberOfCubes+1, len(ppd.PathBuilders))
	for _, ctxType := range GetAllContextTypes() {
		nbIndexes := ctxType.GetNbIndexes()
		for pIdx := 0; pIdx < nbIndexes; pIdx++ {
			growthCtx := ppd.GetGrowthContextByTypeAndIndex(ctxType, pIdx)
			maxOffset := ctxType.GetMaxOffset()
			for offset := 0; offset < maxOffset; offset++ {
				centerPoint := Origin
				for div := uint64(0); div < 8*3; div++ {
					nbRoot := 0
					nbIntemediate := 0
					nbLastInter := 0
					if div != 0 {
						switch div % 3 {
						case 0:
							centerPoint = centerPoint.Add(XFirst)
						case 1:
							centerPoint = centerPoint.Add(YFirst)
						case 2:
							centerPoint = centerPoint.Add(ZFirst)
						}
					}
					assert.Equal(t, div, growthCtx.GetBaseDivByThree(centerPoint), "something wrong with div by three for %s", growthCtx.String())
					pnb := ppd.GetPathNodeBuilder(growthCtx, offset, centerPoint)
					assert.NotNil(t, pnb, "did not find builder for %s %v %v", growthCtx.String(), offset, div)
					rpnb, tok := pnb.(*RootPathNodeBuilder)
					assert.True(t, tok, "%s is not a root builder", pnb.String())
					nbRoot++
					trioIdx := pnb.GetTrioIndex()
					assert.NotEqual(t, NilTrioIndex, trioIdx, "no trio index for builder %s", pnb.String())
					assert.Equal(t, rpnb.trIdx, trioIdx, "trio index mismatch for builder %s", pnb.String())
					assert.True(t, trioIdx.IsBaseTrio(), "trio index is not a base one for builder %s", pnb.String())
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
					td := ppd.GetTrioDetails(trioIdx)
					assert.NotNil(t, td, "did not find trio index %s for path builder %s", trioIdx.String(), pnb.String())
					for i, cd := range td.conns {
						assert.Equal(t, cd.Id, rpnb.pathLinks[i].connId, "connId mismatch at %d for %s", i, pnb.String())
						npnb, np := pnb.GetNextPathNodeBuilder(centerPoint, cd.Id, offset)
						assert.NotNil(t, npnb, "nil next path node at %d for connId %s and pnb %s", i, cd.Id.String(), pnb.String())
						assert.NotEqual(t, NilTrioIndex, npnb.GetTrioIndex(), "no trio index for next path node at %d for connId %s and pnb %s", i, cd.Id.String(), pnb.String())
						assert.False(t, npnb.GetTrioIndex().IsBaseTrio(), "trio index should not a base one for builder %s", npnb.String())
						assert.Equal(t, centerPoint.Add(cd.Vector), np, "failed next point for builder %s", npnb.String())
						ntd := ppd.GetTrioDetails(npnb.GetTrioIndex())
						ipnb, iok := npnb.(*IntermediatePathNodeBuilder)
						nbIntemediate++
						assert.True(t, iok, "next path node not an intermediate at %d for connId %s and pnb %s", i, cd.Id.String(), pnb.String())
						for _, ncd := range ntd.conns {
							if ncd.GetNegId() != cd.GetId() {
								found := false
								var lipnb *LastPathNodeBuilder
								for _, ipl := range ipnb.pathLinks {
									if ipl.connId == ncd.GetId() {
										found = true
										liok := false
										lipnb, liok = ipl.pathNode.(*LastPathNodeBuilder)
										assert.True(t, liok, "next path node not an intermediate at %d for connId %s and pnb %s", i, cd.Id.String(), ipnb.String())
										assert.NotEqual(t, NilTrioIndex, lipnb.GetTrioIndex(), "no trio index for next path node at %d for connId %s and pnb %s", i, cd.Id.String(), lipnb.String())
										assert.False(t, lipnb.GetTrioIndex().IsBaseTrio(), "trio index should not a base one for builder %s", lipnb.String())
									}
								}
								assert.True(t, found, "not found inter cid %s for connId %s and pnb %s", ncd.GetId(), i, cd.Id.String(), pnb.String())
								lastIpnb, lip := ipnb.GetNextPathNodeBuilder(np, ncd.GetId(), offset)
								olipnb, liok := lastIpnb.(*LastPathNodeBuilder)
								nbLastInter++
								assert.True(t, liok, "next path node not an intermediate at %d for connId %s and pnb %s", i, cd.Id.String(), pnb.String())
								assert.Equal(t, lipnb, olipnb, "next path node not an intermediate at %d for connId %s and pnb %s", i, cd.Id.String(), pnb.String())
								assert.NotEqual(t, NilTrioIndex, lastIpnb.GetTrioIndex(), "no trio index for next path node at %d for connId %s and pnb %s", i, cd.Id.String(), pnb.String())
								assert.Equal(t, np.Add(ncd.Vector), lip, "next path node failed points at %d for connId %s and pnb %s", i, cd.Id.String(), pnb.String())

								nextMainPB, nmp := lastIpnb.GetNextPathNodeBuilder(lip, olipnb.nextMainConnId, offset)
								_, tok := nextMainPB.(*RootPathNodeBuilder)
								assert.True(t, tok, "%s is not a root builder", nextMainPB.String())
								assert.True(t, nmp.IsMainPoint(), "last node builder main does not give main for builder %s", nextMainPB.String())
								assert.True(t, nextMainPB.GetTrioIndex().IsBaseTrio(), "trio index is not a base one for builder %s", nextMainPB.String())
								assert.Equal(t, lip.GetNearMainPoint(), nmp, "next path node failed points at %d for connId %s and pnb %s", i, cd.Id.String(), pnb.String())

								// Make sure the way back get same trio
								if Log.IsTrace() {
									Log.Tracef("get back from %s %s %v", nextMainPB.String(), ppd.getBaseTrioDetails(growthCtx, nmp, offset).String(), nmp)
								}
								backIpnb, oLip := nextMainPB.GetNextPathNodeBuilder(nmp, olipnb.nextMainConnId.GetNegId(), offset)
								assert.NotNil(t, backIpnb, "%s next root builder is nil", nextMainPB.String())
								assert.Equal(t, lip, oLip, "%s next root builder does not point back to same point", nextMainPB.String())
								assert.Equal(t, lastIpnb.GetTrioIndex(), backIpnb.GetTrioIndex(), "%s next root builder does not point back to same trio index", nextMainPB.String())

								nextInterPB, nlip := lastIpnb.GetNextPathNodeBuilder(lip, olipnb.nextInterConnId, offset)
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
