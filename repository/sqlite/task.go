package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"

	"github.com/pkg/errors"

	"gokanban/model"
	"gokanban/repository"
)

const (
	makeGapeSQL    = `UPDATE task SET position = position + 1 WHERE column_id = ? AND position >= ?;`
	moveTaskSQL    = `UPDATE task SET position = ?, column_id = ? WHERE id = ?`
	removeGapeSQL  = `UPDATE task SET position = position - 1 WHERE column_id = ? AND position > ?;`
	setPositionSQL = `UPDATE task SET position = (SELECT MAX(position) FROM task WHERE column_id = ? AND deleted_at IS NULL) + 1 WHERE id = ?;`

	getTaskSQL    = `SELECT "id", "title", "description", "column_id", "position", "color", "created_at", "updated_at", "deleted_at" FROM task WHERE deleted_at IS NULL AND id = ?;`
	listTasksSQL  = `SELECT "id", "title", "description", "column_id", "position", "color", "created_at", "updated_at", "deleted_at" FROM task WHERE deleted_at IS NULL AND column_id = ? ORDER BY position ASC;`
	deleteTaskSQL = `UPDATE task SET deleted_at = DATETIME('now') WHERE id = ?;`
	createTaskSQL = `INSERT INTO task("title", "description", "column_id", "position", "color", "created_at", "updated_at") VALUES (?, ?, ?, ?, ?, DATETIME('now'), DATETIME('now'));`
	updateTaskSQL = `UPDATE task SET "title" = ?, "description" = ?, "column_id" = ?, "position" = ?, "color" = ?, "updated_at" = DATETIME('now') WHERE deleted_at IS NULL AND id = ?;`

	deleteTagsForTaskSQL = `DELETE FROM task_tags WHERE task_id = ?;`
	insertTagForTaskSQL  = `INSERT INTO task_tags (task_id, tag_id) VALUES (?, ?);`
)

type sqliteTaskRepository struct {
	db *sqlx.DB
}

func NewSqliteTaskRepository(db *sqlx.DB) *sqliteTaskRepository {
	return &sqliteTaskRepository{
		db: db,
	}
}

func (r *sqliteTaskRepository) List(ctx context.Context, columnID uint) ([]*model.Task, error) {
	rows, err := r.db.QueryContext(ctx, listTasksSQL, columnID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []*model.Task{}
	for rows.Next() {
		task, err := readTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (r *sqliteTaskRepository) Get(ctx context.Context, id int) (*model.Task, error) {
	rows, err := r.db.QueryContext(ctx, getTaskSQL, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var task *model.Task
	for rows.Next() {
		if task != nil {
			return nil, repository.ErrMoreThanOneRow
		}

		var err error
		task, err = readTask(rows)
		if err != nil {
			return nil, err
		}
	}

	return task, nil
}

func (r *sqliteTaskRepository) Save(ctx context.Context, task *model.Task) error {
	tx, err := r.db.Begin()
	if err != nil {
		return errors.Wrap(err, "cannot start transaction")
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, createTaskSQL, task.Title, task.Description, task.ColumnID, task.Position, task.Color)
	if err != nil {
		return errors.Wrapf(err, "cannot save task '%+v'", task)
	}

	lastInsertedID, err := res.LastInsertId()
	if err != nil {
		return errors.Wrapf(err, "cannot fetch last inserted ID for task '%+v'", task)
	}
	task.ID = uint(lastInsertedID)

	if err = r.saveTags(ctx, tx, task); err != nil {
		return errors.Wrapf(err, "cannot save tags for task '%+v'", task)
	}

	if err = tx.Commit(); err != nil {
		return errors.Wrap(err, "cannot commit transaction")
	}

	return nil
}

func (r *sqliteTaskRepository) Update(ctx context.Context, task *model.Task) error {
	tx, err := r.db.Begin()
	if err != nil {
		return errors.Wrap(err, "cannot start transaction")
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, updateTaskSQL, task.Title, task.Description, task.ColumnID, task.Position, task.Color, task.ID)
	if err != nil {
		return err
	}
	if err = r.saveTags(ctx, tx, task); err != nil {
		return errors.Wrapf(err, "cannot save tags for task '%+v'", task)
	}

	if err = tx.Commit(); err != nil {
		return errors.Wrap(err, "cannot commit transaction")
	}

	return nil
}

func (r *sqliteTaskRepository) saveTags(ctx context.Context, tx *sql.Tx, task *model.Task) error {
	if task.ID == 0 {
		return repository.ErrUnsavedTask
	}

	if _, err := tx.ExecContext(ctx, deleteTagsForTaskSQL, task.ID); err != nil {
		return errors.Wrapf(err, "cannot delete old tags for taskID '%d'", task.ID)
	}

	for _, tag := range task.Tags {
		if _, err := tx.ExecContext(ctx, insertTagForTaskSQL, task.ID, tag.ID, task.ID); err != nil {
			return err
		}
	}

	return nil
}

func (r *sqliteTaskRepository) SetPosition(ctx context.Context, columnID, taskID uint) error {
	_, err := r.db.ExecContext(ctx, setPositionSQL, columnID, taskID)
	if err != nil {
		return err
	}
	return nil
}

func (r *sqliteTaskRepository) UpdateTaskPosition(ctx context.Context, task *model.Task, newPosition, newColumnID int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err = tx.ExecContext(ctx, makeGapeSQL, newColumnID, newPosition); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, moveTaskSQL, newPosition, newColumnID, task.ID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, removeGapeSQL, task.ColumnID, task.Position); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (r *sqliteTaskRepository) DeleteTask(ctx context.Context, task *model.Task) error {
	_, err := r.db.ExecContext(ctx, deleteTaskSQL, task.ID)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, removeGapeSQL, task.ColumnID, task.Position)
	if err != nil {
		return err
	}
	return nil
}

func readTask(rows *sql.Rows) (*model.Task, error) {
	var task model.Task
	err := rows.Scan(
		&task.ID, &task.Title, &task.Description, &task.ColumnID, &task.Position, &task.Color,
		&task.CreatedAt, &task.UpdatedAt, &task.DeletedAt,
	)
	if err != nil {
		return nil, err
	}

	return &task, nil
}
