package pointdb

import (
	"github.com/freddy33/qsm-go/backend/m3db"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3point"
	"sort"
)

const (
	TrioCubesTable = "trio_cubes"
)

type CubeListBuilder struct {
	ppd       *ServerPointPackData
	growthCtx m3point.GrowthContext
	allCubes  []CubeOfTrioIndex
}

func init() {
	m3db.AddTableDef(createContextCubesTableDef())
}

func createContextCubesTableDef() *m3db.TableDefinition {
	res := m3db.TableDefinition{}
	res.Name = TrioCubesTable
	res.DdlColumns = "(id smallint PRIMARY KEY," +
		" ctx_id smallint REFERENCES %s (id)," +
		" center smallint REFERENCES %s (id)," +
		" center_faces_PX smallint REFERENCES %s (id), center_faces_MX smallint REFERENCES %s (id)," +
		" center_faces_PY smallint REFERENCES %s (id), center_faces_MY smallint REFERENCES %s (id)," +
		" center_faces_PZ smallint REFERENCES %s (id), center_faces_MZ smallint REFERENCES %s (id)," +
		" middle_edges_PXPY smallint REFERENCES %s (id), middle_edges_PXMY smallint REFERENCES %s (id), middle_edges_PXPZ smallint REFERENCES %s (id), middle_edges_PXMZ smallint REFERENCES %s (id)," +
		" middle_edges_MXPY smallint REFERENCES %s (id), middle_edges_MXMY smallint REFERENCES %s (id), middle_edges_MXPZ smallint REFERENCES %s (id), middle_edges_MXMZ smallint REFERENCES %s (id)," +
		" middle_edges_PYPZ smallint REFERENCES %s (id), middle_edges_PYMZ smallint REFERENCES %s (id), middle_edges_MYPZ smallint REFERENCES %s (id), middle_edges_MYMZ smallint REFERENCES %s (id))"
	res.DdlColumnsRefs = []string{
		GrowthContextsTable,
		TrioDetailsTable,
		TrioDetailsTable, TrioDetailsTable,
		TrioDetailsTable, TrioDetailsTable,
		TrioDetailsTable, TrioDetailsTable,
		TrioDetailsTable, TrioDetailsTable, TrioDetailsTable, TrioDetailsTable,
		TrioDetailsTable, TrioDetailsTable, TrioDetailsTable, TrioDetailsTable,
		TrioDetailsTable, TrioDetailsTable, TrioDetailsTable, TrioDetailsTable}
	res.Insert = "(id, ctx_id, center," +
		" center_faces_PX, center_faces_MX, center_faces_PY, center_faces_MY, center_faces_PZ, center_faces_MZ, " +
		" middle_edges_PXPY, middle_edges_PXMY, middle_edges_PXPZ, middle_edges_PXMZ, " +
		" middle_edges_MXPY, middle_edges_MXMY, middle_edges_MXPZ, middle_edges_MXMZ, " +
		" middle_edges_PYPZ, middle_edges_PYMZ, middle_edges_MYPZ, middle_edges_MYMZ)" +
		" values ($1,$2,$3," +
		" $4,$5,$6,$7,$8,$9," +
		" $10,$11,$12,$13," +
		" $14,$15,$16,$17," +
		" $18,$19,$20,$21)"
	res.SelectAll = "select id, ctx_id, center," +
		" center_faces_PX, center_faces_MX, center_faces_PY, center_faces_MY, center_faces_PZ, center_faces_MZ, " +
		" middle_edges_PXPY, middle_edges_PXMY, middle_edges_PXPZ, middle_edges_PXMZ, " +
		" middle_edges_MXPY, middle_edges_MXMY, middle_edges_MXPZ, middle_edges_MXMZ, " +
		" middle_edges_PYPZ, middle_edges_PYMZ, middle_edges_MYPZ, middle_edges_MYMZ" +
		" from %s"
	res.ExpectedCount = TotalNumberOfCubes
	return &res
}

/***************************************************************/
// ServerPointPackData Functions for Cubes Load and Save
/***************************************************************/

func (ppd *ServerPointPackData) loadContextCubes() error {
	te := ppd.trioCubesTe
	rows, err := te.SelectAllForLoad()
	if err != nil {
		return err
	}
	res := make(map[CubeKeyId]int, te.TableDef.ExpectedCount)

	loaded := 0
	for rows.Next() {
		var cubeId int
		var growthCtxId int
		cube := CubeOfTrioIndex{}
		err := rows.Scan(&cubeId, &growthCtxId, &cube.Center,
			&cube.CenterFaces[0], &cube.CenterFaces[1], &cube.CenterFaces[2], &cube.CenterFaces[3], &cube.CenterFaces[4], &cube.CenterFaces[5],
			&cube.MiddleEdges[0], &cube.MiddleEdges[1], &cube.MiddleEdges[2], &cube.MiddleEdges[3],
			&cube.MiddleEdges[4], &cube.MiddleEdges[5], &cube.MiddleEdges[6], &cube.MiddleEdges[7],
			&cube.MiddleEdges[8], &cube.MiddleEdges[9], &cube.MiddleEdges[10], &cube.MiddleEdges[11])
		if err != nil {
			return m3util.MakeWrapQsmErrorf(err, "failed to load trio context line %d due to %v", loaded, err)
		} else {
			key := CubeKeyId{GrowthCtxId: growthCtxId, Cube: cube}
			res[key] = cubeId
		}
		loaded++
	}

	ppd.cubeIdsPerKey = res
	ppd.cubesLoaded = true

	return nil
}

func (ppd *ServerPointPackData) saveAllContextCubes() (int, error) {
	te := ppd.trioCubesTe
	inserted, toFill, err := te.GetForSaveAll()
	if err != nil {
		return 0, err
	}
	if toFill {
		cubeKeys := ppd.calculateAllContextCubes()
		if Log.IsDebug() {
			Log.Debugf("Populating table %s with %d elements", te.GetFullTableName(), len(cubeKeys))
		}
		for cubeKey, cubeId := range cubeKeys {
			cube := cubeKey.Cube
			err := te.Insert(cubeId, cubeKey.GrowthCtxId, cube.Center,
				cube.CenterFaces[0], cube.CenterFaces[1], cube.CenterFaces[2], cube.CenterFaces[3], cube.CenterFaces[4], cube.CenterFaces[5],
				cube.MiddleEdges[0], cube.MiddleEdges[1], cube.MiddleEdges[2], cube.MiddleEdges[3],
				cube.MiddleEdges[4], cube.MiddleEdges[5], cube.MiddleEdges[6], cube.MiddleEdges[7],
				cube.MiddleEdges[8], cube.MiddleEdges[9], cube.MiddleEdges[10], cube.MiddleEdges[11])
			if err != nil {
				return inserted, err
			} else {
				inserted++
			}
		}
		te.SetFilled()
	}
	return inserted, nil
}

/***************************************************************/
// CubeListBuilder Functions
/***************************************************************/

func (ppd *ServerPointPackData) calculateAllContextCubes() map[CubeKeyId]int {
	res := make(map[CubeKeyId]int, TotalNumberOfCubes)
	cubeIdx := 1
	for _, growthCtx := range ppd.GetAllGrowthContexts() {
		cl := CubeListBuilder{ppd: ppd, growthCtx: growthCtx}
		switch growthCtx.GetGrowthType() {
		case 1:
			cl.populate(1)
		case 3:
			cl.populate(6)
		case 2:
			cl.populate(1)
		case 4:
			cl.populate(4)
		case 8:
			cl.populate(8)
		}
		sort.Slice(cl.allCubes, func(i, j int) bool {
			c1 := cl.allCubes[i]
			c2 := cl.allCubes[j]
			centerDiff := int(c1.Center) - int(c2.Center)
			if centerDiff != 0 {
				return centerDiff < 0
			}
			for cfIdx := 0; cfIdx < len(c1.CenterFaces); cfIdx++ {
				cfDiff := int(c1.CenterFaces[cfIdx]) - int(c2.CenterFaces[cfIdx])
				if cfDiff != 0 {
					return cfDiff < 0
				}
			}
			for meIdx := 0; meIdx < len(c1.MiddleEdges); meIdx++ {
				meDiff := int(c1.MiddleEdges[meIdx]) - int(c2.MiddleEdges[meIdx])
				if meDiff != 0 {
					return meDiff < 0
				}
			}
			return false
		})
		for _, cube := range cl.allCubes {
			key := CubeKeyId{GrowthCtxId: growthCtx.GetId(), Cube: cube}
			_, alreadyIn := res[key]
			if !alreadyIn {
				res[key] = cubeIdx
				cubeIdx++
			}
		}
	}
	return res
}

func (cl *CubeListBuilder) populate(max m3point.CInt) {
	allCubesMap := make(map[CubeOfTrioIndex]int)
	// For center populate for all offsets
	maxOffset := cl.growthCtx.GetGrowthType().GetMaxOffset()
	for offset := 0; offset < maxOffset; offset++ {
		cube := CreateTrioCube(cl.ppd, cl.growthCtx, offset, Origin)
		allCubesMap[cube]++
	}
	// Go through space
	for x := -max; x <= max; x++ {
		for y := -max; y <= max; y++ {
			for z := -max; z <= max; z++ {
				cube := CreateTrioCube(cl.ppd, cl.growthCtx, 0, m3point.Point{x, y, z}.Mul(m3point.THREE))
				allCubesMap[cube]++
			}
		}
	}
	cl.allCubes = make([]CubeOfTrioIndex, len(allCubesMap))
	idx := 0
	for c := range allCubesMap {
		cl.allCubes[idx] = c
		idx++
	}
}

func (cl *CubeListBuilder) exists(offset int, c m3point.Point) bool {
	toFind := CreateTrioCube(cl.ppd, cl.growthCtx, offset, c)
	for _, c := range cl.allCubes {
		if c == toFind {
			return true
		}
	}
	return false
}

func (ppd *ServerPointPackData) getCubeList(growthCtx m3point.GrowthContext) *CubeListBuilder {
	ppd.CheckCubesInitialized()
	res := CubeListBuilder{ppd: ppd, growthCtx: growthCtx, allCubes: make([]CubeOfTrioIndex, 0, 100)}
	for cubeKey := range ppd.cubeIdsPerKey {
		if cubeKey.GrowthCtxId == growthCtx.GetId() {
			res.allCubes = append(res.allCubes, cubeKey.Cube)
		}
	}
	return &res
}
