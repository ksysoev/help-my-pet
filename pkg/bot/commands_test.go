package bot

import (
	"context"
	"fmt"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ksysoev/help-my-pet/pkg/core"
	"github.com/ksysoev/help-my-pet/pkg/i18n"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandleCommand(t *testing.T) {
	tests := []struct {
		name              string
		command           string
		languageCode      string
		chatID            int64
		userID            int64
		mockSetup         func(*MockAIProvider, *i18n.Config)
		expectedMsg       string
		expectedError     error
		expectedParseMode string
	}{
		{
			name:         "start command",
			command:      "/start",
			languageCode: "en",
			chatID:       123,
			userID:       456,
			mockSetup: func(m *MockAIProvider, msgs *i18n.Config) {
				msgs.Languages = map[string]i18n.Messages{
					"en": {
						Start: "Welcome message",
					},
				}
			},
			expectedMsg: "Welcome message",
		},
		{
			name:              "terms command",
			command:           "/terms",
			languageCode:      "en",
			chatID:            123,
			userID:            456,
			mockSetup:         func(m *MockAIProvider, msgs *i18n.Config) {},
			expectedMsg:       termsContent,
			expectedParseMode: "HTML",
		},
		{
			name:         "editprofile command success",
			command:      "/editprofile",
			languageCode: "en",
			chatID:       123,
			userID:       456,
			mockSetup: func(m *MockAIProvider, msgs *i18n.Config) {
				m.On("ProcessEditProfile", mock.Anything, mock.MatchedBy(func(req *core.UserMessage) bool {
					return req.UserID == "456" && req.ChatID == "123"
				})).Return(&core.Response{Message: "Profile updated"}, nil)
			},
			expectedMsg: "Profile updated",
		},
		{
			name:         "editprofile command error",
			command:      "/editprofile",
			languageCode: "en",
			chatID:       123,
			userID:       456,
			mockSetup: func(m *MockAIProvider, msgs *i18n.Config) {
				msgs.Languages = map[string]i18n.Messages{
					"en": {
						Error: "Error occurred",
					},
				}
				m.On("ProcessEditProfile", mock.Anything, mock.Anything).
					Return(nil, fmt.Errorf("processing error"))
			},
			expectedMsg: "Error occurred",
		},
		{
			name:         "unknown command",
			command:      "/unknown",
			languageCode: "en",
			chatID:       123,
			userID:       456,
			mockSetup:    func(m *MockAIProvider, msgs *i18n.Config) {},
			expectedMsg:  "Unknown command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockAI := NewMockAIProvider(t)
			messages := &i18n.Config{
				Languages: make(map[string]i18n.Messages),
			}

			// Configure mocks based on test case
			tt.mockSetup(mockAI, messages)

			// Create service instance
			svc := &ServiceImpl{
				AISvc:    mockAI,
				Messages: messages,
			}

			// Create message with proper command structure
			cmdText := tt.command + " " // Add space after command to ensure proper parsing
			msg := &tgbotapi.Message{
				Text: cmdText,
				From: &tgbotapi.User{
					ID:           tt.userID,
					LanguageCode: tt.languageCode,
				},
				Chat: &tgbotapi.Chat{
					ID: tt.chatID,
				},
				Entities: []tgbotapi.MessageEntity{
					{
						Type:   "bot_command",
						Offset: 0,
						Length: len(tt.command),
					},
				},
			}

			// Execute command
			resp, err := svc.HandleCommand(context.Background(), msg)

			// Verify results
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.chatID, resp.ChatID)
				assert.Equal(t, tt.expectedMsg, resp.Text)
				if tt.expectedParseMode != "" {
					assert.Equal(t, tt.expectedParseMode, resp.ParseMode)
				}
			}

			// Verify all mock expectations were met
			mockAI.AssertExpectations(t)
		})
	}
}
