package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ksysoev/help-my-pet/pkg/core/conversation"
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

// NewConversationRepository creates a new instance of ConversationRepository with the given Redis client.
// It requires a valid redis.Client to interact with Redis as the underlying storage.
// Returns a pointer to the initialized ConversationRepository.
func NewConversationRepository(client *redis.Client) *ConversationRepository {
	return &ConversationRepository{
		client: client,
	}
}

// Save serializes the given conversation and saves it to Redis under a key derived from its ID.
// It overwrites any existing data for the same key and sets a time-to-live based on ConversationTTL.
// Returns an error if serialization fails or if the Redis operation encounters an issue.
func (r *ConversationRepository) Save(ctx context.Context, conversation core.Conversation) error {
	data, err := json.Marshal(conversation)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, r.key(conversation.GetID()), data, ConversationTTL).Err()
}

// FindByID retrieves a conversation from Redis by its ID.
// It fetches the serialized conversation from Redis, unmarshals it, and returns the Conversation object.
// Accepts ctx for request-scoped context and id representing the unique conversation identifier.
// Returns the Conversation object if found or an error if the conversation is not found or unmarshaling fails.
func (r *ConversationRepository) FindByID(ctx context.Context, id string) (core.Conversation, error) {
	data, err := r.client.Get(ctx, r.key(id)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, core.ErrConversationNotFound
		}
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	conv, err := conversation.Unmarshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal conversation with id %s: %w", id, err)
	}

	return conv, nil
}

// FindOrCreate retrieves a conversation by its ID, or creates a new one if it is not found.
// It fetches the conversation from Redis, initializes a new conversation if it doesn't exist,
// and returns any errors encountered during retrieval or creation.
// ctx: request-scoped context for handling deadlines and cancellations
// id: unique identifier for the conversation
// Returns the conversation object and an error if there's an issue retrieving the conversation or creating a new one.
func (r *ConversationRepository) FindOrCreate(ctx context.Context, id string) (core.Conversation, error) {
	conv, err := r.FindByID(ctx, id)
	if err == core.ErrConversationNotFound {
		conv = conversation.NewConversation(id)
	} else if err != nil {
		return nil, err
	}

	return conv, nil
}

// Delete removes the conversation with the specified ID from Redis storage.
// It deletes the key derived from the conversation ID and returns an error if the Redis operation fails.
// ctx is the request-scoped context to manage cancellation and timeouts.
// id is the unique identifier of the conversation to delete.
// Returns an error if the Redis deletion operation fails.
func (r *ConversationRepository) Delete(ctx context.Context, id string) error {
	return r.client.Del(ctx, r.key(id)).Err()
}

// key generates a Redis key for a conversation by prefixing the provided conversation ID with "conversation:".
// Accepts id as the unique identifier for the conversation.
// Returns the fully constructed Redis key as a string.
func (r *ConversationRepository) key(id string) string {
	return "conversation:" + id
}
