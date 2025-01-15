package bot

import (
	"context"
	"fmt"
	"strconv"

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
func (a *AIServiceBotAdapter) GetPetAdvice(ctx context.Context, chatID string, question string) (*core.PetAdviceResponse, error) {
	userID, err := strconv.ParseInt(chatID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid chat ID: %w", err)
	}

	request := &core.PetAdviceRequest{
		UserID:  strconv.FormatInt(userID, 10),
		ChatID:  chatID,
		Message: question,
	}

	return a.service.GetPetAdvice(ctx, request)
}
