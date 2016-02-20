package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

func TaskListView(
	w http.ResponseWriter,
	r *http.Request,
	p map[string]string,
) {
	tasks := []Task{}
	db.Preload("Tags", "Column").Find(&tasks)

	if err := json.NewEncoder(w).Encode(tasks); err != nil {
		log.Println(err)
	}
}

func TaskView(
	w http.ResponseWriter,
	r *http.Request,
	p map[string]string,
) {
	task := Task{}
	log.Println(task.Column)
	id, _ := strconv.Atoi(p["id"])

	db.Where("id = ?", id).Preload("Tags", "Column").Find(&task)

	if err := json.NewEncoder(w).Encode(task); err != nil {
		log.Println(err)
	}
}

func ColumnListView(
	w http.ResponseWriter,
	r *http.Request,
	p map[string]string,
) {
	columns := []Column{}
	db.Order("`order` asc").Find(&columns)

	for i, column := range columns {
		columns[i].Tasks = &[]Task{}
		db.Where("column_id = ?", column.ID).Find(columns[i].Tasks)
	}

	if err := json.NewEncoder(w).Encode(columns); err != nil {
		log.Println(err)
	}
}

func ColumnView(
	w http.ResponseWriter,
	r *http.Request,
	p map[string]string,
) {
	tasks := []Task{}
	column := Column{}
	id, _ := strconv.Atoi(p["id"])

	db.Where("id = ?", id).Find(&column)
	db.Model(&column).Related(&tasks, "Column")

	if err := json.NewEncoder(w).Encode(column); err != nil {
		log.Println(err)
	}
}
