package main

import (
	"log"
	"os"

	"github.com/jinzhu/gorm"

	"github.com/Alkemic/gokanban/app"
	"github.com/Alkemic/gokanban/model"
	"github.com/Alkemic/gokanban/rest"
)

var (
	bindAddr = os.Getenv("GOKANBAN_BIND_ADDR")
	dbName   = os.Getenv("GOKANBAN_DB_FILE")
)

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile|log.Ldate)
	db, err := InitDB(dbName)
	if err != nil {
		logger.Fatalf("Can't instanitize db: %s", err)
	}

	rest_ := rest.NewRestHandler(logger, db)

	application := app.NewApp(logger, rest_)
	application.Run(bindAddr)
}

func InitDB(dbName string) (*gorm.DB, error) {
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

	return db, nil
}
