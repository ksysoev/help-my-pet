package core

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCancelQuestionnaire(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*MockConversationRepository, *MockConversation)
		expectedError error
	}{
		{
			name: "successful cancellation",
			setupMocks: func(repoMock *MockConversationRepository, convMock *MockConversation) {
				repoMock.EXPECT().FindOrCreate(mock.Anything, "chat123").Return(convMock, nil)
				convMock.EXPECT().CancelQuestionnaire().Return()
				repoMock.EXPECT().Save(mock.Anything, convMock).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "find or create fails",
			setupMocks: func(repoMock *MockConversationRepository, _ *MockConversation) {
				repoMock.EXPECT().FindOrCreate(mock.Anything, "chat123").Return(nil, assert.AnError)
			},
			expectedError: fmt.Errorf("failed to get conversation: %w", assert.AnError),
		},
		{
			name: "save conversation fails",
			setupMocks: func(repoMock *MockConversationRepository, convMock *MockConversation) {
				repoMock.EXPECT().FindOrCreate(mock.Anything, "chat123").Return(convMock, nil)
				convMock.EXPECT().CancelQuestionnaire().Return()
				repoMock.EXPECT().Save(mock.Anything, convMock).Return(assert.AnError)
			},
			expectedError: fmt.Errorf("failed to save conversation: %w", assert.AnError),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx := context.Background()
			repoMock := NewMockConversationRepository(t)
			convMock := NewMockConversation(t)
			tt.setupMocks(repoMock, convMock)

			service := &AIService{
				repo: repoMock,
			}

			// Act
			err := service.CancelQuestionnaire(ctx, "chat123")

			// Assert
			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			repoMock.AssertExpectations(t)
			convMock.AssertExpectations(t)
		})
	}
}
