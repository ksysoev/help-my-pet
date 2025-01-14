package anthropic

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/tmc/langchaingo/llms"
)

type MockLLMAdapter struct {
	mock.Mock
}

func NewMockLLMAdapter(t mock.TestingT) *MockLLMAdapter {
	mock := &MockLLMAdapter{}
	mock.Test(t)
	return mock
}

func (m *MockLLMAdapter) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	args := m.Called(ctx, prompt, options)
	return args.String(0), args.Error(1)
}
