package cmd

import (
	"log/slog"

	"github.com/spf13/cobra"
)

type args struct {
	version    string
	LogLevel   string
	ConfigPath string
	TextFormat bool
}

// InitCommands initializes and returns the root command for the application.
func InitCommands(version string) *cobra.Command {
	args := &args{
		version: version,
	}

	cmd := &cobra.Command{
		Use:   "help-my-pet",
		Short: "AI-powered Telegram bot for pet health assistance",
		Long:  "A Telegram bot that uses Anthropic AI to help pet owners with health-related questions about their pets",
	}

	cmd.AddCommand(BotCommand(args))

	cmd.PersistentFlags().StringVar(&args.ConfigPath, "config", "", "config file path")
	cmd.PersistentFlags().StringVar(&args.LogLevel, "loglevel", "info", "log level (debug, info, warn, error)")
	cmd.PersistentFlags().BoolVar(&args.TextFormat, "logtext", false, "log in text format, otherwise JSON")

	return cmd
}

// BotCommand creates a new cobra.Command to start the Telegram bot server.
func BotCommand(arg *args) *cobra.Command {
	return &cobra.Command{
		Use:   "bot",
		Short: "Start the Telegram bot server",
		Long:  "Start the AI-powered Telegram bot server for pet health assistance",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := initLogger(arg); err != nil {
				return err
			}

			slog.Info("Starting Help My Pet bot", slog.String("version", arg.version))

			cfg, err := initConfig(arg)
			if err != nil {
				return err
			}

			return runBot(cmd.Context(), cfg, nil)
		},
	}
}
