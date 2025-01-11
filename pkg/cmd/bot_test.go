package cmd

import (
	"context"
	"errors"
	"testing"

	"github.com/ksysoev/help-my-pet/pkg/bot"
	"github.com/ksysoev/help-my-pet/pkg/core"
	"github.com/stretchr/testify/assert"
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
	mockService := bot.NewMockService(t)

	result := runner.WithBotService(mockService)

	assert.Equal(t, mockService, runner.botService)
	assert.Equal(t, runner, result)
}

func TestBotRunner_WithLLMProvider(t *testing.T) {
	runner := NewBotRunner()
	mockProvider := core.NewMockLLM(t)

	result := runner.WithLLMProvider(mockProvider)

	assert.Equal(t, mockProvider, runner.llmProvider)
	assert.Equal(t, runner, result)
}

func TestBotRunner_RunBot(t *testing.T) {
	tests := []struct {
		setupRunner func() *BotRunner
		cfg         *Config
		name        string
		errMsg      string
		wantErr     bool
	}{
		{
			name: "success with custom bot service",
			setupRunner: func() *BotRunner {
				mockService := bot.NewMockService(t)
				mockService.EXPECT().Run(context.Background()).Return(nil)
				runner := NewBotRunner()
				return runner.WithBotService(mockService)
			},
			cfg:     &Config{},
			wantErr: false,
		},
		{
			name: "success with custom LLM provider",
			setupRunner: func() *BotRunner {
				mockLLMProvider := core.NewMockLLM(t)
				mockService := bot.NewMockService(t)
				mockService.EXPECT().Run(context.Background()).Return(nil)
				runner := NewBotRunner()
				runner.createService = func(cfg *bot.Config, aiSvc bot.AIProvider) (bot.Service, error) {
					return mockService, nil
				}
				return runner.WithLLMProvider(mockLLMProvider)
			},
			cfg:     &Config{},
			wantErr: false,
		},
		{
			name: "error creating bot service",
			setupRunner: func() *BotRunner {
				mockLLMProvider := core.NewMockLLM(t)
				runner := NewBotRunner()
				runner.createService = func(cfg *bot.Config, aiSvc bot.AIProvider) (bot.Service, error) {
					return nil, errors.New("service creation error")
				}
				return runner.WithLLMProvider(mockLLMProvider)
			},
			cfg:     &Config{},
			wantErr: true,
			errMsg:  "failed to create bot service: service creation error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := tt.setupRunner()
			err := runner.RunBot(context.Background(), tt.cfg)

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
