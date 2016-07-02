package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var ColumnListEndPoint,
	ColumnEndPoint,
	TaskListEndPoint,
	TaskEndPoint RESTEndPoint

func init() {
	TaskListEndPoint = RESTEndPoint{
		Get: func(w http.ResponseWriter, r *http.Request, p map[string]string) {
			tasks := []Task{}
			db.Preload("Tags", "Column").Find(&tasks)
			for i, task := range tasks {
				tasks[i].DescriptionRendered = RenderMarkdown(task.Description)
			}

			err := json.NewEncoder(w).Encode(tasks)
			if err != nil {
				log.Println(err)
			}
		},
		Post: func(w http.ResponseWriter, r *http.Request, p map[string]string) {
			db.LogMode(true)
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
			if _, ok := r.Form["ColumnID"]; ok {
				ColumnID, _ := strconv.Atoi(r.Form.Get("ColumnID"))
				db.Where("id = ?", ColumnID).Find(&column)
			} else {
				db.FirstOrCreate(&column, Column{Position: 1})
			}

			task := Task{
				Title:       r.Form.Get("Title"),
				Description: r.Form.Get("Description"),
				Tags:        prepareTags(r.Form.Get("TagsString")),
				Column:      &column,
				ColumnID:    int(column.ID),
			}
			db.Save(&task)
			db.Exec(
				"update task set position = (select max(position) "+
					"from task where column_id = ?) + 1 where id = ?;",
				column.ID, task.ID)
			db.LogMode(false)
		},
	}

	TaskEndPoint = RESTEndPoint{
		Get: func(w http.ResponseWriter, r *http.Request, p map[string]string) {
			id, _ := strconv.Atoi(p["id"])
			task := Task{}

			db.Where("id = ?", id).Preload("Tags", "Column").Find(&task)

			err := json.NewEncoder(w).Encode(task)
			if err != nil {
				log.Println(err)
			}
		},
		Put: func(w http.ResponseWriter, r *http.Request, p map[string]string) {
			id, _ := strconv.Atoi(p["id"])
			task := Task{}
			r.ParseForm()

			db.Where("id = ?", id).Find(&task)

			_, okO := r.Form["Position"]
			_, okC := r.Form["ColumnID"]
			if okO && okC {
				newPosition, _ := strconv.Atoi(r.Form.Get("Position"))
				newColumnID, _ := strconv.Atoi(r.Form.Get("ColumnID"))
				if task.ColumnID != newColumnID {
					// remoce gap in old column
					db.Exec(
						"update task set position = position - 1 "+
							"where column_id = ? and position >= ?;",
						task.ColumnID, task.Position)
					// make space in new column
					db.Exec(
						"update task set position = position + 1 "+
							"where column_id = ? and position >= ?;",
						newColumnID, newPosition)
				} else {
					if newPosition > task.Position {
						// move task between old and new position up
						db.Exec(
							`update task set position = position - 1
							where column_id = ? and position <= ? and
							position >= ?;`,
							task.ColumnID, newPosition, task.Position)
					} else if newPosition < task.Position {
						// move task between old and new position down
						db.Exec(
							`update task set position = position + 1
							where column_id = ? and position <= ? and
							position >= ?;`,
							task.ColumnID, task.Position, newPosition)
					}
					// nop when newPosition == task.Position
				}
				task.Position = newPosition
				task.ColumnID = newColumnID
			} else {
				if _, ok := r.Form["ColumnID"]; ok {
					task.ColumnID, _ = strconv.Atoi(r.Form.Get("ColumnID"))
				}
			}
			if _, ok := r.Form["Title"]; ok {
				task.Title = r.Form.Get("Title")
			}
			if _, ok := r.Form["Description"]; ok {
				task.Description = r.Form.Get("Description")
			}
			if _, ok := r.Form["TagsString"]; ok {
				db.Exec("DELETE FROM task_tags WHERE task_id = ?", task.ID)
				task.Tags = prepareTags(r.Form.Get("TagsString"))
			}
			db.Save(&task)
		},
		Delete: func(w http.ResponseWriter, r *http.Request, p map[string]string) {
			id, _ := strconv.Atoi(p["id"])
			db.Where("id = ?", id).Delete(&Task{})

			err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
			if err != nil {
				log.Println(err)
			}
		},
	}

	ColumnListEndPoint = RESTEndPoint{
		Get: func(w http.ResponseWriter, r *http.Request, p map[string]string) {
			columns := []Column{}
			db.Order("position asc").Find(&columns)

			for i, column := range columns {
				columns[i].Tasks = &[]Task{}
				db.Order("position asc").
					Where("column_id = ?", column.ID).
					Preload("Tags").Find(columns[i].Tasks)

				for j, task := range *(columns[i].Tasks) {
					(*columns[i].Tasks)[j].DescriptionRendered = RenderMarkdown(task.Description)
				}
			}

			err := json.NewEncoder(w).Encode(columns)
			if err != nil {
				log.Println(err)
			}
		},
	}

	ColumnEndPoint = RESTEndPoint{
		Get: func(w http.ResponseWriter, r *http.Request, p map[string]string) {
			column := Column{}
			id, _ := strconv.Atoi(p["id"])

			db.Where("id = ?", id).Find(&column)

			column.Tasks = &[]Task{}
			db.Order("position asc").
				Where("column_id = ?", column.ID).
				Preload("Tags").Find(column.Tasks)

			for i, task := range *column.Tasks {
				(*column.Tasks)[i].DescriptionRendered = RenderMarkdown(task.Description)
			}

			err := json.NewEncoder(w).Encode(column)
			if err != nil {
				log.Println(err)
			}
		},
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
