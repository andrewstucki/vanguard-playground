package server

import (
	"context"
	"errors"
	"time"

	"github.com/andrewstucki/vanguard-playground/internal/models"
)

func RunWorker(ctx context.Context) (ret error) {
	logger, writer := NewLogger()
	defer func() {
		if err := writer.Close(); err != nil {
			ret = errors.Join(ret, err)
		}
	}()

	logger = logger.With().Str("component", "worker").Logger()

	handler := &handler{
		logger: logger,
	}

	backend, err := models.NewBackend(models.BackendConfig{
		Logger:  logger,
		Handler: handler,
		// can only be run in persistent mode
		Persistent: true,
	})
	if err != nil {
		return err
	}
	handler.backend = backend
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		if err := handler.backend.Shutdown(shutdownCtx); err != nil {
			logger.Err(err).Msg("Error shutting processor down cleanly")
			ret = errors.Join(ret, err)
		}
		cancel()
	}()

	if err := handler.backend.Start(ctx); err != nil {
		return err
	}

	<-ctx.Done()

	return nil
}
