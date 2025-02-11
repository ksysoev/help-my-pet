package middleware

import (
	"context"
	"errors"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
)

func TestWithErrorHandling(t *testing.T) {
	tests := []struct {
		expectedError error
		handler       Handler
		message       *tgbotapi.Message
		checkLang     func(t *testing.T, lang string)
		name          string
		expectedMsg   string
	}{
		{
			name: "handles error from handler",
			handler: HandlerFunc(func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
				return tgbotapi.MessageConfig{}, errors.New("handler error")
			}),
			message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: 123},
				From: &tgbotapi.User{LanguageCode: "en"},
			},
			expectedError: nil,
			expectedMsg:   "Sorry, I encountered an error while processing your request. Please try again later.",
		},
		{
			name: "passes through successful response",
			handler: HandlerFunc(func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
				return tgbotapi.NewMessage(123, "success"), nil
			}),
			message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: 123},
				From: &tgbotapi.User{LanguageCode: "en"},
			},
			expectedError: nil,
			expectedMsg:   "success",
		},
		{
			name: "handles message without From field",
			handler: HandlerFunc(func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
				return tgbotapi.MessageConfig{}, errors.New("handler error")
			}),
			message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: 123},
			},
			expectedError: nil,
			expectedMsg:   "Sorry, I encountered an error while processing your request. Please try again later.",
		},
		{
			name: "handles message with empty language code",
			handler: HandlerFunc(func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
				return tgbotapi.MessageConfig{}, errors.New("handler error")
			}),
			message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: 123},
				From: &tgbotapi.User{},
			},
			expectedError: nil,
			expectedMsg:   "Sorry, I encountered an error while processing your request. Please try again later.",
		},
		{
			name: "handles context cancellation",
			handler: HandlerFunc(func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
				return tgbotapi.MessageConfig{}, context.Canceled
			}),
			message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: 123},
				From: &tgbotapi.User{LanguageCode: "en"},
			},
			expectedError: nil,
			expectedMsg:   "Sorry, I encountered an error while processing your request. Please try again later.",
		},
		{
			name: "handles nil chat",
			handler: HandlerFunc(func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
				return tgbotapi.MessageConfig{}, errors.New("handler error")
			}),
			message:       &tgbotapi.Message{},
			expectedError: nil,
			expectedMsg:   "Sorry, I encountered an error while processing your request. Please try again later.",
		},
		{
			name: "handles nil message",
			handler: HandlerFunc(func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
				return tgbotapi.MessageConfig{}, errors.New("handler error")
			}),
			message:       nil,
			expectedError: errors.New("message is nil"),
			expectedMsg:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			handler := WithErrorHandling()(tt.handler)
			msgConfig, err := handler.Handle(context.Background(), tt.message)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedMsg, msgConfig.Text)
			}
		})
	}
}
