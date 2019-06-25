package m3db

import (
	"encoding/json"
	"fmt"
	"github.com/freddy33/qsm-go/m3util"
	_ "github.com/lib/pq"
	"io/ioutil"
	"log"
)

type DbConnDetails struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DbName   string `json:"dbName"`
}

func readDbConf(dbNumber int) DbConnDetails {
	confData, err := ioutil.ReadFile(fmt.Sprintf("%s/dbconn%d.json", m3util.GetConfDir(), dbNumber))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(confData))
	var res DbConnDetails
	err = json.Unmarshal([]byte(confData), &res)
	if err != nil {
		log.Fatal(err)
	}
	return res
}
