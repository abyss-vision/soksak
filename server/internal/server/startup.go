package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Run starts the HTTP server and blocks until a shutdown signal is received.
// It performs a graceful shutdown with a 10-second timeout.
func Run(app *App) error {
	srv := &http.Server{
		Addr:    app.Config.Addr(),
		Handler: app.Router,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	errCh := make(chan error, 1)
	go func() {
		slog.Info("server listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-stop:
		slog.Info("shutting down gracefully")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return err
	}

	if app.DB != nil {
		if err := app.DB.Close(); err != nil {
			slog.Error("error closing db", "err", err)
		}
	}

	slog.Info("server stopped")
	return nil
}
