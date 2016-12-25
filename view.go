package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var ColumnEndPoint,
	TaskEndPoint RESTEndPoint

func TaskEndPointGet(w http.ResponseWriter, r *http.Request, p map[string]string) {
	var err error

	if p["id"] == "" {
		tasks := []Task{}
		db.Preload("Tags", "Column").Find(&tasks)

		tasksMap := loadTasksAsMap(&tasks)

		err = json.NewEncoder(w).Encode(tasksMap)
	} else {
		id, _ := strconv.Atoi(p["id"])
		task := Task{}

		db.Where("id = ?", id).Preload("Tags", "Column").Find(&task)
		err = json.NewEncoder(w).Encode(taskToMap(task))
	}

	if err != nil {
		log.Println(err)
	}
}

func TaskEndPointPost(w http.ResponseWriter, r *http.Request, p map[string]string) {
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
	logTask(int(task.ID), int(column.ID), "create")
}

func TaskEndPointPut(w http.ResponseWriter, r *http.Request, p map[string]string) {
	id, _ := strconv.Atoi(p["id"])
	task := Task{}
	r.ParseForm()

	db.Where("id = ?", id).Find(&task)

	_, okB := r.Form["checkId"]
	_, okO := r.Form["Position"]
	_, okC := r.Form["ColumnID"]
	if okB { // we are toggling checkbox
		checkId, _ := strconv.Atoi(r.Form.Get("checkId"))
		task.Description = toggleCheckbox(task.Description, checkId)
	} else if okO && okC { //
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
			logTask(id, task.ColumnID, "move column")
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
				logTask(id, task.ColumnID, "move position")
			}
			// nop when newPosition == task.Position
		}
		task.Position = newPosition
		task.ColumnID = newColumnID
	} else {
		if _, ok := r.Form["ColumnID"]; ok {
			task.ColumnID, _ = strconv.Atoi(r.Form.Get("ColumnID"))
		}
		logTask(id, task.ColumnID, "update")
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
}

func TaskEndPointDelete(w http.ResponseWriter, r *http.Request, p map[string]string) {
	id, _ := strconv.Atoi(p["id"])
	db.Where("id = ?", id).Delete(&Task{})

	err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	if err != nil {
		log.Println(err)
	}
	logTask(id, 0, "delete")
}

func ColumnListEndPointGet(w http.ResponseWriter, r *http.Request, p map[string]string) {
	var err error
	tasks := []Task{}

	if p["id"] == "" {
		columns := []Column{}
		db.Order("position asc").Find(&columns)

		columnsMap := loadColumnsAsMap(&columns)

		for i, column := range columnsMap {
			tasks = []Task{}
			db.Order("position asc").Where("column_id = ?", column["ID"]).
				Preload("Tags").Find(&tasks)

			columnsMap[i]["Tasks"] = loadTasksAsMap(&tasks)
		}

		err = json.NewEncoder(w).Encode(columnsMap)
	} else {
		column := Column{}
		id, _ := strconv.Atoi(p["id"])

		db.Where("id = ?", id).Find(&column)
		columnMap := columnToMap(&column)

		db.Order("position asc").
			Where("column_id = ?", column.ID).
			Preload("Tags").Find(&tasks)

		columnMap["Tasks"] = loadTasksAsMap(&tasks)

		err = json.NewEncoder(w).Encode(columnMap)
	}

	if err != nil {
		log.Println(err)
	}
}

func init() {
	TaskEndPoint = RESTEndPoint{
		Get:    TaskEndPointGet,
		Put:    TaskEndPointPut,
		Delete: TaskEndPointDelete,
		Post:   TaskEndPointPost,
	}

	ColumnEndPoint = RESTEndPoint{
		Get: ColumnListEndPointGet,
	}
}
