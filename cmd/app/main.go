package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/gmlazutin/comparch-lab-3mod-3/internal/api"
	"github.com/gmlazutin/comparch-lab-3mod-3/internal/api/gin"
	"github.com/gmlazutin/comparch-lab-3mod-3/internal/front"
	"github.com/gmlazutin/comparch-lab-3mod-3/internal/logging"
	"github.com/gmlazutin/comparch-lab-3mod-3/internal/service"
	"github.com/gmlazutin/comparch-lab-3mod-3/internal/service/auth"
	"github.com/gmlazutin/comparch-lab-3mod-3/internal/service/auth/session"
	"github.com/gmlazutin/comparch-lab-3mod-3/internal/service/contactbook"
	"github.com/gmlazutin/comparch-lab-3mod-3/internal/storage"
	"github.com/gmlazutin/comparch-lab-3mod-3/internal/storage/gorm"
)

func getLogLvl() slog.Leveler {
	a, err := strconv.Atoi(os.Getenv("APP_LOG_LVL"))
	if err != nil {
		return slog.LevelInfo
	}

	return slog.Level(a)
}

func getAddr() string {
	return os.Getenv("APP_LISTEN_ADDR")
}

func getDB() (string, string) {
	return os.Getenv("APP_DB"), os.Getenv("APP_DSN")
}

func getJWTSecret() []byte {
	return []byte(os.Getenv("APP_JWT_SECRET"))
}

func getPublicUrl() string {
	origin := os.Getenv("APP_PUBLIC_URL")
	return origin
}

func getSessionExp() time.Duration {
	tm, err := strconv.Atoi(os.Getenv("APP_SESSION_EXPIRE_TIMEOUT"))
	if err != nil {
		return 0
	}
	return time.Duration(tm) * time.Minute
}

func getDBExp() time.Duration {
	tm, err := strconv.Atoi(os.Getenv("APP_DB_FLUSH_INTERVAL"))
	if err != nil {
		return time.Minute * 10
	}
	return time.Duration(tm) * time.Minute
}

func main() {
	logger := logging.InitLogger(getLogLvl())

	var wg sync.WaitGroup
	defer wg.Wait()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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

	if err := gormdb.PerfomMigrations(ctx); err != nil {
		logger.Error("failed to perform GORM migrations", logging.Error(err))
		return
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			err := gormdb.PerformFlush(ctx, gorm.DB_FLUSH_DEFAULT_UNIT)
			if err != nil {
				return
			}
			select {
			case <-time.After(getDBExp()):
				continue
			case <-ctx.Done():
				return
			}
		}
	}()

	sopts := service.ServiceOptions{
		Logger:   logger,
		Transact: gormdb,
	}
	aservice := auth.New(auth.Options{
		Storage:                   gormdb,
		ServiceOpts:               sopts,
		SessionValidatorGenerator: session.NewJWTSessionProvider(getJWTSecret()),
		SessionExpireTimeout:      getSessionExp(),
	})

	cbservice := contactbook.New(contactbook.Options{
		ContactStorage: gormdb,
		PhoneStorage:   gormdb,
		ServiceOpts:    sopts,
	})

	srv, err := gin.NewAPIServer(gin.Options{
		Opts: api.APIServerOptions{
			AuthService:        aservice,
			ContactbookService: cbservice,
			Logger:             logger,
			Addr:               getAddr(),
		},
		PublicUrl: getPublicUrl(),
		StaticFS:  front.FS(),
	})
	if err != nil {
		logger.Error("failed to create GIN api server", logging.Error(err))
		return
	}

	logger.Info("Running web server...", slog.String("addr", srv.Addr()))

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
