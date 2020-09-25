package spacedb

import (
	"github.com/freddy33/qsm-go/backend/m3db"
	"github.com/freddy33/qsm-go/backend/pathdb"
	"github.com/freddy33/qsm-go/m3util"
	"sync"
)

var Log = m3util.NewLogger("spacedb", m3util.INFO)

const (
	SpaceTable       = "spaces"
	SelectSpacePerId = 0
)

func init() {
	m3db.AddTableDef(createSpacesTableDef())
}

func InitializeSpaceDBEnv(env *m3db.QsmDbEnvironment) {
	pathdb.InitializePathDBEnv(env)
	createTablesEnv(env)
}

func createSpacesTableDef() *m3db.TableDefinition {
	res := m3db.TableDefinition{}
	res.Name = SpaceTable
	res.DdlColumns = "(id serial PRIMARY KEY," +
		" name VARCHAR(175) NOT NULL UNIQUE," +
		" active_path_node_threshold smallint NOT NULL," +
		" max_trios_per_point smallint NOT NULL," +
		" max_path_nodes_per_point smallint NOT NULL," +
		" max_coord integer NOT NULL," +
		" max_time integer NOT NULL)"
	allFields := "name,active_path_node_threshold,max_trios_per_point,max_path_nodes_per_point,max_coord,max_time"
	res.Insert = "(" + allFields + ") values ($1,$2,$3,$4,$5,0) returning id"
	res.SelectAll = "select id," + allFields + " from %s"
	res.ExpectedCount = -1
	res.Queries = make([]string, 1)
	res.Queries[SelectSpacePerId] = res.SelectAll + " where id=$1"

	return &res
}

func createTablesEnv(env *m3db.QsmDbEnvironment) {
	_, err := env.GetOrCreateTableExec(SpaceTable)
	if err != nil {
		Log.Fatalf("could not create table %s due to %v", SpaceTable, err)
		return
	}
}

/***************************************************************/
// Utility methods for test
/***************************************************************/

var dbMutex sync.Mutex
var testDbFilled [m3util.MaxNumberOfEnvironments]bool

func GetSpaceDbFullEnv(envId m3util.QsmEnvID) *m3db.QsmDbEnvironment {
	env := pathdb.GetPathDbFullEnv(envId)
	checkEnv(env)
	return env
}

func GetSpaceDbCleanEnv(envId m3util.QsmEnvID) *m3db.QsmDbEnvironment {
	env := pathdb.GetPathDbCleanEnv(envId)
	checkEnv(env)
	return env
}

func checkEnv(env *m3db.QsmDbEnvironment) {
	envId := env.GetId()
	dbMutex.Lock()
	defer dbMutex.Unlock()
	if !testDbFilled[envId] {
		createTablesEnv(env)
		testDbFilled[envId] = true
	}
}
