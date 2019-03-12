package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/Alkemic/go-route/middleware"
	_ "github.com/mattn/go-sqlite3"

	"gokanban/app"
	"gokanban/kanban"
	repository "gokanban/repository/sqlite"
	"gokanban/rest"
)

var (
	bindAddr = os.Getenv("GOKANBAN_BIND_ADDR")
	dbName   = os.Getenv("GOKANBAN_DB_FILE")

	basicAuthUser     = os.Getenv("GOKANBAN_AUTH_USER")
	basicAuthPassword = os.Getenv("GOKANBAN_AUTH_PASSWORD")
)

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile|log.Ldate)

	db, err, closeFn := InitDB(dbName)
	if err != nil {
		logger.Fatalf("Can't instantiate db: %s", err)
	}
	defer closeFn()

	taskRepository := repository.NewSqliteTaskRepository(db)
	columnRepository := repository.NewSqliteColumnRepository(db)
	taskLogRepository := repository.NewSQLiteTaskLogRepository(db)
	kanban := kanban.NewKanban(taskRepository, columnRepository, taskLogRepository)

	var authenticateFunc middleware.AuthFn
	if basicAuthUser != "" && basicAuthPassword != "" {
		logger.Println("using basic authenticate")
		authenticateFunc = middleware.Authenticate(basicAuthUser, basicAuthPassword)
	} else {
		logger.Println("not using authentication")
	}

	rest := rest.NewRestHandler(logger, kanban, authenticateFunc)
	application := app.NewApp(logger, rest)
	application.Run(bindAddr)
}

func InitDB(dbName string) (*sql.DB, error, func()) {
	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		log.Fatal(err)
	}
	if err != nil {
		return nil, err, nil
	}
	return db, nil, func() {
		db.Close()
	}
}
