package kanban

import (
	"context"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"gokanban/markdown"
	"gokanban/model"
)

type taskRepository interface {
	Get(ctx context.Context, id int) (*model.Task, error)
	List(ctx context.Context, columnID uint) ([]*model.Task, error)
	Save(ctx context.Context, task *model.Task) error
	Update(ctx context.Context, task *model.Task) error
	SetPosition(ctx context.Context, columnID, taskID uint) error
	UpdateTaskPosition(ctx context.Context, task *model.Task, newPosition, newColumnID int) error
	DeleteTask(ctx context.Context, task *model.Task) error
	GetOrCreateTag(ctx context.Context, name string) (*model.Tag, error)
	FetchTagsForTask(ctx context.Context, task *model.Task) ([]*model.Tag, error)
}

type columnRepository interface {
	List(ctx context.Context) ([]*model.Column, error)
	Get(ctx context.Context, id int) (*model.Column, error)
}

type taskLogRepository interface {
	LogTask(ctx context.Context, columnID, taskID uint, action string) error
}

type kanban struct {
	taskRepository    taskRepository
	columnRepository  columnRepository
	taskLogRepository taskLogRepository
}

func NewKanban(taskRepository taskRepository, columnRepository columnRepository, taskLogRepository taskLogRepository) *kanban {
	return &kanban{
		taskRepository:    taskRepository,
		columnRepository:  columnRepository,
		taskLogRepository: taskLogRepository,
	}
}

func (k *kanban) ListColumns(ctx context.Context) ([]map[string]interface{}, error) {
	columns, err := k.columnRepository.List(ctx)
	if err != nil {
		return nil, err
	}

	columnsMap := columnsToMap(columns)

	for i, column := range columnsMap {
		tasks, err := k.taskRepository.List(ctx, column["ID"].(uint))
		if err != nil {
			return nil, err
		}

		tasks, err = k.loadTagsForTasks(ctx, tasks)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot load tags for task")
		}
		columnsMap[i]["Tasks"] = tasksToMap(tasks)
	}

	return columnsMap, nil
}

func (k *kanban) GetColumn(ctx context.Context, id int) (map[string]interface{}, error) {
	column, err := k.columnRepository.Get(ctx, id)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot select column '%d'", id)
	}

	columnMap := columnToMap(column)
	tasks, err := k.taskRepository.List(ctx, column.ID)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot fetch tasks list for column '%d'", id)
	}

	tasks, err = k.loadTagsForTasks(ctx, tasks)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot load tags for task")
	}
	columnMap["Tasks"] = tasksToMap(tasks)

	return columnMap, nil
}

func (k *kanban) loadTagsForTasks(ctx context.Context, tasks []*model.Task) ([]*model.Task, error) {
	var err error
	for _, task := range tasks {
		task.Tags, err = k.taskRepository.FetchTagsForTask(ctx, task)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot fetch tags for task '%+v'", task)
		}
	}

	return tasks, nil
}

func (k *kanban) CreateTask(ctx context.Context, data map[string]string) error {
	tags := []*model.Tag{}
	for _, value := range strings.Split(data["TagsString"], ",") {
		if value = strings.TrimSpace(value); value == "" {
			continue
		}

		tag, err := k.taskRepository.GetOrCreateTag(ctx, value)
		if err != nil {
			return errors.Wrapf(err, "cannot gat or create tag '%s'", value)
		}
		tags = append(tags, tag)
	}

	columnID, _ := strconv.Atoi(data["ColumnID"])
	column, err := k.columnRepository.Get(ctx, columnID)
	if err != nil {
		return errors.Wrapf(err, "cannot load column '%d'", columnID)
	}

	task := model.Task{
		Title:       data["Title"],
		Description: data["Description"],
		Tags:        tags,
		Column:      column,
		ColumnID:    int(column.ID),
	}
	_ = task.Color.Scan(data["Color"])
	if err := k.taskRepository.Save(ctx, &task); err != nil {
		return errors.Wrapf(err, "cannot save task '%+v'", task)
	}
	if err := k.taskRepository.SetPosition(ctx, column.ID, task.ID); err != nil {
		return errors.Wrapf(err, "cannot set position for task '%+v'", task)
	}
	if err := k.taskLogRepository.LogTask(ctx, column.ID, task.ID, "create"); err != nil {
		return errors.Wrapf(err, "cannot log task action '%+v'", task)
	}

	return nil
}

func (k *kanban) ToggleCheckbox(ctx context.Context, id, checkboxID int) error {
	task, err := k.taskRepository.Get(ctx, id)
	if err != nil {
		return err
	}
	task.Description = markdown.ToggleCheckbox(task.Description, checkboxID)
	return k.taskRepository.Save(ctx, task)
}

func (k *kanban) MoveTaskTo(ctx context.Context, taskID, newPosition, newColumnID int) error {
	task, err := k.taskRepository.Get(ctx, taskID)
	if err != nil {
		return errors.Wrapf(err, "cannot fetch task '%d'", taskID)
	}

	err = k.taskRepository.UpdateTaskPosition(ctx, task, newPosition, newColumnID)
	if err != nil {
		return errors.Wrapf(err, "cannot update position for task '%+v'", task)
	}
	if err := k.taskLogRepository.LogTask(ctx, uint(taskID), uint(task.ColumnID), "move column"); err != nil {
		return errors.Wrapf(err, "cannot log task action '%+v'", task)
	}

	return nil
}

func (k *kanban) UpdateTask(ctx context.Context, id int, data map[string]string) error {
	task, err := k.taskRepository.Get(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "error getting task '%d'", id)
	}

	if title, ok := data["Title"]; ok {
		task.Title = title
	}
	if description, ok := data["Description"]; ok {
		task.Description = description
	}
	var tags []*model.Tag
	if tagsString, ok := data["TagsString"]; ok {
		for _, value := range strings.Split(tagsString, ",") {
			if value = strings.TrimSpace(value); value == "" {
				continue
			}

			tag, err := k.taskRepository.GetOrCreateTag(ctx, value)
			if err != nil {
				return err
			}
			tags = append(tags, tag)
		}
	}
	task.Tags = tags
	if color, ok := data["Color"]; ok {
		_ = task.Color.Scan(color)
	}
	k.taskLogRepository.LogTask(ctx, uint(id), uint(task.ColumnID), "update task")
	return k.taskRepository.Update(ctx, task)
}

func (k *kanban) DeleteTask(ctx context.Context, id int) error {
	task, err := k.taskRepository.Get(ctx, id)
	if err != nil {
		return err
	}
	if err := k.taskRepository.DeleteTask(ctx, task); err != nil {
		return err
	}
	if err := k.taskLogRepository.LogTask(ctx, uint(id), uint(task.ColumnID), "delete task"); err != nil {
		return nil
	}
	return nil
}
