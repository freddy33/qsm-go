package m3db

import (
	"database/sql"
	"fmt"
	"github.com/freddy33/qsm-go/m3util"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"log"
	"strings"
	"testing"
)

func TestDbConf(t *testing.T) {
	confDir := m3util.GetConfDir()
	assert.True(t, strings.HasSuffix(confDir, "conf"), "conf dir %s does end with conf", confDir)

	connDetails := readDbConf(1)
	fmt.Println(connDetails)
}

func TestDbConnection(t *testing.T) {
	connDetails := readDbConf(1)
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		connDetails.Host, connDetails.Port, connDetails.User, connDetails.Password, connDetails.DbName)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	err = db.Ping()
	assert.True(t, err == nil, "Got ping error %v", err)
}