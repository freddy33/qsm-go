package main

import (
	"fmt"
	"github.com/freddy33/qsm-go/backend/m3db"
	"github.com/freddy33/qsm-go/backend/m3server"
	"github.com/freddy33/qsm-go/utils/m3util"
	"log"
	"net/http"
	"os"
)

func main() {
	others := m3util.ReadVerbose()
	didSomething := false
	runServer := false
	port := "8063"
	for i, o := range others {
		switch o {
		case "server":
			// Run the server at the end
			runServer = true
			didSomething = true
		case "gentxt":
			m3server.GenerateTextFilesEnv(m3util.GetDefaultEnvironment().(*m3db.QsmDbEnvironment))
			didSomething = true
		case "filldb":
			m3server.FillDbEnv(m3util.GetDefaultEnvironment().(*m3db.QsmDbEnvironment))
			didSomething = true
		case "-port":
			port = others[i+1]
		case "-test":
			m3util.SetToTestMode()
		}
	}
	if !didSomething {
		fmt.Println("The commands", others, "are all unknown")
		os.Exit(1)
	}
	if runServer {
		defer m3util.CloseAll()
		app := m3server.MakeApp(m3util.GetDefaultEnvId())
		err := http.ListenAndServe(":"+port, app.Router) // set listen port
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}
}
