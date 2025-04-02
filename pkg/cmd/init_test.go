package cmd

import (
	"context"
	"log/slog"
	"os"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ksysoev/help-my-pet/pkg/bot"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
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
		errContains string
		textFormat  bool
		wantErr     bool
	}{
		{
			name:     "valid configuration",
			logLevel: "info",
			configData: `
bot:
  telegram_token: "test-token"
ai:
  model: "test-model"
  api_key: "test-key"
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
  model: "test-model"
  api_key: "test-key"
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

			cmd := BotCommand(args)
			require.NotNil(t, cmd)

			// For error cases, we don't need to run the bot service
			if tt.wantErr {
				err := cmd.RunE(cmd, []string{})
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			// For successful case, run the bot service with mocks
			updatesChan := make(chan tgbotapi.Update)
			close(updatesChan)

			// Create cancellable context
			ctx, cancel := context.WithCancel(context.Background())
			cmd.SetContext(ctx)

			// Create bot runner with mock service
			runner := NewBotRunner()
			mockService := bot.NewMockService(t)
			mockService.EXPECT().Run(ctx).Return(nil)
			runner.WithBotService(mockService)

			// Override the bot command's RunE to use our runner
			cmd.RunE = func(cmd *cobra.Command, _ []string) error {
				if err := initLogger(args); err != nil {
					return err
				}

				slog.Info("Starting Help My Pet bot", slog.String("version", args.version))

				cfg, err := initConfig(args)
				if err != nil {
					return err
				}

				// Cancel context after a short delay
				go func() {
					cancel()
				}()

				return runner.RunBot(cmd.Context(), cfg)
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
		})
	}
}

func TestBotCommand_ContextCancellation(t *testing.T) {
	// Create temporary config file
	configData := `
bot:
  telegram_token: "test-token"
ai:
  model: "test-model"
  api_key: "test-key"
`
	tmpfile, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)

	_, err = tmpfile.WriteString(configData)
	require.NoError(t, err)
	require.NoError(t, tmpfile.Close())

	args := &args{
		version:    "1.0.0",
		LogLevel:   "info",
		ConfigPath: tmpfile.Name(),
		TextFormat: false,
	}

	// Create mock bot API
	updatesChan := make(chan tgbotapi.Update)
	close(updatesChan)

	cmd := BotCommand(args)
	require.NotNil(t, cmd)

	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	cmd.SetContext(ctx)

	// Create bot runner with mock service
	runner := NewBotRunner()
	mockService := bot.NewMockService(t)
	mockService.EXPECT().Run(ctx).Return(nil)
	runner.WithBotService(mockService)

	// Override the bot command's RunE to use our runner
	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
		if err := initLogger(args); err != nil {
			return err
		}

		slog.Info("Starting Help My Pet bot", slog.String("version", args.version))

		cfg, err := initConfig(args)
		if err != nil {
			return err
		}

		// Cancel context after a short delay
		go func() {
			cancel()
		}()

		return runner.RunBot(cmd.Context(), cfg)
	}

	err = cmd.RunE(cmd, []string{})
	require.NoError(t, err)
}
