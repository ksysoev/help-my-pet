package middleware

import (
	"context"
	"errors"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ksysoev/help-my-pet/pkg/i18n"
	"github.com/stretchr/testify/assert"
)

func TestWithErrorHandling(t *testing.T) {
	tests := []struct {
		expectedError error
		handler       Handler
		getMessage    func(lang string, msgType i18n.Message) string
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
			getMessage: func(lang string, msgType i18n.Message) string {
				return "error message"
			},
			message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: 123},
				From: &tgbotapi.User{LanguageCode: "en"},
			},
			expectedError: nil,
			expectedMsg:   "error message",
			checkLang: func(t *testing.T, lang string) {
				assert.Equal(t, "en", lang)
			},
		},
		{
			name: "passes through successful response",
			handler: HandlerFunc(func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
				return tgbotapi.NewMessage(123, "success"), nil
			}),
			getMessage: func(lang string, msgType i18n.Message) string {
				return "error message"
			},
			message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: 123},
				From: &tgbotapi.User{LanguageCode: "en"},
			},
			expectedError: nil,
			expectedMsg:   "success",
			checkLang:     func(t *testing.T, lang string) {},
		},
		{
			name: "handles message without From field",
			handler: HandlerFunc(func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
				return tgbotapi.MessageConfig{}, errors.New("handler error")
			}),
			getMessage: func(lang string, msgType i18n.Message) string {
				return "error message"
			},
			message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: 123},
			},
			expectedError: nil,
			expectedMsg:   "error message",
			checkLang: func(t *testing.T, lang string) {
				assert.Equal(t, "", lang)
			},
		},
		{
			name: "handles message with empty language code",
			handler: HandlerFunc(func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
				return tgbotapi.MessageConfig{}, errors.New("handler error")
			}),
			getMessage: func(lang string, msgType i18n.Message) string {
				return "error message"
			},
			message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: 123},
				From: &tgbotapi.User{},
			},
			expectedError: nil,
			expectedMsg:   "error message",
			checkLang: func(t *testing.T, lang string) {
				assert.Equal(t, "", lang)
			},
		},
		{
			name: "handles context cancellation",
			handler: HandlerFunc(func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
				return tgbotapi.MessageConfig{}, context.Canceled
			}),
			getMessage: func(lang string, msgType i18n.Message) string {
				return "error message"
			},
			message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: 123},
				From: &tgbotapi.User{LanguageCode: "en"},
			},
			expectedError: nil,
			expectedMsg:   "error message",
			checkLang: func(t *testing.T, lang string) {
				assert.Equal(t, "en", lang)
			},
		},
		{
			name: "handles nil chat",
			handler: HandlerFunc(func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
				return tgbotapi.MessageConfig{}, errors.New("handler error")
			}),
			getMessage: func(lang string, msgType i18n.Message) string {
				return "error message"
			},
			message:       &tgbotapi.Message{},
			expectedError: nil,
			expectedMsg:   "error message",
			checkLang: func(t *testing.T, lang string) {
				assert.Equal(t, "", lang)
			},
		},
		{
			name: "handles nil message",
			handler: HandlerFunc(func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
				return tgbotapi.MessageConfig{}, errors.New("handler error")
			}),
			getMessage: func(lang string, msgType i18n.Message) string {
				return "error message"
			},
			message:       nil,
			expectedError: errors.New("message is nil"),
			expectedMsg:   "",
			checkLang:     func(t *testing.T, lang string) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedLang string
			wrappedGetMessage := func(lang string, msgType i18n.Message) string {
				capturedLang = lang
				return tt.getMessage(lang, msgType)
			}

			handler := WithErrorHandling(wrappedGetMessage)(tt.handler)
			msgConfig, err := handler.Handle(context.Background(), tt.message)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedMsg, msgConfig.Text)
				tt.checkLang(t, capturedLang)
			}
		})
	}
}
