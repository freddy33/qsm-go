package m3path

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3db"
	"github.com/freddy33/qsm-go/m3point"
	"math/rand"
	"strings"
	"sync"
	"time"
)

const (
	PointsTable = "points"
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

func createPathContextsTableDef() *m3db.TableDefinition {
	res := m3db.TableDefinition{}
	res.Name = PathContextsTable
	res.DdlColumns = fmt.Sprintf("(id serial PRIMARY KEY," +
		" growth_ctx_id smallint NOT NULL REFERENCES %s (id),"+
		" growth_offset smallint NOT NULL," +
		" path_builders_id smallint NOT NULL REFERENCES %s (id))",
		m3point.GrowthContextsTable, m3point.PathBuildersTable)
	res.Insert = "(growth_ctx_id, growth_offset, path_builders_id) values ($1,$2,$3) returning id"
	res.SelectAll = fmt.Sprintf("select id, growth_ctx_id, growth_offset, path_builders_id from %s", PathContextsTable)
	res.ExpectedCount = -1
	return &res
}

const (
	SelectPathNodesById = 0
	UpdatePathNode = 1
	SelectPathNodeByCtxAndDistance = 2
	SelectPathNodeByCtx = 3
)

func creatPathNodesTableDef() *m3db.TableDefinition {
	res := m3db.TableDefinition{}
	res.Name = PathNodesTable
	res.DdlColumns = fmt.Sprintf("(id serial PRIMARY KEY," +
		" path_ctx_id integer NOT NULL REFERENCES %s (id)," +
		" path_builders_id smallint NOT NULL REFERENCES %s (id)," +
		" trio_id smallint NOT NULL REFERENCES %s (id)," +
		" point_id bigint NOT NULL REFERENCES %s (id)," +
		" d integer NOT NULL DEFAULT 0," +
		" from1 integer NULL REFERENCES %s (id), from2 integer NULL REFERENCES %s (id), from3 integer NULL REFERENCES %s (id), " +
		" next1 integer NULL REFERENCES %s (id), next2 integer NULL REFERENCES %s (id), next3 integer NULL REFERENCES %s (id))",
		PathContextsTable, m3point.PathBuildersTable, m3point.TrioDetailsTable, PointsTable,
		PathNodesTable, PathNodesTable, PathNodesTable,
		PathNodesTable, PathNodesTable, PathNodesTable)
	res.Insert = "(path_ctx_id, path_builders_id, trio_id, point_id, d," +
		" from1, from2, from3," +
		" next1, next2, next3)" +
		" values ($1,$2,$3,$4,$5," +
		" $6,$7,$8," +
		" $9,$10,$11) returning id"
	res.SelectAll = "not to call select all on node path"
	res.ExpectedCount = -1
	res.Queries = make([]string, 4)
	res.Queries[SelectPathNodesById] = fmt.Sprintf("select path_ctx_id, path_builders_id, trio_id, point_id, d," +
		" from1, from2, from3," +
		" next1, next2, next3 from %s where id = $1", PathNodesTable)
	res.Queries[UpdatePathNode] = fmt.Sprintf("update %s set from1 = $2, from2 = $3, from3 = $4," +
		" next1 = $5, next2 = $6, next3 = $7 where id = $1", PathNodesTable)
	res.Queries[SelectPathNodeByCtxAndDistance] = fmt.Sprintf("select id, path_builders_id, trio_id, point_id," +
		" from1, from2, from3," +
		" next1, next2, next3 from %s where path_ctx_id = $1 and d = $2", PathNodesTable)
	res.Queries[SelectPathNodeByCtx] = fmt.Sprintf("select id, path_builders_id, trio_id, point_id, d," +
		" from1, from2, from3," +
		" next1, next2, next3 from %s where path_ctx_id = $1", PathNodesTable)
	return &res
}

func createTables() {
	createTablesEnv(GetPathEnv())
}

func GetOrCreatePoint(p m3point.Point) int64 {
	return GetOrCreatePointEnv(GetPathEnv(), p)
}

func GetPoint(pointId int64) *m3point.Point {
	return GetPointEnv(GetPathEnv(), pointId)
}

func GetPointEnv(env *m3db.QsmEnvironment, pointId int64) *m3point.Point {
	te, err :=env.GetOrCreateTableExec(PointsTable)
	if err != nil {
		Log.Errorf("could not get points table exec due to %v", err)
		return nil
	}
	rows, err := te.Query(SelectPointPerId, pointId)
	if err != nil {
		Log.Errorf("could not select points table exec due to %v", err)
		return nil
	}
	defer te.CloseRows(rows)
	if rows.Next() {
		res := m3point.Point{}
		err = rows.Scan(&res[0], &res[1], &res[2])
		if err != nil {
			Log.Errorf("Could not read row of %s due to %v", PointsTable, err)
		} else {
			return &res
		}
	}
	return nil
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

func GetOrCreatePointEnv(env *m3db.QsmEnvironment, p m3point.Point) int64 {
	te, err :=env.GetOrCreateTableExec(PointsTable)
	if err != nil {
		Log.Errorf("could not get points table exec due to %v", err)
		return -1
	}
	rows, err := te.Query(FindPointIdPerCoord, p.X(), p.Y(), p.Z())
	if err != nil {
		Log.Errorf("could not select points table exec due to %v", err)
		return -1
	}
	defer te.CloseRows(rows)
	var id int64
	if rows.Next() {
		err = rows.Scan(&id)
		if err != nil {
			Log.Errorf("could not convert points table id for %v due to %v", p, err)
			return -1
		}
		return id
	} else {
		id, err = te.InsertReturnId(p.X(), p.Y(), p.Z())
		if err == nil {
			return id
		} else {
			errorMessage := err.Error()
			if strings.Contains(errorMessage, "duplicate key") && strings.Contains(errorMessage, "points_x_y_z_key") {
				// got concurrent insert, let's just reselect
				rows, err = te.Query(FindPointIdPerCoord, p.X(), p.Y(), p.Z())
				if err != nil {
					Log.Errorf("could not select points table for %v after duplicate key insert exec due to %v", p, err)
					return -1
				}
				defer te.CloseRows(rows)
				if !rows.Next() {
					Log.Errorf("selecting points table for %v after duplicate key returns no rows!", p)
				}
				err = rows.Scan(&id)
				if err != nil {
					Log.Errorf("could not convert points table id for %v due to %v", p, err)
					return -1
				}
				return id
			} else {
				Log.Errorf("got unknown points table for %v error %v", p, err)
				return -1
			}
		}
	}
}

/***************************************************************/
// perf test main
/***************************************************************/
func RunInsertRandomPoints() {
	m3db.SetToTestMode()
	env := GetFullTestDb(m3db.PathTestEnv)
	// increase concurrency chance with low random
	rdMax := m3point.CInt(10)
	nbRoutines := 100
	nbRound := 250
	start := time.Now()
	wg := new(sync.WaitGroup)
	for r := 0; r < nbRoutines; r++ {
		wg.Add(1)
		go func() {
			for i := 0; i < nbRound; i++ {
				randomPoint := RandomPoint(rdMax)
				id := GetOrCreatePointEnv(env, randomPoint)
				if id <= 0 {
					Log.Errorf("failed to insert %v got %d id", randomPoint, id)
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
	Log.Infof("It took %v to create %d points with nb routines=%d max coord %d", time.Now().Sub(start), nbRoutines*nbRound, nbRoutines, rdMax)
}


/***************************************************************/
// Utility methods for test
/***************************************************************/

func RandomPoint(max m3point.CInt) m3point.Point {
	return m3point.Point{RandomCInt(max), RandomCInt(max), RandomCInt(max)}
}

func RandomCInt(max m3point.CInt) m3point.CInt {
	r := m3point.CInt(rand.Int31n(int32(max)))
	if rand.Float32() < 0.5 {
		return -r
	}
	return r
}

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
