package api

import (
	"log/slog"

	"github.com/gmlazutin/comparch-lab-3mod-3/internal/service/auth"
	"github.com/gmlazutin/comparch-lab-3mod-3/internal/service/contactbook"
)

type APIServerOptions struct {
	AuthService        *auth.Service
	ContactbookService *contactbook.Service
	Logger             *slog.Logger
	Addr               string
}
