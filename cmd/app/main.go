package main

import (
	"github.com/gmlazutin/comparch-lab-3mod-3/internal/api"
	"github.com/gmlazutin/comparch-lab-3mod-3/internal/api/gin"
	"github.com/gmlazutin/comparch-lab-3mod-3/internal/logging"
	"github.com/gmlazutin/comparch-lab-3mod-3/internal/service"
	"github.com/gmlazutin/comparch-lab-3mod-3/internal/service/auth"
	"github.com/gmlazutin/comparch-lab-3mod-3/internal/service/auth/session"
	"github.com/gmlazutin/comparch-lab-3mod-3/internal/service/contactbook"
	"github.com/gmlazutin/comparch-lab-3mod-3/internal/storage"
	"github.com/gmlazutin/comparch-lab-3mod-3/internal/storage/gorm"
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func getLogLvl() slog.Leveler {
	a, err := strconv.Atoi(os.Getenv("APP_LOG_LVL"))
	if err != nil {
		return slog.LevelInfo
	}

	return slog.Level(a)
}

func getAddr() string {
	addr := os.Getenv("APP_LISTEN_ADDR")
	if len(addr) == 0 {
		addr = ":8080"
	}
	return addr
}

func getDB() (string, string) {
	return os.Getenv("APP_DB"), os.Getenv("APP_DSN")
}

func getJWTSecret() []byte {
	return []byte(os.Getenv("APP_JWT_SECRET"))
}

func main() {
	logger := logging.InitLogger(getLogLvl())

	db, dsn := getDB()
	gormdb, err := gorm.New(gorm.Options{
		Driver: db,
		Dsn:    dsn,
		Opts: storage.Options{
			Logger: logger,
		},
	})
	if err != nil {
		logger.Error("failed to init GORM db conn", logging.Error(err))
		return
	}
	defer gormdb.Stop()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := gormdb.PerfomMigrations(ctx); err != nil {
		logger.Error("failed to perform GORM migrations", logging.Error(err))
		return
	}

	sopts := service.ServiceOptions{
		Logger: logger,
	}
	aservice := auth.New(auth.Options{
		Storage:                   gormdb,
		ServiceOpts:               sopts,
		SessionValidatorGenerator: session.NewJWTSessionProvider(getJWTSecret()),
		SessionExpireTimeout:      time.Minute * 10,
	})

	cbservice := contactbook.New(contactbook.Options{
		ContactStorage: gormdb,
		PhoneStorage:   gormdb,
		ServiceOpts:    sopts,
	})

	srv := gin.NewAPIServer(gin.Options{
		Opts: api.APIServerOptions{
			AuthService:        aservice,
			ContactbookService: cbservice,
			Logger:             logger,
			Addr:               getAddr(),
		},
	})

	logger.Info("Running web server...", slog.String("addr", getAddr()))

	errCh := make(chan error, 1)
	go func() {
		if err := srv.Start(); err != nil {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		logger.Info("Shutting down server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Stop(shutdownCtx); err != nil {
			logger.Error("unable to shutdown server gracefully", logging.Error(err))
		}

	case err := <-errCh:
		logger.Error("failed to start server", logging.Error(err))
	}
}
