package messages

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/gorilla/websocket"
)

func NewWebSocketListener(logger *slog.Logger, signalClient *SignalClient, wsURL string) *WebSocketListener {
	return &WebSocketListener{
		logger:       logger,
		signalClient: signalClient,
		wsURL:        wsURL,
	}
}

type GotifyMessage struct {
	ID       int       `json:"id"`
	AppID    int       `json:"appid"`
	Message  string    `json:"message"`
	Title    string    `json:"title"`
	Priority int       `json:"priority"`
	Date     time.Time `json:"date"`
}

type WebSocketListener struct {
	logger       *slog.Logger
	signalClient *SignalClient
	wsURL        string
}

func (w WebSocketListener) Listen(ctx context.Context) {
	const createConnectionTimeout = 2 * time.Minute
	timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), createConnectionTimeout)
	defer timeoutCancel()

	conn, err := w.createConnection(timeoutCtx, w.wsURL)
	if err != nil {
		w.logger.ErrorContext(timeoutCtx, "unable to create websocket connection", "error", err)
	}

	defer func() {
		err = conn.Close()
		if err != nil {
			w.logger.Error("unable to close websocket connection", "error", err)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			w.logger.InfoContext(ctx, "websocket listener stopped")
			return
		default:
			var msg GotifyMessage
			err := conn.ReadJSON(&msg)
			if err != nil {
				w.logger.ErrorContext(ctx, "cannot read gotify message from server", "error", err)
			}
			w.logger.DebugContext(ctx, "received gotify message", "message", msg)
			err = w.signalClient.Send(&msg)
			if err != nil {
				w.logger.ErrorContext(ctx, "cannot send message to signal", "error", err)
			}
		}
	}
}

func (w WebSocketListener) createConnection(ctx context.Context, url string) (*websocket.Conn, error) {
	const retryInterval = 5 * time.Second

	w.logger.InfoContext(ctx, "opening websocket connection", "url", url)
	for {
		conn, resp, err := websocket.DefaultDialer.DialContext(ctx, url, nil)
		e := resp.Body.Close()
		if e != nil {
			w.logger.ErrorContext(ctx, "unable to close response body", "error", e)
		}

		if err == nil {
			return conn, nil
		}

		logger.Error("unable to open websocket connection, retrying in 5s", "url", url)
		timer := time.NewTimer(retryInterval)

		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, fmt.Errorf("opening websocket connection failed: %w", ctx.Err())
		case <-timer.C:
			// retry
		}
	}
}
