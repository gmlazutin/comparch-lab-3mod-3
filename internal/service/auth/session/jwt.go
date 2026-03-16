package session

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"time"

	"github.com/gmlazutin/comparch-lab-3mod-3/internal/service"

	"github.com/golang-jwt/jwt/v5"
)

var _ AuthTokenValidatorGenerator = (*JWTSessionProvider)(nil)

type JWTSessionProvider struct {
	secret []byte
}

func NewJWTSessionProvider(secret []byte) *JWTSessionProvider {
	if len(secret) == 0 {
		secret = make([]byte, 32)
		_, err := io.ReadFull(rand.Reader, secret)
		if err != nil {
			panic("jwtSessionProvider: unable to read secret for JWT from cryptosource (because secret param is empty): " + err.Error())
		}
	}
	return &JWTSessionProvider{
		secret: secret,
	}
}

func (p *JWTSessionProvider) wrapErr(err error) error {
	return fmt.Errorf("jwtSessionProvider: %w", err)
}

func (p *JWTSessionProvider) Validate(ctx context.Context, token string, ts time.Time) (*Session, error) {
	tkn, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return p.secret, nil
	})
	if err != nil {
		return nil, p.wrapErr(fmt.Errorf("error while parsing token: %w", service.ErrInvalidToken))
	}

	if !tkn.Valid {
		return nil, p.wrapErr(service.ErrInvalidToken)
	}

	//todo: clarify each error reason in fmt.Errorf before passing into wrapErr

	claims, ok := tkn.Claims.(jwt.MapClaims)
	if !ok {
		return nil, p.wrapErr(service.ErrInvalidToken)
	}
	subFloat, ok := claims["sub"].(float64)
	if !ok {
		return nil, p.wrapErr(service.ErrInvalidToken)
	}
	expFloat, ok := claims["exp"].(float64)
	if !ok {
		return nil, p.wrapErr(service.ErrInvalidToken)
	}

	session := &Session{
		UserID:  uint(subFloat),
		Expires: time.Unix(int64(expFloat), 0),
	}

	if ts.After(session.Expires) {
		return nil, p.wrapErr(service.ErrInvalidToken)
	}

	return session, nil
}

func (p *JWTSessionProvider) Generate(ctx context.Context, session Session) (string, error) {
	claims := jwt.MapClaims{
		"sub": session.UserID,
		"exp": session.Expires.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(p.secret)
	if err != nil {
		return "", p.wrapErr(fmt.Errorf("error while signing token: %w", err))
	}
	return signed, nil
}
