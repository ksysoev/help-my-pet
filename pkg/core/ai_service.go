package core

import (
	"context"
	"fmt"
	"log/slog"
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

func (s *AIService) GetPetAdvice(ctx context.Context, chatID string, userInput string) (*PetAdviceResponse, error) {
	slog.Info("getting pet advice", "chat_id", chatID, "input", userInput)

	conversation, err := s.repo.FindOrCreate(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	// Handle questionnaire state if active
	if conversation.State == StateQuestioning {
		return s.handleQuestionnaireResponse(ctx, conversation, userInput)
	}

	// Handle new question flow
	return s.handleNewQuestion(ctx, conversation, userInput)
}

// handleNewQuestion processes a new question from the user
func (s *AIService) handleNewQuestion(ctx context.Context, conversation *Conversation, question string) (*PetAdviceResponse, error) {
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

	response, err := s.llm.Call(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI response: %w", err)
	}

	// Add AI's response to conversation
	conversation.AddMessage("assistant", response.Text)

	// Store follow-up questions if any
	if len(response.Questions) > 0 {
		// Store questions in a format that can be retrieved later
		questionsStr := "\nFollow-up questions:"
		for i, q := range response.Questions {
			questionsStr += fmt.Sprintf("\n%d. %s", i+1, q.Text)
			if len(q.Answers) > 0 {
				questionsStr += "\nOptions:"
				for _, answer := range q.Answers {
					questionsStr += fmt.Sprintf("\n- %s", answer)
				}
			}
		}
		conversation.AddMessage("assistant_questions", questionsStr)

		// Initialize questionnaire
		conversation.StartQuestionnaire(response.Text, response.Questions)

		// Get the first question
		currentQuestion, err := conversation.GetCurrentQuestion()
		if err != nil {
			return nil, fmt.Errorf("failed to get first question: %w", err)
		}

		// Save conversation state
		if err := s.repo.Save(ctx, conversation); err != nil {
			return nil, fmt.Errorf("failed to save conversation: %w", err)
		}

		// Return response with first question
		answers := []string{}
		if len(currentQuestion.Answers) > 0 {
			answers = currentQuestion.Answers
		}
		return NewPetAdviceResponse(
			response.Text+"\n\n"+currentQuestion.Text,
			answers,
		), nil
	}

	// Save conversation state
	if err := s.repo.Save(ctx, conversation); err != nil {
		return nil, fmt.Errorf("failed to save conversation: %w", err)
	}

	return NewPetAdviceResponse(response.Text, []string{}), nil
}

// handleQuestionnaireResponse processes a response to a follow-up question
func (s *AIService) handleQuestionnaireResponse(ctx context.Context, conversation *Conversation, answer string) (*PetAdviceResponse, error) {
	// Add user's answer to conversation
	conversation.AddMessage("user", answer)

	// Store the answer and check if questionnaire is complete
	isComplete, err := conversation.AddQuestionAnswer(answer)
	if err != nil {
		return nil, fmt.Errorf("failed to add question answer: %w", err)
	}

	if isComplete {
		// Get all collected answers
		initialPrompt, answers, err := conversation.GetQuestionnaireResult()
		if err != nil {
			return nil, fmt.Errorf("failed to get questionnaire result: %w", err)
		}

		// Build prompt with all answers
		prompt := initialPrompt + "\n\nFollow-up information:\n"
		for i, q := range conversation.Questionnaire.Questions {
			prompt += fmt.Sprintf("%s: %s\n", q.Text, answers[i])
		}

		// Get final response from LLM
		response, err := s.llm.Call(ctx, prompt)
		if err != nil {
			return nil, fmt.Errorf("failed to get AI response: %w", err)
		}

		// Add AI's final response to conversation
		conversation.AddMessage("assistant", response.Text)

		// Save conversation state
		if err := s.repo.Save(ctx, conversation); err != nil {
			return nil, fmt.Errorf("failed to save conversation: %w", err)
		}

		return NewPetAdviceResponse(response.Text, []string{}), nil
	}

	// Get next question
	currentQuestion, err := conversation.GetCurrentQuestion()
	if err != nil {
		return nil, fmt.Errorf("failed to get next question: %w", err)
	}

	// Save conversation state
	if err := s.repo.Save(ctx, conversation); err != nil {
		return nil, fmt.Errorf("failed to save conversation: %w", err)
	}

	return NewPetAdviceResponse(currentQuestion.Text, currentQuestion.Answers), nil
}
