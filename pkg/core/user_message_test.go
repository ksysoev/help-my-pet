package core

import (
	"testing"
)

func TestNewUserMessage(t *testing.T) {
	tests := []struct {
		name    string
		userID  string
		chatID  string
		text    string
		wantErr bool
	}{
		{
			name:    "valid input",
			userID:  "user123",
			chatID:  "chat456",
			text:    "Hello, world!",
			wantErr: false,
		},
		{
			name:    "empty userID",
			userID:  "",
			chatID:  "chat456",
			text:    "Hello, world!",
			wantErr: true,
		},
		{
			name:    "empty chatID",
			userID:  "user123",
			chatID:  "",
			text:    "Hello, world!",
			wantErr: true,
		},
		{
			name:    "empty text",
			userID:  "user123",
			chatID:  "chat456",
			text:    "",
			wantErr: true,
		},
		{
			name:    "invalid userID and chatID",
			userID:  "",
			chatID:  "",
			text:    "Hello, world!",
			wantErr: true,
		},
		{
			name:    "text too long",
			userID:  "user123",
			chatID:  "chat456",
			text:    string(make([]rune, 10001)),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewUserMessage(tt.userID, tt.chatID, tt.text)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewUserMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
