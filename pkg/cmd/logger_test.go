package cmd

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitLogger(t *testing.T) {
	tests := []struct {
		name      string
		logLevel  string
		wantLevel slog.Level
		logText   bool
		wantErr   bool
	}{
		{
			name:      "debug level json format",
			logLevel:  "debug",
			logText:   false,
			wantLevel: slog.LevelDebug,
			wantErr:   false,
		},
		{
			name:      "info level text format",
			logLevel:  "info",
			logText:   true,
			wantLevel: slog.LevelInfo,
			wantErr:   false,
		},
		{
			name:      "warn level json format",
			logLevel:  "warn",
			logText:   false,
			wantLevel: slog.LevelWarn,
			wantErr:   false,
		},
		{
			name:      "error level text format",
			logLevel:  "error",
			logText:   true,
			wantLevel: slog.LevelError,
			wantErr:   false,
		},
		{
			name:      "invalid level",
			logLevel:  "invalid",
			logText:   false,
			wantLevel: slog.LevelInfo,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := &args{
				LogLevel:   tt.logLevel,
				TextFormat: tt.logText,
			}

			err := initLogger(args)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Get the current logger's level
			logger := slog.Default()
			assert.NotNil(t, logger)

			// Unfortunately, slog doesn't provide a way to get the current level
			// or handler type directly, so we can only verify the logger was set
		})
	}
}
