package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

const (
	makeGapeSQL    = "update task set position = position + 1 where column_id = ? and position >= ?;"
	moveTaskSQL    = "update task set position = ?, column_id = ? where id = ?;"
	removeGapeSQL  = "update task set position = position - 1 where column_id = ? and position > ?;"
	setPositionSQL = "update task set position = (select max(position) from task where column_id = ? and deleted_at is null) + 1 where id = ?;"
)

func (a *app) TaskEndPointGet(w http.ResponseWriter, r *http.Request, p map[string]string) {
	var err error

	if p["id"] == "" {
		tasks := []Task{}
		a.db.Preload("Tags", "Column").Find(&tasks)

		tasksMap := loadTasksAsMap(&tasks)

		err = json.NewEncoder(w).Encode(tasksMap)
	} else {
		id, _ := strconv.Atoi(p["id"])
		task := Task{}

		a.db.Where("id = ?", id).Preload("Tags", "Column").Find(&task)
		err = json.NewEncoder(w).Encode(taskToMap(task))
	}

	if err != nil {
		a.logger.Println(err)
	}
}

func (a *app) TaskEndPointPost(w http.ResponseWriter, r *http.Request, p map[string]string) {
	r.ParseForm()

	tags := []Tag{}
	for _, value := range strings.Split(r.Form.Get("TagsString"), ",") {
		if value == "" {
			continue
		}

		tag := Tag{}
		a.db.FirstOrCreate(&tag, Tag{Name: strings.TrimSpace(value)})
		tags = append(tags, tag)
	}

	column := Column{}
	if _, ok := r.Form["ColumnID"]; ok {
		ColumnID, _ := strconv.Atoi(r.Form.Get("ColumnID"))
		a.db.Where("id = ?", ColumnID).Find(&column)
	} else {
		a.db.FirstOrCreate(&column, Column{Position: 1})
	}

	task := Task{
		Title:       r.Form.Get("Title"),
		Description: r.Form.Get("Description"),
		Tags:        prepareTags(a.db, r.Form.Get("TagsString")),
		Column:      &column,
		ColumnID:    int(column.ID),
		Color:       r.Form.Get("Color"),
	}
	a.db.Save(&task)
	a.db.Exec(setPositionSQL, column.ID, task.ID)
	logTask(a.db, int(task.ID), int(column.ID), "create")
}

func (a *app) TaskEndPointPut(w http.ResponseWriter, r *http.Request, p map[string]string) {
	id, _ := strconv.Atoi(p["id"])
	task := Task{}
	r.ParseForm()

	a.db.Where("id = ?", id).Find(&task)

	_, okB := r.Form["checkId"]
	_, okO := r.Form["Position"]
	_, okC := r.Form["ColumnID"]
	if okB { // we are toggling checkbox
		checkID, _ := strconv.Atoi(r.Form.Get("checkId"))
		task.Description = toggleCheckbox(task.Description, checkID)
	} else if okO && okC {
		newPosition, _ := strconv.Atoi(r.Form.Get("Position"))
		newColumnID, _ := strconv.Atoi(r.Form.Get("ColumnID"))

		a.db.Exec(makeGapeSQL, newColumnID, newPosition)
		a.db.Exec(moveTaskSQL, newPosition, newColumnID, task.ID)
		a.db.Exec(removeGapeSQL, task.ColumnID, task.Position)
		logTask(a.db, id, task.ColumnID, "move column")

		task.Position = newPosition
		task.ColumnID = newColumnID
	} else {
		if _, ok := r.Form["Title"]; ok {
			task.Title = r.Form.Get("Title")
		}
		if _, ok := r.Form["Description"]; ok {
			task.Description = r.Form.Get("Description")
		}
		if _, ok := r.Form["TagsString"]; ok {
			a.db.Exec("DELETE FROM task_tags WHERE task_id = ?", task.ID)
			task.Tags = prepareTags(a.db, r.Form.Get("TagsString"))
		}
		if _, ok := r.Form["Color"]; ok {
			task.Color = r.Form.Get("Color")
		} else {
			task.Color = ""
		}
		logTask(a.db, id, task.ColumnID, "update task")
	}
	a.db.Save(&task)
}

func (a *app) TaskEndPointDelete(w http.ResponseWriter, r *http.Request, p map[string]string) {
	id, _ := strconv.Atoi(p["id"])

	task := Task{}
	a.db.Where("id = ?", id).Find(&task)
	a.db.Delete(task)

	a.db.Exec(removeGapeSQL, task.ColumnID, task.Position)

	err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	if err != nil {
		a.logger.Println(err)
	}
	logTask(a.db, id, 0, "delete")
}

func (a *app) ColumnListEndPointGet(w http.ResponseWriter, r *http.Request, p map[string]string) {
	var err error
	tasks := []Task{}

	if p["id"] == "" {
		columns := []Column{}
		a.db.Order("position asc").Find(&columns)

		columnsMap := loadColumnsAsMap(&columns)

		for i, column := range columnsMap {
			tasks = []Task{}
			a.db.Order("position asc").Where("column_id = ?", column["ID"]).
				Preload("Tags").Find(&tasks)

			columnsMap[i]["Tasks"] = loadTasksAsMap(&tasks)
		}

		err = json.NewEncoder(w).Encode(columnsMap)
	} else {
		column := Column{}
		id, _ := strconv.Atoi(p["id"])

		a.db.Where("id = ?", id).Find(&column)
		columnMap := columnToMap(&column)

		a.db.Order("position asc").
			Where("column_id = ?", column.ID).
			Preload("Tags").Find(&tasks)

		columnMap["Tasks"] = loadTasksAsMap(&tasks)

		err = json.NewEncoder(w).Encode(columnMap)
	}

	if err != nil {
		a.logger.Println(err)
	}
}
