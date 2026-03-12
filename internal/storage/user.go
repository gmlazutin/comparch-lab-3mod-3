package storage

import (
	"context"
)

const (
	UserField string = "user"
)

type UserPassword struct {
	Hash string
	Algo string
}

type User struct {
	ID       uint
	Login    string
	Password UserPassword
}

type AddUserData struct {
	//ID will be overwritten
	User User
}

type GetUserData struct {
	ID              uint
	Login           string
	WithCredentials bool
}

type UserStorage interface {
	AddUser(ctx context.Context, data AddUserData) (*User, error)
	GetUser(ctx context.Context, data GetUserData) (*User, error)
	//UpdateUser(ctx context.Context, data UserUpdateData) (*UserProfile, error)
	//DeleteUser(ctx context.Context, id uint) error
}
