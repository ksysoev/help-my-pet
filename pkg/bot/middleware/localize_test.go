package middleware

import (
	"context"
	"errors"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockHandler struct {
	mock.Mock
}

func (m *MockHandler) Handle(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	args := m.Called(ctx, msg)
	return args.Get(0).(tgbotapi.MessageConfig), args.Error(1)
}

func TestWithLocalization(t *testing.T) {
	tests := []struct {
		name            string
		message         *tgbotapi.Message
		expectedMessage tgbotapi.MessageConfig
		expectedError   error
	}{
		{
			name: "message with language code",
			message: &tgbotapi.Message{
				From: &tgbotapi.User{LanguageCode: "en"},
			},
			expectedMessage: tgbotapi.MessageConfig{Text: "test reply"},
			expectedError:   nil,
		},
		{
			name: "message without language code",
			message: &tgbotapi.Message{
				From: nil,
			},
			expectedMessage: tgbotapi.MessageConfig{Text: "default reply"},
			expectedError:   nil,
		},
		{
			name: "handler returns an error",
			message: &tgbotapi.Message{
				From: &tgbotapi.User{LanguageCode: "fr"},
			},
			expectedMessage: tgbotapi.MessageConfig{},
			expectedError:   errors.New("handler error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockHandler := new(MockHandler)
			localizedHandler := WithLocalization()(mockHandler)

			mockHandler.On("Handle", mock.Anything, tt.message).Return(tt.expectedMessage, tt.expectedError)

			// Act
			msgConfig, err := localizedHandler.Handle(context.Background(), tt.message)

			// Assert
			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedMessage, msgConfig)
			}

			mockHandler.AssertExpectations(t)
		})
	}
}
