package bot

import (
	"context"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ksysoev/help-my-pet/pkg/conversation"
	"github.com/ksysoev/help-my-pet/pkg/core"
	"github.com/ksysoev/help-my-pet/pkg/i18n"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewService(t *testing.T) {
	tests := []struct {
		cfg     *Config
		aiSvc   AIProvider
		name    string
		wantErr bool
	}{
		{
			name:    "nil config",
			cfg:     nil,
			aiSvc:   NewMockAIProvider(t),
			wantErr: true,
		},
		{
			name: "empty token",
			cfg: &Config{
				TelegramToken: "",
				Messages:      &i18n.Config{},
			},
			aiSvc:   NewMockAIProvider(t),
			wantErr: true,
		},
		{
			name: "nil AIProvider",
			cfg: &Config{
				TelegramToken: "test-token",
				Messages:      &i18n.Config{},
			},
			aiSvc:   nil,
			wantErr: true,
		},
		{
			name: "nil messages",
			cfg: &Config{
				TelegramToken: "test-token",
				Messages:      nil,
			},
			aiSvc:   NewMockAIProvider(t),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewService(tt.cfg, tt.aiSvc)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestServiceImpl_ProcessMessage(t *testing.T) {
	tests := []struct {
		setupMocks  func(*MockBotAPI, *MockAIProvider)
		message     *tgbotapi.Message
		name        string
		expectError bool
	}{
		{
			name: "successful message processing",
			message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: 123},
				From: &tgbotapi.User{
					ID:           456,
					LanguageCode: "en",
				},
				Text: "test message",
			},
			setupMocks: func(mockBot *MockBotAPI, mockAI *MockAIProvider) {
				mockBot.EXPECT().Request(mock.Anything).Return(&tgbotapi.APIResponse{}, nil)
				mockAI.EXPECT().GetPetAdvice(mock.Anything, &core.PetAdviceRequest{
					ChatID:  "123",
					UserID:  "456",
					Message: "test message",
				}).Return(&core.PetAdviceResponse{
					Message: "AI response",
				}, nil)
				mockBot.EXPECT().Send(mock.MatchedBy(func(msg tgbotapi.MessageConfig) bool {
					return msg.Text == "AI response"
				})).Return(tgbotapi.Message{}, nil)
			},
			expectError: false,
		},
		{
			name: "empty message",
			message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: 123},
				Text: "",
			},
			setupMocks: func(mockBot *MockBotAPI, mockAI *MockAIProvider) {
				mockBot.EXPECT().Request(mock.Anything).Return(&tgbotapi.APIResponse{}, nil)
			},
			expectError: false,
		},
		{
			name: "start command",
			message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: 123},
				From: &tgbotapi.User{
					ID:           456,
					LanguageCode: "en",
				},
				Text: "/start",
			},
			setupMocks: func(mockBot *MockBotAPI, mockAI *MockAIProvider) {
				mockBot.EXPECT().Request(mock.Anything).Return(&tgbotapi.APIResponse{}, nil)
				mockBot.EXPECT().Send(mock.MatchedBy(func(msg tgbotapi.MessageConfig) bool {
					return msg.ChatID == 123
				})).Return(tgbotapi.Message{}, nil)
			},
			expectError: false,
		},
		{
			name: "nil From field",
			message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: 123},
				Text: "test message",
			},
			setupMocks: func(mockBot *MockBotAPI, mockAI *MockAIProvider) {
				mockBot.EXPECT().Request(mock.Anything).Return(&tgbotapi.APIResponse{}, nil)
				// Expect error message to be sent
				mockBot.EXPECT().Send(mock.MatchedBy(func(msg tgbotapi.MessageConfig) bool {
					return msg.ChatID == 123
				})).Return(tgbotapi.Message{}, nil)
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBot := NewMockBotAPI(t)
			mockAI := NewMockAIProvider(t)

			service := &ServiceImpl{
				Bot:      mockBot,
				AISvc:    mockAI,
				Messages: &i18n.Config{},
				reqMgr:   conversation.NewRequestManager(),
			}

			tt.setupMocks(mockBot, mockAI)

			ctx := context.Background()
			service.processMessage(ctx, tt.message)

			mockBot.AssertExpectations(t)
			mockAI.AssertExpectations(t)
		})
	}
}

func TestServiceImpl_Run(t *testing.T) {
	mockBot := NewMockBotAPI(t)
	mockAI := NewMockAIProvider(t)

	service := &ServiceImpl{
		Bot:      mockBot,
		AISvc:    mockAI,
		Messages: &i18n.Config{},
		reqMgr:   conversation.NewRequestManager(),
	}

	updates := make(chan tgbotapi.Update)
	mockBot.EXPECT().GetUpdatesChan(mock.Anything).Return((<-chan tgbotapi.Update)(updates))
	mockBot.EXPECT().StopReceivingUpdates().Return()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	go func() {
		err := service.Run(ctx)
		assert.NoError(t, err)
		close(done)
	}()

	// Wait for either completion or timeout
	select {
	case <-done:
		// Test completed successfully
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Test timed out")
	}

	mockBot.AssertExpectations(t)
}
