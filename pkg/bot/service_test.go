package bot

import (
	"context"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ksysoev/help-my-pet/pkg/core/message"
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
			},
			aiSvc:   NewMockAIProvider(t),
			wantErr: true,
		},
		{
			name: "nil AIProvider",
			cfg: &Config{
				TelegramToken: "test-token",
			},
			aiSvc:   nil,
			wantErr: true,
		},
		{
			name: "nil messages",
			cfg: &Config{
				TelegramToken: "test-token",
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
		ctx         context.Context
		setupMocks  func(*MockBotAPI, *MockAIProvider)
		update      *tgbotapi.Update
		name        string
		expectError bool
	}{
		{
			name: "successful update processing",
			ctx:  context.Background(),
			update: &tgbotapi.Update{
				Message: &tgbotapi.Message{
					Chat: &tgbotapi.Chat{ID: 123},
					From: &tgbotapi.User{
						ID:           456,
						LanguageCode: "en",
					},
					Text: "test update",
				},
			},
			setupMocks: func(mockBot *MockBotAPI, mockAI *MockAIProvider) {
				mockBot.EXPECT().Request(mock.Anything).Return(&tgbotapi.APIResponse{}, nil)
				mockAI.EXPECT().ProcessMessage(mock.Anything, &message.UserMessage{
					ChatID: "123",
					UserID: "456",
					Text:   "test update",
				}).Return(&message.Response{
					Message: "AI response",
				}, nil)
				mockBot.EXPECT().Send(mock.MatchedBy(func(msg tgbotapi.MessageConfig) bool {
					return msg.Text == "AI response"
				})).Return(tgbotapi.Message{}, nil)
			},
			expectError: false,
		},
		{
			name: "failed typing action",
			ctx:  context.Background(),
			update: &tgbotapi.Update{
				Message: &tgbotapi.Message{
					Chat: &tgbotapi.Chat{ID: 123},
					From: &tgbotapi.User{
						ID:           456,
						LanguageCode: "en",
					},
					Text: "test update",
				},
			},
			setupMocks: func(mockBot *MockBotAPI, mockAI *MockAIProvider) {
				mockBot.EXPECT().Request(mock.Anything).Return(nil, assert.AnError)
				mockAI.EXPECT().ProcessMessage(mock.Anything, &message.UserMessage{
					ChatID: "123",
					UserID: "456",
					Text:   "test update",
				}).Return(&message.Response{
					Message: "AI response",
				}, nil)
				mockBot.EXPECT().Send(mock.MatchedBy(func(msg tgbotapi.MessageConfig) bool {
					return msg.Text == "AI response"
				})).Return(tgbotapi.Message{}, nil)
			},
			expectError: false,
		},
		{
			name: "context cancelled",
			ctx:  context.Background(),
			update: &tgbotapi.Update{
				Message: &tgbotapi.Message{
					Chat: &tgbotapi.Chat{ID: 123},
					From: &tgbotapi.User{
						ID:           456,
						LanguageCode: "en",
					},
					Text: "test update",
				},
			},
			setupMocks: func(mockBot *MockBotAPI, mockAI *MockAIProvider) {
				mockBot.EXPECT().Request(mock.Anything).Return(&tgbotapi.APIResponse{}, nil)
				mockAI.EXPECT().ProcessMessage(mock.Anything, &message.UserMessage{
					ChatID: "123",
					UserID: "456",
					Text:   "test update",
				}).Return(nil, context.Canceled)
				mockBot.EXPECT().Send(mock.MatchedBy(func(msg tgbotapi.MessageConfig) bool {
					return msg.ChatID == 123 && msg.Text != ""
				})).Return(tgbotapi.Message{}, nil)
			},
			expectError: false,
		},
		{
			name: "failed to send update",
			ctx:  context.Background(),
			update: &tgbotapi.Update{
				Message: &tgbotapi.Message{
					Chat: &tgbotapi.Chat{ID: 123},
					From: &tgbotapi.User{
						ID:           456,
						LanguageCode: "en",
					},
					Text: "test update",
				},
			},
			setupMocks: func(mockBot *MockBotAPI, mockAI *MockAIProvider) {
				mockBot.EXPECT().Request(mock.Anything).Return(&tgbotapi.APIResponse{}, nil)
				mockAI.EXPECT().ProcessMessage(mock.Anything, &message.UserMessage{
					ChatID: "123",
					UserID: "456",
					Text:   "test update",
				}).Return(&message.Response{
					Message: "AI response",
				}, nil)
				mockBot.EXPECT().Send(mock.MatchedBy(func(msg tgbotapi.MessageConfig) bool {
					return msg.Text == "AI response"
				})).Return(tgbotapi.Message{}, assert.AnError)
			},
			expectError: false,
		},
		{
			name: "empty response",
			ctx:  context.Background(),
			update: &tgbotapi.Update{
				Message: &tgbotapi.Message{
					Chat: &tgbotapi.Chat{ID: 123},
					From: &tgbotapi.User{
						ID:           456,
						LanguageCode: "en",
					},
					Text: "test update",
				},
			},
			setupMocks: func(mockBot *MockBotAPI, mockAI *MockAIProvider) {
				mockBot.EXPECT().Request(mock.Anything).Return(&tgbotapi.APIResponse{}, nil)
				mockAI.EXPECT().ProcessMessage(mock.Anything, &message.UserMessage{
					ChatID: "123",
					UserID: "456",
					Text:   "test update",
				}).Return(&message.Response{
					Message: "",
				}, nil)
			},
			expectError: false,
		},
		{
			name: "empty update",
			ctx:  context.Background(),
			update: &tgbotapi.Update{
				Message: &tgbotapi.Message{
					Chat: &tgbotapi.Chat{ID: 123},
					From: &tgbotapi.User{
						ID:           456,
						LanguageCode: "en",
					},
					Text: "",
				},
			},
			setupMocks: func(mockBot *MockBotAPI, mockAI *MockAIProvider) {
				mockBot.EXPECT().Request(mock.Anything).Return(&tgbotapi.APIResponse{}, nil)
			},
			expectError: false,
		},
		{
			name: "nil From field",
			ctx:  context.Background(),
			update: &tgbotapi.Update{
				Message: &tgbotapi.Message{
					Chat: &tgbotapi.Chat{ID: 123},
					Text: "test update",
				},
			},
			setupMocks: func(mockBot *MockBotAPI, mockAI *MockAIProvider) {
				mockBot.EXPECT().Request(mock.Anything).Return(&tgbotapi.APIResponse{}, nil)
				// Expect error update to be sent
				mockBot.EXPECT().Send(mock.MatchedBy(func(msg tgbotapi.MessageConfig) bool {
					return msg.ChatID == 123 && msg.Text != ""
				})).Return(tgbotapi.Message{}, nil)
			},
			expectError: false,
		},
		{
			name: "error returned from handler",
			ctx:  context.Background(),
			update: &tgbotapi.Update{
				Message: &tgbotapi.Message{
					Chat: &tgbotapi.Chat{ID: 789},
					From: &tgbotapi.User{ID: 456},
					Text: "handle this",
				},
			},
			setupMocks: func(mockBot *MockBotAPI, mockAI *MockAIProvider) {
				mockBot.EXPECT().Request(mock.Anything).Return(&tgbotapi.APIResponse{}, nil)
				mockBot.EXPECT().Send(mock.Anything).Return(tgbotapi.Message{}, nil).Maybe()
				mockAI.EXPECT().ProcessMessage(mock.Anything, mock.Anything).Return(nil, assert.AnError)
			},
			expectError: false,
		},
		{
			name: "skipped update type",
			ctx:  context.Background(),
			update: &tgbotapi.Update{
				InlineQuery: &tgbotapi.InlineQuery{}, // Unsupported update type
			},
			setupMocks: func(mockBot *MockBotAPI, mockAI *MockAIProvider) {
				// Nothing should happen since the handler skips non-message updates
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBot := NewMockBotAPI(t)
			mockAI := NewMockAIProvider(t)

			service := &ServiceImpl{
				Bot:   mockBot,
				AISvc: mockAI,
			}

			service.handler = service.setupHandler()

			tt.setupMocks(mockBot, mockAI)

			service.processUpdate(tt.ctx, tt.update)

			mockBot.AssertExpectations(t)
			mockAI.AssertExpectations(t)
		})
	}
}

func TestServiceImpl_Run(t *testing.T) {
	mockBot := NewMockBotAPI(t)
	mockAI := NewMockAIProvider(t)

	service := &ServiceImpl{
		Bot:   mockBot,
		AISvc: mockAI,
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

func TestServiceImpl_SendTyping(t *testing.T) {
	tests := []struct {
		name        string
		chatID      int64
		setupMock   func(*MockBotAPI)
		expectError bool
	}{
		{
			name:   "successful send typing action",
			chatID: 12345,
			setupMock: func(mockBot *MockBotAPI) {
				typing := tgbotapi.NewChatAction(12345, tgbotapi.ChatTyping)
				mockBot.EXPECT().Request(typing).Return(&tgbotapi.APIResponse{}, nil)
			},
			expectError: false,
		},
		{
			name:   "failed to send typing action",
			chatID: 12345,
			setupMock: func(mockBot *MockBotAPI) {
				typing := tgbotapi.NewChatAction(12345, tgbotapi.ChatTyping)
				mockBot.EXPECT().Request(typing).Return(nil, assert.AnError)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockBot := NewMockBotAPI(t)

			service := &ServiceImpl{
				Bot: mockBot,
			}

			tt.setupMock(mockBot)
			service.sendTyping(ctx, tt.chatID)

			mockBot.AssertExpectations(t)
		})
	}
}

func TestServiceImpl_KeepTyping(t *testing.T) {
	tests := []struct {
		name          string
		chatID        int64
		setupMock     func(mockBot *MockBotAPI)
		cancelContext bool
		waitDuration  time.Duration
	}{
		{
			name:   "successful typing",
			chatID: 12345,
			setupMock: func(mockBot *MockBotAPI) {
				mockBot.EXPECT().Request(mock.MatchedBy(func(c tgbotapi.Chattable) bool {
					action, ok := c.(tgbotapi.ChatActionConfig)
					return ok && action.ChatID == 12345 && action.Action == tgbotapi.ChatTyping
				})).Return(&tgbotapi.APIResponse{}, nil).Times(2) // Initial + 1 tick
			},
			cancelContext: false,
			waitDuration:  6 * time.Millisecond, // 1 second more than ticker
		},
		{
			name:   "error in typing action",
			chatID: 67890,
			setupMock: func(mockBot *MockBotAPI) {
				mockBot.EXPECT().Request(mock.MatchedBy(func(c tgbotapi.Chattable) bool {
					action, ok := c.(tgbotapi.ChatActionConfig)
					return ok && action.ChatID == 67890 && action.Action == tgbotapi.ChatTyping
				})).Return(nil, assert.AnError).Times(2) // Initial + 1 tick
			},
			cancelContext: false,
			waitDuration:  6 * time.Millisecond,
		},
		{
			name:   "context canceled",
			chatID: 34567,
			setupMock: func(mockBot *MockBotAPI) {
				mockBot.EXPECT().Request(mock.MatchedBy(func(c tgbotapi.Chattable) bool {
					action, ok := c.(tgbotapi.ChatActionConfig)
					return ok && action.ChatID == 34567 && action.Action == tgbotapi.ChatTyping
				})).Return(&tgbotapi.APIResponse{}, nil).Times(1) // Only initial send
			},
			cancelContext: true,
			waitDuration:  1 * time.Millisecond, // Ensure cancellation before tick
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			mockBot := NewMockBotAPI(t)
			service := &ServiceImpl{
				Bot: mockBot,
			}

			if tt.cancelContext {
				go func() {
					time.Sleep(1 * time.Millisecond) // Cancel before next typing action
					cancel()
				}()
			} else {
				defer cancel()
			}

			tt.setupMock(mockBot)
			service.keepTyping(ctx, tt.chatID, 5*time.Millisecond)

			time.Sleep(tt.waitDuration)

			mockBot.AssertExpectations(t)
		})
	}
}
