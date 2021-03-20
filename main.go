package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"gokanban/account"
	"gokanban/app"
	"gokanban/kanban"
	repository "gokanban/repository/sqlite"
	"gokanban/rest"
)

var (
	bindAddr = os.Getenv("GOKANBAN_BIND_ADDR")
	dbName   = os.Getenv("GOKANBAN_DB_FILE")
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

	settingsRepository := repository.NewSettingsRepository(db)
	sessionRepository := repository.NewSessionRepository(28 * 24 * time.Hour)
	authenticateHandler := account.NewAuthenticateHandler(logger, settingsRepository, sessionRepository)
	authenticateMiddleware := account.NewAuthenticateMiddleware(logger, settingsRepository, sessionRepository)

	rest := rest.NewRestHandler(logger, kanban, authenticateHandler, authenticateMiddleware)
	application := app.NewApp(logger, rest)
	application.Run(bindAddr)
}

func InitDB(dbName string) (*sqlx.DB, error, func()) {
	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		return nil, err, nil
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging db: %w", err), nil
	}
	return sqlx.NewDb(db, "mysql"), nil, func() {
		db.Close()
	}
}
