package app

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"r3-ti-faceattend/backend/internal/config"
	"r3-ti-faceattend/backend/internal/database"
	"r3-ti-faceattend/backend/internal/health"
)

func Run() {
	if err := run(); err != nil {
		log.Fatalf("api stopped: %v", err)
	}
}

func run() error {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db := database.New(cfg.Database)
	if err := db.Ping(ctx); err != nil {
		log.Print("database is not ready")
	}
	defer db.Close()

	handler := newHTTPHandler(cfg, db)
	server := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("api listening on %s", server.Addr)
		errCh <- server.ListenAndServe()
	}()

	return waitForServer(ctx, server, errCh)
}

func waitForServer(ctx context.Context, server *http.Server, errCh <-chan error) error {
	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

func newHTTPHandler(cfg config.Config, db health.DatabasePinger) http.Handler {
	mux := http.NewServeMux()
	healthHandler := health.NewHandler(cfg.AppEnv, db)
	mux.Handle("/health", healthHandler)
	mux.Handle("/api/v1/health", healthHandler)

	return mux
}
