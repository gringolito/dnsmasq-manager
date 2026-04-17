package config_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/gringolito/dnsmasq-manager/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slog"
)

const testConfigName = "dmm-test-config"

func TestInitConfigFileNotFoundWarning(t *testing.T) {
	tests := []struct {
		name            string
		setupConfigFile bool
		expectWarn      bool
	}{
		{
			name:            "Config file not found emits WARN",
			setupConfigFile: false,
			expectWarn:      true,
		},
		{
			name:            "Config file found emits no WARN",
			setupConfigFile: true,
			expectWarn:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn})
			originalLogger := slog.Default()
			slog.SetDefault(slog.New(handler))
			t.Cleanup(func() { slog.SetDefault(originalLogger) })

			if tt.setupConfigFile {
				err := os.WriteFile(testConfigName+".yaml", []byte("# minimal test config\n"), 0644)
				require.NoError(t, err, "Failed to create test config file")
				t.Cleanup(func() { os.Remove(testConfigName + ".yaml") })
			}

			cfg, err := config.Init(testConfigName)
			require.NoError(t, err)
			assert.NotNil(t, cfg)

			logOutput := buf.String()
			if tt.expectWarn {
				assert.Contains(t, logOutput, "WARN")
				assert.Contains(t, logOutput, "No configuration file found, running with defaults")
			} else {
				assert.NotContains(t, logOutput, "WARN")
			}
		})
	}
}
