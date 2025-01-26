package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/ksysoev/help-my-pet/pkg/core"
)

const (
	// ConversationTTL defines how long conversations are stored (1 week)
	ConversationTTL = 7 * 24 * time.Hour
)

// ConversationRepository implements core.ConversationRepository using Redis
type ConversationRepository struct {
	client *redis.Client
}

// NewConversationRepository creates a new Redis-backed conversation repository
func NewConversationRepository(client *redis.Client) *ConversationRepository {
	return &ConversationRepository{
		client: client,
	}
}

// Save stores a conversation in Redis with TTL
func (r *ConversationRepository) Save(ctx context.Context, conversation *core.Conversation) error {
	data, err := json.Marshal(conversation)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, r.key(conversation.ID), data, ConversationTTL).Err()
}

// FindByID retrieves a conversation by its ID
func (r *ConversationRepository) FindByID(ctx context.Context, id string) (*core.Conversation, error) {
	data, err := r.client.Get(ctx, r.key(id)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, core.ErrConversationNotFound
		}
		return nil, err
	}

	var conversation core.Conversation
	if err := json.Unmarshal(data, &conversation); err != nil {
		return nil, err
	}

	return &conversation, nil
}

// FindOrCreate retrieves a conversation by ID or creates a new one if it doesn't exist
func (r *ConversationRepository) FindOrCreate(ctx context.Context, id string) (*core.Conversation, error) {
	conversation, err := r.FindByID(ctx, id)
	if err == core.ErrConversationNotFound {
		conversation = core.NewConversation(id)
		if err := r.Save(ctx, conversation); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	return conversation, nil
}

// key generates a Redis key for a conversation ID
func (r *ConversationRepository) key(id string) string {
	return "conversation:" + id
}
