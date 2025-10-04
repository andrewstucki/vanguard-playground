package models

import (
	"context"
	"database/sql"
	"errors"

	"github.com/andrewstucki/protoc-states/workflows"
	"github.com/rs/zerolog"

	playgroundv1 "github.com/andrewstucki/vanguard-playground/internal/gen/playground/v1"
)

type Backend struct {
	*Queries
	*workflows.WorkflowProcessor
	db      *sql.DB
	cleanup func()
}

type BackendConfig struct {
	Logger     zerolog.Logger
	Persistent bool
	Handler    playgroundv1.SendMessageStateWorkflowHandler
}

func (c BackendConfig) validate() error {
	if c.Handler == nil {
		return errors.New("handler must not be nil")
	}
	return nil
}

func NewBackend(config BackendConfig) (*Backend, error) {
	if err := config.validate(); err != nil {
		return nil, err
	}

	builder := workflows.NewWorkflowProcessorBuilder().WithLogger(&zerologBackendLogger{logger: config.Logger}).Register(
		playgroundv1.NewSendMessageStateWorkflowRegistration(config.Handler),
	)

	var err error
	var db *sql.DB
	var cleanup func()
	if config.Persistent {
		dbBuilder := workflows.NewLibSQLBackendBuilder()
		db, cleanup, err = dbBuilder.DB()
		if err != nil {
			return nil, err
		}
		builder.WithBackendFactory(dbBuilder.Build)
	} else {
		db, cleanup, err = workflows.MemoryDB()
		if err != nil {
			return nil, err
		}
	}

	if err := EnsureSchema(db); err != nil {
		return nil, err
	}

	return &Backend{
		Queries:           New(db),
		WorkflowProcessor: builder.Build(),
		db:                db,
		cleanup:           cleanup,
	}, nil
}

func (b *Backend) Tx(ctx context.Context) (*sql.Tx, *Queries, error) {
	tx, err := b.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}
	return tx, b.WithTx(tx), nil
}

func (b *Backend) Start(ctx context.Context) error {
	return b.WorkflowProcessor.Start(ctx)
}

func (b *Backend) Shutdown(ctx context.Context) error {
	if b.cleanup != nil {
		b.cleanup()
	}
	if err := b.WorkflowProcessor.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}
