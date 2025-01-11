package core

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
)

type AIService struct {
	llm   LLM
	model string
}

func NewAIService(apiKey, model string) *AIService {
	llm, err := anthropic.New(anthropic.WithToken(apiKey))
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize Anthropic LLM: %v", err))
	}

	return &AIService{
		model: model,
		llm:   llm,
	}
}

func (s *AIService) GetPetAdvice(ctx context.Context, question string) (string, error) {
	prompt := fmt.Sprintf(`You are a helpful veterinary AI assistant. Please provide accurate, helpful, and compassionate advice for the following pet-related question. If the question involves a serious medical condition, always recommend consulting with a veterinarian.

Question: %s

Please provide a clear and informative response:`, question)

	completion, err := s.llm.Call(ctx, prompt,
		llms.WithModel(s.model),
		llms.WithMaxTokens(1000),
	)
	if err != nil {
		return "", fmt.Errorf("failed to get AI response: %w", err)
	}

	return completion, nil
}