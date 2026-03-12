package storage

import "errors"

type Selector struct {
	Offset uint
	Limit  uint
}

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)
