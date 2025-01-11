package bot

import (
	"context"
	"testing"
)

type mockAIProvider struct {
	response string
	err      error
}

func (m *mockAIProvider) GetPetAdvice(ctx context.Context, question string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.response, nil
}

func TestNewService(t *testing.T) {
	t.Skip("Skipping test as it requires Telegram token")

	mockAI := &mockAIProvider{
		response: "Test response",
		err:      nil,
	}

	svc := NewService("test-token", mockAI)
	if svc == nil {
		t.Error("NewService() returned nil")
	}
}
