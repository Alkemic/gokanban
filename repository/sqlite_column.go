package repository

import (
	"github.com/jinzhu/gorm"

	"gokanban/model"
)

type sqliteColumnRepository struct {
	db *gorm.DB
}

func NewSqliteColumnRepository(db *gorm.DB) *sqliteColumnRepository {
	return &sqliteColumnRepository{
		db: db,
	}
}

func (r *sqliteColumnRepository) List() ([]*model.Column, error) {
	columns := []*model.Column{}
	q := r.db.Order("position asc").Find(&columns)
	if q.Error != nil {
		return nil, q.Error
	}
	return columns, nil
}

func (r *sqliteColumnRepository) Get(id int) (*model.Column, error) {
	var column model.Column
	if err := r.db.Where("id = ?", id).Find(&column).Error; err != nil {
		return nil, err
	}
	return &column, nil
}

func (r *sqliteColumnRepository) Init() {
	r.db.FirstOrCreate(&model.Column{}, &model.Column{Name: "Backlog", Limit: 10, Position: 1})
	r.db.FirstOrCreate(&model.Column{}, &model.Column{Name: "To Do", Limit: 10, Position: 2})
	r.db.FirstOrCreate(&model.Column{}, &model.Column{Name: "WiP", Limit: 10, Position: 3})
	r.db.FirstOrCreate(&model.Column{}, &model.Column{Name: "Done", Limit: 10, Position: 4})
}
