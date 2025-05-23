package cmd

import (
	"context"
	"errors"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ksysoev/help-my-pet/pkg/bot"
	"github.com/ksysoev/help-my-pet/pkg/prov/anthropic"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

func TestNewBotRunner(t *testing.T) {
	runner := NewBotRunner()
	assert.NotNil(t, runner)
	assert.NotNil(t, runner.createService)
	assert.Nil(t, runner.botService)
	assert.Nil(t, runner.llmProvider)
}

func TestBotRunner_WithBotService(t *testing.T) {
	runner := NewBotRunner()
	mockService := NewMockBotService(t)

	result := runner.WithBotService(mockService)

	assert.Equal(t, mockService, runner.botService)
	assert.Equal(t, runner, result)
}

func TestBotRunner_RunBot(t *testing.T) {
	tests := []struct {
		setupRunner func(t *testing.T) *BotRunner
		cfg         *Config
		name        string
		errMsg      string
		wantErr     bool
	}{
		{
			name: "success with custom bot service",
			setupRunner: func(t *testing.T) *BotRunner {
				mockService := NewMockBotService(t)
				mockService.On("Run", mock.Anything).Return(nil)
				runner := NewBotRunner()
				return runner.WithBotService(mockService)
			},
			cfg:     &Config{},
			wantErr: false,
		},
		{
			name: "success with custom LLM provider",
			setupRunner: func(t *testing.T) *BotRunner {
				mockBotAPI := bot.NewMockBotAPI(t)
				ch := make(chan tgbotapi.Update)
				mockBotAPI.On("GetUpdatesChan", mock.Anything).Return(tgbotapi.UpdatesChannel(ch)).Once()
				mockBotAPI.On("StopReceivingUpdates").Return().Once()

				runner := NewBotRunner()
				runner.createService = func(cfg *bot.Config, aiSvc bot.AIProvider) (*bot.ServiceImpl, error) {
					svc := &bot.ServiceImpl{
						Bot:   mockBotAPI,
						AISvc: aiSvc,
					}
					return svc, nil
				}
				return runner
			},
			cfg: &Config{
				Bot: bot.Config{},
				AI: anthropic.Config{
					APIKey: "test",
				},
			},
			wantErr: false,
		},
		{
			name: "error creating bot service",
			setupRunner: func(t *testing.T) *BotRunner {
				runner := NewBotRunner()
				runner.createService = func(cfg *bot.Config, aiSvc bot.AIProvider) (*bot.ServiceImpl, error) {
					return nil, errors.New("service creation error")
				}
				return runner
			},
			cfg: &Config{
				AI: anthropic.Config{
					APIKey: "test",
				},
			},
			wantErr: true,
			errMsg:  "failed to create bot service: service creation error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			runner := tt.setupRunner(t)

			go func() {
				// Cancel context after a short delay to simulate shutdown
				time.Sleep(10 * time.Millisecond)
				cancel()
			}()

			err := runner.RunBot(ctx, tt.cfg)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Equal(t, tt.errMsg, err.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
