package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidatePositive(t *testing.T) {
	config := Config{
		Signal: SignalConfig{
			FromNumber: "+49123456789",
			ToNumber:   "+49987654321",
		},
	}
	require.NoError(t, config.Validate())
}

func TestValidateNoCountryCode(t *testing.T) {
	config := Config{
		Signal: SignalConfig{
			FromNumber: "0123456789",
			ToNumber:   "0987654321",
		},
	}
	require.Error(t, config.Validate())
}

func TestSetDefaultsForMissingValues(t *testing.T) {
	config := Config{
		Signal: SignalConfig{},
		Gotify: GotifyConfig{},
	}

	config.SetDefaultsForMissingValues()

	require.Equal(t, "<your phone number here>", config.Signal.FromNumber)
	require.Equal(t, "<your phone number here>", config.Signal.FromNumber)
	require.Equal(t, "http://localhost:8089", config.Signal.APIHost)
	require.Equal(t, "<your client token here>", config.Gotify.ClientToken)
	require.Equal(t, "ws://localhost:8080", config.Gotify.Host)
}
