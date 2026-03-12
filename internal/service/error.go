package service

import "errors"

var (
	ErrIncorrectPassword = errors.New("incorrect password")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")

	ErrInvalidToken = errors.New("invalid token")

	ErrContactNotFound = errors.New("contact not found")

	ErrIncorrectSelectorValues = errors.New("incorrect selector")
	ErrIncorrectPhone = errors.New("incorrect phone number")
	ErrIncorrectBirthday = errors.New("incorrect birthday date")
)
