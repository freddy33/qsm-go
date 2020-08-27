package m3db

import (
	"testing"

	config "github.com/freddy33/qsm-go/backend/conf"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestDbConf(t *testing.T) {
	config := config.Config{
		DBHost:     "test-host",
		DBPort:     1234,
		DBUser:     "test-user",
		DBPassword: "test-password",
		DBName:     "test-db",
	}

	env := NewQsmDbEnvironment(config)

	connDetails := env.GetDbConf()
	assert.Equal(t, config.DBHost, connDetails.Host)
	assert.Equal(t, config.DBPort, connDetails.Port)
	assert.Equal(t, config.DBUser, connDetails.User)
	assert.Equal(t, config.DBPassword, connDetails.Password)
	assert.Equal(t, config.DBName, connDetails.DbName)
}
