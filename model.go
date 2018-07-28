package main

import (
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

type Task struct {
	gorm.Model

	Title               string `sql:"size:255"`
	Description         string
	DescriptionRendered string `gorm:"-"`
	Color               string `sql:"size:7"`

	TaskProgress map[string]int `gorm:"-"`

	Tags []Tag `gorm:"many2many:task_tags;"`

	Column   *Column
	ColumnID int

	Position int `sql:"DEFAULT:0"`
}

type Tag struct {
	ID   uint   `gorm:"primary_key"`
	Name string `sql:"size:127"`
}

type Column struct {
	gorm.Model

	Name  string `sql:"size:127"`
	Limit int    `sql:"DEFAULT:10"`

	Position int `sql:"DEFAULT:0"`

	Tasks *[]Task `sql:"-"`
}

type TaskLog struct {
	gorm.Model

	Action string

	Task   Task
	TaskID int // `sql:"type:int(10) unsigned;not null"`

	OldColumn   Column
	OldColumnID int // `sql:"type:int(10) unsigned;not null"`
}
