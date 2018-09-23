package use_case

import (
	"strconv"
	"strings"

	"github.com/Alkemic/gokanban/helper"
	"github.com/Alkemic/gokanban/model"
)

type taskRepository interface {
	Get(id int) (*model.Task, error)
	List(args ...interface{}) ([]*model.Task, error)
	GetOrCreateTag(name string) (*model.Tag, error)
	Save(task *model.Task) error
	SetPosition(columnID, taskID uint) error
	LogTask(columnID, taskID uint, action string) error
	UpdateTaskPosition(task *model.Task, newPosition, newColumnID int) error
	DeleteTask(task *model.Task) error
}

type columnRepository interface {
	List() ([]*model.Column, error)
	Get(id int) (*model.Column, error)
}

type useCase struct {
	taskRepository   taskRepository
	columnRepository columnRepository
}

func NewUseCase(taskRepository taskRepository, columnRepository columnRepository) *useCase {
	return &useCase{
		taskRepository:   taskRepository,
		columnRepository: columnRepository,
	}
}

func (uc *useCase) ListColumns() ([]map[string]interface{}, error) {
	columns, err := uc.columnRepository.List()
	if err != nil {
		return nil, err
	}

	columnsMap := helper.LoadColumnsAsMap(columns)

	for i, column := range columnsMap {
		tasks, err := uc.taskRepository.List("column_id = ?", column["ID"])
		if err != nil {
			return nil, err
		}

		columnsMap[i]["Tasks"] = helper.LoadTasksAsMap(tasks)
	}

	return columnsMap, nil
}

func (uc *useCase) GetColumn(id int) (map[string]interface{}, error) {
	column, err := uc.columnRepository.Get(id)
	if err != nil {
		return nil, err
	}

	columnMap := helper.ColumnToMap(column)
	tasks, err := uc.taskRepository.List("column_id = ?", column.ID)
	if err != nil {
		return nil, err
	}

	columnMap["Tasks"] = helper.LoadTasksAsMap(tasks)

	return columnMap, nil
}

func (uc *useCase) CreateTask(data map[string]string) error {
	tags := []model.Tag{}
	for _, value := range strings.Split(data["TagsString"], ",") {
		if value = strings.TrimSpace(value); value == "" {
			continue
		}

		tag, err := uc.taskRepository.GetOrCreateTag(value)
		if err != nil {
			return err
		}
		tags = append(tags, *tag)
	}

	columnID, _ := strconv.Atoi(data["ColumnID"])
	column, err := uc.columnRepository.Get(columnID)
	if err != nil {
		return err
	}

	task := model.Task{
		Title:       data["Title"],
		Description: data["Description"],
		Tags:        tags,
		Column:      column,
		ColumnID:    int(column.ID),
		Color:       data["Color"],
	}
	if err := uc.taskRepository.Save(&task); err != nil {
		return err
	}
	if err := uc.taskRepository.SetPosition(column.ID, task.ID); err != nil {
		return err
	}
	if err := uc.taskRepository.LogTask(column.ID, task.ID, "create"); err != nil {
		return err
	}

	return nil
}

func (uc *useCase) ToggleCheckbox(id, checkboxID int) error {
	task, err := uc.taskRepository.Get(id)
	if err != nil {
		return err
	}
	task.Description = helper.ToggleCheckbox(task.Description, checkboxID)
	return uc.taskRepository.Save(task)
}

func (uc *useCase) MoveTaskTo(id, newPosition, newColumnID int) error {
	task, err := uc.taskRepository.Get(id)
	if err != nil {
		return err
	}

	err = uc.taskRepository.UpdateTaskPosition(task, newPosition, newColumnID)
	if err != nil {
		return err
	}
	uc.taskRepository.LogTask(uint(id), uint(task.ColumnID), "move column")
	task.Position = newPosition
	task.ColumnID = newColumnID
	return uc.taskRepository.Save(task)
}

func (uc *useCase) UpdateTask(id int, data map[string]string) error {
	task, err := uc.taskRepository.Get(id)
	if err != nil {
		return err
	}

	if title, ok := data["Title"]; ok {
		task.Title = title
	}
	if description, ok := data["Description"]; ok {
		task.Description = description
	}
	var tags []model.Tag
	if tagsString, ok := data["TagsString"]; ok {
		for _, value := range strings.Split(tagsString, ",") {
			if value = strings.TrimSpace(value); value == "" {
				continue
			}

			tag, err := uc.taskRepository.GetOrCreateTag(value)
			if err != nil {
				return err
			}
			tags = append(tags, *tag)
		}
	}
	task.Tags = tags
	if color, ok := data["Color"]; ok {
		task.Color = color
	}
	uc.taskRepository.LogTask(uint(id), uint(task.ColumnID), "update task")
	return uc.taskRepository.Save(task)
}

func (uc *useCase) DeleteTask(id int) error {
	task, err := uc.taskRepository.Get(id)
	if err != nil {
		return err
	}
	if err := uc.taskRepository.DeleteTask(task); err != nil {
		return err
	}
	if err := uc.taskRepository.LogTask(uint(id), uint(task.ColumnID), "delete task"); err != nil {
		return nil
	}
	return nil
}
