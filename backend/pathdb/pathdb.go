package pathdb

import (
	"github.com/freddy33/qsm-go/backend/m3db"
	"github.com/freddy33/qsm-go/backend/pointdb"
	"github.com/freddy33/qsm-go/m3util"
	"sync"
)

var Log = m3util.NewLogger("pathdb", m3util.INFO)

const (
	PointsTable       = "points"
	PathContextsTable = "path_contexts"
	PathNodesTable    = "path_nodes"
)

func init() {
	m3db.AddTableDef(createPointsTableDef())
	m3db.AddTableDef(createPathContextsTableDef())
	m3db.AddTableDef(creatPathNodesTableDef())
}

const (
	FindPointIdPerCoord = 0
	SelectPointPerId    = 1
)

func createPointsTableDef() *m3db.TableDefinition {
	res := m3db.TableDefinition{}
	res.Name = PointsTable
	res.DdlColumns = "(id bigserial PRIMARY KEY," +
		" x integer NOT NULL, y integer NOT NULL, z integer NOT NULL," +
		" CONSTRAINT points_x_y_z_key UNIQUE (x,y,z))"
	res.Insert = "(x,y,z) values ($1,$2,$3) returning id"
	res.SelectAll = "not to call select all on points"
	res.ExpectedCount = -1
	res.Queries = make([]string, 2)
	res.Queries[FindPointIdPerCoord] = "select id from %s where x=$1 and y=$2 and z=$3"
	res.Queries[SelectPointPerId] = "select x,y,z from %s where id=$1"

	res.ErrorFilter = func(err error) bool {
		return err.Error() == "pq: duplicate key value violates unique constraint \"points_x_y_z_key\""
	}
	return &res
}

const (
	UpdatePathBuilderId = 0
	UpdateMaxDist       = 1
)

func createPathContextsTableDef() *m3db.TableDefinition {
	res := m3db.TableDefinition{}
	res.Name = PathContextsTable
	res.DdlColumns = "(id serial PRIMARY KEY," +
		" growth_ctx_id smallint NOT NULL REFERENCES %s (id)," +
		" growth_offset smallint NOT NULL," +
		" path_builders_id smallint NULL REFERENCES %s (id)," +
		" max_dist integer NOT NULL DEFAULT 0," +
		" CONSTRAINT path_ctx_growth_param_key UNIQUE (growth_ctx_id, growth_offset))"
	res.DdlColumnsRefs = []string{
		pointdb.GrowthContextsTable, pointdb.PathBuildersTable}
	res.Insert = "(growth_ctx_id, growth_offset, path_builders_id) values ($1,$2,NULL) returning id"
	allFields := "id, growth_ctx_id, growth_offset, path_builders_id, max_dist"
	res.SelectAll = "select " + allFields + " from %s"
	res.ExpectedCount = 200
	res.Queries = make([]string, 2)
	res.Queries[UpdatePathBuilderId] = "update %s set path_builders_id = $2 where id = $1"
	res.Queries[UpdateMaxDist] = "update %s set max_dist = $2 where id = $1"
	return &res
}

const (
	SelectPathNodesById int = iota
	UpdatePathNode
	CountPathNodesByCtxAndDistance
	SelectPathNodesByCtxAndDistance
	CountPathNodesByCtxAndBetweenDistance
	SelectPathNodesByCtxAndBetweenDistance
	CountPathNodesByCtx
	SelectPathNodeIdByCtxAndPointId
)

func creatPathNodesTableDef() *m3db.TableDefinition {
	res := m3db.TableDefinition{}
	res.Name = PathNodesTable
	res.DdlColumns = "(id bigserial PRIMARY KEY," +
		" path_ctx_id integer NOT NULL REFERENCES %s (id)," +
		" path_builders_id smallint NOT NULL REFERENCES %s (id)," +
		" path_builder_idx smallint NOT NULL," +
		" trio_id smallint NOT NULL REFERENCES %s (id)," +
		" point_id bigint NOT NULL REFERENCES %s (id)," +
		" d integer NOT NULL DEFAULT 0," +
		" connection_mask smallint NOT NULL DEFAULT 0," +
		" path_node1 bigint NULL REFERENCES %s (id), path_node2 bigint NULL REFERENCES %s (id), path_node3 bigint NULL REFERENCES %s (id)," +
		" CONSTRAINT unique_point_per_path_ctx UNIQUE (path_ctx_id, point_id))"
	res.DdlColumnsRefs = []string{
		PathContextsTable, pointdb.PathBuildersTable, pointdb.TrioDetailsTable, PointsTable,
		PathNodesTable, PathNodesTable, PathNodesTable}
	res.ErrorFilter = func(err error) bool {
		return err.Error() == "pq: duplicate key value violates unique constraint \"unique_point_per_path_ctx\""
	}
	res.Insert = "(path_ctx_id, path_builders_id, path_builder_idx, trio_id, point_id, d," +
		" connection_mask," +
		" path_node1, path_node2, path_node3)" +
		" values ($1,$2,$3,$4,$5,$6," +
		" $7," +
		" $8,$9,$10) returning id"
	res.SelectAll = "not to call select all on node path"
	res.ExpectedCount = -1
	res.Queries = make([]string, 8)
	res.QueryTableRefs = make(map[int][]string, 1)
	selectAllFields := "id, path_ctx_id, path_builders_id, path_builder_idx, trio_id, point_id, d," +
		" connection_mask," +
		" path_node1, path_node2, path_node3"
	res.Queries[SelectPathNodesById] = "select " +
		selectAllFields +
		" from %s where id = $1"
	res.Queries[UpdatePathNode] = "update %s set" +
		" connection_mask = $2," +
		" path_node1 = $3, path_node2 = $4, path_node3 = $5" +
		" where id = $1"

	res.Queries[CountPathNodesByCtxAndDistance] = "select count(id)" +
		" from %s where path_ctx_id = $1 and d = $2"
	res.Queries[SelectPathNodesByCtxAndDistance] =
		"select " + PathNodesTable + "." + selectAllFields + ", x, y, z" +
			" from %s join %s on " + PointsTable + ".id = " + PathNodesTable + ".point_id" +
			" where path_ctx_id = $1 and d = $2"
	res.QueryTableRefs[SelectPathNodesByCtxAndDistance] = []string{PointsTable}

	res.Queries[CountPathNodesByCtxAndBetweenDistance] = "select count(id)" +
		" from %s where path_ctx_id = $1 and d >= $2 and d <= $3"
	res.Queries[SelectPathNodesByCtxAndBetweenDistance] =
		"select " + PathNodesTable + "." + selectAllFields + ", x, y, z" +
			" from %s join %s on " + PointsTable + ".id = " + PathNodesTable + ".point_id" +
			" where path_ctx_id = $1 and d >= $2 and d <= $3"
	res.QueryTableRefs[SelectPathNodesByCtxAndBetweenDistance] = []string{PointsTable}

	res.Queries[CountPathNodesByCtx] = "select count(*)" +
		" from %s where path_ctx_id = $1"

	res.Queries[SelectPathNodeIdByCtxAndPointId] = "select id " +
		" from %s where path_ctx_id = $1 and point_id = $2"

	return &res
}

func (pathData *ServerPathPackData) createTables() {
	tableNames := [3]string{PointsTable, PathContextsTable, PathNodesTable}
	pathTableExecs := [3]*m3db.TableExec{}

	// IMPORTANT: Create ALL the tables before preparing the queries
	var err error

	for i := 0; i < len(pathTableExecs); i++ {
		pathTableExecs[i], err = pathData.env.GetOrCreateTableExec(tableNames[i])
		if err != nil {
			Log.Fatal(err)
			return
		}
	}

	for i := 0; i < len(pathTableExecs); i++ {
		err = pathTableExecs[i].PrepareQueries()
		if err != nil {
			Log.Fatal(err)
			return
		}
	}

	pathData.pointsTe = pathTableExecs[0]
	pathData.pathCtxTe = pathTableExecs[1]
	pathData.pathNodesTe = pathTableExecs[2]
}

var dbMutex sync.Mutex
var testDbFilled [m3util.MaxNumberOfEnvironments]bool

func GetPathDbFullEnv(envId m3util.QsmEnvID) *m3db.QsmDbEnvironment {
	env := pointdb.GetPointDbFullEnv(envId)
	checkEnv(env)
	return env
}

func GetPathDbCleanEnv(envId m3util.QsmEnvID) *m3db.QsmDbEnvironment {
	env := pointdb.GetPointDbCleanEnv(envId)
	checkEnv(env)
	return env
}

func checkEnv(env *m3db.QsmDbEnvironment) {
	envId := env.GetId()
	dbMutex.Lock()
	defer dbMutex.Unlock()
	if !testDbFilled[envId] {
		pointData := pointdb.GetServerPointPackData(env)
		pointData.FillDb()
		GetServerPathPackData(env).createTables()
		testDbFilled[envId] = true
	}
}
