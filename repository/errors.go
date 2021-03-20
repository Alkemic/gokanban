package repository

import (
	"errors"
)

var (
	ErrMoreThanOneRow = errors.New("statement returned more than one row")
	ErrUnsavedTask    = errors.New("task needs to be saved")
	ErrNotFound       = errors.New("not found")
)
