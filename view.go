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
	db.Preload("Tags", "InColumn").Find(&tasks)

	if err := json.NewEncoder(w).Encode(tasks); err != nil {
		panic(err)
	}
}

func TaskView(
	w http.ResponseWriter,
	r *http.Request,
	p map[string]string,
) {
	task := Task{}
	log.Println(task.InColumn)
	id, _ := strconv.Atoi(p["id"])

	db.Where("id = ?", id).Preload("Tags", "InColumn").Find(&task)

	if err := json.NewEncoder(w).Encode(task); err != nil {
		panic(err)
	}
}

func ColumnListView(
	w http.ResponseWriter,
	r *http.Request,
	p map[string]string,
) {
	columns := []Column{}
	db.Find(&columns)

	if err := json.NewEncoder(w).Encode(columns); err != nil {
		panic(err)
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
	db.Model(&column).Related(&tasks, "InColumn")
	log.Println(tasks)

	if err := json.NewEncoder(w).Encode(column); err != nil {
		panic(err)
	}
}
