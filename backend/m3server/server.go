package m3server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/freddy33/qsm-go/backend/m3db"
	"github.com/freddy33/qsm-go/backend/spacedb"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3api"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

type QsmApp struct {
	HttpServerDone *sync.WaitGroup
	Server         *http.Server
	Router         *mux.Router
	Env            *m3db.QsmDbEnvironment
}

func (app *QsmApp) AddHandler(path string, handleFunc func(http.ResponseWriter, *http.Request)) *mux.Route {
	return app.Router.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		envId := app.Env.GetId()
		fromHeader := r.Header.Get(m3api.HttpEnvIdKey)
		if fromHeader == "" {
			r.Header.Add(m3api.HttpEnvIdKey, app.Env.GetEnvNumber())
		} else {
			envId = m3util.ReadEnvId(fmt.Sprintf("header var %q", m3api.HttpEnvIdKey), fromHeader)
		}
		ctx := context.WithValue(r.Context(), m3api.HttpEnvIdKey, envId)
		handleFunc(w, r.WithContext(ctx))
	})
}

func GetEnvId(r *http.Request) m3util.QsmEnvID {
	return r.Context().Value(m3api.HttpEnvIdKey).(m3util.QsmEnvID)
}

func GetEnvironment(r *http.Request) *m3db.QsmDbEnvironment {
	return m3db.GetEnvironment(GetEnvId(r))
}

func SendResponse(w http.ResponseWriter, status int, format string, args ...interface{}) {
	if status >= 400 {
		Log.Errorf(format, args...)
	}
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "text/plain")
	_, err := fmt.Fprintf(w, format, args...)
	if err != nil {
		log.Printf("failed to send data to response due to %q", err.Error())
	}
}

func getRequestType(w http.ResponseWriter, r *http.Request) string {
	reqContentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(reqContentType, "application/json") {
		return "json"
	} else if strings.HasPrefix(reqContentType, "application/x-protobuf") {
		return "proto"
	} else {
		SendResponse(w, http.StatusBadRequest, "unsupported content type %q", reqContentType)
		return "error"
	}
}

/*
Return true if an error occurred and the response already filed
*/
func ReadRequestMsg(w http.ResponseWriter, r *http.Request, reqMsg proto.Message) bool {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		SendResponse(w, http.StatusBadRequest, "req body could not be read req body due to: %s", err.Error())
		return true
	}
	reqContentType := getRequestType(w, r)
	if reqContentType == "json" {
		err = json.Unmarshal(b, reqMsg)
	} else if reqContentType == "proto" {
		err = proto.Unmarshal(b, reqMsg)
	} else {
		return true
	}
	if err != nil {
		SendResponse(w, http.StatusBadRequest, "req body could not be parsed due to: %s", err.Error())
		return true
	}
	return false
}

func WriteResponseMsg(w http.ResponseWriter, r *http.Request, resMsg proto.Message) {
	var useProtobuf, useJson bool

	reqContentType := getRequestType(w, r)
	// Return same type has request payload by default
	if reqContentType == "json" {
		useProtobuf = false
		useJson = true
	} else if reqContentType == "proto" {
		useProtobuf = true
		useJson = false
	} else {
		return
	}
	// If accept tells me ok to use proto switch to it
	acceptContents := r.Header.Values("Accept")
	for _, ac := range acceptContents {
		if strings.HasPrefix(ac, "application/x-protobuf") {
			useProtobuf = true
			useJson = false
			break
		}
	}

	typeName := reflect.TypeOf(resMsg).String()
	typeName = strings.TrimPrefix(typeName, "*")

	var data []byte
	var err error
	if useProtobuf {
		data, err = proto.Marshal(resMsg)
		w.Header().Set("Content-Type", "application/x-protobuf; messageType="+typeName)
	} else if useJson {
		data, err = json.Marshal(resMsg)
		w.Header().Set("Content-Type", "application/json; messageType="+typeName)
	} else {
		SendResponse(w, http.StatusBadRequest, "No acceptable content type for response found in %v", acceptContents)
	}

	if err != nil {
		SendResponse(w, http.StatusInternalServerError, "Failed to marshal %q due to: %s", typeName, err.Error())
		return
	}

	_, err = w.Write(data)
	if err != nil {
		Log.Errorf("failed to send data to response due to %q", err.Error())
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	SendResponse(w, http.StatusOK, "Using env id=%d\nMethod=%s\n", r.Context().Value(m3api.HttpEnvIdKey), r.Method)
}

func drop(w http.ResponseWriter, r *http.Request) {
	env := GetEnvironment(r)
	envId := env.GetId()
	env.Destroy()
	SendResponse(w, http.StatusOK, "Test env id %d was deleted", envId)
}

func initialize(w http.ResponseWriter, r *http.Request) {
	envId := GetEnvId(r)
	spacedb.GetSpaceDbFullEnv(envId)
	SendResponse(w, http.StatusCreated, "Test env id %d was initialized", envId)
}

func logLevel(w http.ResponseWriter, r *http.Request) {
	Log.Infof("Receive logLevel")

	values := r.URL.Query()
	if len(values) == 0 {
		SendResponse(w, http.StatusBadRequest, "Please provide a new level for packages as query parameter!")
		return
	}
	for packName, listLevels := range values {
		if len(listLevels) != 1 {
			SendResponse(w, http.StatusBadRequest, "Please provide a specific level for package name %q in your query parameter!", packName)
			return
		}
		foundLevel := m3util.LogLevel(-1)
		newLevel := strings.ToUpper(listLevels[0])
		for _, lv := range m3util.GetAllLogLevels() {
			if m3util.GetLevelName(lv) == newLevel {
				foundLevel = lv
				break
			}
		}
		if foundLevel < 0 {
			intVal, err := strconv.Atoi(newLevel)
			if err != nil {
				SendResponse(w, http.StatusBadRequest, "The level provided %q for package name %q is not valid.", newLevel, packName)
				return
			}
			for _, lv := range m3util.GetAllLogLevels() {
				if intVal == int(lv) {
					foundLevel = lv
					break
				}
			}
		}
		if foundLevel < 0 {
			SendResponse(w, http.StatusBadRequest, "The level provided %q for package name %q is not valid.", newLevel, packName)
			return
		}
		if packName == "all" {
			m3util.SetLogLevelForAll(foundLevel)
		} else if packName == "services" {
			m3util.SetLoggerLevel("pointdb", foundLevel)
			m3util.SetLoggerLevel("pathdb", foundLevel)
			m3util.SetLoggerLevel("spacedb", foundLevel)
		} else if packName == "space" {
			m3util.SetLoggerLevel("m3space", foundLevel)
			m3util.SetLoggerLevel("spacedb", foundLevel)
		} else if packName == "path" {
			m3util.SetLoggerLevel("m3path", foundLevel)
			m3util.SetLoggerLevel("pathdb", foundLevel)
		} else if packName == "point" {
			m3util.SetLoggerLevel("m3point", foundLevel)
			m3util.SetLoggerLevel("pointdb", foundLevel)
		} else {
			m3util.SetLoggerLevel(packName, foundLevel)
		}
	}

	// TODO: Send in response the log levels updated
	response := "Updated Log Levels"
	SendResponse(w, http.StatusOK, response)
}

func MakeApp(envId m3util.QsmEnvID) *QsmApp {
	if envId == m3util.NoEnv {
		envId = m3util.GetDefaultEnvId()
	}
	env := m3db.GetEnvironment(envId)
	r := mux.NewRouter()
	app := &QsmApp{Router: r, Env: env}
	app.AddHandler("/", home)

	// TODO: MAke also a getter to list current log level
	app.AddHandler("/log", logLevel).Methods("POST")

	app.AddHandler("/point-data", retrievePointData).Methods("GET")

	app.AddHandler("/init-env", initialize).Methods("POST")
	app.AddHandler("/drop-env", drop).Methods("DELETE")

	app.AddHandler("/create-path-ctx", createPathContext).Methods("PUT")
	app.AddHandler("/next-nodes", moveToNextNode).Methods("POST")

	app.AddHandler("/space", createSpace).Methods("PUT")
	app.AddHandler("/space", getSpaces).Methods("GET")
	app.AddHandler("/event", createEvent).Methods("PUT")

	return app
}
