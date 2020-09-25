package spacedb

import (
	"github.com/freddy33/qsm-go/backend/m3db"
	"github.com/freddy33/qsm-go/backend/pathdb"
	"github.com/freddy33/qsm-go/m3util"
	"sync"
)

var Log = m3util.NewLogger("spacedb", m3util.INFO)

const (
	SpacesTable         = "spaces"
	EventsTable         = "events"
	NodesTable          = "nodes"
	SelectSpacePerId    = 0
	SelectEventPerId    = 0
	SelectActiveEvents  = 1
	SelectNodePerId     = 0
	SelectNodesPerD     = 1
	SelectNodesPerPoint = 2
)

func init() {
	m3db.AddTableDef(createSpacesTableDef())
	m3db.AddTableDef(createEventsTableDef())
	m3db.AddTableDef(createNodesTableDef())
}

func InitializeSpaceDBEnv(env *m3db.QsmDbEnvironment) {
	pathdb.InitializePathDBEnv(env)
	GetServerSpacePackData(env).createTables()
}

func createSpacesTableDef() *m3db.TableDefinition {
	res := m3db.TableDefinition{}
	res.Name = SpacesTable
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

func createEventsTableDef() *m3db.TableDefinition {
	res := m3db.TableDefinition{}
	res.Name = EventsTable
	// End time set equal to creation time when alive
	res.DdlColumns = "(id serial PRIMARY KEY," +
		" space_id integer NOT NULL REFERENCES %s (id)," +
		" path_ctx_id integer NOT NULL REFERENCES %s (id)," +
		" creation_time integer NOT NULL," +
		" color smallint NOT NULL," +
		" end_time integer NOT NULL)"
	res.DdlColumnsRefs = []string{SpacesTable, pathdb.PathContextsTable}

	allFields := "space_id, path_ctx_id, creation_time, color, end_time"
	res.Insert = "(" + allFields + ") values ($1,$2,$3,$4,$5) returning id"
	res.SelectAll = "select id," + allFields + " from %s"
	res.ExpectedCount = -1
	res.Queries = make([]string, 1)
	res.Queries[SelectEventPerId] = res.SelectAll + " where id=$1"

	return &res
}

func createNodesTableDef() *m3db.TableDefinition {
	res := m3db.TableDefinition{}
	res.Name = NodesTable
	res.DdlColumns = "(id bigserial PRIMARY KEY," +
		" event_id integer NOT NULL REFERENCES %s (id)," +
		" path_node_id bigint NOT NULL REFERENCES %s (id)," +
		" point_id bigint NOT NULL REFERENCES %s (id)," +
		" d integer NOT NULL," +
		" creation_time integer NOT NULL," +
		" connection_mask smallint NOT NULL DEFAULT 0," +
		" node1 bigint NULL REFERENCES %s (id), node2 bigint NULL REFERENCES %s (id), node3 bigint NULL REFERENCES %s (id)," +
		" CONSTRAINT unique_point_per_event_node UNIQUE (event_id, point_id))"
	res.DdlColumnsRefs = []string{EventsTable, pathdb.PathNodesTable, pathdb.PointsTable,
		NodesTable, NodesTable, NodesTable}

	allFields := "event_id, path_node_id, point_id, d, creation_time, connection_mask, node1, node2, node3"
	res.Insert = "(" + allFields + ") values ($1,$2,$3,$4,$5,$6,$7,$8,$9) returning id"
	res.SelectAll = "select id," + allFields + " from %s"
	res.ExpectedCount = -1
	res.Queries = make([]string, 3)
	res.Queries[SelectNodePerId] = res.SelectAll + " where id=$1"
	res.Queries[SelectNodesPerD] = res.SelectAll + " where event_id=$1 and d=$2"
	res.Queries[SelectNodesPerPoint] = res.SelectAll + " where point_id=$1"
	return &res
}

func (spd *ServerSpacePackData) createTables() {
	var err error
	spd.spacesTe, err = spd.env.GetOrCreateTableExec(SpacesTable)
	if err != nil {
		Log.Fatalf("could not create table %s due to %v", SpacesTable, err)
		return
	}
	spd.eventsTe, err = spd.env.GetOrCreateTableExec(EventsTable)
	if err != nil {
		Log.Fatalf("could not create table %s due to %v", EventsTable, err)
		return
	}
	spd.nodesTe, err = spd.env.GetOrCreateTableExec(NodesTable)
	if err != nil {
		Log.Fatalf("could not create table %s due to %v", NodesTable, err)
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
		GetServerSpacePackData(env).createTables()
		testDbFilled[envId] = true
	}
}
