package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/validate"
	"connectrpc.com/vanguard"
	"github.com/google/uuid"
	"github.com/microsoft/durabletask-go/backend"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	"github.com/andrewstucki/protoc-states/workflows"
	playgroundv1 "github.com/andrewstucki/vanguard-playground/internal/gen/playground/v1"
	"github.com/andrewstucki/vanguard-playground/internal/gen/playground/v1/playgroundv1connect"
	"github.com/andrewstucki/vanguard-playground/internal/models"
)

type handler struct {
	logger zerolog.Logger

	db        *sql.DB
	queries   *models.Queries
	processor *workflows.WorkflowProcessor
}

var _ playgroundv1connect.MessageServiceHandler = (*handler)(nil)

func (h *handler) GetMessage(ctx context.Context, req *connect.Request[playgroundv1.GetMessageRequest]) (*connect.Response[playgroundv1.GetMessageResponse], error) {
	message, err := h.queries.GetMessage(ctx, req.Msg.MessageId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("message with ID %q not found", req.Msg.MessageId))
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&playgroundv1.GetMessageResponse{
		Message: &playgroundv1.Message{
			MessageId: message.ID,
			Text:      message.Text,
		},
	}), nil
}

func (h *handler) CreateMessage(ctx context.Context, req *connect.Request[playgroundv1.CreateMessageRequest]) (*connect.Response[playgroundv1.CreateMessageResponse], error) {
	id := uuid.New().String()

	message, err := h.queries.CreateMessage(ctx, models.CreateMessageParams{
		ID:   id,
		Text: req.Msg.Text,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&playgroundv1.CreateMessageResponse{
		MessageId: message.ID,
	}), nil
}

func (h *handler) DeleteMessage(ctx context.Context, req *connect.Request[playgroundv1.DeleteMessageRequest]) (*connect.Response[playgroundv1.DeleteMessageResponse], error) {
	if err := h.queries.DeleteMessage(ctx, req.Msg.MessageId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("message with ID %q not found", req.Msg.MessageId))
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&playgroundv1.DeleteMessageResponse{}), nil
}

func (h *handler) ListMessages(ctx context.Context, _ *connect.Request[playgroundv1.ListMessagesRequest]) (*connect.Response[playgroundv1.ListMessagesResponse], error) {
	queried, err := h.queries.ListMessages(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	var messages []*playgroundv1.Message
	for _, model := range queried {
		messages = append(messages, &playgroundv1.Message{
			MessageId: model.ID,
			Text:      model.Text,
		})
	}

	return connect.NewResponse(&playgroundv1.ListMessagesResponse{
		Messages: messages,
	}), nil
}

func (h *handler) SendMessage(ctx context.Context, req *connect.Request[playgroundv1.SendMessageRequest]) (*connect.Response[playgroundv1.SendMessageResponse], error) {
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	defer tx.Rollback()

	queries := h.queries.WithTx(tx)

	message, err := queries.GetMessage(ctx, req.Msg.MessageId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("message with ID %q not found", req.Msg.MessageId))
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	operationID := uuid.New().String()

	_, err = queries.CreateSentMessage(ctx, models.CreateSentMessageParams{
		ID:        operationID,
		MessageID: message.ID,
		Text:      message.Text,
		Result:    playgroundv1.MessageState_SENDING.String(),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	if err := tx.Commit(); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	_, err = h.processor.RunWorkflow(context.Background(), playgroundv1.SendMessageStateWorkflow, playgroundv1.SendMessageState{
		OperationId:     operationID,
		SimulateFailure: req.Msg.SimulateFailure,
		State:           playgroundv1.MessageState_SENDING,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("error scheduling workflow: %w", err))
	}

	return connect.NewResponse(&playgroundv1.SendMessageResponse{
		MessageId:   req.Msg.MessageId,
		OperationId: operationID,
	}), nil
}

func (h *handler) MessageStatus(ctx context.Context, req *connect.Request[playgroundv1.MessageStatusRequest]) (*connect.Response[playgroundv1.MessageStatusResponse], error) {
	operation, err := h.queries.GetSentMessage(ctx, models.GetSentMessageParams{
		ID:        req.Msg.OperationId,
		MessageID: req.Msg.MessageId,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("message with ID %q has no operation with ID %q not found", req.Msg.MessageId, req.Msg.OperationId))
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&playgroundv1.MessageStatusResponse{
		State: operation.Result,
	}), nil
}

func (h *handler) Do(io *playgroundv1.SendMessageState) error {
	ctx := context.Background()

	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	queries := h.queries.WithTx(tx)

	msg, err := queries.GetSentMessageByID(ctx, io.OperationId)
	if err != nil {
		return err
	}

	if msg.Result != playgroundv1.MessageState_SENDING.String() {
		// no-op since this is already processed
		return nil
	}

	nextState := playgroundv1.MessageState_SUCCEEDED
	if io.SimulateFailure {
		nextState = playgroundv1.MessageState_FAILED
	}

	io.State = nextState

	_, err = queries.UpdateSentMessage(ctx, models.UpdateSentMessageParams{
		ID:     io.OperationId,
		Result: nextState.String(),
	})
	if err != nil {
		return err
	}

	return tx.Commit()
}

func Run(ctx context.Context, port int, persist bool) error {
	logger := log.With().Str("component", "server").Logger()

	validator, err := validate.NewInterceptor()
	if err != nil {
		logger.Err(err).Msg("error creating interceptor")
		return err
	}

	handler := &handler{
		logger: logger,
	}

	builder := workflows.NewWorkflowProcessorBuilder().WithLogger(backend.DefaultLogger()).Register(
		playgroundv1.NewSendMessageStateWorkflowRegistration(handler),
	)

	var db *sql.DB
	var cleanup func()
	if persist {
		logger.Info().Msg("using persisten database")
		dbBuilder := workflows.NewLibSQLBackendBuilder()
		db, cleanup, err = dbBuilder.DB()
		if err != nil {
			return err
		}

		builder.WithBackendFactory(dbBuilder.Build)
	} else {
		logger.Info().Msg("using in-memory database")
		db, cleanup, err = workflows.MemoryDB()
		if err != nil {
			return err
		}
	}

	defer cleanup()
	if err := models.EnsureSchema(db); err != nil {
		return err
	}

	handler.db = db
	handler.queries = models.New(db)
	handler.processor = builder.Build()

	service := vanguard.NewService(playgroundv1connect.NewMessageServiceHandler(handler, connect.WithInterceptors(validator)))

	transcoder, err := vanguard.NewTranscoder([]*vanguard.Service{service})
	if err != nil {
		logger.Err(err).Msg("Error creating transcoder")
		return err
	}

	server := &http.Server{Addr: fmt.Sprintf("localhost:%d", port), Handler: transcoder}

	group, groupCtx := errgroup.WithContext(ctx)

	group.Go(func() error {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	})

	group.Go(func() error {
		return handler.processor.Start(groupCtx)
	})

	<-groupCtx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	logger.Debug().Msg("Shutting down server")
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Err(err).Msg("Error shutting server down cleanly")
		return err
	}

	if err := handler.processor.Shutdown(shutdownCtx); err != nil {
		log.Err(err).Msg("Error shutting processor down cleanly")
		return err
	}

	return group.Wait()
}
