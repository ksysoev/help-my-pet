package bot

import (
	"context"
	"fmt"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ksysoev/help-my-pet/pkg/core/message"
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
		mockSetup         func(*MockAIProvider)
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
			mockSetup: func(m *MockAIProvider) {
				m.On("ResetUserConversation", mock.Anything, "456", "123").Return(nil)
			},
			expectedMsg: "Welcome to Help My Pet Bot",
		},
		{
			name:         "start command with error",
			command:      "/start",
			languageCode: "en",
			chatID:       123,
			userID:       456,
			mockSetup: func(m *MockAIProvider) {
				m.On("ResetUserConversation", mock.Anything, "456", "123").Return(assert.AnError)
			},
			expectedMsg: "Welcome to Help My Pet Bot",
		},
		{
			name:              "terms command",
			command:           "/terms",
			languageCode:      "en",
			chatID:            123,
			userID:            456,
			mockSetup:         func(m *MockAIProvider) {},
			expectedMsg:       "Terms and Conditions",
			expectedParseMode: "HTML",
		},
		{
			name:         "editprofile command success",
			command:      "/editprofile",
			languageCode: "en",
			chatID:       123,
			userID:       456,
			mockSetup: func(m *MockAIProvider) {
				m.On("ProcessEditProfile", mock.Anything, mock.MatchedBy(func(req *message.UserMessage) bool {
					return req.UserID == "456" && req.ChatID == "123"
				})).Return(&message.Response{Message: "Profile updated"}, nil)
			},
			expectedMsg: "Profile updated",
		},
		{
			name:         "editprofile command error",
			command:      "/editprofile",
			languageCode: "en",
			chatID:       123,
			userID:       456,
			mockSetup: func(m *MockAIProvider) {
				m.On("ProcessEditProfile", mock.Anything, mock.Anything).
					Return(nil, fmt.Errorf("processing error"))
			},
			expectedError: fmt.Errorf("failed to process edit profile request: processing error"),
		},
		{
			name:         "unknown command",
			command:      "/unknown",
			languageCode: "en",
			chatID:       123,
			userID:       456,
			mockSetup:    func(m *MockAIProvider) {},
			expectedMsg:  "Unknown command",
		},
		{
			name:         "cancel command success",
			command:      "/cancel",
			languageCode: "en",
			chatID:       123,
			userID:       456,
			mockSetup: func(m *MockAIProvider) {
				m.EXPECT().CancelQuestionnaire(mock.Anything, "123").Return(nil)
			},
			expectedMsg: "Questionary is cancelled",
		},
		{
			name:         "cancel command error",
			command:      "/cancel",
			languageCode: "en",
			chatID:       123,
			userID:       456,
			mockSetup: func(m *MockAIProvider) {
				m.EXPECT().CancelQuestionnaire(mock.Anything, "123").Return(assert.AnError)
			},
			expectedError: fmt.Errorf("failed to reset conversation: %w", assert.AnError),
		},
		{
			name:         "help command",
			command:      "/help",
			languageCode: "en",
			chatID:       123,
			userID:       456,
			mockSetup:    func(m *MockAIProvider) {},
			expectedMsg:  "Help My Pet Bot Commands",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockAI := NewMockAIProvider(t)

			// Configure mocks based on test case
			tt.mockSetup(mockAI)

			// Create service instance
			svc := &ServiceImpl{
				AISvc: mockAI,
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
				assert.Contains(t, resp.Text, tt.expectedMsg)
				if tt.expectedParseMode != "" {
					assert.Equal(t, tt.expectedParseMode, resp.ParseMode)
				}
			}

			// Verify all mock expectations were met
			mockAI.AssertExpectations(t)
		})
	}
}
