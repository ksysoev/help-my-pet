package core

import (
	"time"
)

// Conversation represents a chat conversation with its context and messages.
type Conversation struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	ID        string
	Messages  []Message
}

// Message represents a single message in a conversation.
type Message struct {
	Timestamp time.Time
	Role      string
	Content   string
}

// NewConversation creates a new conversation with the given ID.
func NewConversation(id string) *Conversation {
	now := time.Now()
	return &Conversation{
		ID:        id,
		Messages:  make([]Message, 0),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// AddMessage adds a new message to the conversation.
func (c *Conversation) AddMessage(role, content string) {
	c.Messages = append(c.Messages, Message{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	})
	c.UpdatedAt = time.Now()
}

// GetContext returns all messages in the conversation as context.
func (c *Conversation) GetContext() []Message {
	return c.Messages
}
