package main

import (
	"github.com/freddy33/qsm-go/backend/m3server"
	"github.com/freddy33/qsm-go/utils/m3db"
	"log"
	"net/http"
)

func main() {
	app := m3server.MakeApp(m3db.NoEnv)
	err := http.ListenAndServe(":8063", app.Router) // set listen port
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
