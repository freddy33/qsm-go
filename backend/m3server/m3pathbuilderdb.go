package m3server

import (
	"fmt"
	"github.com/freddy33/qsm-go/backend/m3db"
	"github.com/freddy33/qsm-go/model/m3point"
)

const (
	PathBuildersTable = "path_builders"
)

func init() {
	m3db.AddTableDef(createPathBuilderContextTableDef())
}

func createPathBuilderContextTableDef() *m3db.TableDefinition {
	res := m3db.TableDefinition{}
	res.Name = PathBuildersTable
	res.DdlColumns = fmt.Sprintf("(id smallint PRIMARY KEY REFERENCES %s (id),"+
		" ctx_id smallint NOT NULL REFERENCES %s (id),"+
		" root smallint NOT NULL REFERENCES %s (id),"+
		" inter1 smallint NOT NULL REFERENCES %s (id), inter2 smallint NOT NULL REFERENCES %s (id), inter3 smallint NOT NULL REFERENCES %s (id),"+
		" conn11 smallint NOT NULL REFERENCES %s (id), last_inter11 smallint NOT NULL REFERENCES %s (id), next_main_conn11 smallint NOT NULL REFERENCES %s (id), next_inter_conn11 smallint NOT NULL REFERENCES %s (id),"+
		" conn12 smallint NOT NULL REFERENCES %s (id), last_inter12 smallint NOT NULL REFERENCES %s (id), next_main_conn12 smallint NOT NULL REFERENCES %s (id), next_inter_conn12 smallint NOT NULL REFERENCES %s (id),"+
		" conn21 smallint NOT NULL REFERENCES %s (id), last_inter21 smallint NOT NULL REFERENCES %s (id), next_main_conn21 smallint NOT NULL REFERENCES %s (id), next_inter_conn21 smallint NOT NULL REFERENCES %s (id),"+
		" conn22 smallint NOT NULL REFERENCES %s (id), last_inter22 smallint NOT NULL REFERENCES %s (id), next_main_conn22 smallint NOT NULL REFERENCES %s (id), next_inter_conn22 smallint NOT NULL REFERENCES %s (id),"+
		" conn31 smallint NOT NULL REFERENCES %s (id), last_inter31 smallint NOT NULL REFERENCES %s (id), next_main_conn31 smallint NOT NULL REFERENCES %s (id), next_inter_conn31 smallint NOT NULL REFERENCES %s (id),"+
		" conn32 smallint NOT NULL REFERENCES %s (id), last_inter32 smallint NOT NULL REFERENCES %s (id), next_main_conn32 smallint NOT NULL REFERENCES %s (id), next_inter_conn32 smallint NOT NULL REFERENCES %s (id))",
		TrioCubesTable,
		GrowthContextsTable,
		TrioDetailsTable,
		TrioDetailsTable, TrioDetailsTable, TrioDetailsTable,
		ConnectionDetailsTable, TrioDetailsTable, ConnectionDetailsTable, ConnectionDetailsTable,
		ConnectionDetailsTable, TrioDetailsTable, ConnectionDetailsTable, ConnectionDetailsTable,
		ConnectionDetailsTable, TrioDetailsTable, ConnectionDetailsTable, ConnectionDetailsTable,
		ConnectionDetailsTable, TrioDetailsTable, ConnectionDetailsTable, ConnectionDetailsTable,
		ConnectionDetailsTable, TrioDetailsTable, ConnectionDetailsTable, ConnectionDetailsTable,
		ConnectionDetailsTable, TrioDetailsTable, ConnectionDetailsTable, ConnectionDetailsTable)
	res.Insert = "(id, ctx_id, root," +
		" inter1, inter2, inter3, " +
		" conn11, last_inter11, next_main_conn11, next_inter_conn11," +
		" conn12, last_inter12, next_main_conn12, next_inter_conn12," +
		" conn21, last_inter21, next_main_conn21, next_inter_conn21," +
		" conn22, last_inter22, next_main_conn22, next_inter_conn22," +
		" conn31, last_inter31, next_main_conn31, next_inter_conn31," +
		" conn32, last_inter32, next_main_conn32, next_inter_conn32)" +
		" values ($1,$2,$3," +
		" $4,$5,$6," +
		" $7,$8,$9,$10," +
		" $11,$12,$13,$14," +
		" $15,$16,$17,$18," +
		" $19,$20,$21,$22," +
		" $23,$24,$25,$26," +
		" $27,$28,$29,$30)"
	res.SelectAll = fmt.Sprintf("select id, ctx_id, root,"+
		" inter1, inter2, inter3, "+
		" conn11, last_inter11, next_main_conn11, next_inter_conn11,"+
		" conn12, last_inter12, next_main_conn12, next_inter_conn12,"+
		" conn21, last_inter21, next_main_conn21, next_inter_conn21,"+
		" conn22, last_inter22, next_main_conn22, next_inter_conn22,"+
		" conn31, last_inter31, next_main_conn31, next_inter_conn31,"+
		" conn32, last_inter32, next_main_conn32, next_inter_conn32"+
		" from %s", PathBuildersTable)
	res.ExpectedCount = m3point.TotalNumberOfCubes
	return &res
}

/***************************************************************/
// trio Contexts Load and Save
/***************************************************************/

func (ppd *PointPackData) loadPathBuilders() []*m3point.RootPathNodeBuilder {
	_, rows := ppd.Env.SelectAllForLoad(PathBuildersTable)
	res := make([]*m3point.RootPathNodeBuilder, m3point.TotalNumberOfCubes+1)

	for rows.Next() {
		var cubeId, trioIndexId int
		var rootTrIdx int
		var intersTrIdx [3]int
		var connIds [3][2]int
		var lastIntersTrIdx [3][2]int
		var nextMainConnIds [3][2]int
		var nextInterConnIds [3][2]int
		err := rows.Scan(&cubeId, &trioIndexId, &rootTrIdx,
			&intersTrIdx[0], &intersTrIdx[1], &intersTrIdx[2],
			&connIds[0][0], &lastIntersTrIdx[0][0], &nextMainConnIds[0][0], &nextInterConnIds[0][0],
			&connIds[0][1], &lastIntersTrIdx[0][1], &nextMainConnIds[0][1], &nextInterConnIds[0][1],
			&connIds[1][0], &lastIntersTrIdx[1][0], &nextMainConnIds[1][0], &nextInterConnIds[1][0],
			&connIds[1][1], &lastIntersTrIdx[1][1], &nextMainConnIds[1][1], &nextInterConnIds[1][1],
			&connIds[2][0], &lastIntersTrIdx[2][0], &nextMainConnIds[2][0], &nextInterConnIds[2][0],
			&connIds[2][1], &lastIntersTrIdx[2][1], &nextMainConnIds[2][1], &nextInterConnIds[2][1])
		if err != nil {
			Log.Errorf("failed to load path builder line %d", len(res))
		} else {
			pathBuilderCtx := m3point.PathBuilderContext{GrowthCtx: ppd.GetGrowthContextById(trioIndexId), CubeId: cubeId}
			builder := m3point.RootPathNodeBuilder{}
			builder.Ctx = &pathBuilderCtx
			rootTd := ppd.GetTrioDetails(m3point.TrioIndex(rootTrIdx))
			builder.TrIdx = rootTd.GetId()
			for i, interTrIdx := range intersTrIdx {
				interPathNode := m3point.IntermediatePathNodeBuilder{}
				interPathNode.Ctx = builder.Ctx
				interPathNode.TrIdx = m3point.TrioIndex(interTrIdx)
				for j := 0; j < 2; j++ {
					lastPathNode := m3point.LastPathNodeBuilder{}
					lastPathNode.Ctx = builder.Ctx
					lastPathNode.TrIdx = m3point.TrioIndex(lastIntersTrIdx[i][j])
					lastPathNode.NextMainConnId = m3point.ConnectionId(nextMainConnIds[i][j])
					lastPathNode.NextInterConnId = m3point.ConnectionId(nextInterConnIds[i][j])
					interPathNode.PathLinks[j] = m3point.PathLinkBuilder{ConnId: m3point.ConnectionId(connIds[i][j]), PathNode: &lastPathNode}
				}
				builder.PathLinks[i] = m3point.PathLinkBuilder{ConnId: rootTd.Conns[i].GetId(), PathNode: &interPathNode}
			}
			res[cubeId] = &builder
		}
	}
	return res
}

func (ppd *PointPackData) saveAllPathBuilders() (int, error) {
	te, inserted, toFill, err := ppd.Env.GetForSaveAll(PathBuildersTable)
	if err != nil {
		return 0, err
	}
	if toFill {
		builders := ppd.calculateAllPathBuilders()
		if Log.IsDebug() {
			Log.Debugf("Populating table %s with %d elements", te.TableDef.Name, len(builders)-1)
		}
		for cubeId, rootNode := range builders {
			if cubeId == 0 {
				continue
			}
			interPNs := [3]*m3point.IntermediatePathNodeBuilder{}
			interConnIds := [3][2]m3point.ConnectionId{}
			lastInterPNs := [3][2]*m3point.LastPathNodeBuilder{}
			for i, pl := range rootNode.PathLinks {
				ipn, ok := pl.PathNode.(*m3point.IntermediatePathNodeBuilder)
				if !ok {
					err = m3db.MakeQsmErrorf("trying to convert path node to intermediate failed for %v", pl)
					return 0, err
				}
				interPNs[i] = ipn
				for j := 0; j < 2; j++ {
					ipl := ipn.PathLinks[j]
					interConnIds[i][j] = ipl.ConnId
					lipn, ok := ipl.PathNode.(*m3point.LastPathNodeBuilder)
					if !ok {
						err = m3db.MakeQsmErrorf("trying to convert path node to last intermediate failed for %v", ipl)
						return 0, err
					}
					lastInterPNs[i][j] = lipn
				}
			}
			err := te.Insert(cubeId, rootNode.Ctx.GrowthCtx.GetId(), rootNode.TrIdx,
				interPNs[0].TrIdx, interPNs[1].TrIdx, interPNs[2].TrIdx,
				interConnIds[0][0], lastInterPNs[0][0].TrIdx, lastInterPNs[0][0].NextMainConnId, lastInterPNs[0][0].NextInterConnId,
				interConnIds[0][1], lastInterPNs[0][1].TrIdx, lastInterPNs[0][1].NextMainConnId, lastInterPNs[0][1].NextInterConnId,
				interConnIds[1][0], lastInterPNs[1][0].TrIdx, lastInterPNs[1][0].NextMainConnId, lastInterPNs[1][0].NextInterConnId,
				interConnIds[1][1], lastInterPNs[1][1].TrIdx, lastInterPNs[1][1].NextMainConnId, lastInterPNs[1][1].NextInterConnId,
				interConnIds[2][0], lastInterPNs[2][0].TrIdx, lastInterPNs[2][0].NextMainConnId, lastInterPNs[2][0].NextInterConnId,
				interConnIds[2][1], lastInterPNs[2][1].TrIdx, lastInterPNs[2][1].NextMainConnId, lastInterPNs[2][1].NextInterConnId)
			if err != nil {
				Log.Error(err)
			} else {
				inserted++
			}
		}
	}
	return inserted, nil
}

func (ppd *PointPackData) calculateAllPathBuilders() []*m3point.RootPathNodeBuilder {
	ppd.CheckCubesInitialized()
	res := make([]*m3point.RootPathNodeBuilder, m3point.TotalNumberOfCubes+1)
	res[0] = nil
	for cubeKey, cubeId := range ppd.CubeIdsPerKey {
		root := m3point.RootPathNodeBuilder{}
		root.Ctx = &m3point.PathBuilderContext{GrowthCtx: ppd.GetGrowthContextById(cubeKey.GrowthCtxId), CubeId: cubeId}
		ppd.Populate(&root)
		res[cubeId] = &root
	}
	return res
}

func (ppd *PointPackData) Populate(rpnb *m3point.RootPathNodeBuilder) {
	growthCtx := rpnb.Ctx.GrowthCtx
	cubeKey := ppd.GetCubeById(rpnb.Ctx.CubeId)
	cube := cubeKey.Cube
	rpnb.TrIdx = cube.Center
	td := ppd.GetTrioDetails(rpnb.TrIdx)
	for i, cd := range td.Conns {
		// We are talking about the intermediate point here
		ip := cd.Vector

		// From each center out connection there 2 last PNB
		// They can be filled from the 2 unit directions of the base vector
		nextMains := [2]m3point.NextMainPathNode{}
		for j, ud := range cd.GetDirections() {
			nextMains[j].Ud = ud
			nmp := ud.GetFirstPoint()
			nextTrIdx := cube.GetCenterFaceTrio(ud)
			nextTd := ppd.GetTrioDetails(nextTrIdx)
			backConn := nextTd.GetOppositeConn(ud)
			nextMains[j].Lip = nmp.Add(backConn.Vector)
			nextMains[j].BackConn = backConn
			lipnb := m3point.LastPathNodeBuilder{}
			lipnb.Ctx = rpnb.Ctx
			lipnb.NextMainConnId = backConn.GetNegId()
			nextMains[j].Lipnb = &lipnb
		}

		// We have all the last nodes let's create the intermediate one
		// We have the three connections from ip to find the correct trio
		var iTd *m3point.TrioDetails
		ipConns := [2]*m3point.ConnectionDetails{ppd.GetConnDetailsByPoints(ip, nextMains[0].Lip), ppd.GetConnDetailsByPoints(ip, nextMains[1].Lip)}
		for _, possTd := range ppd.AllTrioDetails {
			if possTd.HasConnections(cd.GetNegId(), ipConns[0].GetId(), ipConns[1].GetId()) {
				iTd = possTd
				break
			}
		}
		if iTd == nil {
			Log.Fatalf("did not find any trio details matching %s %s %s in %s cube %s", cd.GetNegId(), ipConns[0].GetId(), ipConns[1].GetId(), growthCtx.String(), cube.String())
			return
		}

		ipnb := m3point.IntermediatePathNodeBuilder{}
		ipnb.Ctx = rpnb.Ctx
		ipnb.TrIdx = iTd.GetId()

		// Find the trio index for filling the last intermediate
		for j, nm := range nextMains {
			backUds := nm.BackConn.GetDirections()
			foundUd := false
			for _, backUd := range backUds {
				if backUd.GetOpposite() == nm.Ud {
					foundUd = true
				} else {
					nextInterTrIdx := cube.GetMiddleEdgeTrio(nm.Ud, backUd)
					nextInterTd := ppd.GetTrioDetails(nextInterTrIdx)
					nextInterBackConn := nextInterTd.GetOppositeConn(backUd)
					nextInterNearMainPoint := nm.Ud.GetFirstPoint().Add(backUd.GetFirstPoint()).Add(nextInterBackConn.Vector)
					lipToOtherConn := ppd.GetConnDetailsByPoints(nm.Lip, nextInterNearMainPoint)
					nm.Lipnb.NextInterConnId = lipToOtherConn.GetId()

					var liTd *m3point.TrioDetails
					for _, possTd := range ppd.AllTrioDetails {
						if possTd.HasConnections(ipConns[j].GetNegId(), nm.Lipnb.NextInterConnId, nm.Lipnb.NextMainConnId) {
							liTd = possTd
							break
						}
					}
					if liTd == nil {
						Log.Fatalf("did not find any trio details matching %s %s %s in %s cube %s", ipConns[j].GetNegId(), nm.Lipnb.NextInterConnId, nm.Lipnb.NextMainConnId, growthCtx.String(), cube.String())
						return
					}
					nm.Lipnb.TrIdx = liTd.GetId()
				}
			}
			if !foundUd {
				Log.Fatalf("direction mess between trio details %s %s and %d %v", td.String(), iTd.String(), nm.Ud, backUds)
			}
			nm.Lipnb.Verify()
			ipnb.PathLinks[j] = m3point.PathLinkBuilder{ConnId: ipConns[j].GetId(), PathNode: nm.Lipnb}
		}
		ipnb.Verify()

		rpnb.PathLinks[i] = m3point.PathLinkBuilder{ConnId: cd.Id, PathNode: &ipnb}
	}
	rpnb.Verify()
}

func (ppd *PointPackData) GetPathNodeBuilder(growthCtx m3point.GrowthContext, offset int, c m3point.Point) m3point.PathNodeBuilder {
	ppd.CheckPathBuildersInitialized()
	// TODO: Verify the key below stay local and is not staying in memory
	key := m3point.CubeKeyId{GrowthCtxId: growthCtx.GetId(), Cube: m3point.CreateTrioCube(ppd, growthCtx, offset, c)}
	cubeId := ppd.GetCubeIdByKey(key)
	return ppd.GetPathNodeBuilderById(cubeId)
}
