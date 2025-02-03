package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/go-redis/redismock/v9"
	"github.com/ksysoev/help-my-pet/pkg/core/conversation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ksysoev/help-my-pet/pkg/core"
)

func TestConversationRepository_Save(t *testing.T) {
	db, mock := redismock.NewClientMock()
	repo := NewConversationRepository(db)
	ctx := context.Background()

	t.Run("save valid conversation", func(t *testing.T) {
		conv := conversation.NewConversation("test-id")
		conv.AddMessage("user", "hello")

		data, err := json.Marshal(conv)
		require.NoError(t, err)

		mock.ExpectSet("conversation:test-id", data, ConversationTTL).SetVal("OK")

		err = repo.Save(ctx, conv)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestConversationRepository_FindByID(t *testing.T) {
	db, mock := redismock.NewClientMock()
	repo := NewConversationRepository(db)
	ctx := context.Background()

	t.Run("find existing conversation", func(t *testing.T) {
		conv := conversation.NewConversation("test-id")
		conv.AddMessage("user", "hello")

		data, err := json.Marshal(conv)
		require.NoError(t, err)

		mock.ExpectGet("conversation:test-id").SetVal(string(data))

		found, err := repo.FindByID(ctx, conv.ID)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, conv.ID, found.ID)
		assert.Equal(t, len(conv.Messages), len(found.Messages))
		assert.Equal(t, conv.Messages[0].Content, found.Messages[0].Content)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("find non-existing conversation", func(t *testing.T) {
		mock.ExpectGet("conversation:non-existing").RedisNil()

		found, err := repo.FindByID(ctx, "non-existing")
		assert.ErrorIs(t, err, core.ErrConversationNotFound)
		assert.Nil(t, found)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Get invalid JSON", func(t *testing.T) {
		mock.ExpectGet("conversation:invalid-json").SetVal("invalid json")

		found, err := repo.FindByID(ctx, "invalid-json")
		assert.Error(t, err)
		assert.Nil(t, found)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestConversationRepository_ComplexConversation(t *testing.T) {
	db, mock := redismock.NewClientMock()
	repo := NewConversationRepository(db)
	ctx := context.Background()

	// Create a conversation with questionnaire state
	conv := conversation.NewConversation("test-id")
	conv.AddMessage("user", "hello")

	err := conv.StartFollowUpQuestions("initial prompt", []conversation.Question{
		{Text: "question 1"},
		{Text: "question 2"},
	})

	require.NoError(t, err)

	data, err := json.Marshal(conv)
	require.NoError(t, err)

	mock.ExpectSet("conversation:test-id", data, ConversationTTL).SetVal("OK")
	require.NoError(t, repo.Save(ctx, conv))

	mock.ExpectGet("conversation:test-id").SetVal(string(data))
	found, err := repo.FindByID(ctx, conv.ID)
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, conv.ID, found.ID)
	assert.Equal(t, conv.State, found.State)
	assert.Equal(t, conv.Messages[0].Content, found.Messages[0].Content)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationRepository_FindOrCreate(t *testing.T) {
	db, mock := redismock.NewClientMock()
	repo := NewConversationRepository(db)
	ctx := context.Background()

	t.Run("find existing conversation", func(t *testing.T) {
		conv := conversation.NewConversation("test-id")
		conv.AddMessage("user", "hello")

		data, err := json.Marshal(conv)
		require.NoError(t, err)

		mock.ExpectGet("conversation:test-id").SetVal(string(data))

		found, err := repo.FindOrCreate(ctx, conv.ID)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, conv.ID, found.ID)
		assert.Equal(t, len(conv.Messages), len(found.Messages))
		assert.Equal(t, conv.Messages[0].Content, found.Messages[0].Content)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("create new conversation when not found", func(t *testing.T) {
		mock.ExpectGet("conversation:new-id").RedisNil()

		found, err := repo.FindOrCreate(ctx, "new-id")
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, "new-id", found.ID)
		assert.Empty(t, found.Messages)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error on FindByID", func(t *testing.T) {
		mock.ExpectGet("conversation:error-id").SetErr(fmt.Errorf("redis error"))

		found, err := repo.FindOrCreate(ctx, "error-id")
		assert.Error(t, err)
		assert.Nil(t, found)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
