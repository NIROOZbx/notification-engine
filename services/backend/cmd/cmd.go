package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NIROOZbx/notification-engine/services/backend/app"
)

func Run(a *app.App, port string) error {
	signals := []os.Signal{os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP}
	ctx, cancel := signal.NotifyContext(context.Background(), signals...)

	defer cancel()

	errChan := make(chan error, 1)

	go func() {
		if err := a.Server.Listen(port); err != nil {
			errChan <- fmt.Errorf("problem in starting app server")
			return
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		a.Logger.Info().Msg("shutdown signal received")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	a.Logger.Info().Msg("closing server and connections...")
	if err := shutdown(a, shutdownCtx); err != nil {
		return fmt.Errorf("shutdown failed: %w", err)
	}
	return nil 
}

func shutdown(a *app.App, ctx context.Context) error {
	a.Logger.Info().Msg("shutting down server")
	if err := a.Server.ShutdownWithContext(ctx); err != nil {
		a.Logger.Error().Err(err).Msg("fiber shutdown error")
	}

	if err := a.Redis.Close(); err != nil {
		a.Logger.Error().Err(err).Msg("redis close error")
	}
	a.DB.Close()

	a.Logger.Info().Msg("shutdown complete")

	return nil

}
