package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitConfig(t *testing.T) {
	tests := []struct {
		envVars     map[string]string
		name        string
		configFile  string
		configData  string
		errContains string
		wantErr     bool
	}{
		{
			name: "valid config file",
			configData: `
bot:
  telegram_token: "test-token"
ai:
  model: "test-model"
  api_key: "test-key"
`,
			wantErr: false,
		},
		{
			name: "missing telegram token",
			configData: `
ai:
  model: "test-model"
  api_key: "test-key"
`,
			wantErr:     true,
			errContains: "telegram token is required",
		},
		{
			name: "missing anthropic key",
			configData: `
bot:
  telegram_token: "test-token"
ai:
  model: "test-model"
`,
			wantErr:     true,
			errContains: "anthropic API key is required",
		},
		{
			name: "env vars override",
			configData: `
bot:
  telegram_token: "test-token"
ai:
  model: "test-model"
  api_key: "test-key"
`,
			envVars: map[string]string{
				"BOT_TELEGRAM_TOKEN": "env-token",
				"AI_API_KEY":         "env-key",
				"AI_MODEL":           "env-model",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp config file if config data is provided
			var configPath string
			if tt.configData != "" {
				tmpfile, err := os.CreateTemp("", "config-*.yaml")
				require.NoError(t, err)

				_, err = tmpfile.WriteString(tt.configData)
				require.NoError(t, err)
				require.NoError(t, tmpfile.Close())
				configPath = tmpfile.Name()
			}

			// Set environment variables
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			args := &args{
				ConfigPath: configPath,
			}

			cfg, err := initConfig(args)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, cfg)

			// Check if env vars were properly applied
			if len(tt.envVars) > 0 {
				assert.Equal(t, tt.envVars["BOT_TELEGRAM_TOKEN"], cfg.Bot.TelegramToken)
				assert.Equal(t, tt.envVars["AI_API_KEY"], cfg.AI.APIKey)
				assert.Equal(t, tt.envVars["AI_MODEL"], cfg.AI.Model)
			}
		})
	}
}
