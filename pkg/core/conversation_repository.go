package core

import (
	"context"
	"fmt"
)

// ErrConversationNotFound is returned when a conversation is not found.
var ErrConversationNotFound = fmt.Errorf("conversation not found")

// ConversationRepository defines the interface for conversation storage operations.
type ConversationRepository interface {
	// Save stores a conversation in the repository.
	Save(ctx context.Context, conversation *Conversation) error

	// FindByID retrieves a conversation by its ID.
	FindByID(ctx context.Context, id string) (*Conversation, error)

	// FindOrCreate retrieves a conversation by ID or creates a new one if it doesn't exist.
	FindOrCreate(ctx context.Context, id string) (*Conversation, error)
}
