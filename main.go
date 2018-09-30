package main

import (
	"log"
	"os"

	"github.com/jinzhu/gorm"

	"github.com/Alkemic/gokanban/app"
	"github.com/Alkemic/gokanban/kanban"
	"github.com/Alkemic/gokanban/model"
	"github.com/Alkemic/gokanban/repository"
	"github.com/Alkemic/gokanban/rest"
)

var (
	bindAddr = os.Getenv("GOKANBAN_BIND_ADDR")
	dbName   = os.Getenv("GOKANBAN_DB_FILE")
	debug    = os.Getenv("GOKANBAN_DEBUG_SQL")
)

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile|log.Ldate)
	db, err := InitDB(dbName, debug == "true")
	if err != nil {
		logger.Fatalf("Can't instantiate db: %s", err)
	}
	taskRepository := repository.NewSqliteTaskRepository(db)
	columnRepository := repository.NewSqliteColumnRepository(db)
	columnRepository.Init()
	kanban := kanban.NewKanban(taskRepository, columnRepository)
	rest := rest.NewRestHandler(logger, db, kanban)
	application := app.NewApp(logger, rest)
	application.Run(bindAddr)
}

func InitDB(dbName string, debug bool) (*gorm.DB, error) {
	db, err := gorm.Open("sqlite3", dbName)
	if err != nil {
		return nil, err
	}

	err = db.DB().Ping()
	if err != nil {
		return nil, err
	}
	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(100)

	// Disable table name's pluralization
	db.SingularTable(true)

	db.AutoMigrate(&model.Column{}, &model.Task{}, &model.Tag{}, &model.TaskLog{})

	db.Model(&model.TaskLog{}).AddForeignKey("task_id", "tasks(id)", "RESTRICT", "RESTRICT")
	db.Model(&model.TaskLog{}).AddForeignKey("old_column_id", "columns(id)", "RESTRICT", "RESTRICT")
	if debug {
		db = db.Debug()
	}
	return db, nil
}
