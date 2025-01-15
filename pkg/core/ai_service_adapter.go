package core

import (
	"context"
	"fmt"
)

// AIServiceAdapter adapts AIService to the AIProvider interface
type AIServiceAdapter struct {
	service *AIService
}

// NewAIServiceAdapter creates a new AIServiceAdapter
func NewAIServiceAdapter(service *AIService) *AIServiceAdapter {
	return &AIServiceAdapter{
		service: service,
	}
}

// Start initializes the AI service
func (a *AIServiceAdapter) Start(ctx context.Context) (string, error) {
	return a.service.Start(ctx)
}

// GetPetAdvice gets advice for a pet-related question
func (a *AIServiceAdapter) GetPetAdvice(ctx context.Context, chatID string, question string) (string, error) {
	request := &PetAdviceRequest{
		UserID:  chatID, // Using chatID as userID for now, should be updated when proper user ID is available
		ChatID:  chatID,
		Message: question,
	}

	response, err := a.service.GetPetAdvice(ctx, request)
	if err != nil {
		return "", fmt.Errorf("failed to get pet advice: %w", err)
	}

	// Format response with answers if available
	if len(response.Answers) > 0 {
		result := response.Message + "\n\nOptions:"
		for _, answer := range response.Answers {
			result += fmt.Sprintf("\n- %s", answer)
		}
		return result, nil
	}

	return response.Message, nil
}
