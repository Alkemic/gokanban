package main

import (
	"log"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

var db gorm.DB

func init() {
	var err error
	db, err = gorm.Open("sqlite3", "./db.sqlite")
	if err != nil {
		log.Println(err)
	}
	db.DB()

	// Then you could invoke `*sql.DB`'s functions with it
	db.DB().Ping()
	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(100)

	// Disable table name's pluralization
	db.SingularTable(true)

	db.AutoMigrate(&Task{}, &Tag{}, &Column{})
}

type Task struct {
	gorm.Model

	Title       string `sql:"size:255"`
	Description string

	Tags []Tag `gorm:"many2many:task_tags;"`

	InColumn *Column
}

type Tag struct {
	ID   uint   `gorm:"primary_key"`
	Name string `sql:"size:127"`
}

type Column struct {
	gorm.Model

	Name  string `sql:"size:127"`
	Limit int    `sql:"DEFAULT:10"`

	Order int `sql:"DEFAULT:0"`
}
