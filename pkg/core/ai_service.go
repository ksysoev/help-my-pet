package core

import (
	"context"
	"errors"
	"fmt"

	"github.com/ksysoev/help-my-pet/pkg/core/conversation"
	"github.com/ksysoev/help-my-pet/pkg/core/message"
	"github.com/ksysoev/help-my-pet/pkg/core/pet"
)

var (
	// ErrConversationNotFound is returned when a conversation is not found.
	ErrConversationNotFound = fmt.Errorf("conversation not found")

	// ErrRateLimit is returned when the API rate limit is exceeded
	ErrRateLimit = errors.New("rate limit exceeded")

	// ErrGlobalLimit is returned when the global daily request limit is exceeded
	ErrGlobalLimit = errors.New("global request limit exceeded for today, please try again tomorrow")

	ErrProfileNotFound = fmt.Errorf("pet profile not found")
)

type Conversation interface {
	GetID() string
	GetState() conversation.ConversationState
	AddMessage(role, content string)
	History(skip int) string
	StartFollowUpQuestions(initialPrompt string, questions []message.Question) error
	StartProfileQuestions(ctx context.Context) error
	GetCurrentQuestion() (*message.Question, error)
	AddQuestionAnswer(answer string) (bool, error)
	GetQuestionnaireResult() ([]conversation.QuestionAnswer, error)
	CancelQuestionnaire()
}

// ConversationRepository defines the interface for conversation storage operations.
type ConversationRepository interface {
	// Save stores a conversation in the repository.
	Save(ctx context.Context, conversation Conversation) error

	// FindByID retrieves a conversation by its id.
	FindByID(ctx context.Context, id string) (Conversation, error)

	// FindOrCreate retrieves a conversation by id or creates a new one if it doesn't exist.
	FindOrCreate(ctx context.Context, id string) (Conversation, error)
}

// RateLimiter defines the interface for rate limiting functionality
type RateLimiter interface {
	// IsNewQuestionAllowed checks if a user is allowed to ask a new question
	IsNewQuestionAllowed(ctx context.Context, userID string) (bool, error)
	// RecordNewQuestion records that a user has asked a new question
	RecordNewQuestion(ctx context.Context, userID string) error
}

// PetProfileRepository defines the interface for pet profile storage operations
type PetProfileRepository interface {
	SaveProfile(ctx context.Context, userID string, profile *pet.Profile) error
	GetCurrentProfile(ctx context.Context, userID string) (*pet.Profile, error)
}

// LLM interface represents the language model capabilities
type LLM interface {
	Analyze(ctx context.Context, prompt string, imgs []*message.Image) (*message.LLMResult, error)
	Report(ctx context.Context, request string) (*message.LLMResult, error)
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
