package bot

import (
	"context"

	"github.com/ksysoev/help-my-pet/pkg/core"
)

// AIServiceAdapter adapts core.AIService to bot.AIProvider interface
type AIServiceAdapter struct {
	service *core.AIService
}

// NewAIServiceAdapter creates a new AIServiceAdapter
func NewAIServiceAdapter(service *core.AIService) *AIServiceAdapter {
	return &AIServiceAdapter{
		service: service,
	}
}

// GetPetAdvice implements AIProvider interface
func (a *AIServiceAdapter) GetPetAdvice(ctx context.Context, chatID string, question string) (*core.PetAdviceResponse, error) {
	return a.service.GetPetAdvice(ctx, chatID, question)
}

// Start implements AIProvider interface
func (a *AIServiceAdapter) Start(ctx context.Context) (string, error) {
	return a.service.Start(ctx)
}
