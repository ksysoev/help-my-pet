package core

import (
	"context"
	"fmt"
	"log/slog"
)

type AIService struct {
	llm         LLM
	repo        ConversationRepository
	profileRepo PetProfileRepository
	rateLimiter RateLimiter
}

func NewAIService(llm LLM, repo ConversationRepository, profileRepo PetProfileRepository, rateLimiter RateLimiter) *AIService {
	return &AIService{
		llm:         llm,
		repo:        repo,
		profileRepo: profileRepo,
		rateLimiter: rateLimiter,
	}
}

func (s *AIService) GetPetAdvice(ctx context.Context, request *UserMessage) (*PetAdviceResponse, error) {
	slog.DebugContext(ctx, "getting pet advice", "input", request.Text)

	conversation, err := s.repo.FindOrCreate(ctx, request.ChatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	// Handle questionnaire state if active
	if conversation.State == StateQuestioning {
		return s.handleQuestionnaireResponse(ctx, conversation, request.Text)
	}

	// Handle new question flow
	return s.handleNewQuestion(ctx, request, conversation)
}

// handleNewQuestion processes a new question from the user
func (s *AIService) handleNewQuestion(ctx context.Context, request *UserMessage, conversation *Conversation) (*PetAdviceResponse, error) {
	// Fetch pet profile from repository
	petProfiles, err := s.profileRepo.GetProfiles(request.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pet profiles: %w", err)
	}

	// Check rate limit for new questions
	if s.rateLimiter != nil {
		allowed, err := s.rateLimiter.IsNewQuestionAllowed(ctx, request.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to check rate limit: %w", err)
		}
		if !allowed {
			return nil, fmt.Errorf("rate limit exceeded for user %s", request.UserID)
		}

		if err := s.rateLimiter.RecordNewQuestion(ctx, request.UserID); err != nil {
			return nil, fmt.Errorf("failed to record rate limit: %w", err)
		}
	}

	// Add user's question to conversation
	conversation.AddMessage("user", request.Text)

	// Save conversation immediately after adding user's message
	if err := s.repo.Save(ctx, conversation); err != nil {
		return nil, fmt.Errorf("failed to save conversation: %w", err)
	}

	// Build prompt with pet profiles and conversation context
	prompt := fmt.Sprintf("Pet profiles: %+v\n\n", petProfiles)
	if len(conversation.GetContext()) <= 1 {
		prompt += request.Text
	} else {
		// Include conversation history
		prompt += "Previous conversation:\n"
		for _, msg := range conversation.GetContext()[:len(conversation.GetContext())-1] {
			prompt += fmt.Sprintf("%s: %s\n", msg.Role, msg.Content)
		}
		prompt += fmt.Sprintf("\nCurrent question: %s", request.Text)
	}

	response, err := s.llm.Call(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI response: %w", err)
	}

	// Add AI's response to conversation
	conversation.AddMessage("assistant", response.Text)

	// Handle follow-up questions if any
	if len(response.Questions) > 0 {
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
	// Store the answer and check if questionnaire is complete
	isComplete, err := conversation.AddQuestionAnswer(answer)
	if err != nil {
		return nil, fmt.Errorf("failed to add question answer: %w", err)
	}

	// Save conversation after adding answer
	if err := s.repo.Save(ctx, conversation); err != nil {
		return nil, fmt.Errorf("failed to save conversation: %w", err)
	}

	if isComplete {
		// Get all collected question-answer pairs
		qaPairs, err := conversation.GetQuestionnaireResult()
		if err != nil {
			return nil, fmt.Errorf("failed to get questionnaire result: %w", err)
		}

		// Build prompt with conversation history and question-answer pairs
		prompt := "Previous conversation:\n"
		history := conversation.GetContext()
		for _, msg := range history[:len(history)-1] {
			prompt += fmt.Sprintf("%s: %s\n", msg.Role, msg.Content)
		}

		prompt += "\nFollow-up information:\n"
		for _, qa := range qaPairs {
			prompt += fmt.Sprintf("Question: %s\nAnswer: %s\n", qa.Question.Text, qa.Answer)
		}

		// Get final response from LLM
		response, err := s.llm.Call(ctx, prompt)
		if err != nil {
			return nil, fmt.Errorf("failed to get AI response: %w", err)
		}

		// Add AI's response to conversation history
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
