package server

import (
	"context"
	"fmt"
	"net/http"
	"sync"
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
)

type messageSend struct {
	state           playgroundv1.MessageState
	simulateFailure bool
	id              string
}

type handler struct {
	logger zerolog.Logger

	processor *workflows.WorkflowProcessor

	messages map[string]string
	mutex    sync.RWMutex

	sends     map[string]messageSend
	sendMutex sync.RWMutex
}

var _ playgroundv1connect.MessageServiceHandler = (*handler)(nil)

func (h *handler) GetMessage(_ context.Context, req *connect.Request[playgroundv1.GetMessageRequest]) (*connect.Response[playgroundv1.GetMessageResponse], error) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if text, ok := h.messages[req.Msg.MessageId]; ok {
		return connect.NewResponse(&playgroundv1.GetMessageResponse{
			Message: &playgroundv1.Message{
				MessageId: req.Msg.MessageId,
				Text:      text,
			},
		}), nil
	}

	return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("message with ID %q not found", req.Msg.MessageId))
}

func (h *handler) CreateMessage(_ context.Context, req *connect.Request[playgroundv1.CreateMessageRequest]) (*connect.Response[playgroundv1.CreateMessageResponse], error) {
	id := uuid.New().String()

	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.messages[id] = req.Msg.Text
	return connect.NewResponse(&playgroundv1.CreateMessageResponse{
		MessageId: id,
	}), nil
}

func (h *handler) DeleteMessage(_ context.Context, req *connect.Request[playgroundv1.DeleteMessageRequest]) (*connect.Response[playgroundv1.DeleteMessageResponse], error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	delete(h.messages, req.Msg.MessageId)
	return connect.NewResponse(&playgroundv1.DeleteMessageResponse{}), nil
}

func (h *handler) ListMessages(context.Context, *connect.Request[playgroundv1.ListMessagesRequest]) (*connect.Response[playgroundv1.ListMessagesResponse], error) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	var messages []*playgroundv1.Message
	for id, text := range h.messages {
		messages = append(messages, &playgroundv1.Message{
			MessageId: id,
			Text:      text,
		})
	}

	return connect.NewResponse(&playgroundv1.ListMessagesResponse{
		Messages: messages,
	}), nil
}

func (h *handler) SendMessage(ctx context.Context, req *connect.Request[playgroundv1.SendMessageRequest]) (*connect.Response[playgroundv1.SendMessageResponse], error) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if _, ok := h.messages[req.Msg.MessageId]; ok {
		operationID := uuid.New().String()

		h.sendMutex.Lock()
		defer h.sendMutex.Unlock()
		h.sends[operationID] = messageSend{
			state:           playgroundv1.MessageState_SENDING,
			id:              req.Msg.MessageId,
			simulateFailure: req.Msg.SimulateFailure,
		}

		_, err := h.processor.RunWorkflow(context.Background(), playgroundv1.SendMessageStateWorkflow, playgroundv1.SendMessageState{
			OperationId: operationID,
			State:       playgroundv1.MessageState_SENDING,
		})
		if err != nil {
			delete(h.sends, operationID)
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("error scheduling workflow: %w", err))
		}

		return connect.NewResponse(&playgroundv1.SendMessageResponse{
			MessageId:   req.Msg.MessageId,
			OperationId: operationID,
		}), nil
	}

	return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("message with ID %q not found", req.Msg.MessageId))
}

func (h *handler) MessageStatus(ctx context.Context, req *connect.Request[playgroundv1.MessageStatusRequest]) (*connect.Response[playgroundv1.MessageStatusResponse], error) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if _, ok := h.messages[req.Msg.MessageId]; !ok {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("message with ID %q not found", req.Msg.MessageId))
	}

	h.sendMutex.RLock()
	defer h.sendMutex.RUnlock()
	state, ok := h.sends[req.Msg.OperationId]
	if !ok || state.id != req.Msg.MessageId {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("message with ID %q has no operation with ID %q not found", req.Msg.MessageId, req.Msg.OperationId))
	}

	return connect.NewResponse(&playgroundv1.MessageStatusResponse{
		State: state.state,
	}), nil
}

func (h *handler) Do(io *playgroundv1.SendMessageState) error {
	h.sendMutex.Lock()
	defer h.sendMutex.Unlock()

	state, ok := h.sends[io.OperationId]
	if !ok {
		return fmt.Errorf("operation not found: %s", io.OperationId)
	}

	if state.state != playgroundv1.MessageState_SENDING {
		// no-op since this is already processed
		return nil
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()
	message, ok := h.messages[state.id]
	if !ok {
		return fmt.Errorf("message not found: %s", state.id)
	}

	h.logger.Info().Str("message", message).Msg("processing")

	nextState := playgroundv1.MessageState_SUCCEEDED
	if state.simulateFailure {
		nextState = playgroundv1.MessageState_FAILED
	}

	io.State = nextState
	state.state = nextState
	h.sends[io.OperationId] = state

	return nil
}

func Run(ctx context.Context, port int) error {
	logger := log.With().Str("component", "server").Logger()

	validator, err := validate.NewInterceptor()
	if err != nil {
		logger.Err(err).Msg("error creating interceptor")
		return err
	}

	handler := &handler{
		logger:   logger,
		messages: make(map[string]string),
		sends:    make(map[string]messageSend),
	}

	handler.processor = workflows.NewWorkflowProcessorBuilder().WithLogger(backend.DefaultLogger()).Register(
		playgroundv1.NewSendMessageStateWorkflowRegistration(handler),
	).Build()

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
