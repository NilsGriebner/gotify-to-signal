package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/NilsGriebner/gotify-to-signal/config"
	"github.com/NilsGriebner/gotify-to-signal/messages"
	"github.com/caarlos0/env/v11"
	"github.com/gotify/plugin-api"
)

// GetGotifyPluginInfo returns gotify plugin info.
func GetGotifyPluginInfo() plugin.Info {
	return plugin.Info{
		ModulePath:  "https://github.com/NilsGriebner/gotify-to-signal",
		Version:     "0.0.1",
		Author:      "Nils Griebner",
		Website:     "https://github.com/NilsGriebner/gotify-to-signal",
		Description: "This plugin forwards received messages to Signal",
		License:     "MIT",
		Name:        "Signal message forwarder",
	}
}

// SignalForwarder is the gotify plugin instance.
type SignalForwarder struct {
	logger             *slog.Logger
	MessageHandler     plugin.MessageHandler
	config             *config.Config
	listenerCancelFunc context.CancelFunc
}

// Enable enables the plugin.
func (s *SignalForwarder) Enable() error {
	websocketURL := fmt.Sprintf("%s/stream?token=%s", s.config.Gotify.Host, s.config.Gotify.ClientToken)

	listenerCtx, listenerCancel := context.WithCancel(context.Background())
	s.listenerCancelFunc = listenerCancel

	signalClient := messages.NewSignalClient(
		s.config.Signal.FromNumber, s.config.Signal.ToNumber, s.config.Signal.APIHost, s.logger)

	listener := messages.NewWebSocketListener(s.logger, signalClient, websocketURL)

	s.logger.Info("starting websocket listener")
	go listener.Listen(listenerCtx)

	return nil
}

// Disable disables the plugin.
func (s *SignalForwarder) Disable() error {
	s.listenerCancelFunc()
	return nil
}

func (s *SignalForwarder) DefaultConfig() interface{} {
	cfg := config.NewConfig()
	err := env.Parse(cfg)
	if err != nil {
		// Since we don't require any envs, this is just informational
		s.logger.Warn("error parsing env vars, falling back to default", "error", err)
	}

	cfg.SetDefaultsForMissingValues()

	return cfg
}

func (s *SignalForwarder) ValidateAndSetConfig(c interface{}) error {
	cfg, ok := c.(*config.Config)
	if !ok {
		s.logger.Warn("invalid config type, falling back to default")
		return fmt.Errorf("invalid config type: %T", c)
	}

	err := cfg.Validate()
	if err != nil {
		s.logger.Warn("failed to validate config", "error", err)
		return err
	}

	s.config = cfg

	return nil
}

// SetMessageHandler implements plugin.Messenger
// Invoked during initialization.
func (s *SignalForwarder) SetMessageHandler(h plugin.MessageHandler) {
	s.MessageHandler = h
}

// NewGotifyPluginInstance creates a plugin instance for a user context.
func NewGotifyPluginInstance(_ plugin.UserContext) plugin.Plugin {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).
		With(slog.String("plugin", "gotify-to-signal"))
	return &SignalForwarder{
		logger: logger,
	}
}

func main() {
	panic("this should be built as go plugin")
}
