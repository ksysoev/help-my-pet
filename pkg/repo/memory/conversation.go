package memory

import (
	"context"
	"sync"

	"github.com/ksysoev/help-my-pet/pkg/core"
)

// ConversationRepository implements core.ConversationRepository interface
// using in-memory storage.
type ConversationRepository struct {
	conversations map[string]*core.Conversation
	mu            sync.RWMutex
}

// NewConversationRepository creates a new instance of ConversationRepository.
func NewConversationRepository() *ConversationRepository {
	return &ConversationRepository{
		conversations: make(map[string]*core.Conversation),
	}
}

// Save stores a conversation in the repository.
func (r *ConversationRepository) Save(_ context.Context, conversation *core.Conversation) error {
	if conversation == nil {
		return core.ErrConversationNotFound
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.conversations[conversation.ID] = conversation
	return nil
}

// FindByID retrieves a conversation by its ID.
func (r *ConversationRepository) FindByID(_ context.Context, id string) (*core.Conversation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	conversation, exists := r.conversations[id]
	if !exists {
		return nil, core.ErrConversationNotFound
	}

	return conversation, nil
}

// FindOrCreate retrieves a conversation by ID or creates a new one if it doesn't exist.
func (r *ConversationRepository) FindOrCreate(_ context.Context, id string) (*core.Conversation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if conversation, exists := r.conversations[id]; exists {
		return conversation, nil
	}

	conversation := core.NewConversation(id)
	r.conversations[id] = conversation
	return conversation, nil
}
