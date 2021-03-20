package rest

import (
	"context"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/Alkemic/go-route"
	"github.com/Alkemic/go-route/middleware"

	"gokanban/account"
	"gokanban/repository"
)

type restHandler struct {
	logger                 *log.Logger
	kanban                 kanban
	authenticateHandler    *account.AuthenticateHandler
	authenticateMiddleware *account.Middleware
}

type kanban interface {
	ListColumns(ctx context.Context) ([]map[string]interface{}, error)
	GetColumn(ctx context.Context, id int) (map[string]interface{}, error)
	CreateTask(ctx context.Context, data map[string]string) error
	ToggleCheckbox(ctx context.Context, id, checkboxID int) error
	MoveTaskTo(ctx context.Context, id, newPosition, newColumnID int) error
	UpdateTask(ctx context.Context, id int, data map[string]string) error
	DeleteTask(ctx context.Context, id int) error
}

func NewRestHandler(logger *log.Logger, kanban kanban, authenticateHandler *account.AuthenticateHandler, authenticateMiddleware *account.Middleware) *restHandler {
	return &restHandler{
		logger:                 logger,
		kanban:                 kanban,
		authenticateMiddleware: authenticateMiddleware,
		authenticateHandler:    authenticateHandler,
	}
}

func (r *restHandler) TaskEndPointPost(rw http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	ctx := req.Context()
	data := r.toMap(req.Form)
	if err := r.kanban.CreateTask(ctx, data); err != nil {
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
	ctx := req.Context()
	var err error
	if okB { // we are toggling checkbox
		checkID, _ := strconv.Atoi(req.Form.Get("checkId"))
		err = r.kanban.ToggleCheckbox(ctx, id, checkID)
	} else if okO && okC {
		newPosition, _ := strconv.Atoi(req.Form.Get("Position"))
		newColumnID, _ := strconv.Atoi(req.Form.Get("ColumnID"))
		err = r.kanban.MoveTaskTo(ctx, id, newPosition, newColumnID)
	} else {
		err = r.kanban.UpdateTask(ctx, id, r.toMap(req.Form))
	}

	if err != nil {
		panic(err)
	}
}

func (r *restHandler) TaskEndPointDelete(rw http.ResponseWriter, req *http.Request) {
	p := route.GetParams(req)
	id, _ := strconv.Atoi(p["id"])
	ctx := req.Context()
	if err := r.kanban.DeleteTask(ctx, id); err != nil {
		panic(err)
	}

	if err := json.NewEncoder(rw).Encode(map[string]string{"status": "ok"}); err != nil {
		panic(err)
	}
}

func (r *restHandler) ColumnList(rw http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	columns, err := r.kanban.ListColumns(ctx)
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
	ctx := req.Context()
	column, err := r.kanban.GetColumn(ctx, id)
	if err != nil {
		panic(err)
	}

	if err = json.NewEncoder(rw).Encode(column); err != nil {
		panic(err)
	}
}

func (r *restHandler) GetMux() *http.ServeMux {
	TaskResource := RESTEndPoint{
		Put:    r.TaskEndPointPut,
		Delete: r.TaskEndPointDelete,
	}
	TaskCollection := RESTEndPoint{
		Post: r.TaskEndPointPost,
	}

	ColumnResource := RESTEndPoint{
		Get: r.ColumnGet,
	}
	ColumnCollection := RESTEndPoint{
		Get: r.ColumnList,
	}

	timeTrackDecorator := middleware.TimeTrack(r.logger)
	panicInterceptor := middleware.PanicInterceptorWithLogger(r.logger)
	authenticate := r.authenticateMiddleware.LoginRequiredMiddleware
	mux := http.NewServeMux()

	serveStatic := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static", serveStatic))

	TaskRouting := route.RegexpRouter{}
	TaskRouting.Add(`^/task/?$`, TaskCollection.Dispatch)
	TaskRouting.Add(`^/task/(?P<id>\d+)/$`, TaskResource.Dispatch)
	mux.HandleFunc("/task/", timeTrackDecorator(panicInterceptor(authenticate(TaskRouting.ServeHTTP))))

	ColumnRouting := route.RegexpRouter{}
	ColumnRouting.Add(`^/column/?$`, ColumnCollection.Dispatch)
	ColumnRouting.Add(`^/column/(?P<id>\d+)/$`, ColumnResource.Dispatch)
	mux.HandleFunc("/column/", timeTrackDecorator(panicInterceptor(authenticate(ColumnRouting.ServeHTTP))))

	mux.HandleFunc("/", timeTrackDecorator(authenticate(func(rw http.ResponseWriter, req *http.Request) {
		tmpl := template.Must(template.New("index.html").Funcs(map[string]interface{}{
			"marshal": func(v interface{}) template.JS {
				a, _ := json.Marshal(v)
				return template.JS(a)
			},
		}).Delims("[[", "]]").ParseFiles("frontend/templates/index.html"))
		tmplData := struct {
			User      repository.User
			LogoutURL string
		}{
			User:      account.GetUser(req),
			LogoutURL: account.LogoutPageURL,
		}
		if err := tmpl.ExecuteTemplate(rw, "index.html", tmplData); err != nil {
			r.logger.Println("cannot execute template:", err)
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})))

	mux.HandleFunc(account.LoginPageURL, r.authenticateHandler.Login)
	mux.HandleFunc(account.LogoutPageURL, r.authenticateHandler.Login)
	mux.HandleFunc("/user/", authenticate(r.authenticateHandler.Edit))

	return mux
}
