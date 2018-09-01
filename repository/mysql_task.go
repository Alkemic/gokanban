package repository

import (
	"github.com/jinzhu/gorm"

	"github.com/Alkemic/gokanban/model"
)

type mysqlTaskRepository struct {
	db *gorm.DB
}

func NewMysqlTaskRepository(db *gorm.DB) *mysqlTaskRepository {
	return &mysqlTaskRepository{
		db: db,
	}
}

func (r *mysqlTaskRepository) List(args ...interface{}) ([]*model.Task, error) {
	tasks := []*model.Task{}
	q := r.db.Order("position asc")
	if len(args) > 0 {
		q = q.Where(args[0], args[1:]...)
	}
	q.Preload("Tags").Find(&tasks)

	if err := q.Error; err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *mysqlTaskRepository) Get(id int) (*model.Task, error) {
	task := &model.Task{}
	r.db.Where("id = ?", id).Preload("Tags", "Column").Find(task)
	if err := r.db.Error; err != nil {
		return nil, err
	}
	return task, nil
}
