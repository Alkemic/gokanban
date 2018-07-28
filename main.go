package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/jinzhu/gorm"
)

var (
	bindHost = flag.String("host", "", "")
	dbName   = flag.String("db-name", "./db.sqlite", "")
	bindPort = flag.Int("port", 8080, "")
)

type app struct {
	logger   *log.Logger
	bindHost string
	bindPort int
	db       *gorm.DB
}

// NewApp returns new instance of app
func NewApp(logger *log.Logger, bindHost string, bindPort int, dbName string) *app {
	app := &app{
		logger:   logger,
		bindHost: bindHost,
		bindPort: bindPort,
	}

	app.InitDB(dbName)
	app.InitRouting()

	return app
}

func (a *app) InitDB(dbName string) {
	db, err := gorm.Open("sqlite3", dbName)
	if err != nil {
		a.logger.Println("can't open db", err)
	}

	// Then you could invoke `*sql.DB`'s functions with it
	db.DB().Ping()
	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(100)

	// Disable table name's pluralization
	db.SingularTable(true)

	db.AutoMigrate(&Column{}, &Task{}, &Tag{}, &TaskLog{})

	db.Model(&TaskLog{}).AddForeignKey("task_id", "tasks(id)", "RESTRICT", "RESTRICT")
	db.Model(&TaskLog{}).AddForeignKey("old_column_id", "columns(id)", "RESTRICT", "RESTRICT")

	a.db = db
}

func (a *app) InitRouting() {
	// todo: refactor this so it'll use regexp routing saved as a field in app
	// and then use in http.ListenAndServe as a handler
	TaskEndPoint := RESTEndPoint{
		Get:    a.TaskEndPointGet,
		Put:    a.TaskEndPointPut,
		Delete: a.TaskEndPointDelete,
		Post:   a.TaskEndPointPost,
	}

	ColumnEndPoint := RESTEndPoint{
		Get: a.ColumnListEndPointGet,
	}

	timeTrackDecorator := TimeTrackDecorator(a.logger)

	serveStatic := http.FileServer(http.Dir("."))
	http.Handle("/frontend/", serveStatic)

	TaskRouting := RegexpHandler{}
	TaskRouting.HandleFunc(`^/task/((?P<id>\d+)/)?$`, TaskEndPoint.Dispatch)
	http.HandleFunc("/task/", timeTrackDecorator(TaskRouting.ServeHTTP))

	ColumnRouting := RegexpHandler{}
	ColumnRouting.HandleFunc(`^/column/((?P<id>\d+)/)?$`, ColumnEndPoint.Dispatch)
	http.HandleFunc("/column/", timeTrackDecorator(ColumnRouting.ServeHTTP))

	http.HandleFunc("/",
		timeTrackDecorator(func(w http.ResponseWriter, r *http.Request) {
			index, _ := ioutil.ReadFile("./frontend/templates/index.html")
			io.WriteString(w, string(index))
		}))
}

func (a *app) Run() {
	bindAddress := fmt.Sprintf("%s:%d", a.bindHost, a.bindPort)
	a.logger.Printf("Server starting on: %s\n", bindAddress)
	a.logger.Fatal(http.ListenAndServe(bindAddress, nil))
}

func main() {
	flag.Parse()
	logger := log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile|log.Ldate)

	application := NewApp(logger, *bindHost, *bindPort, *dbName)
	application.Run()
}
