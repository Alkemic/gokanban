package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"

	"gokanban/model"
	"gokanban/repository"
)

const (
	listColumnsSql = `SELECT "id", "name", "limit", "position", "created_at", "updated_at", "deleted_at" FROM "column" ORDER BY "position" ASC;`
	getColumnSql   = `SELECT "id", "name", "limit", "position", "created_at", "updated_at", "deleted_at" FROM "column" WHERE id = ?;`
)

type sqliteColumnRepository struct {
	db *sql.DB
}

func NewSqliteColumnRepository(dbpool *sql.DB) *sqliteColumnRepository {
	return &sqliteColumnRepository{
		db: dbpool,
	}
}

func (r *sqliteColumnRepository) List(ctx context.Context) ([]*model.Column, error) {
	rows, err := r.db.QueryContext(ctx, listColumnsSql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns := []*model.Column{}
	for rows.Next() {
		column, err := readColumnFromStatement(rows)
		if err != nil {
			return nil, err
		}
		columns = append(columns, column)
	}

	return columns, nil
}

func (r *sqliteColumnRepository) Get(ctx context.Context, id int) (*model.Column, error) {
	rows, err := r.db.QueryContext(ctx, getColumnSql, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var column *model.Column
	for rows.Next() {
		if column != nil {
			return nil, repository.ErrMoreThanOneRow
		}

		var err error
		column, err = readColumnFromStatement(rows)
		if err != nil {
			return nil, err
		}
	}

	return column, nil
}

func readColumnFromStatement(rows *sql.Rows) (*model.Column, error) {
	//var createdAtValue, updateAtValue string
	//var deleteAtValue model.NullTime
	var column model.Column
	err := rows.Scan(&column.ID, &column.Name, &column.Limit, &column.Position, &column.CreatedAt, &column.UpdatedAt, &column.DeletedAt)
	if err != nil {
		return nil, errors.Wrap(err, "cannot scan column from row")
	}

	//if deleteAtValue.Valid {
	//	//var err error
	//	*column.DeletedAt = deleteAtValue.Time
	//	//if *column.DeletedAt, err = parseTime(deleteAtValue); err != nil {
	//	//	return nil, errors.Wrapf(err, "error parsing 'column.deleted_at' = '%s'", deleteAtValue)
	//	//}
	//}
	//column.CreatedAt, err = parseTime(createdAtValue)
	//if err != nil {
	//	return nil, errors.Wrapf(err, "error parsing 'column.deleted_at' = '%s'", createdAtValue)
	//}
	//column.UpdatedAt, err = parseTime(updateAtValue)
	//if err != nil {
	//	return nil, errors.Wrapf(err, "error parsing 'column.deleted_at' = '%s'", updateAtValue)
	//}
	return &column, nil
}

func parseTime(v string) (time.Time, error) {
	//const longTimeFormat = "2006-01-02 15:04:05.999999999+07:00"
	//const longTimeFormat = "2006-01-02T15:04:05Z"
	const longTimeFormat = time.RFC3339
	parsed, err := time.Parse(longTimeFormat, v)
	if err != nil {
		return time.Time{}, err
	}
	return parsed, nil
}
