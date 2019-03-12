package repository

import (
	"context"
	"database/sql"
)

const (
	saveTaskLogSQL = `INSERT INTO task_log ("task_id", "old_column_id", "action", "created_at") VALUES (?, ?, ?, DATETIME('now'));`
)

type sqliteTaskLogRepository struct {
	db *sql.DB
}

func NewSQLiteTaskLogRepository(db *sql.DB) *sqliteTaskLogRepository {
	return &sqliteTaskLogRepository{
		db: db,
	}
}

func (r *sqliteTaskLogRepository) LogTask(ctx context.Context, columnID, taskID uint, action string) error {
	_, err := r.db.ExecContext(ctx, saveTaskLogSQL, taskID, columnID, action)
	if err != nil {
		return err
	}
	return nil
}
