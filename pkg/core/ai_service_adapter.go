package core

import (
	"context"
	"fmt"
)

// AIServiceAdapter adapts AIService to match the bot.AIProvider interface
type AIServiceAdapter struct {
	service *AIService
}

// NewAIServiceAdapter creates a new adapter for AIService
func NewAIServiceAdapter(service *AIService) *AIServiceAdapter {
	return &AIServiceAdapter{
		service: service,
	}
}

func (a *AIServiceAdapter) Start(ctx context.Context) (string, error) {
	return a.service.Start(ctx)
}

func (a *AIServiceAdapter) GetPetAdvice(ctx context.Context, question string) (string, error) {
	// Use a default chat ID since the bot interface doesn't provide it
	// We'll use the hash of the question as a unique identifier
	chatID := fmt.Sprintf("chat_%d", hashString(question))
	return a.service.GetPetAdvice(ctx, chatID, question)
}

// hashString creates a simple hash of a string
func hashString(s string) uint32 {
	var hash uint32
	for i := 0; i < len(s); i++ {
		hash = hash*31 + uint32(s[i])
	}
	return hash
}
