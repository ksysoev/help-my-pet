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

// NewConversationRepository creates a new Redis-backed conversation repository
func NewConversationRepository(client *redis.Client) *ConversationRepository {
	return &ConversationRepository{
		client: client,
	}
}

// Save stores a conversation in Redis with TTL
func (r *ConversationRepository) Save(ctx context.Context, conversation *conversation.Conversation) error {
	data, err := json.Marshal(conversation)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, r.key(conversation.ID), data, ConversationTTL).Err()
}

// FindByID retrieves a conversation by its ID
func (r *ConversationRepository) FindByID(ctx context.Context, id string) (*conversation.Conversation, error) {
	data, err := r.client.Get(ctx, r.key(id)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, core.ErrConversationNotFound
		}
		return nil, err
	}

	var tmpConv struct {
		ID            string
		State         conversation.ConversationState
		Messages      []conversation.Message
		Questionnaire json.RawMessage `json:"questionnaire"`
	}

	if err := json.Unmarshal(data, &tmpConv); err != nil {
		return nil, err
	}

	switch tmpConv.State {
	case conversation.StateNormal:
		return &conversation.Conversation{
			ID:       tmpConv.ID,
			State:    conversation.StateNormal,
			Messages: tmpConv.Messages,
		}, nil
	case conversation.StatePetProfileQuestioning:
		var q conversation.PetProfileStateImpl
		if err := json.Unmarshal(tmpConv.Questionnaire, &q); err != nil {
			return nil, fmt.Errorf("failed to unmarshal pet profile questionnaire: %w", err)
		}

		return &conversation.Conversation{
			ID:            tmpConv.ID,
			State:         conversation.StatePetProfileQuestioning,
			Messages:      tmpConv.Messages,
			Questionnaire: q,
		}, nil
	case conversation.StateFollowUpQuestioning:
		var q conversation.FollowUpQuestionnaireState
		if err := json.Unmarshal(tmpConv.Questionnaire, &q); err != nil {
			return nil, fmt.Errorf("failed to unmarshal follow-up questionnaire: %w", err)
		}

		return &conversation.Conversation{
			ID:            tmpConv.ID,
			State:         conversation.StateFollowUpQuestioning,
			Messages:      tmpConv.Messages,
			Questionnaire: q,
		}, nil
	default:
		return nil, fmt.Errorf("unknown conversation state: %s", tmpConv.State)
	}
}

// FindOrCreate retrieves a conversation by ID or creates a new one if it doesn't exist
func (r *ConversationRepository) FindOrCreate(ctx context.Context, id string) (*conversation.Conversation, error) {
	conv, err := r.FindByID(ctx, id)
	if err == core.ErrConversationNotFound {
		conv = conversation.NewConversation(id)
		if err := r.Save(ctx, conv); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return conv, nil
}

// key generates a Redis key for a conversation ID
func (r *ConversationRepository) key(id string) string {
	return "conversation:" + id
}
