package api_test

import (
	"bytes"
	"testing"

	"github.com/gringolito/dnsmasq-manager/api"
	"github.com/gringolito/dnsmasq-manager/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slog"
)

func TestNewMiddlewareAuthDisabledWarning(t *testing.T) {
	tests := []struct {
		name       string
		authMethod string
		authKey    string
		expectWarn bool
	}{
		{
			name:       "NoAuth emits WARN",
			authMethod: config.NoAuth,
			authKey:    "",
			expectWarn: true,
		},
		{
			name:       "Configured auth emits no WARN",
			authMethod: config.AuthHS256,
			authKey:    "test-secret-key",
			expectWarn: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn})
			originalLogger := slog.Default()
			slog.SetDefault(slog.New(handler))
			t.Cleanup(func() { slog.SetDefault(originalLogger) })

			cfg := &config.Config{}
			cfg.Auth.Method = tt.authMethod
			cfg.Auth.Key = tt.authKey

			mw, err := api.NewMiddleware(nil, cfg)

			require.NoError(t, err)
			assert.NotNil(t, mw)

			logOutput := buf.String()
			if tt.expectWarn {
				assert.Contains(t, logOutput, "WARN")
				assert.Contains(t, logOutput, "Authentication is disabled")
			} else {
				assert.NotContains(t, logOutput, "WARN")
			}
		})
	}
}
