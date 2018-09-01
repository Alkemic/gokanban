package app

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/Alkemic/go-route"

	"github.com/Alkemic/gokanban/helper"
)

type restHandler interface {
	TaskEndPointGet(rw http.ResponseWriter, req *http.Request, p map[string]string)
	TaskEndPointPost(rw http.ResponseWriter, req *http.Request, p map[string]string)
	TaskEndPointPut(rw http.ResponseWriter, req *http.Request, p map[string]string)
	TaskEndPointDelete(rw http.ResponseWriter, req *http.Request, p map[string]string)
	ColumnListEndPointGet(rw http.ResponseWriter, req *http.Request, p map[string]string)
}

type app struct {
	logger *log.Logger
	rest   restHandler
}

// NewApp returns new instance of app
func NewApp(logger *log.Logger, rest restHandler) *app {
	app := &app{
		logger: logger,
		rest:   rest,
	}

	app.initRouting()

	return app
}

func (a *app) initRouting() {
	// todo: refactor this so it'll use regexp routing saved as a field in app
	// and then use in http.ListenAndServe as a handler
	TaskEndPoint := helper.RESTEndPoint{
		Get:    a.rest.TaskEndPointGet,
		Put:    a.rest.TaskEndPointPut,
		Delete: a.rest.TaskEndPointDelete,
		Post:   a.rest.TaskEndPointPost,
	}

	ColumnEndPoint := helper.RESTEndPoint{
		Get: a.rest.ColumnListEndPointGet,
	}

	timeTrackDecorator := helper.TimeTrack(a.logger)

	serveStatic := http.FileServer(http.Dir("."))
	http.Handle("/frontend/", serveStatic)

	TaskRouting := route.RegexpRouter{}
	TaskRouting.Add(`^/task/((?P<id>\d+)/)?$`, TaskEndPoint.Dispatch)
	http.HandleFunc("/task/", timeTrackDecorator(TaskRouting.ServeHTTP))

	ColumnRouting := route.RegexpRouter{}
	ColumnRouting.Add(`^/column/((?P<id>\d+)/)?$`, ColumnEndPoint.Dispatch)
	http.HandleFunc("/column/", timeTrackDecorator(ColumnRouting.ServeHTTP))

	http.HandleFunc("/",
		timeTrackDecorator(func(w http.ResponseWriter, r *http.Request) {
			index, _ := ioutil.ReadFile("./frontend/templates/index.html")
			io.WriteString(w, string(index))
		}))
}

func (a *app) Run(bindAddr string) {
	a.logger.Printf("Server starting on: %s\n", bindAddr)
	a.logger.Fatal(http.ListenAndServe(bindAddr, nil))
}
