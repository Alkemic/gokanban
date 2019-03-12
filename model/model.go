package model

import (
	"time"
)

type Task struct {
	ID                  uint
	Title               string
	Description         string
	DescriptionRendered string
	Color               NullString `json:"omitempty"`
	TaskProgress        map[string]int
	Tags                []*Tag
	Column              *Column
	ColumnID            int
	Position            int
	CreatedAt           time.Time
	UpdatedAt           time.Time
	DeletedAt           NullTime `json:"-"`
}

func (t *Task) MoveToColumn(id int) error {
	return nil
}

type Tag struct {
	ID   uint
	Name string
}

type Column struct {
	ID        uint
	Name      string
	Limit     int
	Position  int
	Tasks     []*Task
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt NullTime `json:"-"`
}

type TaskLog struct {
	ID          uint
	Action      string
	Task        Task
	TaskID      int
	OldColumn   Column
	OldColumnID int
	CreatedAt   time.Time
}
