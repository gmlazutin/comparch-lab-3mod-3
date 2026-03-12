package session

import (
	"context"
	"time"
)

type Session struct {
	UserID  uint
	Expires time.Time
}

type AuthTokenValidator interface {
	Validate(ctx context.Context, token string, ts time.Time) (*Session, error)
}

type AuthTokenGenerator interface {
	Generate(ctx context.Context, session Session) (string, error)
}

type AuthTokenValidatorGenerator interface {
	AuthTokenValidator
	AuthTokenGenerator
}
