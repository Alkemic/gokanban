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
