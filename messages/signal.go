package messages

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

type SignalSendRequest struct {
	Message  string   `json:"message"`
	Number   string   `json:"number"`
	Receipts []string `json:"recipients"`
}

func NewSignalClient(fromNumber string, toNumber string, apiHost string, logger *slog.Logger) *SignalClient {
	return &SignalClient{
		logger:     logger,
		FromNumber: fromNumber,
		ToNumber:   toNumber,
		APIHost:    apiHost,
	}
}

type SignalClient struct {
	logger     *slog.Logger
	FromNumber string
	ToNumber   string
	APIHost    string
}

func (c *SignalClient) Send(message *GotifyMessage) error {
	reqBody := SignalSendRequest{
		Number:   c.FromNumber,
		Receipts: []string{c.ToNumber},
		Message:  fmt.Sprintf("%s\n%s", message.Title, message.Message),
	}

	reqBodyJSON, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	reqURL := fmt.Sprintf("%s/v2/send", c.APIHost)
	req, err := http.NewRequestWithContext(context.TODO(), http.MethodPost, reqURL, bytes.NewReader(reqBodyJSON))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		err = resp.Body.Close()
		if err != nil {
			c.logger.Error("unable to close response body", "error", err)
		}
	}()

	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		c.logger.Info("successfully forwarded message to signal")
		return nil
	}
	return fmt.Errorf("signal api returned code: %d", resp.StatusCode)
}
