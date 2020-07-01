package m3path

import (
	"fmt"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/freddy33/qsm-go/utils/m3db"
	"github.com/freddy33/qsm-go/utils/m3util"
	"sync"
)

var Log = m3util.NewLogger("m3path", m3util.INFO)

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

func InitializeDBEnv(env *m3db.QsmDbEnvironment) {
	m3point.InitializeDBEnv(env, true)
	createTablesEnv(env)
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
	res.Queries[FindPointIdPerCoord] = fmt.Sprintf("select id from %s where x=$1 and y=$2 and z=$3", PointsTable)
	res.Queries[SelectPointPerId] = fmt.Sprintf("select x,y,z from %s where id=$1", PointsTable)

	res.ErrorFilter = func(err error) bool {
		return err.Error() == "pq: duplicate key value violates unique constraint \"points_x_y_z_key\""
	}
	return &res
}

const (
	SelectPathContextById = 0
	UpdatePathBuilderId   = 1
)

func createPathContextsTableDef() *m3db.TableDefinition {
	res := m3db.TableDefinition{}
	res.Name = PathContextsTable
	res.DdlColumns = fmt.Sprintf("(id serial PRIMARY KEY,"+
		" growth_ctx_id smallint NOT NULL REFERENCES %s (id),"+
		" growth_offset smallint NOT NULL,"+
		" path_builders_id smallint NULL REFERENCES %s (id))",
		m3point.GrowthContextsTable, m3point.PathBuildersTable)
	res.Insert = "(growth_ctx_id, growth_offset, path_builders_id) values ($1,$2,NULL) returning id"
	res.SelectAll = fmt.Sprintf("select id, growth_ctx_id, growth_offset, path_builders_id from %s", PathContextsTable)
	res.ExpectedCount = -1
	res.Queries = make([]string, 2)
	res.Queries[SelectPathContextById] = fmt.Sprintf("select growth_ctx_id, growth_offset, path_builders_id from %s where id = $1", PathContextsTable)
	res.Queries[UpdatePathBuilderId] = fmt.Sprintf("update %s set path_builders_id = $2 where id = $1", PathContextsTable)
	return &res
}

const (
	SelectPathNodesById int = iota
	UpdatePathNode
	SelectPathNodesByCtxAndDistance
	PathNodeIdsBefore
	ConnectedPathNodeIds
	CountPathNodesByCtx
	SelectPathNodeIdByCtxAndPointId
	SelectPathNodesByPoint
)

func creatPathNodesTableDef() *m3db.TableDefinition {
	res := m3db.TableDefinition{}
	res.Name = PathNodesTable
	res.DdlColumns = fmt.Sprintf("(id bigserial PRIMARY KEY,"+
		" path_ctx_id integer NOT NULL REFERENCES %s (id),"+
		" path_builders_id smallint NOT NULL REFERENCES %s (id),"+
		" trio_id smallint NOT NULL REFERENCES %s (id),"+
		" point_id bigint NOT NULL REFERENCES %s (id),"+
		" d integer NOT NULL DEFAULT 0,"+
		" connection_mask smallint NOT NULL DEFAULT 0,"+
		" path_node1 bigint NULL REFERENCES %s (id), path_node2 bigint NULL REFERENCES %s (id), path_node3 bigint NULL REFERENCES %s (id),"+
		" CONSTRAINT unique_point_per_path_ctx UNIQUE (path_ctx_id, point_id))",
		PathContextsTable, m3point.PathBuildersTable, m3point.TrioDetailsTable, PointsTable,
		PathNodesTable, PathNodesTable, PathNodesTable)
	res.ErrorFilter = func(err error) bool {
		return err.Error() == "pq: duplicate key value violates unique constraint \"unique_point_per_path_ctx\""
	}
	res.Insert = "(path_ctx_id, path_builders_id, trio_id, point_id, d," +
		" connection_mask," +
		" path_node1, path_node2, path_node3)" +
		" values ($1,$2,$3,$4,$5," +
		" $6," +
		" $7,$8,$9) returning id"
	res.SelectAll = "not to call select all on node path"
	res.ExpectedCount = -1
	res.Queries = make([]string, 8)
	selectAllFields := " id, path_ctx_id, path_builders_id, trio_id, point_id, d," +
		" connection_mask," +
		" path_node1, path_node2, path_node3"
	res.Queries[SelectPathNodesById] = fmt.Sprintf("select "+
		selectAllFields+
		" from %s where id = $1", PathNodesTable)
	res.Queries[UpdatePathNode] = fmt.Sprintf("update %s set"+
		" connection_mask = $2,"+
		" path_node1 = $3, path_node2 = $4, path_node3 = $5"+
		" where id = $1", PathNodesTable)
	res.Queries[SelectPathNodesByCtxAndDistance] = fmt.Sprintf("select "+
		selectAllFields+
		" from %s where path_ctx_id = $1 and d = $2", PathNodesTable)
	res.Queries[PathNodeIdsBefore] = fmt.Sprintf("select point_id, id, d"+
		" from %s where path_ctx_id = $1 and d < $2 and d >= $3", PathNodesTable)
	res.Queries[PathNodeIdsBefore] = fmt.Sprintf("select point_id, id, d"+
		" from %s where path_ctx_id = $1 and d < $2 and d >= $3", PathNodesTable)
	res.Queries[ConnectedPathNodeIds] = fmt.Sprintf("select id"+
		" from %s where path_ctx_id = $1 and d = $2 and (path_node1 = $3 or path_node2 = $3 or path_node3 = $3)", PathNodesTable)
	res.Queries[CountPathNodesByCtx] = fmt.Sprintf("select count(*)"+
		" from %s where path_ctx_id = $1", PathNodesTable)
	res.Queries[SelectPathNodeIdByCtxAndPointId] = fmt.Sprintf("select id "+
		" from %s where path_ctx_id = $1 and point_id = $2", PathNodesTable)
	res.Queries[SelectPathNodesByPoint] = fmt.Sprintf("select "+
		selectAllFields+
		" from %s where point_id = $1", PathNodesTable)
	return &res
}

func createTablesEnv(env *m3db.QsmDbEnvironment) {
	_, err := env.GetOrCreateTableExec(PointsTable)
	if err != nil {
		Log.Fatalf("could not create table %s due to %v", PointsTable, err)
		return
	}
	_, err = env.GetOrCreateTableExec(PathContextsTable)
	if err != nil {
		Log.Fatalf("could not create table %s due to %v", PathContextsTable, err)
		return
	}
	_, err = env.GetOrCreateTableExec(PathNodesTable)
	if err != nil {
		Log.Fatalf("could not create table %s due to %v", PathNodesTable, err)
		return
	}
}

/***************************************************************/
// Utility methods for test
/***************************************************************/

var dbMutex sync.Mutex
var testDbFilled [m3util.MaxNumberOfEnvironments]bool

func GetFullTestDb(envId m3util.QsmEnvID) *m3db.QsmDbEnvironment {
	env := m3point.GetFullTestDb(envId)
	checkEnv(env)
	return env
}

func GetCleanTempDb(envId m3util.QsmEnvID) *m3db.QsmDbEnvironment {
	env := m3point.GetCleanTempDb(envId)
	checkEnv(env)
	return env
}

func checkEnv(env *m3db.QsmDbEnvironment) {
	envId := env.GetId()
	dbMutex.Lock()
	defer dbMutex.Unlock()
	if !testDbFilled[envId] {
		m3point.FillDbEnv(env)
		createTablesEnv(env)
		testDbFilled[envId] = true
	}
}
