package storage

import (
	"fmt"
)

type Selector struct {
	Offset uint
	Limit  uint
}

type MaxCountError struct {
	Field string
}

func (e MaxCountError) Error() string {
	return fmt.Sprintf("limit exceeded for %q", e.Field)
}

type MinCountError struct {
	Field string
}

func (e MinCountError) Error() string {
	return fmt.Sprintf("cannot delete %q due to minimum count constraint violation", e.Field)
}

type NotFoundError struct {
	Field string
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("%q not found", e.Field)
}

type AlreadyExistsError struct {
	Field string
}

func (e AlreadyExistsError) Error() string {
	return fmt.Sprintf("%q already exists", e.Field)
}
