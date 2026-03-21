package storage

import "context"

type TransactFunc func(ctx context.Context) error

type Transact interface {
	Transact(ctx context.Context, fc TransactFunc) error
}
