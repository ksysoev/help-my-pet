package core

import (
	"context"
	"testing"
)

func TestAIService_GetPetAdvice(t *testing.T) {
	// Skip if no API key is provided
	t.Skip("Skipping test as it requires Anthropic API key")

	svc := NewAIService("test-key", "claude-2")
	ctx := context.Background()

	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{
			name:    "basic question",
			query:   "What food is good for cats?",
			wantErr: false,
		},
		{
			name:    "empty question",
			query:   "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.GetPetAdvice(ctx, tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPetAdvice() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
