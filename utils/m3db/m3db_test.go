package m3db

import (
	"github.com/freddy33/qsm-go/utils/m3util"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TODO: test creation of DB tables

func silentDeleteFile(path string) {
	err := os.Remove(path)
	if err != nil {
		Log.Warnf("could not delete file %s due to %v", path, err)
	}
}

func TestDbConf(t *testing.T) {
	confDir := m3util.GetConfDir()
	assert.True(t, strings.HasSuffix(confDir, "conf"), "conf dir %s does end with conf", confDir)

	testConfFile := filepath.Join(confDir, "dbconn1234.json")
	dbTestFile := filepath.Join(confDir, "db-test.json")
	CopyFile(dbTestFile, testConfFile)

	defer silentDeleteFile(testConfFile)

	testConfEnv := m3util.QsmEnvID(1234)
	env := new(QsmDbEnvironment)
	env.Id = testConfEnv

	env.fillDbConf()
	connDetails := env.GetDbConf()
	assert.Equal(t, "hostTest", connDetails.Host, "fails reading %v", connDetails)
	assert.Equal(t, 1234, connDetails.Port, "fails reading %v", connDetails)
	assert.Equal(t, "userTest", connDetails.User, "fails reading %v", connDetails)
	assert.Equal(t, "passwordTest", connDetails.Password, "fails reading %v", connDetails)
	assert.Equal(t, "dbNameTest", connDetails.DbName, "fails reading %v", connDetails)
}

func TestEnvCreationAndDestroy(t *testing.T) {
	Log.SetDebug()
	env := m3util.GetEnvironment(m3util.DbTempEnv).(*QsmDbEnvironment)
	if env == nil {
		assert.NotNil(t, env, "could not create environment %d", m3util.DbTempEnv)
		return
	}
	defer env.Destroy()
	err := env.GetConnection().Ping()
	assert.True(t, err == nil, "Got ping error %v", err)
}
