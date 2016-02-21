package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func TaskListView(
	w http.ResponseWriter,
	r *http.Request,
	p map[string]string,
) {
	if r.Method == "POST" {
		r.ParseForm()

		tags := []Tag{}
		for _, value := range strings.Split(r.Form.Get("TagsString"), ",") {
			if value == "" {
				continue
			}

			tag := Tag{}
			db.FirstOrCreate(&tag, Tag{Name: strings.TrimSpace(value)})
			tags = append(tags, tag)
		}
		column := Column{}
		db.FirstOrCreate(&column, Column{Order: 1})
		task := Task{
			Title:       r.Form.Get("Title"),
			Description: r.Form.Get("Description"),
			Tags:        prepareTags(r.Form.Get("TagsString")),
			Column:      &column,
			ColumnID:    int(column.ID),
		}
		db.Save(&task)
	} else if r.Method == "GET" {
		tasks := []Task{}
		db.Preload("Tags", "Column").Find(&tasks)

		err := json.NewEncoder(w).Encode(tasks)
		if err != nil {
			log.Println(err)
		}
	}
}

func TaskView(
	w http.ResponseWriter,
	r *http.Request,
	p map[string]string,
) {
	id, _ := strconv.Atoi(p["id"])
	task := Task{}
	if r.Method == "PUT" {
		r.ParseForm()

		db.Where("id = ?", id).Find(&task)

		if _, ok := r.Form["Title"]; ok {
			task.Title = r.Form.Get("Title")
		}
		if _, ok := r.Form["Description"]; ok {
			task.Description = r.Form.Get("Description")
		}
		if _, ok := r.Form["ColumnID"]; ok {
			task.ColumnID, _ = strconv.Atoi(r.Form.Get("ColumnID"))
		}
		if _, ok := r.Form["TagsString"]; ok {
			db.Exec("DELETE FROM task_tags WHERE task_id = ?", task.ID)
			task.Tags = prepareTags(r.Form.Get("TagsString"))
		}
		db.Save(&task)
	} else if r.Method == "DELETE" {
		db.Where("id = ?", id).Delete(&Task{})

		err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		if err != nil {
			log.Println(err)
		}
	} else if r.Method == "GET" {
		db.Where("id = ?", id).Preload("Tags", "Column").Find(&task)

		err := json.NewEncoder(w).Encode(task)
		if err != nil {
			log.Println(err)
		}
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
		db.Where("column_id = ?", column.ID).Preload("Tags").Find(columns[i].Tasks)
	}

	err := json.NewEncoder(w).Encode(columns)
	if err != nil {
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

	err := json.NewEncoder(w).Encode(column)
	if err != nil {
		log.Println(err)
	}
}

func prepareTags(s string) (tags []Tag) {
	for _, value := range strings.Split(s, ",") {
		if value == "" {
			continue
		}

		tag := Tag{}
		db.FirstOrCreate(&tag, Tag{Name: strings.TrimSpace(value)})
		tags = append(tags, tag)
	}

	return tags
}
