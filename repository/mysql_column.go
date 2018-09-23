package repository

import (
	"github.com/jinzhu/gorm"

	"github.com/Alkemic/gokanban/model"
)

type mySQLColumnRepository struct {
	db *gorm.DB
}

func NewMySQLColumnRepository(db *gorm.DB) *mySQLColumnRepository {
	return &mySQLColumnRepository{
		db: db,
	}
}

func (r *mySQLColumnRepository) List() ([]*model.Column, error) {
	columns := []*model.Column{}
	q := r.db.Order("position asc").Find(&columns)
	if q.Error != nil {
		return nil, q.Error
	}
	return columns, nil
}

func (r *mySQLColumnRepository) Get(id int) (*model.Column, error) {
	column := &model.Column{}
	q := r.db.Find(column)
	if q.Error != nil {
		return nil, q.Error
	}
	return column, nil
}

func (r *mySQLColumnRepository) Init() {
	r.db.FirstOrCreate(&model.Column{}, &model.Column{Name: "Backlog", Limit: 10, Position: 1})
	r.db.FirstOrCreate(&model.Column{}, &model.Column{Name: "To Do", Limit: 10, Position: 2})
	r.db.FirstOrCreate(&model.Column{}, &model.Column{Name: "WiP", Limit: 10, Position: 3})
	r.db.FirstOrCreate(&model.Column{}, &model.Column{Name: "Done", Limit: 10, Position: 4})
}
