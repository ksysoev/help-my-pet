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

func (s *AIService) Start(ctx context.Context) (string, error) {
	return `Welcome to Help My Pet Bot! 🐾

I'm your personal pet care assistant, ready to help you take better care of your furry friend. I can assist you with:

• Pet health and behavior questions
• Diet and nutrition advice
• Training tips and techniques
• General pet care guidance
• Emergency situation advice

Simply type your question or concern about your pet, and I'll provide helpful, informative answers based on reliable veterinary knowledge. Remember, while I can offer guidance, for serious medical conditions, always consult with a veterinarian.

To get started, just ask me any question about your pet!`, nil
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
