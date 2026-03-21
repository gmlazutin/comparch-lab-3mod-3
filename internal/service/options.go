package service

import (
	"log/slog"

	"github.com/gmlazutin/comparch-lab-3mod-3/internal/storage"
)

type ServiceOptions struct {
	Logger   *slog.Logger
	Transact storage.Transact
}
