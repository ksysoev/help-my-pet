package bot

import (
	"context"

	"github.com/ksysoev/help-my-pet/pkg/core"
)

// AIServiceBotAdapter adapts AIService to the AIProvider interface
type AIServiceBotAdapter struct {
	service *core.AIService
}

// NewAIServiceAdapter creates a new AIServiceBotAdapter
func NewAIServiceAdapter(service *core.AIService) AIProvider {
	return &AIServiceBotAdapter{
		service: service,
	}
}

// Start initializes the AI service
func (a *AIServiceBotAdapter) Start(ctx context.Context) (string, error) {
	return a.service.Start(ctx)
}

// GetPetAdvice gets advice for a pet-related question
func (a *AIServiceBotAdapter) GetPetAdvice(ctx context.Context, userID string, question string) (*core.PetAdviceResponse, error) {
	request := &core.PetAdviceRequest{
		UserID:  userID,
		ChatID:  userID, // Using userID for chatID as they represent the same conversation
		Message: question,
	}

	return a.service.GetPetAdvice(ctx, request)
}
