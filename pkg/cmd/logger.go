package cmd

import (
	"context"
	"log/slog"
	"os"
)

type ContextHandler struct {
	slog.Handler
}

func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if requestID, ok := ctx.Value("req_id").(string); ok {
		r.AddAttrs(slog.String("req_id", requestID))
	}
	if userID, ok := ctx.Value("chat_id").(string); ok {
		r.AddAttrs(slog.String("chat_id", userID))
	}

	return h.Handler.Handle(ctx, r)
}

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

	ctxHandler := &ContextHandler{Handler: logHandler}

	logger := slog.New(ctxHandler).With(
		slog.String("ver", arg.version),
		slog.String("app", "help-my-pet"),
	)

	slog.SetDefault(logger)

	return nil
}
