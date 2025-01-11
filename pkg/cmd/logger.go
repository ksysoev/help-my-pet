package cmd

import (
	"log/slog"
	"os"
)

// initLogger initializes the default logger for the application using slog.
func initLogger(arg *args) error {
	var logLevel slog.Level
	if err := logLevel.UnmarshalText([]byte(arg.LogLevel)); err != nil {
		return err
	}

	options := &slog.HandlerOptions{
		Level: logLevel,
	}

	var logHandler slog.Handler
	if arg.TextFormat {
		logHandler = slog.NewTextHandler(os.Stdout, options)
	} else {
		logHandler = slog.NewJSONHandler(os.Stdout, options)
	}

	logger := slog.New(logHandler).With(
		slog.String("ver", arg.version),
		slog.String("app", "help-my-pet"),
	)

	slog.SetDefault(logger)

	return nil
}
