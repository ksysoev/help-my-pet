package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewConversation(t *testing.T) {
	id := "test-id"
	conv := NewConversation(id)

	assert.Equal(t, id, conv.ID)
	assert.Empty(t, conv.Messages)
	assert.NotZero(t, conv.CreatedAt)
	assert.Equal(t, conv.CreatedAt, conv.UpdatedAt)
}

func TestConversation_AddMessage(t *testing.T) {
	tests := []struct {
		name    string
		role    string
		content string
	}{
		{
			name:    "add user message",
			role:    "user",
			content: "Hello",
		},
		{
			name:    "add assistant message",
			role:    "assistant",
			content: "Hi there!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv := NewConversation("test-id")
			beforeAdd := time.Now()
			time.Sleep(time.Millisecond) // Ensure time difference

			conv.AddMessage(tt.role, tt.content)

			assert.Len(t, conv.Messages, 1)
			msg := conv.Messages[0]
			assert.Equal(t, tt.role, msg.Role)
			assert.Equal(t, tt.content, msg.Content)
			assert.NotZero(t, msg.Timestamp)
			assert.True(t, msg.Timestamp.After(beforeAdd))
			assert.True(t, conv.UpdatedAt.After(conv.CreatedAt))
		})
	}
}

func TestConversation_GetContext(t *testing.T) {
	conv := NewConversation("test-id")
	messages := []struct {
		role    string
		content string
	}{
		{"user", "Hello"},
		{"assistant", "Hi there!"},
		{"user", "How are you?"},
	}

	for _, msg := range messages {
		conv.AddMessage(msg.role, msg.content)
	}

	context := conv.GetContext()
	assert.Len(t, context, len(messages))

	for i, msg := range messages {
		assert.Equal(t, msg.role, context[i].Role)
		assert.Equal(t, msg.content, context[i].Content)
	}
}
