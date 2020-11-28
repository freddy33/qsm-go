package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/freddy33/qsm-go/backend/spacedb"
	"github.com/rs/cors"

	config "github.com/freddy33/qsm-go/backend/conf"

	"github.com/freddy33/qsm-go/backend/m3server"
	"github.com/freddy33/qsm-go/m3util"
)

var runningApp *m3server.QsmApp

func listenSignals() {
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan,
		// https://www.gnu.org/software/libc/manual/html_node/Termination-Signals.html
		syscall.SIGTERM, // "the normal way to politely ask a program to terminate"
		syscall.SIGINT,  // Ctrl+C
		syscall.SIGQUIT, // Ctrl-\
		syscall.SIGKILL, // "always fatal", "SIGKILL and SIGSTOP may not be caught by a program"
		syscall.SIGHUP,  // "terminal is disconnected"
	)
	for {
		fmt.Println("Starting to wait for signals")
		s := <-sigchan
		sig := s.(syscall.Signal)
		sigInt := int(sig)
		fmt.Println("Received", sigInt, s.String())
		switch sig {
		case syscall.SIGQUIT:
			// Print the stack traces of all go routines
			b := make([]byte, 1<<16)
			l := runtime.Stack(b, true)
			fmt.Println("Received 0x03 signal:\n", string(b[:l]))
		case syscall.SIGHUP:
			fmt.Println("Keeping run after disconnecting for calling terminal.")
		default:
			fmt.Println("Shutting down QSM Backend Server on signal")
			killServer()
			return
		}
	}
}

func createAppAndListen(port string) {
	defer m3util.CloseAll()
	runningApp = m3server.MakeApp(m3util.GetDefaultEnvId())

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders: []string{"QsmEnvId"},
	})

	handler := c.Handler(runningApp.Router)

	runningApp.Server = &http.Server{Addr: ":" + port, Handler: handler}
	runningApp.HttpServerDone = &sync.WaitGroup{}
	runningApp.HttpServerDone.Add(1)
	log.Printf("Starting server on port=%s", port)
	go launchServer()
	runningApp.HttpServerDone.Wait()
}

func launchServer() {
	err := runningApp.Server.ListenAndServe()
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func killServer() {
	fmt.Println("Kill server called")
	defer runningApp.HttpServerDone.Done()
	if runningApp != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := runningApp.Server.Shutdown(ctx); err != nil {
			log.Fatal("failed to call shutdown on server", err)
		}
	}
}

func main() {
	others := m3util.ReadVerbose()
	didSomething := false
	runServer := false
	for i, o := range others {
		switch o {
		case "server":
			// Run the server at the end
			runServer = true
			didSomething = true
		case "gentxt":
			// TODO: Make a REST API and UI for retrieving this data
			m3server.GenerateTextFilesEnv(spacedb.GetSpaceDbFullEnv(m3util.GetDefaultEnvId()))
			didSomething = true
		case "-env":
			m3util.SetDefaultEnvId(m3util.ReadEnvId("backend main", others[i+1]))
		}
	}
	if !didSomething {
		fmt.Println("The commands", others, "are all unknown")
		os.Exit(1)
	}
	if runServer {
		go listenSignals()
		serverConfig := config.NewServerConfig()
		createAppAndListen(serverConfig.ServerPort)
		fmt.Println("Exiting main")
	}
}
