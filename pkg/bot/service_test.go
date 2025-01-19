package bot

import (
	"context"
	"fmt"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ksysoev/help-my-pet/pkg/core"
	"github.com/ksysoev/help-my-pet/pkg/i18n"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

	messages := &i18n.Config{
		Languages: map[string]i18n.Messages{
			"en": {
				Error:       "Sorry, I encountered an error while processing your request. Please try again later.",
				Start:       "Welcome to Help My Pet Bot!",
				RateLimit:   "You have reached the maximum number of requests per hour. Please try again later.",
				GlobalLimit: "We have reached our daily request limit. Please come back tomorrow when our budget is refreshed.",
			},
		},
	}

	svc := &ServiceImpl{
		Bot:      mockBot,
		AISvc:    mockAI,
		Messages: messages,
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
		GetPetAdvice(mock.Anything, &core.PetAdviceRequest{
			UserID:  "123",
			ChatID:  "123",
			Message: "test message",
		}).
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

	messages := &i18n.Config{
		Languages: map[string]i18n.Messages{
			"en": {
				Error:       "Sorry, I encountered an error while processing your request. Please try again later.",
				Start:       "Welcome to Help My Pet Bot!",
				RateLimit:   "You have reached the maximum number of requests per hour. Please try again later.",
				GlobalLimit: "We have reached our daily request limit. Please come back tomorrow when our budget is refreshed.",
			},
		},
	}

	svc := &ServiceImpl{
		Bot:      mockBot,
		AISvc:    mockAI,
		Messages: messages,
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

func TestNewService(t *testing.T) {
	mockAI := NewMockAIProvider(t)
	messages := &i18n.Config{
		Languages: map[string]i18n.Messages{
			"en": {
				Error:       "Sorry, I encountered an error while processing your request. Please try again later.",
				Start:       "Welcome to Help My Pet Bot!",
				RateLimit:   "You have reached the maximum number of requests per hour. Please try again later.",
				GlobalLimit: "We have reached our daily request limit. Please come back tomorrow when our budget is refreshed.",
			},
		},
	}

	t.Run("invalid token", func(t *testing.T) {
		cfg := &Config{
			TelegramToken: "test-token",
			Messages:      messages,
		}
		svc, err := NewService(cfg, mockAI)
		assert.Error(t, err)
		assert.Nil(t, svc)
	})

	t.Run("empty token", func(t *testing.T) {
		cfg := &Config{
			Messages: messages,
		}
		svc, err := NewService(cfg, mockAI)
		assert.Error(t, err)
		assert.Nil(t, svc)
	})

	t.Run("with nil AI provider", func(t *testing.T) {
		cfg := &Config{
			TelegramToken: "test-token",
			Messages:      messages,
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
			Messages:      messages,
		}
		svc, err := NewService(cfg, mockAI)
		assert.Error(t, err)
		assert.Nil(t, svc)
	})

	t.Run("with valid token and no rate limiter", func(t *testing.T) {
		cfg := &Config{
			TelegramToken: "1234567890:ABCdefGHIjklMNOpqrsTUVwxyz", // Valid format but invalid token
			Messages:      messages,
		}
		svc, err := NewService(cfg, mockAI)
		assert.Error(t, err) // Error because token is invalid
		assert.Nil(t, svc)
	})

	t.Run("with invalid token format", func(t *testing.T) {
		cfg := &Config{
			TelegramToken: "invalid_token_format",
			Messages:      messages,
		}
		svc, err := NewService(cfg, mockAI)
		assert.Error(t, err)
		assert.Nil(t, svc)
	})
}
