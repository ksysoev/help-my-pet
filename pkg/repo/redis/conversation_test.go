package redis

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ksysoev/help-my-pet/pkg/core"
)

func TestConversationRepository_Save(t *testing.T) {
	db, mock := redismock.NewClientMock()
	repo := NewConversationRepository(db)
	ctx := context.Background()

	t.Run("save valid conversation", func(t *testing.T) {
		conv := core.NewConversation("test-id")
		conv.AddMessage("user", "hello")

		data, err := json.Marshal(conv)
		require.NoError(t, err)

		mock.ExpectSet("conversation:test-id", data, 0).SetVal("OK")

		err = repo.Save(ctx, conv)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("save nil conversation", func(t *testing.T) {
		err := repo.Save(ctx, nil)
		assert.ErrorIs(t, err, core.ErrConversationNotFound)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestConversationRepository_FindByID(t *testing.T) {
	db, mock := redismock.NewClientMock()
	repo := NewConversationRepository(db)
	ctx := context.Background()

	t.Run("find existing conversation", func(t *testing.T) {
		conv := core.NewConversation("test-id")
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
}

func TestConversationRepository_FindOrCreate(t *testing.T) {
	db, mock := redismock.NewClientMock()
	repo := NewConversationRepository(db)
	ctx := context.Background()

	t.Run("find existing conversation", func(t *testing.T) {
		conv := core.NewConversation("test-id")
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
}

func TestConversationRepository_ComplexConversation(t *testing.T) {
	db, mock := redismock.NewClientMock()
	repo := NewConversationRepository(db)
	ctx := context.Background()

	// Create a conversation with questionnaire state
	conv := core.NewConversation("test-id")
	conv.AddMessage("user", "hello")
	conv.StartQuestionnaire("initial prompt", []core.Question{
		{Text: "question 1"},
		{Text: "question 2"},
	})

	data, err := json.Marshal(conv)
	require.NoError(t, err)

	mock.ExpectSet("conversation:test-id", data, 0).SetVal("OK")
	require.NoError(t, repo.Save(ctx, conv))

	mock.ExpectGet("conversation:test-id").SetVal(string(data))
	found, err := repo.FindByID(ctx, conv.ID)
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, conv.ID, found.ID)
	assert.Equal(t, conv.State, found.State)
	assert.Equal(t, conv.Messages[0].Content, found.Messages[0].Content)
	assert.Equal(t, conv.Questionnaire.InitialPrompt, found.Questionnaire.InitialPrompt)
	assert.Equal(t, len(conv.Questionnaire.Questions), len(found.Questionnaire.Questions))
	assert.Equal(t, conv.Questionnaire.Questions[0].Text, found.Questionnaire.Questions[0].Text)
	assert.NoError(t, mock.ExpectationsWereMet())
}
