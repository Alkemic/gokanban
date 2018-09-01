package use_case

import (
	"github.com/Alkemic/gokanban/helper"
	"github.com/Alkemic/gokanban/model"
)

type taskRepository interface {
	List(args ...interface{}) ([]*model.Task, error)
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
