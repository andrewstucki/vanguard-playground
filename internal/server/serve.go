package server

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/vanguard"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	playgroundv1 "github.com/andrewstucki/vanguard-playground/internal/gen/playground/v1"
	"github.com/andrewstucki/vanguard-playground/internal/gen/playground/v1/playgroundv1connect"
)

type handler struct {
	logger zerolog.Logger

	messages map[string]string
	mutex    sync.RWMutex
}

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

func Run(ctx context.Context, port int) error {
	logger := log.With().Str("component", "server").Logger()

	service := vanguard.NewService(playgroundv1connect.NewMessageServiceHandler(&handler{
		logger:   logger,
		messages: make(map[string]string),
	}))

	transcoder, err := vanguard.NewTranscoder([]*vanguard.Service{service})
	if err != nil {
		logger.Err(err).Msg("Error creating transcoder")
		return err
	}

	server := &http.Server{Addr: fmt.Sprintf("localhost:%d", port), Handler: transcoder}

	chErr := make(chan error, 1)
	logger.Debug().Int("port", port).Msg("Starting server")
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			chErr <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		logger.Debug().Msg("Shutting down server")
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Err(err).Msg("Error shutting server down cleanly")
			return err
		}
	case err := <-chErr:
		logger.Err(err).Msg("Error running server")
		return err
	}
	return nil
}
