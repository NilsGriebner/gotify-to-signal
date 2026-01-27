package config

import (
	"fmt"

	"github.com/nyaruka/phonenumbers"
)

func NewConfig() *Config {
	return &Config{}
}

type SignalConfig struct {
	FromNumber string `yaml:"fromNumber" env:"SIGNAL_FROM_NUMBER"`
	ToNumber   string `yaml:"toNumber" env:"SIGNAL_TO_NUMBER"`
	APIHost    string `yaml:"apiHost" env:"SIGNAL_API_HOST"`
}

type GotifyConfig struct {
	Host        string `yaml:"host" env:"GOTIFY_HOST"`
	ClientToken string `yaml:"clientToken" env:"SIGNAL_CLIENT_TOKEN"`
}

type Config struct {
	Signal SignalConfig `yaml:"signal"`
	Gotify GotifyConfig `yaml:"gotify"`
}

func (c *Config) SetDefaultsForMissingValues() {
	if c.Signal.FromNumber == "" {
		c.Signal.FromNumber = "<your phone number here>"
	}

	if c.Signal.ToNumber == "" {
		c.Signal.ToNumber = "<your phone number here>"
	}

	if c.Signal.APIHost == "" {
		c.Signal.APIHost = "http://localhost:8089"
	}

	if c.Gotify.Host == "" {
		c.Gotify.Host = "ws://localhost:8080"
	}

	if c.Gotify.ClientToken == "" {
		c.Gotify.ClientToken = "<your client token here>"
	}
}

func (c *Config) Validate() error {
	err := validatePhoneNumbers(c)
	if err != nil {
		return err
	}

	return nil
}

func validatePhoneNumbers(c *Config) error {
	numbersToParse := []string{c.Signal.FromNumber, c.Signal.ToNumber}
	for _, numberString := range numbersToParse {
		_, err := phonenumbers.Parse(numberString, "ZZ")
		if err != nil {
			return fmt.Errorf("invalid phone number %s: %w", numberString, err)
		}
	}

	return nil
}
