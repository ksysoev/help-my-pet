package bot

import (
	"context"
	"fmt"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_handleMessage(t *testing.T) {
	tests := []struct {
		aiErr         error
		name          string
		message       string
		aiResponse    string
		expectError   bool
		mockSendError bool
	}{
		{
			name:        "successful response",
			message:     "What food is good for cats?",
			aiResponse:  "Cats need a balanced diet...",
			aiErr:       nil,
			expectError: false,
		},
		{
			name:        "empty message",
			message:     "",
			aiResponse:  "",
			aiErr:       nil,
			expectError: false,
		},
		{
			name:        "ai error",
			message:     "What food is good for cats?",
			aiResponse:  "",
			aiErr:       fmt.Errorf("ai error"),
			expectError: true,
		},
		{
			name:          "send error",
			message:       "What food is good for cats?",
			aiResponse:    "Cats need a balanced diet...",
			aiErr:         nil,
			expectError:   true,
			mockSendError: true,
		},
		{
			name:          "ai error with send error",
			message:       "What food is good for cats?",
			aiResponse:    "",
			aiErr:         fmt.Errorf("ai error"),
			expectError:   true,
			mockSendError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAI := NewMockAIProvider(t)
			mockBot := NewMockBotAPI(t)

			if tt.message != "" {
				mockAI.EXPECT().
					GetPetAdvice(context.Background(), tt.message).
					Return(tt.aiResponse, tt.aiErr)

				var sendErr error
				if tt.mockSendError {
					sendErr = fmt.Errorf("send error")
				}

				mockBot.EXPECT().
					Send(tgbotapi.NewChatAction(int64(123), tgbotapi.ChatTyping)).
					Return(tgbotapi.Message{}, sendErr)

				if tt.aiErr != nil {
					mockBot.EXPECT().
						Send(tgbotapi.NewMessage(int64(123), "Sorry, I encountered an error while processing your request. Please try again later.")).
						Return(tgbotapi.Message{}, sendErr)
				} else {
					msg := tgbotapi.NewMessage(int64(123), tt.aiResponse)
					msg.ReplyToMessageID = 456
					mockBot.EXPECT().
						Send(msg).
						Return(tgbotapi.Message{}, sendErr)
				}
			}

			msg := &tgbotapi.Message{
				Text: tt.message,
				Chat: &tgbotapi.Chat{
					ID: 123,
				},
				MessageID: 456,
			}

			svc := &ServiceImpl{
				Bot:   mockBot,
				AISvc: mockAI,
			}
			svc.handleMessage(context.Background(), msg)
		})
	}
}

func TestService_Run_SuccessfulMessageHandling(t *testing.T) {
	mockAI := NewMockAIProvider(t)
	mockBot := NewMockBotAPI(t)

	updates := make(chan tgbotapi.Update)
	mockBot.EXPECT().
		GetUpdatesChan(tgbotapi.UpdateConfig{Offset: 0, Timeout: 30}).
		Return(updates)

	mockBot.EXPECT().
		Send(mock.MatchedBy(func(c tgbotapi.Chattable) bool {
			_, ok := c.(tgbotapi.ChatActionConfig)
			return ok
		})).
		Return(tgbotapi.Message{}, nil)

	mockAI.EXPECT().
		GetPetAdvice(mock.Anything, "test message").
		Return("test response", nil)

	mockBot.EXPECT().
		Send(mock.MatchedBy(func(c tgbotapi.Chattable) bool {
			msg, ok := c.(tgbotapi.MessageConfig)
			return ok && msg.Text == "test response"
		})).
		Return(tgbotapi.Message{}, nil)

	mockBot.EXPECT().
		StopReceivingUpdates().
		Return()

	svc := &ServiceImpl{
		Bot:   mockBot,
		AISvc: mockAI,
	}

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error)

	go func() {
		errCh <- svc.Run(ctx)
	}()

	updates <- tgbotapi.Update{
		Message: &tgbotapi.Message{
			Text: "test message",
			Chat: &tgbotapi.Chat{
				ID: 123,
			},
			MessageID: 456,
		},
	}

	time.Sleep(100 * time.Millisecond)
	cancel()
	err := <-errCh
	assert.NoError(t, err)
}

func TestService_Run_EmptyUpdateMessage(t *testing.T) {
	mockAI := NewMockAIProvider(t)
	mockBot := NewMockBotAPI(t)

	updates := make(chan tgbotapi.Update)
	mockBot.EXPECT().
		GetUpdatesChan(tgbotapi.UpdateConfig{Offset: 0, Timeout: 30}).
		Return(updates)

	mockBot.EXPECT().
		StopReceivingUpdates().
		Return()

	svc := &ServiceImpl{
		Bot:   mockBot,
		AISvc: mockAI,
	}

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error)

	go func() {
		errCh <- svc.Run(ctx)
	}()

	updates <- tgbotapi.Update{
		Message: nil,
	}

	time.Sleep(100 * time.Millisecond)
	cancel()
	err := <-errCh
	assert.NoError(t, err)
}

func TestService_Run_SendError(t *testing.T) {
	mockAI := NewMockAIProvider(t)
	mockBot := NewMockBotAPI(t)

	updates := make(chan tgbotapi.Update)
	mockBot.EXPECT().
		GetUpdatesChan(tgbotapi.UpdateConfig{Offset: 0, Timeout: 30}).
		Return(updates)

	mockBot.EXPECT().
		Send(mock.MatchedBy(func(c tgbotapi.Chattable) bool {
			_, ok := c.(tgbotapi.ChatActionConfig)
			return ok
		})).
		Return(tgbotapi.Message{}, fmt.Errorf("send error"))

	mockAI.EXPECT().
		GetPetAdvice(mock.Anything, "test message").
		Return("test response", nil)

	mockBot.EXPECT().
		Send(mock.MatchedBy(func(c tgbotapi.Chattable) bool {
			msg, ok := c.(tgbotapi.MessageConfig)
			return ok && msg.Text == "test response"
		})).
		Return(tgbotapi.Message{}, fmt.Errorf("send error"))

	mockBot.EXPECT().
		StopReceivingUpdates().
		Return()

	svc := &ServiceImpl{
		Bot:   mockBot,
		AISvc: mockAI,
	}

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error)

	go func() {
		errCh <- svc.Run(ctx)
	}()

	updates <- tgbotapi.Update{
		Message: &tgbotapi.Message{
			Text: "test message",
			Chat: &tgbotapi.Chat{
				ID: 123,
			},
			MessageID: 456,
		},
	}

	time.Sleep(100 * time.Millisecond)
	cancel()
	err := <-errCh
	assert.NoError(t, err)
}

func TestNewService(t *testing.T) {
	mockAI := NewMockAIProvider(t)
	cfg := &Config{
		TelegramToken: "test-token",
	}

	t.Run("invalid token", func(t *testing.T) {
		svc, err := NewService(cfg, mockAI)
		assert.Error(t, err)
		assert.Nil(t, svc)
	})

	t.Run("empty token", func(t *testing.T) {
		cfg := &Config{}
		svc, err := NewService(cfg, mockAI)
		assert.Error(t, err)
		assert.Nil(t, svc)
	})
}
