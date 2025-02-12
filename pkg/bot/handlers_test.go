package bot

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ksysoev/help-my-pet/pkg/bot/media"
	"github.com/ksysoev/help-my-pet/pkg/core"
	"github.com/ksysoev/help-my-pet/pkg/core/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_handleMessage(t *testing.T) {
	tests := []struct {
		aiResponse   *message.Response
		aiErr        error
		name         string
		message      string
		expectedText string
		langCode     string
		userID       int64
		expectError  bool
		isStart      bool
		hasPhoto     bool
	}{
		{
			name:         "successful response with keyboard",
			message:      "What food is good for cats?",
			aiResponse:   message.NewResponse("Cats need a balanced diet...", []string{"Yes", "No"}),
			aiErr:        nil,
			expectError:  false,
			userID:       123,
			langCode:     "en",
			expectedText: "Cats need a balanced diet...",
		},
		{
			name:         "successful response without keyboard",
			message:      "What food is good for cats?",
			aiResponse:   message.NewResponse("Cats need a balanced diet...", []string{}),
			aiErr:        nil,
			expectError:  false,
			userID:       123,
			langCode:     "ru",
			expectedText: "Cats need a balanced diet...",
		},
		{
			name:         "empty message",
			message:      "",
			aiResponse:   message.NewResponse("", []string{}),
			aiErr:        nil,
			expectError:  false,
			userID:       123,
			langCode:     "en",
			expectedText: "",
		},
		{
			name:         "ai error",
			message:      "What food is good for cats?",
			aiResponse:   message.NewResponse("", []string{}),
			aiErr:        fmt.Errorf("ai error"),
			expectError:  true,
			userID:       123,
			langCode:     "es",
			expectedText: "",
		},
		{
			name:         "rate limit error",
			message:      "What food is good for cats?",
			aiResponse:   message.NewResponse("", []string{}),
			aiErr:        core.ErrRateLimit,
			expectError:  false,
			userID:       123,
			langCode:     "en",
			expectedText: "You have reached the maximum number of requests per hour. Please try again later.",
		},
		{
			name:         "global limit error",
			message:      "What food is good for cats?",
			aiResponse:   message.NewResponse("", []string{}),
			aiErr:        core.ErrGlobalLimit,
			expectError:  false,
			userID:       123,
			langCode:     "en",
			expectedText: "We have reached our daily request limit. Please come back tomorrow when our budget is refreshed.",
		},
		{
			name:         "message without From field",
			message:      "What food is good for cats?",
			aiResponse:   nil,
			aiErr:        nil,
			expectError:  true,
			userID:       0,
			langCode:     "en",
			expectedText: "",
		},
		{
			name:         "message with photo",
			message:      "",
			hasPhoto:     true,
			aiResponse:   nil,
			aiErr:        nil,
			expectError:  false,
			userID:       123,
			langCode:     "en",
			expectedText: "Photo received",
		},
		{
			name:         "message too long",
			message:      "What food is good for cats? " + strings.Repeat("Very long message. ", 1000),
			aiResponse:   nil,
			aiErr:        nil,
			expectError:  false,
			userID:       123,
			langCode:     "en",
			expectedText: "I apologize, but your message is too long for me to process. Please try to make it shorter and more concise.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAI := NewMockAIProvider(t)
			mockBot := NewMockBotAPI(t)

			svc := &ServiceImpl{
				Bot:       mockBot,
				AISvc:     mockAI,
				collector: media.NewCollector(),
			}

			msg := &tgbotapi.Message{
				Text: tt.message,
				Chat: &tgbotapi.Chat{
					ID: 123,
				},
				MessageID: 456,
			}

			// Set Photo field if hasPhoto is true
			if tt.hasPhoto {
				msg.Photo = []tgbotapi.PhotoSize{{FileID: "test-photo"}}
			}

			// Set From field only if userID is not 0
			if tt.userID != 0 {
				msg.From = &tgbotapi.User{
					ID:           tt.userID,
					LanguageCode: tt.langCode,
				}
			}

			if !tt.isStart && tt.message != "" && msg.From != nil && !tt.hasPhoto && !strings.Contains(tt.name, "message too long") {
				expectedRequest := &message.UserMessage{
					UserID: "123",
					ChatID: "123",
					Text:   tt.message,
				}
				mockAI.EXPECT().ProcessMessage(mock.Anything, expectedRequest).Return(tt.aiResponse, tt.aiErr)
			}

			msgConfig, err := svc.Handle(context.Background(), msg)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			if tt.message == "" && !tt.hasPhoto {
				assert.Equal(t, tgbotapi.MessageConfig{}, msgConfig)
				return
			}

			assert.Equal(t, tt.expectedText, msgConfig.Text)
			assert.Equal(t, int64(123), msgConfig.ChatID)

			if tt.aiResponse != nil && len(tt.aiResponse.Answers) > 0 {
				keyboard, ok := msgConfig.ReplyMarkup.(tgbotapi.ReplyKeyboardMarkup)
				assert.True(t, ok)
				assert.Equal(t, len(tt.aiResponse.Answers), len(keyboard.Keyboard))
				for i, row := range keyboard.Keyboard {
					assert.Equal(t, tt.aiResponse.Answers[i], row[0].Text)
				}
			}
		})
	}
}

func TestService_Run_SuccessfulMessageHandling(t *testing.T) {
	mockAI := NewMockAIProvider(t)
	mockBot := NewMockBotAPI(t)

	svc := &ServiceImpl{
		Bot:   mockBot,
		AISvc: mockAI,
	}

	svc.handler = svc.setupHandler()

	updates := make(chan tgbotapi.Update)
	mockBot.EXPECT().
		GetUpdatesChan(tgbotapi.UpdateConfig{Offset: 0, Timeout: 30}).
		Return(updates)

	// Expect typing action
	mockBot.EXPECT().
		Request(mock.MatchedBy(func(c tgbotapi.Chattable) bool {
			action, ok := c.(tgbotapi.ChatActionConfig)
			return ok && action.ChatID == 123 && action.Action == tgbotapi.ChatTyping
		})).
		Return(&tgbotapi.APIResponse{}, nil)

	// Expect AI request
	mockAI.EXPECT().
		ProcessMessage(mock.Anything, &message.UserMessage{
			UserID: "123",
			ChatID: "123",
			Text:   "test message",
		}).
		Return(message.NewResponse("test response", []string{}), nil)

	// Expect message send
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
			Text: "test message",
			Chat: &tgbotapi.Chat{
				ID: 123,
			},
			MessageID: 456,
			From: &tgbotapi.User{
				ID:           123,
				LanguageCode: "en",
			},
		},
	}

	time.Sleep(100 * time.Millisecond)
	cancel()
	err := <-errCh
	assert.NoError(t, err)
}

func TestService_handleProcessingError(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		langCode     string
		expectedText string
	}{
		{
			name:         "rate limit error",
			err:          core.ErrRateLimit,
			langCode:     "en",
			expectedText: "You have reached the maximum number of requests per hour. Please try again later.",
		},
		{
			name:         "global limit error",
			err:          core.ErrGlobalLimit,
			langCode:     "en",
			expectedText: "We have reached our daily request limit. Please come back tomorrow when our budget is refreshed.",
		},
		{
			name:         "unhandled error",
			err:          fmt.Errorf("unknown error"),
			langCode:     "en",
			expectedText: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBot := NewMockBotAPI(t)

			svc := &ServiceImpl{
				Bot: mockBot,
			}

			msg := &tgbotapi.Message{
				Chat: &tgbotapi.Chat{
					ID: 123,
				},
				From: &tgbotapi.User{
					LanguageCode: tt.langCode,
				},
			}

			msgConfig, err := svc.handleProcessingError(context.Background(), tt.err, msg)

			if tt.name == "unhandled error" {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedText, msgConfig.Text)
				assert.Equal(t, int64(123), msgConfig.ChatID)
			}
		})
	}
}
