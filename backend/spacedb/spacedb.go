package spacedb

import (
	"github.com/freddy33/qsm-go/backend/m3db"
	"github.com/freddy33/qsm-go/backend/pathdb"
	"github.com/freddy33/qsm-go/backend/pointdb"
	"github.com/freddy33/qsm-go/m3util"
	"sync"
)

var Log = m3util.NewLogger("spacedb", m3util.INFO)

const (
	SpacesTable      = "spaces"
	EventsTable      = "events"
	NodesTable       = "nodes"
	SelectSpacePerId = 0
)

func init() {
	m3db.AddTableDef(createSpacesTableDef())
	m3db.AddTableDef(createEventsTableDef())
	m3db.AddTableDef(createNodesTableDef())
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
		" max_node_time integer NOT NULL)"
	allFields := "name, active_path_node_threshold, max_trios_per_point, max_path_nodes_per_point, max_coord, max_node_time"
	res.Insert = "(" + allFields + ") values ($1,$2,$3,$4,$5,$6) returning id"
	res.SelectAll = "select id," + allFields + " from %s"
	res.ExpectedCount = -1
	res.Queries = make([]string, 1)
	res.Queries[SelectSpacePerId] = res.SelectAll + " where id=$1"

	return &res
}

const (
	SelectEventPerId int = iota
	SelectEventsPerSpace
	UpdateMaxNodeTime
)

func createEventsTableDef() *m3db.TableDefinition {
	res := m3db.TableDefinition{}
	res.Name = EventsTable
	// End time set equal to creation time when alive
	res.DdlColumns = "(id serial PRIMARY KEY," +
		" space_id integer NOT NULL REFERENCES %s (id)," +
		" path_ctx_id integer NOT NULL REFERENCES %s (id)," +
		" creation_time integer NOT NULL," +
		" color smallint NOT NULL," +
		" end_time integer NOT NULL," +
		" max_node_time integer NOT NULL)"
	res.DdlColumnsRefs = []string{SpacesTable, pathdb.PathContextsTable}

	res.Insert = "(space_id, path_ctx_id, creation_time, color, end_time, max_node_time) values ($1,$2,$3,$4,$5,$6) returning id"
	res.SelectAll = "no select all events"
	res.ExpectedCount = -1
	res.Queries = make([]string, 3)
	res.QueryTableRefs = make(map[int][]string, 1)
	res.Queries[SelectEventPerId] = "select id, space_id, path_ctx_id, creation_time, color, end_time, max_node_time from %s where id=$1"
	res.Queries[SelectEventsPerSpace] =
		"select " + EventsTable + ".id, path_ctx_id, " + EventsTable + ".creation_time," +
			" color, end_time, max_node_time, " +
			NodesTable + ".id, path_node_id, trio_id, point_id," +
			" connection_mask, node1, node2, node3, " +
			" x, y, z" +
			" from %s" +
			" join %s on " + NodesTable + ".event_id = " + EventsTable + ".id " +
			" join %s on " + pathdb.PointsTable + ".id = " + NodesTable + ".point_id " +
			" where " + EventsTable + ".space_id = $1 and " + NodesTable + ".d = 0"
	res.QueryTableRefs[SelectEventsPerSpace] = []string{NodesTable, pathdb.PointsTable}
	res.Queries[UpdateMaxNodeTime] = "update %s set max_node_time = $2 where id = $1"

	return &res
}

const (
	SelectNodesAt      int = iota
	SelectNodesBetween int = iota
)

/*
How to query graph in PG: https://www.postgresql.org/docs/current/queries-with.html
*/
func createNodesTableDef() *m3db.TableDefinition {
	res := m3db.TableDefinition{}
	res.Name = NodesTable
	res.DdlColumns = "(id bigserial PRIMARY KEY," +
		" event_id integer NOT NULL REFERENCES %s (id)," +
		" path_node_id bigint NOT NULL REFERENCES %s (id)," +
		" trio_id smallint NOT NULL REFERENCES %s (id)," +
		" point_id bigint NOT NULL REFERENCES %s (id)," +
		" d integer NOT NULL," +
		" creation_time integer NOT NULL," +
		" connection_mask smallint NOT NULL DEFAULT 0," +
		" node1 bigint NULL REFERENCES %s (id), node2 bigint NULL REFERENCES %s (id), node3 bigint NULL REFERENCES %s (id)," +
		" CONSTRAINT unique_point_per_event_node UNIQUE (event_id, point_id))"
	res.DdlColumnsRefs = []string{EventsTable, pathdb.PathNodesTable, pointdb.TrioDetailsTable, pathdb.PointsTable,
		NodesTable, NodesTable, NodesTable}

	allFields := "event_id, path_node_id, trio_id, point_id, d, creation_time," +
		" connection_mask, node1, node2, node3"

	res.Insert = "(" + allFields + ") values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) returning id"
	res.SelectAll = "no select all for nodes"
	res.ExpectedCount = -1
	res.Queries = make([]string, 2)
	res.QueryTableRefs = make(map[int][]string, 2)
	selAll := "select " + NodesTable + ".id," + allFields + ", x, y, z" +
		" from %s" +
		" join %s on " + pathdb.PointsTable + ".id = " + NodesTable + ".point_id "
	res.Queries[SelectNodesAt] = selAll +
		" where event_id=$1 and creation_time = $2"
	res.QueryTableRefs[SelectNodesAt] = []string{pathdb.PointsTable}
	res.Queries[SelectNodesBetween] = selAll +
		" where event_id=$1 and creation_time >= $2 and creation_time <= $3"
	res.QueryTableRefs[SelectNodesBetween] = []string{pathdb.PointsTable}
	return &res
}

func (spaceData *ServerSpacePackData) createTables() {
	tableNames := [3]string{SpacesTable, EventsTable, NodesTable}
	spaceTableExecs := [3]*m3db.TableExec{}

	// IMPORTANT: Create ALL the tables before preparing the queries
	var err error

	for i := 0; i < len(tableNames); i++ {
		spaceTableExecs[i], err = spaceData.env.GetOrCreateTableExec(tableNames[i])
		if err != nil {
			Log.Fatal(err)
			return
		}
	}

	for i := 0; i < len(tableNames); i++ {
		err = spaceTableExecs[i].PrepareQueries()
		if err != nil {
			Log.Fatal(err)
			return
		}
	}

	spaceData.spacesTe = spaceTableExecs[0]
	spaceData.eventsTe = spaceTableExecs[1]
	spaceData.nodesTe = spaceTableExecs[2]

}

var dbMutex sync.Mutex
var testDbFilled [m3util.MaxNumberOfEnvironments]bool

func GetSpaceDbFullEnv(envId m3util.QsmEnvID) *m3db.QsmDbEnvironment {
	env := pathdb.GetPathDbFullEnv(envId)
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
