package main

import (
	"github.com/freddy33/qsm-go/backend/m3server"
	"github.com/freddy33/qsm-go/utils/m3util"
	"log"
	"net/http"
)

func main() {
	defer m3util.CloseAll()
	app := m3server.MakeApp(m3util.NoEnv)
	err := http.ListenAndServe(":8063", app.Router) // set listen port
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
