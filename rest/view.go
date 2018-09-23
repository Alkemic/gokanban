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
	"github.com/jinzhu/gorm"

	"github.com/Alkemic/gokanban/helper"
)

type restHandler struct {
	logger  *log.Logger
	db      *gorm.DB
	useCase useCase
}

type useCase interface {
	ListColumns() ([]map[string]interface{}, error)
	GetColumn(id int) (map[string]interface{}, error)
	CreateTask(data map[string]string) error
	ToggleCheckbox(id, checkboxID int) error
	MoveTaskTo(id, newPosition, newColumnID int) error
	UpdateTask(id int, data map[string]string) error
	DeleteTask(id int) error
}

func NewRestHandler(logger *log.Logger, db *gorm.DB, useCase useCase) *restHandler {
	return &restHandler{
		logger:  logger,
		db:      db,
		useCase: useCase,
	}
}

func (r *restHandler) TaskEndPointPost(rw http.ResponseWriter, req *http.Request, p map[string]string) {
	req.ParseForm()
	data := r.toMap(req.Form)
	if err := r.useCase.CreateTask(data); err != nil {
		r.logger.Println(err)
		helper.Handle500(rw)
	}
}

func (r *restHandler) toMap(formData url.Values) map[string]string {
	data := map[string]string{}
	for key, _ := range formData {
		data[key] = string(formData.Get(key))
	}
	return data
}

func (r *restHandler) TaskEndPointPut(rw http.ResponseWriter, req *http.Request, p map[string]string) {
	id, _ := strconv.Atoi(p["id"])
	req.ParseForm()

	var err error
	_, okB := req.Form["checkId"]
	_, okO := req.Form["Position"]
	_, okC := req.Form["ColumnID"]
	if okB { // we are toggling checkbox
		checkID, _ := strconv.Atoi(req.Form.Get("checkId"))
		err = r.useCase.ToggleCheckbox(id, checkID)
	} else if okO && okC {
		newPosition, _ := strconv.Atoi(req.Form.Get("Position"))
		newColumnID, _ := strconv.Atoi(req.Form.Get("ColumnID"))
		err = r.useCase.MoveTaskTo(id, newPosition, newColumnID)
	} else {
		err = r.useCase.UpdateTask(id, r.toMap(req.Form))
	}

	if err != nil {
		r.logger.Println(err)
		helper.Handle500(rw)
	}
}

func (r *restHandler) TaskEndPointDelete(rw http.ResponseWriter, req *http.Request, p map[string]string) {
	id, _ := strconv.Atoi(p["id"])
	if err := r.useCase.DeleteTask(id); err != nil {
		r.logger.Println(err)
		helper.Handle500(rw)
		return
	}

	if err := json.NewEncoder(rw).Encode(map[string]string{"status": "ok"}); err != nil {
		r.logger.Println(err)
	}
}

func (r *restHandler) ColumnListEndPointGet(rw http.ResponseWriter, req *http.Request, p map[string]string) {
	var err error
	if p["id"] == "" {
		columns, err := r.useCase.ListColumns()
		if err != nil {
			helper.Handle500(rw)
			r.logger.Println(err)
			return
		}
		err = json.NewEncoder(rw).Encode(columns)
	} else {
		id, _ := strconv.Atoi(p["id"])

		column, err := r.useCase.GetColumn(id)
		if err != nil {
			helper.Handle500(rw)
			r.logger.Println(err)
			return
		}
		err = json.NewEncoder(rw).Encode(column)
	}

	if err != nil {
		r.logger.Println(err)
		helper.Handle500(rw)
	}
}

func (r *restHandler) GetMux() *http.ServeMux {
	// todo: refactor this so it'll use regexp routing saved as a field in app
	// and then use in http.ListenAndServe as a handler
	TaskEndPoint := helper.RESTEndPoint{
		//Get:    r.TaskEndPointGet,
		Put:    r.TaskEndPointPut,
		Delete: r.TaskEndPointDelete,
		Post:   r.TaskEndPointPost,
	}

	ColumnEndPoint := helper.RESTEndPoint{
		Get: r.ColumnListEndPointGet,
	}

	timeTrackDecorator := helper.TimeTrack(r.logger)

	mux := http.NewServeMux()

	serveStatic := http.FileServer(http.Dir("."))
	mux.Handle("/frontend/", serveStatic)

	TaskRouting := route.RegexpRouter{}
	TaskRouting.Add(`^/task/((?P<id>\d+)/)?$`, TaskEndPoint.Dispatch)
	mux.HandleFunc("/task/", timeTrackDecorator(TaskRouting.ServeHTTP))

	ColumnRouting := route.RegexpRouter{}
	ColumnRouting.Add(`^/column/((?P<id>\d+)/)?$`, ColumnEndPoint.Dispatch)
	mux.HandleFunc("/column/", timeTrackDecorator(ColumnRouting.ServeHTTP))

	mux.HandleFunc("/",
		timeTrackDecorator(func(w http.ResponseWriter, _ *http.Request) {
			index, _ := ioutil.ReadFile("./frontend/templates/index.html")
			io.WriteString(w, string(index))
		}))

	return mux
}
