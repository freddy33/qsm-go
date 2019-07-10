package m3path

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3db"
	"github.com/freddy33/qsm-go/m3point"
	"sync"
)

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

var pathEnvId m3db.QsmEnvID
var pathEnv *m3db.QsmEnvironment

func GetPathEnv() *m3db.QsmEnvironment {
	if pathEnv == nil || pathEnv.GetConnection() == nil {
		if pathEnvId == m3db.NoEnv {
			pathEnvId = m3db.GetDefaultEnvId()
		}
		pathEnv = m3db.GetEnvironment(pathEnvId)
	}
	return pathEnv
}

func InitializeDB() {
	m3point.InitializeDB()
	createTablesEnv(GetPathEnv())
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
	SelectPathNodesById            = 0
	UpdatePathNode                 = 1
	SelectPathNodeByCtxAndDistance = 2
	SelectPathNodeByCtx            = 3
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
		" blocked_mask smallint NOT NULL DEFAULT 0,"+
		" from1 bigint NULL REFERENCES %s (id), from2 bigint NULL REFERENCES %s (id), from3 bigint NULL REFERENCES %s (id),"+
		" next1 bigint NULL REFERENCES %s (id), next2 bigint NULL REFERENCES %s (id), next3 bigint NULL REFERENCES %s (id),"+
		" CONSTRAINT unique_point_per_path_ctx UNIQUE (path_ctx_id, point_id))",
		PathContextsTable, m3point.PathBuildersTable, m3point.TrioDetailsTable, PointsTable,
		PathNodesTable, PathNodesTable, PathNodesTable,
		PathNodesTable, PathNodesTable, PathNodesTable)
	res.Insert = "(path_ctx_id, path_builders_id, trio_id, point_id, d," +
		" blocked_mask," +
		" from1, from2, from3," +
		" next1, next2, next3)" +
		" values ($1,$2,$3,$4,$5," +
		" $6," +
		" $7,$8,$9," +
		" $10,$11,$12) returning id"
	res.SelectAll = "not to call select all on node path"
	res.ExpectedCount = -1
	res.Queries = make([]string, 4)
	selectAllFields := " id, path_ctx_id, path_builders_id, trio_id, point_id, d," +
		" blocked_mask," +
		" from1, from2, from3," +
		" next1, next2, next3 "
	res.Queries[SelectPathNodesById] = fmt.Sprintf("select "+
		selectAllFields +
		" from %s where id = $1", PathNodesTable)
	res.Queries[UpdatePathNode] = fmt.Sprintf("update %s set"+
		" blocked_mask = $2,"+
		" from1 = $3, from2 = $4, from3 = $5,"+
		" next1 = $6, next2 = $7, next3 = $8"+
		" where id = $1", PathNodesTable)
	res.Queries[SelectPathNodeByCtxAndDistance] = fmt.Sprintf("select "+
		selectAllFields +
		" from %s where path_ctx_id = $1 and d = $2", PathNodesTable)
	res.Queries[SelectPathNodeByCtx] = fmt.Sprintf("select "+
		selectAllFields +
		" from %s where path_ctx_id = $1", PathNodesTable)
	return &res
}

func createTablesEnv(env *m3db.QsmEnvironment) {
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
var testDbFilled [m3db.MaxNumberOfEnvironments]bool

func GetFullTestDb(envId m3db.QsmEnvID) *m3db.QsmEnvironment {
	env := m3point.GetFullTestDb(envId)
	pathEnvId = envId
	pathEnv = nil
	checkEnv(env)
	return env
}

func GetCleanTempDb(envId m3db.QsmEnvID) *m3db.QsmEnvironment {
	env := m3point.GetCleanTempDb(envId)
	checkEnv(env)
	return env
}

func checkEnv(env *m3db.QsmEnvironment) {
	envId := env.GetId()
	dbMutex.Lock()
	defer dbMutex.Unlock()
	if !testDbFilled[envId] {
		m3point.FillDbEnv(env)
		createTablesEnv(env)
		testDbFilled[envId] = true
	}
}
