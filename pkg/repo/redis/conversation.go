package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"

	"github.com/ksysoev/help-my-pet/pkg/core"
)

const (
	conversationPrefix = "conversation:"
)

// ConversationRepository implements core.ConversationRepository interface
// using Redis as storage.
type ConversationRepository struct {
	client *redis.Client
}

// NewConversationRepository creates a new instance of ConversationRepository.
func NewConversationRepository(client *redis.Client) *ConversationRepository {
	return &ConversationRepository{
		client: client,
	}
}

// Save stores a conversation in Redis.
func (r *ConversationRepository) Save(ctx context.Context, conversation *core.Conversation) error {
	if conversation == nil {
		return core.ErrConversationNotFound
	}

	data, err := json.Marshal(conversation)
	if err != nil {
		return fmt.Errorf("failed to marshal conversation: %w", err)
	}

	key := r.getKey(conversation.ID)
	if err := r.client.Set(ctx, key, data, 0).Err(); err != nil {
		return fmt.Errorf("failed to save conversation: %w", err)
	}

	return nil
}

// FindByID retrieves a conversation by its ID from Redis.
func (r *ConversationRepository) FindByID(ctx context.Context, id string) (*core.Conversation, error) {
	key := r.getKey(id)
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, core.ErrConversationNotFound
		}
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	var conversation core.Conversation
	if err := json.Unmarshal(data, &conversation); err != nil {
		return nil, fmt.Errorf("failed to unmarshal conversation: %w", err)
	}

	return &conversation, nil
}

// FindOrCreate retrieves a conversation by ID or creates a new one if it doesn't exist.
func (r *ConversationRepository) FindOrCreate(ctx context.Context, id string) (*core.Conversation, error) {
	conversation, err := r.FindByID(ctx, id)
	if err != nil && err != core.ErrConversationNotFound {
		return nil, err
	}

	if conversation != nil {
		return conversation, nil
	}

	conversation = core.NewConversation(id)
	if err := r.Save(ctx, conversation); err != nil {
		return nil, fmt.Errorf("failed to save new conversation: %w", err)
	}

	return conversation, nil
}

func (r *ConversationRepository) getKey(id string) string {
	return conversationPrefix + id
}
