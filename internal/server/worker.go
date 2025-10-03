package server

import (
	"context"

	"github.com/andrewstucki/protoc-states/workflows"
	playgroundv1 "github.com/andrewstucki/vanguard-playground/internal/gen/playground/v1"
	"github.com/andrewstucki/vanguard-playground/internal/models"
	"github.com/microsoft/durabletask-go/backend"
	"github.com/rs/zerolog/log"
)

func RunWorker(ctx context.Context) error {
	logger := log.With().Str("component", "worker").Logger()

	handler := &handler{
		logger: logger,
	}

	builder := workflows.NewWorkflowProcessorBuilder().WithLogger(backend.DefaultLogger()).Register(
		playgroundv1.NewSendMessageStateWorkflowRegistration(handler),
	)

	// can only be run in persistent mode
	logger.Info().Msg("using persistent database")
	dbBuilder := workflows.NewLibSQLBackendBuilder()
	db, cleanup, err := dbBuilder.DB()
	if err != nil {
		return err
	}

	builder.WithBackendFactory(dbBuilder.Build)
	defer cleanup()
	if err := models.EnsureSchema(db); err != nil {
		return err
	}

	handler.db = db
	handler.queries = models.New(db)
	handler.processor = builder.Build()

	if err := handler.processor.Start(ctx); err != nil {
		return err
	}

	<-ctx.Done()
	return handler.processor.Shutdown(context.Background())
}
