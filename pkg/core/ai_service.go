package core

import (
	"context"
	"fmt"
)

type AIService struct {
	llm LLM
}

func NewAIService(llm LLM) *AIService {
	return &AIService{
		llm: llm,
	}
}

func (s *AIService) GetPetAdvice(ctx context.Context, question string) (string, error) {
	prompt := fmt.Sprintf(`You are a helpful veterinary AI assistant. Please provide accurate, helpful, and compassionate advice for the following pet-related question. If the question involves a serious medical condition, always recommend consulting with a veterinarian.

Question: %s

Please provide a clear and informative response:`, question)

	completion, err := s.llm.Call(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to get AI response: %w", err)
	}

	return completion, nil
}
