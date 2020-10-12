package m3db

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	config "github.com/freddy33/qsm-go/backend/conf"
	"github.com/freddy33/qsm-go/m3util"
	_ "github.com/lib/pq"
)

var Log = m3util.NewLogger("m3db", m3util.INFO)

const (
	maxRetryCount = 5
	retryInterval = 5 * time.Second
)

type DbConnDetails struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DbName   string `json:"dbName"`
}

type QsmDbEnvironment struct {
	m3util.BaseQsmEnvironment
	dbDetails        DbConnDetails
	schemaName       string
	schemaChecked    bool
	db               *sql.DB
	createTableMutex sync.Mutex
	tableExecs       map[string]*TableExec

	dataCheckMutex [m3util.MaxDataEntry]sync.Mutex
	dataChecked    [m3util.MaxDataEntry]bool
}

func NewQsmDbEnvironment(config config.Config) *QsmDbEnvironment {
	env := QsmDbEnvironment{
		dbDetails: DbConnDetails{
			Host:     config.DBHost,
			Port:     config.DBPort,
			User:     config.DBUser,
			Password: config.DBPassword,
			DbName:   config.DBName,
		},
	}

	return &env
}

func (env *QsmDbEnvironment) GetConnection() *sql.DB {
	return env.db
}

func (env *QsmDbEnvironment) GetDbConf() DbConnDetails {
	return env.dbDetails
}

func (env *QsmDbEnvironment) GetSchemaName() string {
	return env.schemaName
}

func createNewDbEnv(envId m3util.QsmEnvID) m3util.QsmEnvironment {
	dbConf := config.NewDBConfig()
	env := NewQsmDbEnvironment(dbConf)

	env.Id = envId
	env.schemaName = "qsm" + envId.String()
	env.tableExecs = make(map[string]*TableExec)

	err := env.OpenDb()
	if err != nil {
		Log.Fatalf("Env %d failed to open DB: %v", envId, err)
	}

	if !env.Ping() {
		Log.Fatalf("Could not ping DB %d", envId)
	}

	return env
}

func GetEnvironment(envId m3util.QsmEnvID) *QsmDbEnvironment {
	return m3util.GetEnvironmentWithCreator(envId, createNewDbEnv).(*QsmDbEnvironment)
}

func (env *QsmDbEnvironment) OpenDb() error {
	connDetails := env.GetDbConf()
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		connDetails.Host, connDetails.Port, connDetails.User, connDetails.Password, connDetails.DbName)
	if Log.IsDebug() {
		Log.Debugf("Opening DB for environment %d is user=%s dbName=%s", env.GetId(), env.dbDetails.User, env.dbDetails.DbName)
	}
	var err error
	env.db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		return m3util.MakeWrapQsmErrorf(err, "fail to open DB for environment %d with user=%s and dbName=%s due to %v", env.GetId(), env.dbDetails.User, env.dbDetails.DbName, err)
	}
	if Log.IsDebug() {
		Log.Debugf("DB opened for environment %d is user=%s dbName=%s", env.GetId(), env.dbDetails.User, env.dbDetails.DbName)
	}
	return nil
}

func (env *QsmDbEnvironment) CloseDb() {
	db := env.db
	env.db = nil
	if db != nil {
		err := db.Close()
		if err != nil {
			Log.Errorf("Error while closing environment %d : %v", env.Id, err)
		}
	}
}

func (env *QsmDbEnvironment) Close() {
	envId := env.GetId()
	Log.Infof("Closing DB environment %d", envId)
	defer m3util.RemoveEnvFromMap(envId)
	env.CleanAllData()
	// clean table exec
	for tn, te := range env.tableExecs {
		err := te.Close()
		if err != nil {
			Log.Warnf("Closing table exec of envId=%d table=%s generated '%s'", env.GetId(), tn, err.Error())
		}
		delete(env.tableExecs, tn)
	}
}

func (env *QsmDbEnvironment) CheckSchema() error {
	if env.schemaChecked {
		return nil
	}

	// Check and create schema if needed
	dbName := env.dbDetails.DbName
	schemaName := env.schemaName

	db := env.GetConnection()
	if db == nil {
		return m3util.MakeQsmErrorf("Got a nil connection for %s", dbName)
	}

	if Log.IsDebug() {
		Log.Debugf("Creating schema %s", schemaName)
	}
	createQuery := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schemaName)
	_, err := db.Exec(createQuery)
	if err != nil {
		return m3util.MakeWrapQsmErrorf(err, "could not create schema %s using '%s' due to error %v", schemaName, createQuery, err)
	}
	err = env.setSearchPath()
	if err != nil {
		return err
	}
	if Log.IsDebug() {
		Log.Debugf("Schema %s created", schemaName)
	}
	env.schemaChecked = true
	return nil
}

func (env *QsmDbEnvironment) setSearchPath() error {
	_, err := env.db.Exec("SET search_path='" + env.schemaName + "'")
	if err != nil {
		return m3util.MakeWrapQsmErrorf(err, "could not set the search path for schema %s due to error %v", env.schemaName, err)
	}
	return nil
}

func (env *QsmDbEnvironment) dropSchema() {
	// Check and create schema if needed
	dbName := env.dbDetails.DbName
	schemaName := env.schemaName

	db := env.GetConnection()
	if db == nil {
		Log.Errorf("Got a nil connection for %s while trying to drop schema", dbName)
		return
	}

	// Set recheck right on
	env.schemaChecked = false

	if Log.IsDebug() {
		Log.Debugf("Dropping schema %s", schemaName)
	}
	dropQuery := fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName)
	_, err := db.Exec(dropQuery)
	if err != nil {
		Log.Errorf("could not drop schema %s using '%s' due to error %v", schemaName, dropQuery, err)
	}
	if Log.IsDebug() {
		Log.Debugf("Schema %s dropped", schemaName)
	}
}

func (env *QsmDbEnvironment) Destroy() {
	env.resetAllCheck()
	defer env.releaseAllCheck()
	env.dropSchema()
	env.Close()
}

func (env *QsmDbEnvironment) Ping() bool {
	currentRetryCount := 1
	for {
		err := env.GetConnection().Ping()
		if err == nil {
			Log.Debugf("ping for environment %d successful", env.GetId())
			return true
		}

		if currentRetryCount > maxRetryCount {
			Log.Errorf("failed to ping env %d on DB %s: %v", env.GetId(), env.dbDetails.DbName, err)
			return false
		}

		time.Sleep(retryInterval)
		Log.Warnf("retry(%d/%d): ping env %d on DB %s", currentRetryCount, maxRetryCount, env.GetId(), env.dbDetails.DbName)
		currentRetryCount += 1
	}

}

func (env *QsmDbEnvironment) DataChecked(dataIdx int) bool {
	return env.dataChecked[dataIdx]
}

func (env *QsmDbEnvironment) releaseAllCheck() {
	for i, _ := range env.dataCheckMutex {
		env.dataCheckMutex[i].Unlock()
	}
}

func (env *QsmDbEnvironment) resetAllCheck() {
	for i, _ := range env.dataCheckMutex {
		env.dataCheckMutex[i].Lock()
		env.dataChecked[i] = false
	}
}

func (env *QsmDbEnvironment) ExecOnce(dataIdx int, doInit func() error) error {
	if env.dataChecked[dataIdx] {
		return nil
	}
	env.dataCheckMutex[dataIdx].Lock()
	defer env.dataCheckMutex[dataIdx].Unlock()
	if env.dataChecked[dataIdx] {
		return nil
	}
	err := doInit()
	env.dataChecked[dataIdx] = true
	return err
}

