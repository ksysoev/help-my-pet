package core

//go:generate mockery --name ConversationRepository

import (
	"context"
	"fmt"

	"github.com/ksysoev/help-my-pet/pkg/core/conversation"
)

// ErrConversationNotFound is returned when a conversation is not found.
var ErrConversationNotFound = fmt.Errorf("conversation not found")

// ConversationRepository defines the interface for conversation storage operations.
type ConversationRepository interface {
	// Save stores a conversation in the repository.
	Save(ctx context.Context, conversation *conversation.Conversation) error

	// FindByID retrieves a conversation by its ID.
	FindByID(ctx context.Context, id string) (*conversation.Conversation, error)

	// FindOrCreate retrieves a conversation by ID or creates a new one if it doesn't exist.
	FindOrCreate(ctx context.Context, id string) (*conversation.Conversation, error)
}
