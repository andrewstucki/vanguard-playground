package client

import (
	"fmt"
	"net/http"

	"github.com/andrewstucki/vanguard-playground/internal/gen/playground/v1/playgroundv1connect"
)

type Client struct {
	playgroundv1connect.MessageServiceClient
}

func NewClient(port int) *Client {
	return &Client{
		MessageServiceClient: playgroundv1connect.NewMessageServiceClient(
			http.DefaultClient,
			fmt.Sprintf("http://localhost:%d", port),
		),
	}
}
