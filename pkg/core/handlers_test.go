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

func TestResetUserConversation(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*MockPetProfileRepository, *MockConversationRepository)
		expectedError error
	}{
		{
			name: "successful reset",
			setupMocks: func(profileMock *MockPetProfileRepository, repoMock *MockConversationRepository) {
				profileMock.EXPECT().RemoveUserProfiles(mock.Anything, "user123").Return(nil)
				repoMock.EXPECT().Delete(mock.Anything, "chat123").Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "remove user profiles fails",
			setupMocks: func(profileMock *MockPetProfileRepository, _ *MockConversationRepository) {
				profileMock.EXPECT().RemoveUserProfiles(mock.Anything, "user123").Return(assert.AnError)
			},
			expectedError: fmt.Errorf("failed to remove user profiles: %w", assert.AnError),
		},
		{
			name: "delete conversation fails",
			setupMocks: func(profileMock *MockPetProfileRepository, repoMock *MockConversationRepository) {
				profileMock.EXPECT().RemoveUserProfiles(mock.Anything, "user123").Return(nil)
				repoMock.EXPECT().Delete(mock.Anything, "chat123").Return(assert.AnError)
			},
			expectedError: fmt.Errorf("failed to remove conversation: %w", assert.AnError),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx := context.Background()
			profileMock := NewMockPetProfileRepository(t)
			repoMock := NewMockConversationRepository(t)
			tt.setupMocks(profileMock, repoMock)

			service := &AIService{
				profileRepo: profileMock,
				repo:        repoMock,
			}

			// Act
			err := service.ResetUserConversation(ctx, "user123", "chat123")

			// Assert
			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			profileMock.AssertExpectations(t)
			repoMock.AssertExpectations(t)
		})
	}
}
