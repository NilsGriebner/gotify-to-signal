package messages

import (
	"context"
	"log/slog"
	"time"

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
	const (
		createConnectionTimeout = 2 * time.Minute
		retryInterval           = 5 * time.Second
	)

	for {
		select {
		case <-ctx.Done():
			w.logger.InfoContext(ctx, "websocket listener stopped")
			return
		default:
		}

		dialCtx, dialCancel := context.WithTimeout(ctx, createConnectionTimeout)
		conn, err := w.createConnection(dialCtx, w.wsURL)
		dialCancel()
		if err != nil {
			w.logger.WarnContext(ctx, "unable to create websocket connection, retrying in 5s", "error", err)
			time.Sleep(retryInterval)
			continue
		}

		w.logger.InfoContext(ctx, "server websocket connection created")

		connCtx, connCancel := context.WithCancel(ctx)
		w.startKeepAlive(connCtx, conn)

		readErr := w.readFromServer(ctx, conn)

		connCancel()

		if closeErr := conn.Close(); closeErr != nil {
			w.logger.ErrorContext(ctx, "unable to close websocket connection", "error", closeErr)
		}

		if readErr != nil {
			w.logger.WarnContext(ctx, "websocket connection closed, reconnecting in 5s", "error", readErr)
		}
		time.Sleep(retryInterval)
	}
}

func (w WebSocketListener) createConnection(ctx context.Context, url string) (*websocket.Conn, error) {
	conn, resp, err := websocket.DefaultDialer.DialContext(ctx, url, nil)
	if err != nil {
		return nil, err
	}

	if resp.Body != nil {
		if err = resp.Body.Close(); err != nil {
			w.logger.ErrorContext(ctx, "unable to close response body", "error", err)
		}
	}
	return conn, nil
}

func (w WebSocketListener) readFromServer(ctx context.Context, serverConn *websocket.Conn) error {
	for {
		select {
		case <-ctx.Done():
			w.logger.InfoContext(ctx, "websocket listener stopped")
			return nil
		default:
			var msg GotifyMessage
			err := serverConn.ReadJSON(&msg)
			if err != nil {
				w.logger.ErrorContext(ctx, "cannot read gotify message from server", "error", err)
				return err
			}
			w.logger.DebugContext(ctx, "received gotify message", "message", msg)
			err = w.signalClient.Send(&msg)
			if err != nil {
				w.logger.ErrorContext(ctx, "cannot send message to signal", "error", err)
				return err
			}
		}
	}
}

func (w WebSocketListener) startKeepAlive(ctx context.Context, conn *websocket.Conn) {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					w.logger.WarnContext(ctx, "ping failed, stopping keepalive", "error", err)
					return
				}
			}
		}
	}()
}
