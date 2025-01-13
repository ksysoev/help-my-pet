package memory

import (
	"context"
	"testing"

	"github.com/ksysoev/help-my-pet/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestNewConversationRepository(t *testing.T) {
	repo := NewConversationRepository()
	assert.NotNil(t, repo)
	assert.NotNil(t, repo.conversations)
}

func TestConversationRepository_Save(t *testing.T) {
	tests := []struct {
		conversation *core.Conversation
		name         string
		wantErr      bool
	}{
		{
			name:         "save valid conversation",
			conversation: core.NewConversation("test-id"),
			wantErr:      false,
		},
		{
			name:         "save nil conversation",
			conversation: nil,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewConversationRepository()
			err := repo.Save(context.Background(), tt.conversation)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, core.ErrConversationNotFound, err)
			} else {
				assert.NoError(t, err)
				saved, err := repo.FindByID(context.Background(), tt.conversation.ID)
				assert.NoError(t, err)
				assert.Equal(t, tt.conversation, saved)
			}
		})
	}
}

func TestConversationRepository_FindOrCreate(t *testing.T) {
	tests := []struct {
		name      string
		setupRepo func(*ConversationRepository)
		id        string
	}{
		{
			name: "find existing conversation",
			setupRepo: func(r *ConversationRepository) {
				conv := core.NewConversation("test-id")
				r.conversations[conv.ID] = conv
			},
			id: "test-id",
		},
		{
			name:      "create new conversation",
			setupRepo: func(r *ConversationRepository) {},
			id:        "new-id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewConversationRepository()
			tt.setupRepo(repo)

			conv, err := repo.FindOrCreate(context.Background(), tt.id)
			assert.NoError(t, err)
			assert.NotNil(t, conv)
			assert.Equal(t, tt.id, conv.ID)

			// Verify the conversation is stored
			stored, err := repo.FindByID(context.Background(), tt.id)
			assert.NoError(t, err)
			assert.Equal(t, conv, stored)
		})
	}
}

func TestConversationRepository_FindByID(t *testing.T) {
	tests := []struct {
		expectedError error
		setupRepo     func(*ConversationRepository)
		name          string
		id            string
		wantErr       bool
	}{
		{
			name: "find existing conversation",
			setupRepo: func(r *ConversationRepository) {
				conv := core.NewConversation("test-id")
				r.conversations[conv.ID] = conv
			},
			id:      "test-id",
			wantErr: false,
		},
		{
			name:          "find non-existing conversation",
			setupRepo:     func(r *ConversationRepository) {},
			id:            "non-existing-id",
			wantErr:       true,
			expectedError: core.ErrConversationNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewConversationRepository()
			tt.setupRepo(repo)

			conv, err := repo.FindByID(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, conv)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, conv)
				assert.Equal(t, tt.id, conv.ID)
			}
		})
	}
}
