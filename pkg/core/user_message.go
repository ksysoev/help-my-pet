package core

import (
	"errors"
	"unicode/utf8"
)

var (
	// ErrEmptyUserID is returned when the user ID is empty
	ErrEmptyUserID = errors.New("empty user ID")
	// ErrEmptyChatID is returned when the chat ID is empty
	ErrEmptyChatID = errors.New("empty chat ID")
	// ErrEmptyText is returned when the text is empty
	ErrEmptyText = errors.New("empty text")
	// ErrTextTooLong is returned when the text is too long
	ErrTextTooLong = errors.New("text is too long")
)

const MaxTextLength = 2000

// UserMessage represents a message sent by a user in a specific chat context.
// It includes the ID of the user, the ID of the chat, and the content of the message.
type UserMessage struct {
	UserID string
	ChatID string
	Text   string
}

// NewUserMessage creates a new UserMessage instance after validating its fields.
// It returns an error if any field is empty or if the text exceeds the maximum allowed length.
func NewUserMessage(userID, chatID, text string) (*UserMessage, error) {
	m := &UserMessage{
		UserID: userID,
		ChatID: chatID,
		Text:   text,
	}

	if err := m.validate(); err != nil {
		return nil, err
	}

	return m, nil
}

// validate checks the validity of a UserMessage instance.
// It ensures that UserID, ChatID, and Text fields are non-empty and that Text does not exceed MaxTextLength.
// Returns an error if any of the fields are invalid.
func (m UserMessage) validate() error {
	if m.UserID == "" {
		return ErrEmptyUserID
	}
	if m.ChatID == "" {
		return ErrEmptyChatID
	}

	if m.Text == "" {
		return ErrEmptyText
	}

	if utf8.RuneCountInString(m.Text) > MaxTextLength {
		return ErrTextTooLong
	}

	return nil
}
