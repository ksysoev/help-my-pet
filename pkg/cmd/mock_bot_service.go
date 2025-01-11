package cmd

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockBotService struct {
	mock.Mock
}

func (m *MockBotService) Run(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func NewMockBotService() *MockBotService {
	return &MockBotService{}
}
