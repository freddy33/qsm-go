package m3server

import (
	"fmt"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/freddy33/qsm-go/utils/m3db"
	"net/http"
	"strings"
)

func GetPointPackData(env *m3db.QsmEnvironment) *m3point.PointPackData {
	if env.GetData(m3db.PointIdx) == nil {
		ppd := new(m3point.PointPackData)
		ppd.env = env
		env.SetData(m3db.PointIdx, ppd)
		// do not return ppd but always the pointer in env data array
	}
	return env.GetData(m3db.PointIdx).(*m3point.PointPackData)
}

func getAllConnections(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	env := m3db.GetDefaultEnvironment()
	ppd := GetPointPackData(env)

	fmt.Println(r.Form) // print form information in server side
	fmt.Println("path", r.URL.Path)
	fmt.Println("scheme", r.URL.Scheme)
	fmt.Println(r.Form["url_long"])
	for k, v := range r.Form {
		fmt.Println("key:", k)
		fmt.Println("val:", strings.Join(v, ""))
	}
	fmt.Fprintf(w, "Hello astaxie!\n") // send data to client side
}
