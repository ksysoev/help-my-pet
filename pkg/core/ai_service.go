package core

import (
	"context"
	"fmt"
)

type AIService struct {
	llm  LLM
	repo ConversationRepository
}

func NewAIService(llm LLM, repo ConversationRepository) *AIService {
	return &AIService{
		llm:  llm,
		repo: repo,
	}
}

func (s *AIService) Start(ctx context.Context) (string, error) {
	return `Welcome to Help My Pet Bot! 🐾

I'm your personal pet care assistant, ready to help you take better care of your furry friend. I can assist you with:

• Pet health and behavior questions
• Diet and nutrition advice
• Training tips and techniques
• General pet care guidance

Simply type your question or concern about your pet, and I'll provide helpful, informative answers based on reliable veterinary knowledge. Remember, while I can offer guidance, for serious medical conditions, always consult with a veterinarian.

To get started, just ask me any question about your pet!`, nil
}

func (s *AIService) GetPetAdvice(ctx context.Context, chatID string, question string) (string, error) {
	conversation, err := s.repo.FindOrCreate(ctx, chatID)
	if err != nil {
		return "", fmt.Errorf("failed to get conversation: %w", err)
	}

	// Add user's question to conversation
	conversation.AddMessage("user", question)

	// Build prompt with conversation context
	var prompt string
	if len(conversation.GetContext()) <= 1 {
		prompt = question
	} else {
		// Include conversation history
		prompt = "Previous conversation:\n"
		for _, msg := range conversation.GetContext()[:len(conversation.GetContext())-1] {
			prompt += fmt.Sprintf("%s: %s\n", msg.Role, msg.Content)
		}
		prompt += fmt.Sprintf("\nCurrent question: %s", question)
	}

	completion, err := s.llm.Call(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to get AI response: %w", err)
	}

	// Add AI's response to conversation
	conversation.AddMessage("assistant", completion)

	// Save updated conversation
	if err := s.repo.Save(ctx, conversation); err != nil {
		return "", fmt.Errorf("failed to save conversation: %w", err)
	}

	return completion, nil
}
