package cmd

import (
	"context"
	"os"
	"testing"

	"github.com/ksysoev/help-my-pet/pkg/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestInitCommands(t *testing.T) {
	version := "1.0.0"
	rootCmd := InitCommands(version)

	// Test root command
	assert.Equal(t, "help-my-pet", rootCmd.Use)
	assert.NotEmpty(t, rootCmd.Short)
	assert.NotEmpty(t, rootCmd.Long)

	// Test persistent flags
	flags := rootCmd.PersistentFlags()
	configFlag := flags.Lookup("config")
	require.NotNil(t, configFlag)
	assert.Equal(t, "", configFlag.DefValue)

	logLevelFlag := flags.Lookup("loglevel")
	require.NotNil(t, logLevelFlag)
	assert.Equal(t, "info", logLevelFlag.DefValue)

	logTextFlag := flags.Lookup("logtext")
	require.NotNil(t, logTextFlag)
	assert.Equal(t, "false", logTextFlag.DefValue)

	// Test bot subcommand
	botCmd, _, err := rootCmd.Find([]string{"bot"})
	require.NoError(t, err)
	assert.Equal(t, "bot", botCmd.Use)
	assert.NotEmpty(t, botCmd.Short)
	assert.NotEmpty(t, botCmd.Long)
}

func TestBotCommand(t *testing.T) {
	tests := []struct {
		name        string
		logLevel    string
		configPath  string
		configData  string
		textFormat  bool
		wantErr     bool
		errContains string
	}{
		{
			name:     "valid configuration",
			logLevel: "info",
			configData: `
bot:
  telegram_token: "test-token"
ai:
  anthropic_key: "test-key"
  model: "test-model"
`,
			textFormat: false,
			wantErr:    false,
		},
		{
			name:     "invalid log level",
			logLevel: "invalid",
			configData: `
bot:
  telegram_token: "test-token"
ai:
  anthropic_key: "test-key"
  model: "test-model"
`,
			textFormat:  false,
			wantErr:     true,
			errContains: "slog: level string \"invalid\": unknown name",
		},
		{
			name:        "invalid config path",
			logLevel:    "info",
			configPath:  "nonexistent.yaml",
			textFormat:  false,
			wantErr:     true,
			errContains: "failed to read config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var configPath string
			if tt.configData != "" {
				tmpfile, err := os.CreateTemp("", "config-*.yaml")
				require.NoError(t, err)
				defer os.Remove(tmpfile.Name())

				_, err = tmpfile.WriteString(tt.configData)
				require.NoError(t, err)
				require.NoError(t, tmpfile.Close())
				configPath = tmpfile.Name()
			} else {
				configPath = tt.configPath
			}

			args := &args{
				version:    "1.0.0",
				LogLevel:   tt.logLevel,
				ConfigPath: configPath,
				TextFormat: tt.textFormat,
			}

			// Create mock bot service
			mockBot := NewMockBotService()
			mockBot.On("Run", mock.Anything).Return(nil)

			// Create mock factory
			mockFactory := func(token string, aiSvc *core.AIService) BotService {
				return mockBot
			}

			cmd := BotCommand(args)
			require.NotNil(t, cmd)

			// Override runBot function
			oldRunBot := runBot
			defer func() { runBot = oldRunBot }()
			runBot = func(ctx context.Context, cfg *Config, factory BotServiceFactory) error {
				return oldRunBot(ctx, cfg, mockFactory)
			}

			err := cmd.RunE(cmd, []string{})
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			mockBot.AssertExpectations(t)
		})
	}
}

func TestBotCommand_ContextCancellation(t *testing.T) {
	// Create temporary config file
	configData := `
bot:
  telegram_token: "test-token"
ai:
  anthropic_key: "test-key"
  model: "test-model"
`
	tmpfile, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.WriteString(configData)
	require.NoError(t, err)
	require.NoError(t, tmpfile.Close())

	args := &args{
		version:    "1.0.0",
		LogLevel:   "info",
		ConfigPath: tmpfile.Name(),
		TextFormat: false,
	}

	// Create mock bot service
	mockBot := NewMockBotService()
	mockBot.On("Run", mock.Anything).Return(nil)

	// Create mock factory
	mockFactory := func(token string, aiSvc *core.AIService) BotService {
		return mockBot
	}

	cmd := BotCommand(args)
	require.NotNil(t, cmd)

	// Override runBot function
	oldRunBot := runBot
	defer func() { runBot = oldRunBot }()
	runBot = func(ctx context.Context, cfg *Config, factory BotServiceFactory) error {
		return oldRunBot(ctx, cfg, mockFactory)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cmd.SetContext(ctx)

	// Cancel the context immediately
	cancel()

	err = cmd.RunE(cmd, []string{})
	require.NoError(t, err)
	mockBot.AssertExpectations(t)
}