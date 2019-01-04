package repository

import (
	"strings"

	"github.com/jinzhu/gorm"

	"gokanban/model"
)

const (
	makeGapeSQL    = "update task set position = position + 1 where column_id = ? and position >= ?"
	moveTaskSQL    = "update task set position = ?, column_id = ? where id = ?"
	removeGapeSQL  = "update task set position = position - 1 where column_id = ? and position > ?"
	setPositionSQL = "update task set position = (select max(position) from task where column_id = ? and deleted_at is null) + 1 where id = ?"
)

type sqliteTaskRepository struct {
	db *gorm.DB
}

func NewSqliteTaskRepository(db *gorm.DB) *sqliteTaskRepository {
	return &sqliteTaskRepository{
		db: db,
	}
}

func (r *sqliteTaskRepository) List(args ...interface{}) ([]*model.Task, error) {
	tasks := []*model.Task{}
	q := r.db.Order("position asc")
	if len(args) > 0 {
		q = q.Where(args[0], args[1:]...)
	}
	q.Preload("Tags").Find(&tasks)

	return tasks, q.Error
}

func (r *sqliteTaskRepository) Get(id int) (*model.Task, error) {
	task := &model.Task{}
	q := r.db.Where("id = ?", id).Preload("Tags").Find(task)
	return task, q.Error
}

func (r *sqliteTaskRepository) GetOrCreateTag(name string) (*model.Tag, error) {
	var tag model.Tag
	q := r.db.FirstOrCreate(&tag, model.Tag{Name: strings.TrimSpace(name)})
	return &tag, q.Error
}

func (r *sqliteTaskRepository) Save(task *model.Task) error {
	return r.db.Save(task).Error
}

func (r *sqliteTaskRepository) SetPosition(columnID, taskID uint) error {
	return r.db.Exec(setPositionSQL, columnID, taskID).Error
}

func (r *sqliteTaskRepository) LogTask(columnID, taskID uint, action string) error {
	return r.db.Save(&model.TaskLog{TaskID: int(taskID), OldColumnID: int(columnID), Action: action}).Error
}

func (r *sqliteTaskRepository) UpdateTaskPosition(task *model.Task, newPosition, newColumnID int) error {
	if q := r.db.Exec(makeGapeSQL, newColumnID, newPosition); q.Error != nil {
		return q.Error
	}
	if q := r.db.Exec(moveTaskSQL, newPosition, newColumnID, task.ID); q.Error != nil {
		return q.Error
	}
	return r.db.Exec(removeGapeSQL, task.ColumnID, task.Position).Error
}

func (r *sqliteTaskRepository) DeleteTask(task *model.Task) error {
	r.db.Delete(task)
	if err := r.db.Error; err != nil {
		return err
	}
	r.db.Exec(removeGapeSQL, task.ColumnID, task.Position)
	if err := r.db.Error; err != nil {
		return err
	}
	return nil
}
