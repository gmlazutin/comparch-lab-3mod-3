package util

import (
	"github.com/gmlazutin/comparch-lab-3mod-3/internal/service"
	"github.com/gmlazutin/comparch-lab-3mod-3/internal/service/auth"
	"github.com/gmlazutin/comparch-lab-3mod-3/internal/service/auth/session"
	"context"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3filter"
)

func ValidateAuthTkn(ctx context.Context, header string, ai *openapi3filter.AuthenticationInput, authservice *auth.Service) (*session.Session, error) {
	scheme := ai.SecurityScheme
	if scheme.Type != "http" || scheme.Scheme != "bearer" || scheme.BearerFormat != "JWT" {
		return nil, service.ErrInvalidToken
	}
	if len(header) == 0 {
		return nil, service.ErrInvalidToken
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return nil, service.ErrInvalidToken
	}

	token := parts[1]

	sess, err := authservice.CheckUserSession(ctx, token, time.Now())
	if err != nil {
		return nil, service.ErrInvalidToken
	}

	return sess, nil
}
