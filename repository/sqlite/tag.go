package repository

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"

	"gokanban/model"
)

const (
	getTagByNameSQL     = `SELECT id, name FROM tag WHERE name = ?;`
	fetchTagsForTaskSQL = `SELECT tag.id, tag.name FROM task_tags LEFT JOIN task ON task.id == task_tags.task_id LEFT JOIN tag ON tag.id == task_tags.tag_id WHERE task.id = ?;`
	saveTagSQL          = `INSERT INTO tag ("name") VALUES (?);`
)

func (r *sqliteTaskRepository) GetOrCreateTag(ctx context.Context, name string) (*model.Tag, error) {
	tag, err := r.GetTagByName(ctx, name)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			tag = &model.Tag{Name: name}
			tag, err = r.SaveTag(ctx, tag)
			if err != nil {
				return nil, errors.Wrapf(err, "cannot create new tag '%s'", name)
			}
			return tag, nil
		}
		return nil, errors.Wrapf(err, "cannot fetch tag '%s'", name)
	}

	return tag, nil
}

func (r *sqliteTaskRepository) GetTagByName(ctx context.Context, name string) (*model.Tag, error) {
	row := r.db.QueryRowContext(ctx, getTagByNameSQL, name)
	tag, err := readTagFromRow(row)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot read tag from rows")
	}

	return tag, nil
}

func (r *sqliteTaskRepository) SaveTag(ctx context.Context, tag *model.Tag) (*model.Tag, error) {
	res, err := r.db.ExecContext(ctx, saveTagSQL, tag.Name)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot save tag '%+v'", tag)
	}

	lastInsertedID, err := res.LastInsertId()
	if err != nil {
		return nil, errors.Wrapf(err, "cannot fetch last inserted ID for tag '%+v'", tag)
	}
	tag.ID = uint(lastInsertedID)

	return tag, nil
}

func (r *sqliteTaskRepository) FetchTagsForTask(ctx context.Context, task *model.Task) ([]*model.Tag, error) {
	rows, err := r.db.QueryContext(ctx, fetchTagsForTaskSQL, task.ID)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot fetch tags for task '%+v'", task)
	}
	defer rows.Close()

	var tags []*model.Tag
	for rows.Next() {
		tag, err := readTag(rows)
		if err != nil {
			return nil, errors.Wrap(err, "cannot read tag from row")
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

func readTag(rows *sql.Rows) (*model.Tag, error) {
	var tag model.Tag
	err := rows.Scan(&tag.ID, &tag.Name)
	if err != nil {
		return nil, errors.Wrap(err, "cannot scan data from row")
	}

	return &tag, nil
}

func readTagFromRow(row *sql.Row) (*model.Tag, error) {
	var tag model.Tag
	err := row.Scan(&tag.ID, &tag.Name)
	if err != nil {
		return nil, errors.Wrap(err, "cannot scan data from row")
	}

	return &tag, nil
}
