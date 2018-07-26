package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/jinzhu/gorm"
)

var (
	bindHost = flag.String("host", "", "")
	dbName   = flag.String("db-name", "./db.sqlite", "")
	bindPort = flag.Int("port", 8080, "")
)

type app struct {
	bindHost string
	bindPort int
	db       *gorm.DB
}

func NewApp(bindHost string, bindPort int, dbName string) *app {
	app := &app{
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
		log.Println("can't open db", err)
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

	serveStatic := http.FileServer(http.Dir("."))
	http.Handle("/frontend/", serveStatic)

	TaskRouting := RegexpHandler{}
	TaskRouting.HandleFunc(`^/task/((?P<id>\d+)/)?$`, TaskEndPoint.Dispatch)
	http.HandleFunc("/task/", TimeTrackDecorator(TaskRouting.ServeHTTP))

	ColumnRouting := RegexpHandler{}
	ColumnRouting.HandleFunc(`^/column/((?P<id>\d+)/)?$`, ColumnEndPoint.Dispatch)
	http.HandleFunc("/column/", TimeTrackDecorator(ColumnRouting.ServeHTTP))

	http.HandleFunc("/",
		TimeTrackDecorator(func(w http.ResponseWriter, r *http.Request) {
			index, _ := ioutil.ReadFile("./frontend/templates/index.html")
			io.WriteString(w, string(index))
		}))
}

func (a *app) Run() {
	bindAddress := fmt.Sprintf("%s:%d", bindHost, bindPort)
	log.Printf("Server starting on: %s\n", bindAddress)
	log.Fatal(http.ListenAndServe(bindAddress, nil))
}

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	application := NewApp(*bindHost, *bindPort, *dbName)
	application.Run()
}
