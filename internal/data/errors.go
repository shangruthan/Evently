package data

import "errors"

var (
	ErrConflict  = errors.New("edit conflict")
	ErrDuplicate = errors.New("duplicate record")
	ErrNotFound  = errors.New("record not found")
)
