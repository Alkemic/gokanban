package main

import (
	"log"
	"os"

	"github.com/jinzhu/gorm"
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

	application := NewApp(logger, bindAddr, db)
	application.Run()
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

	db.AutoMigrate(&Column{}, &Task{}, &Tag{}, &TaskLog{})

	db.Model(&TaskLog{}).AddForeignKey("task_id", "tasks(id)", "RESTRICT", "RESTRICT")
	db.Model(&TaskLog{}).AddForeignKey("old_column_id", "columns(id)", "RESTRICT", "RESTRICT")

	return db, nil
}
