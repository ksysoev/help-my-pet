package middleware

import (
	"context"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
)

func TestWithMetrics(t *testing.T) {
	tests := []struct {
		handler        Handler
		message        *tgbotapi.Message
		name           string
		expectedResult tgbotapi.MessageConfig
		expectedError  bool
	}{
		{
			name: "successful handler execution",
			handler: HandlerFunc(func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
				return tgbotapi.MessageConfig{BaseChat: tgbotapi.BaseChat{ChatID: msg.Chat.ID}}, nil
			}),
			message:       &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: 12345}},
			expectedError: false,
		},
		{
			name: "handler execution with error",
			handler: HandlerFunc(func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
				return tgbotapi.MessageConfig{}, assert.AnError
			}),
			message:       &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: 12345}},
			expectedError: true,
		},
		{
			name: "nil message",
			handler: HandlerFunc(func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
				if msg == nil {
					return tgbotapi.MessageConfig{}, assert.AnError
				}
				return tgbotapi.MessageConfig{}, nil
			}),
			message:       nil,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := WithMetrics()
			wrappedHandler := middleware(tt.handler)

			start := time.Now()
			_, err := wrappedHandler.Handle(context.Background(), tt.message)
			duration := time.Since(start)

			if (err != nil) != tt.expectedError {
				t.Errorf("expected error: %v, got: %v", tt.expectedError, err)
			}

			if duration <= 0 {
				t.Error("expected positive duration, got:", duration)
			}
		})
	}
}
