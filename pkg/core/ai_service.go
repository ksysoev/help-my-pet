package core

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/ksysoev/help-my-pet/pkg/core/conversation"
)

// ErrConversationNotFound is returned when a conversation is not found.
var ErrConversationNotFound = fmt.Errorf("conversation not found")

// ErrRateLimit is returned when the API rate limit is exceeded
var ErrRateLimit = errors.New("rate limit exceeded")

// ErrGlobalLimit is returned when the global daily request limit is exceeded
var ErrGlobalLimit = errors.New("global request limit exceeded for today, please try again tomorrow")

// ConversationRepository defines the interface for conversation storage operations.
type ConversationRepository interface {
	// Save stores a conversation in the repository.
	Save(ctx context.Context, conversation *conversation.Conversation) error

	// FindByID retrieves a conversation by its ID.
	FindByID(ctx context.Context, id string) (*conversation.Conversation, error)

	// FindOrCreate retrieves a conversation by ID or creates a new one if it doesn't exist.
	FindOrCreate(ctx context.Context, id string) (*conversation.Conversation, error)
}

// LLMResult represents a structured response from the LLM
type LLMResult struct {
	Text      string                  `json:"text"`      // Main response text
	Questions []conversation.Question `json:"questions"` // Optional follow-up questions
}

// LLM interface represents the language model capabilities
type LLM interface {
	// Call sends a prompt to the LLM and returns a structured response
	Call(ctx context.Context, prompt string) (*LLMResult, error)
}

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

func (s *AIService) ProcessMessage(ctx context.Context, request *UserMessage) (*Response, error) {
	slog.DebugContext(ctx, "getting pet advice", "input", request.Text)

	conv, err := s.repo.FindOrCreate(ctx, request.ChatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	switch conv.State {
	case conversation.StateNormal:
		return s.handleNewQuestion(ctx, conv, request)
	case conversation.StatePetProfileQuestioning:
		return s.ProcessProfileAnswer(ctx, conv, request)
	case conversation.StateFollowUpQuestioning:
		return s.ProcessFollowUpAnswer(ctx, conv, request)
	default:
		return nil, fmt.Errorf("unknown conversation state: %s", conv.State)
	}
}

// handleNewQuestion processes a new question from the user
func (s *AIService) handleNewQuestion(ctx context.Context, conv *conversation.Conversation, request *UserMessage) (*Response, error) {
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

	// Add user's question to conv
	conv.AddMessage("user", request.Text)

	// Save conv immediately after adding user's message
	if err := s.repo.Save(ctx, conv); err != nil {
		return nil, fmt.Errorf("failed to save conversation: %w", err)
	}

	// Build prompt with pet profiles and conv context
	var prompt string

	// Fetch pet profile from repository
	petProfile, err := s.profileRepo.GetCurrentProfile(ctx, request.UserID)
	if errors.Is(err, ErrProfileNotFound) {
		// If no profile found, do not include pet profiles in prompt
	} else if err != nil {
		return nil, fmt.Errorf("failed to fetch pet profiles: %w", err)
	} else {
		// Include pet profiles in prompt
		prompt += fmt.Sprintf("%s\n\n", petProfile.String())
	}

	convCtx := conv.GetContext()
	fmt.Println(convCtx)
	if len(convCtx) <= 1 {
		prompt += request.Text
	} else {
		// Include conv history
		prompt += "Previous conversation:\n"
		for _, msg := range convCtx[:len(convCtx)-1] {
			prompt += fmt.Sprintf("%s: %s\n", msg.Role, msg.Content)
		}
		prompt += fmt.Sprintf("\nCurrent question: %s", request.Text)
	}

	response, err := s.llm.Call(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI response: %w", err)
	}

	// Add AI's response to conv
	conv.AddMessage("assistant", response.Text)

	// Handle follow-up questions if any
	if len(response.Questions) > 0 {
		// Initialize questionnaire
		if err := conv.StartFollowUpQuestions(response.Text, response.Questions); err != nil {
			return nil, fmt.Errorf("failed to start follow-up questions: %w", err)
		}

		// Get the first question
		currentQuestion, err := conv.GetCurrentQuestion()
		if err != nil {
			return nil, fmt.Errorf("failed to get first question: %w", err)
		}

		// Save conv state
		if err := s.repo.Save(ctx, conv); err != nil {
			return nil, fmt.Errorf("failed to save conversation: %w", err)
		}

		// Return response with first question
		answers := []string{}
		if len(currentQuestion.Answers) > 0 {
			answers = currentQuestion.Answers
		}
		return NewResponse(
			response.Text+"\n\n"+currentQuestion.Text,
			answers,
		), nil
	}

	// Save conv state
	if err := s.repo.Save(ctx, conv); err != nil {
		return nil, fmt.Errorf("failed to save conversation: %w", err)
	}

	return NewResponse(response.Text, []string{}), nil
}
