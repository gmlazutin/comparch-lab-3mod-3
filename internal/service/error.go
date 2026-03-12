package service

import "errors"

var (
	ErrIncorrectPassword = errors.New("incorrect password")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")

	ErrInvalidToken = errors.New("invalid token")

	ErrContactNotFound = errors.New("contact not found")
)
