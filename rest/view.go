package rest

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/Alkemic/go-route"
	"github.com/Alkemic/go-route/middleware"
	"github.com/jinzhu/gorm"

	"github.com/Alkemic/gokanban/helper"
)

type restHandler struct {
	logger *log.Logger
	db     *gorm.DB
	kanban kanban
}

type kanban interface {
	ListColumns() ([]map[string]interface{}, error)
	GetColumn(id int) (map[string]interface{}, error)
	CreateTask(data map[string]string) error
	ToggleCheckbox(id, checkboxID int) error
	MoveTaskTo(id, newPosition, newColumnID int) error
	UpdateTask(id int, data map[string]string) error
	DeleteTask(id int) error
}

func NewRestHandler(logger *log.Logger, db *gorm.DB, kanban kanban) *restHandler {
	return &restHandler{
		logger: logger,
		db:     db,
		kanban: kanban,
	}
}

func (r *restHandler) TaskEndPointPost(rw http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	data := r.toMap(req.Form)
	if err := r.kanban.CreateTask(data); err != nil {
		panic(err)
	}
}

func (r *restHandler) toMap(formData url.Values) map[string]string {
	data := map[string]string{}
	for key, _ := range formData {
		data[key] = string(formData.Get(key))
	}
	return data
}

func (r *restHandler) TaskEndPointPut(rw http.ResponseWriter, req *http.Request) {
	p := route.GetParams(req)
	id, _ := strconv.Atoi(p["id"])
	req.ParseForm()
	_, okB := req.Form["checkId"]
	_, okO := req.Form["Position"]
	_, okC := req.Form["ColumnID"]
	var err error
	if okB { // we are toggling checkbox
		checkID, _ := strconv.Atoi(req.Form.Get("checkId"))
		err = r.kanban.ToggleCheckbox(id, checkID)
	} else if okO && okC {
		newPosition, _ := strconv.Atoi(req.Form.Get("Position"))
		newColumnID, _ := strconv.Atoi(req.Form.Get("ColumnID"))
		err = r.kanban.MoveTaskTo(id, newPosition, newColumnID)
	} else {
		err = r.kanban.UpdateTask(id, r.toMap(req.Form))
	}

	if err != nil {
		panic(err)
	}
}

func (r *restHandler) TaskEndPointDelete(rw http.ResponseWriter, req *http.Request) {
	p := route.GetParams(req)
	id, _ := strconv.Atoi(p["id"])
	if err := r.kanban.DeleteTask(id); err != nil {
		panic(err)
	}

	if err := json.NewEncoder(rw).Encode(map[string]string{"status": "ok"}); err != nil {
		panic(err)
	}
}

func (r *restHandler) ColumnList(rw http.ResponseWriter, req *http.Request) {
	columns, err := r.kanban.ListColumns()
	if err != nil {
		panic(err)
	}

	if err := json.NewEncoder(rw).Encode(columns); err != nil {
		panic(err)
	}
}

func (r *restHandler) ColumnGet(rw http.ResponseWriter, req *http.Request) {
	p := route.GetParams(req)
	id, _ := strconv.Atoi(p["id"])
	column, err := r.kanban.GetColumn(id)
	if err != nil {
		panic(err)
	}

	if err = json.NewEncoder(rw).Encode(column); err != nil {
		panic(err)
	}
}

func (r *restHandler) GetMux() *http.ServeMux {
	TaskResource := helper.RESTEndPoint{
		Put:    r.TaskEndPointPut,
		Delete: r.TaskEndPointDelete,
	}
	TaskCollection := helper.RESTEndPoint{
		Post: r.TaskEndPointPost,
	}

	ColumnResource := helper.RESTEndPoint{
		Get: r.ColumnGet,
	}
	ColumnCollection := helper.RESTEndPoint{
		Get: r.ColumnList,
	}

	timeTrackDecorator := middleware.TimeTrack(r.logger)
	panicInterceptor := middleware.PanicInterceptorWithLogger(r.logger)

	mux := http.NewServeMux()

	serveStatic := http.FileServer(http.Dir("."))
	mux.Handle("/frontend/", serveStatic)

	TaskRouting := route.RegexpRouter{}
	TaskRouting.Add(`^/task/?$`, TaskCollection.Dispatch)
	TaskRouting.Add(`^/task/(?P<id>\d+)/$`, TaskResource.Dispatch)
	mux.HandleFunc("/task/", timeTrackDecorator(panicInterceptor(TaskRouting.ServeHTTP)))

	ColumnRouting := route.RegexpRouter{}
	ColumnRouting.Add(`^/column/?$`, ColumnCollection.Dispatch)
	ColumnRouting.Add(`^/column/(?P<id>\d+)/$`, ColumnResource.Dispatch)
	mux.HandleFunc("/column/", timeTrackDecorator(panicInterceptor(ColumnRouting.ServeHTTP)))

	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		index, _ := ioutil.ReadFile("./frontend/templates/index.html")
		io.WriteString(w, string(index))
	})

	return mux
}
