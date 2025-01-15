package bot

import (
	"context"
	"fmt"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ksysoev/help-my-pet/pkg/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_handleMessage(t *testing.T) {
	tests := []struct {
		aiErr         error
		aiResponse    *core.PetAdviceResponse
		name          string
		message       string
		userID        int64
		expectError   bool
		mockSendError bool
		isStart       bool
	}{
		{
			name:        "successful response",
			message:     "What food is good for cats?",
			aiResponse:  core.NewPetAdviceResponse("Cats need a balanced diet...", []string{"Yes", "No"}),
			aiErr:       nil,
			expectError: false,
			userID:      123,
		},
		{
			name:        "empty message",
			message:     "",
			aiResponse:  core.NewPetAdviceResponse("", []string{}),
			aiErr:       nil,
			expectError: false,
			userID:      123,
		},
		{
			name:        "ai error",
			message:     "What food is good for cats?",
			aiResponse:  core.NewPetAdviceResponse("", []string{}),
			aiErr:       fmt.Errorf("ai error"),
			expectError: true,
			userID:      123,
		},
		{
			name:          "send error",
			message:       "What food is good for cats?",
			aiResponse:    core.NewPetAdviceResponse("Cats need a balanced diet...", []string{}),
			aiErr:         nil,
			expectError:   true,
			mockSendError: true,
			userID:        123,
		},
		{
			name:          "ai error with send error",
			message:       "What food is good for cats?",
			aiResponse:    core.NewPetAdviceResponse("", []string{}),
			aiErr:         fmt.Errorf("ai error"),
			expectError:   true,
			mockSendError: true,
			userID:        123,
		},
		{
			name:        "start command",
			message:     "/start",
			aiResponse:  core.NewPetAdviceResponse("Welcome to Help My Pet Bot!", []string{}),
			aiErr:       nil,
			expectError: false,
			isStart:     true,
			userID:      123,
		},
		{
			name:          "start command with error",
			message:       "/start",
			aiResponse:    core.NewPetAdviceResponse("", []string{}),
			aiErr:         fmt.Errorf("start error"),
			expectError:   true,
			isStart:       true,
			mockSendError: false,
			userID:        123,
		},
		{
			name:        "message without From field",
			message:     "What food is good for cats?",
			aiResponse:  core.NewPetAdviceResponse("Cats need a balanced diet...", []string{}),
			aiErr:       nil,
			expectError: false,
			userID:      0,
		},
		{
			name:          "error sending start message",
			message:       "/start",
			aiResponse:    core.NewPetAdviceResponse("Welcome to Help My Pet Bot!", []string{}),
			aiErr:         nil,
			expectError:   true,
			mockSendError: true,
			isStart:       true,
			userID:        123,
		},
		{
			name:          "error sending error message",
			message:       "What food is good for cats?",
			aiResponse:    core.NewPetAdviceResponse("", []string{}),
			aiErr:         fmt.Errorf("ai error"),
			expectError:   true,
			mockSendError: true,
			userID:        123,
		},
		{
			name:          "typing action error but successful response",
			message:       "What food is good for cats?",
			aiResponse:    core.NewPetAdviceResponse("Cats need a balanced diet...", []string{}),
			aiErr:         nil,
			expectError:   false,
			mockSendError: true,
			userID:        123,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAI := NewMockAIProvider(t)
			mockBot := NewMockBotAPI(t)

			var sendErr error
			if tt.mockSendError {
				sendErr = fmt.Errorf("send error")
			}

			svc := &ServiceImpl{
				Bot:   mockBot,
				AISvc: mockAI,
			}

			msg := &tgbotapi.Message{
				Text: tt.message,
				Chat: &tgbotapi.Chat{
					ID: 123,
				},
				MessageID: 456,
			}

			// Set From field only if userID is not 0
			if tt.userID != 0 {
				msg.From = &tgbotapi.User{
					ID: tt.userID,
				}
			}

			// Set up expectations based on the test case
			if tt.message == "" {
				svc.handleMessage(context.Background(), msg)
				return
			}

			if tt.isStart {
				mockAI.EXPECT().
					Start(context.Background()).
					Return(tt.aiResponse.Message, tt.aiErr)

				if tt.aiErr != nil {
					mockBot.EXPECT().
						Send(tgbotapi.NewMessage(int64(123), "Sorry, I encountered an error while processing your request. Please try again later.")).
						Return(tgbotapi.Message{}, sendErr)
				} else {
					msg := tgbotapi.NewMessage(int64(123), tt.aiResponse.Message)
					mockBot.EXPECT().
						Send(msg).
						Return(tgbotapi.Message{}, sendErr)
				}
			} else {
				mockAI.EXPECT().
					GetPetAdvice(context.Background(), "123", tt.message).
					Return(tt.aiResponse, tt.aiErr)

				mockBot.EXPECT().
					Send(tgbotapi.NewChatAction(int64(123), tgbotapi.ChatTyping)).
					Return(tgbotapi.Message{}, sendErr)

				if tt.aiErr != nil {
					mockBot.EXPECT().
						Send(tgbotapi.NewMessage(int64(123), "Sorry, I encountered an error while processing your request. Please try again later.")).
						Return(tgbotapi.Message{}, sendErr)
				} else {
					responseMsg := tgbotapi.NewMessage(int64(123), tt.aiResponse.Message)
					responseMsg.ReplyToMessageID = 456

					// Add reply keyboard if there are answers
					if len(tt.aiResponse.Answers) > 0 {
						keyboard := make([][]tgbotapi.KeyboardButton, len(tt.aiResponse.Answers))
						for i, answer := range tt.aiResponse.Answers {
							keyboard[i] = []tgbotapi.KeyboardButton{
								{Text: answer},
							}
						}
						responseMsg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
							Keyboard:        keyboard,
							OneTimeKeyboard: true,
							ResizeKeyboard:  true,
						}
					}

					mockBot.EXPECT().
						Send(responseMsg).
						Return(tgbotapi.Message{}, sendErr)
				}
			}

			svc.handleMessage(context.Background(), msg)
		})
	}
}

func TestService_Run_SuccessfulMessageHandling(t *testing.T) {
	tests := []struct {
		name    string
		message string
		userID  int64
	}{
		{
			name:    "successful message",
			message: "test message",
			userID:  123,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAI := NewMockAIProvider(t)
			mockBot := NewMockBotAPI(t)
			svc := &ServiceImpl{
				Bot:   mockBot,
				AISvc: mockAI,
			}

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
				GetPetAdvice(mock.Anything, fmt.Sprintf("%d", tt.userID), tt.message).
				Return(core.NewPetAdviceResponse("test response", []string{}), nil)

			mockBot.EXPECT().
				Send(mock.MatchedBy(func(c tgbotapi.Chattable) bool {
					msg, ok := c.(tgbotapi.MessageConfig)
					return ok && msg.Text == "test response"
				})).
				Return(tgbotapi.Message{}, nil)

			mockBot.EXPECT().
				StopReceivingUpdates().
				Return()

			ctx, cancel := context.WithCancel(context.Background())
			errCh := make(chan error)

			go func() {
				errCh <- svc.Run(ctx)
			}()

			updates <- tgbotapi.Update{
				Message: &tgbotapi.Message{
					Text: tt.message,
					Chat: &tgbotapi.Chat{
						ID: tt.userID,
					},
					MessageID: 456,
					From: &tgbotapi.User{
						ID: tt.userID,
					},
				},
			}

			time.Sleep(100 * time.Millisecond)
			cancel()
			err := <-errCh
			assert.NoError(t, err)
		})
	}
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
		GetPetAdvice(mock.Anything, "123", "test message").
		Return(core.NewPetAdviceResponse("test response", []string{}), nil)

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
			From: &tgbotapi.User{
				ID: 123,
			},
		},
	}

	time.Sleep(100 * time.Millisecond)
	cancel()
	err := <-errCh
	assert.NoError(t, err)
}

func TestNewService(t *testing.T) {
	mockAI := NewMockAIProvider(t)

	t.Run("invalid token", func(t *testing.T) {
		cfg := &Config{
			TelegramToken: "test-token",
		}
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

	t.Run("with nil AI provider", func(t *testing.T) {
		cfg := &Config{
			TelegramToken: "test-token",
		}
		svc, err := NewService(cfg, nil)
		assert.Error(t, err)
		assert.Nil(t, svc)
	})

	t.Run("with nil config", func(t *testing.T) {
		svc, err := NewService(nil, mockAI)
		assert.Error(t, err)
		assert.Nil(t, svc)
	})

	t.Run("with valid token but NewBotAPI fails", func(t *testing.T) {
		cfg := &Config{
			TelegramToken: "invalid:token:format", // This format should trigger a validation error in NewBotAPI
		}
		svc, err := NewService(cfg, mockAI)
		assert.Error(t, err)
		assert.Nil(t, svc)
	})

	t.Run("with valid token and no rate limiter", func(t *testing.T) {
		cfg := &Config{
			TelegramToken: "1234567890:ABCdefGHIjklMNOpqrsTUVwxyz", // Valid format but invalid token
		}
		svc, err := NewService(cfg, mockAI)
		assert.Error(t, err) // Error because token is invalid
		assert.Nil(t, svc)
	})

	t.Run("with invalid token format", func(t *testing.T) {
		cfg := &Config{
			TelegramToken: "invalid_token_format",
		}
		svc, err := NewService(cfg, mockAI)
		assert.Error(t, err)
		assert.Nil(t, svc)
	})

}
