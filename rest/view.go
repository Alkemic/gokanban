package rest

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/Alkemic/go-route"
	"github.com/jinzhu/gorm"
	"gitlab.com/Alkemic/gowks/core/middleware"

	"github.com/Alkemic/gokanban/helper"
	"github.com/Alkemic/gokanban/model"
)

const (
	makeGapeSQL    = "update task set position = position + 1 where column_id = ? and position >= ?;"
	moveTaskSQL    = "update task set position = ?, column_id = ? where id = ?;"
	removeGapeSQL  = "update task set position = position - 1 where column_id = ? and position > ?;"
	setPositionSQL = "update task set position = (select max(position) from task where column_id = ? and deleted_at is null) + 1 where id = ?;"
)

type restHandler struct {
	logger  *log.Logger
	db      *gorm.DB
	useCase useCase
}

type useCase interface {
	ListColumns() ([]map[string]interface{}, error)
	GetColumn(id int) (map[string]interface{}, error)
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

	tags := []model.Tag{}
	for _, value := range strings.Split(req.Form.Get("TagsString"), ",") {
		if value == "" {
			continue
		}

		tag := model.Tag{}
		r.db.FirstOrCreate(&tag, model.Tag{Name: strings.TrimSpace(value)})
		tags = append(tags, tag)
	}

	column := model.Column{}
	if _, ok := req.Form["ColumnID"]; ok {
		ColumnID, _ := strconv.Atoi(req.Form.Get("ColumnID"))
		r.db.Where("id = ?", ColumnID).Find(&column)
	} else {
		r.db.FirstOrCreate(&column, model.Column{Position: 1})
	}

	task := model.Task{
		Title:       req.Form.Get("Title"),
		Description: req.Form.Get("Description"),
		Tags:        helper.PrepareTags(r.db, req.Form.Get("TagsString")),
		Column:      &column,
		ColumnID:    int(column.ID),
		Color:       req.Form.Get("Color"),
	}
	r.db.Save(&task)
	r.db.Exec(setPositionSQL, column.ID, task.ID)
	helper.LogTask(r.db, int(task.ID), int(column.ID), "create")
}

func (r *restHandler) TaskEndPointPut(rw http.ResponseWriter, req *http.Request, p map[string]string) {
	id, _ := strconv.Atoi(p["id"])
	task := model.Task{}
	req.ParseForm()

	r.db.Where("id = ?", id).Find(&task)

	_, okB := req.Form["checkId"]
	_, okO := req.Form["Position"]
	_, okC := req.Form["ColumnID"]
	if okB { // we are toggling checkbox
		checkID, _ := strconv.Atoi(req.Form.Get("checkId"))
		task.Description = helper.ToggleCheckbox(task.Description, checkID)
	} else if okO && okC {
		newPosition, _ := strconv.Atoi(req.Form.Get("Position"))
		newColumnID, _ := strconv.Atoi(req.Form.Get("ColumnID"))

		r.db.Exec(makeGapeSQL, newColumnID, newPosition)
		r.db.Exec(moveTaskSQL, newPosition, newColumnID, task.ID)
		r.db.Exec(removeGapeSQL, task.ColumnID, task.Position)
		helper.LogTask(r.db, id, task.ColumnID, "move column")

		task.Position = newPosition
		task.ColumnID = newColumnID
	} else {
		if _, ok := req.Form["Title"]; ok {
			task.Title = req.Form.Get("Title")
		}
		if _, ok := req.Form["Description"]; ok {
			task.Description = req.Form.Get("Description")
		}
		if _, ok := req.Form["TagsString"]; ok {
			r.db.Exec("DELETE FROM task_tags WHERE task_id = ?", task.ID)
			task.Tags = helper.PrepareTags(r.db, req.Form.Get("TagsString"))
		}
		if _, ok := req.Form["Color"]; ok {
			task.Color = req.Form.Get("Color")
		} else {
			task.Color = ""
		}
		helper.LogTask(r.db, id, task.ColumnID, "update task")
	}
	r.db.Save(&task)
}

func (r *restHandler) TaskEndPointDelete(rw http.ResponseWriter, req *http.Request, p map[string]string) {
	id, _ := strconv.Atoi(p["id"])

	task := model.Task{}
	r.db.Where("id = ?", id).Find(&task)
	r.db.Delete(task)

	r.db.Exec(removeGapeSQL, task.ColumnID, task.Position)

	err := json.NewEncoder(rw).Encode(map[string]string{"status": "ok"})
	if err != nil {
		r.logger.Println(err)
	}
	helper.LogTask(r.db, id, 0, "delete")
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

	timeTrackDecorator := middleware.TimeTrack(r.logger)

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

//func (r *restHandler) GetMux() *http.ServeMux {
//}
