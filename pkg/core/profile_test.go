package core

import (
	"context"
	"testing"
	"time"

	"github.com/ksysoev/help-my-pet/pkg/core/conversation"
	"github.com/ksysoev/help-my-pet/pkg/core/message"
	"github.com/ksysoev/help-my-pet/pkg/core/pet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProcessEditProfile(t *testing.T) {
	tests := []struct {
		name          string
		request       *message.UserMessage
		setupMocks    func(*MockConversationRepository)
		expectedText  string
		expectedError string
	}{
		{
			name: "successful profile creation",
			request: &message.UserMessage{
				ChatID: "123",
				Text:   "edit profile",
			},
			setupMocks: func(repo *MockConversationRepository) {
				conv := conversation.NewConversation("123")
				conv.State = conversation.StateNormal
				repo.EXPECT().FindOrCreate(mock.Anything, "123").Return(conv, nil)
				repo.EXPECT().Save(mock.Anything, mock.MatchedBy(func(c *conversation.Conversation) bool {
					return c.State == conversation.StatePetProfileQuestioning
				})).Return(nil)
			},
			expectedText: "What is your pet's name?", // First question from PetProfileQuestionnaire
		},
		{
			name: "repository error",
			request: &message.UserMessage{
				ChatID: "123",
				Text:   "edit profile",
			},
			setupMocks: func(repo *MockConversationRepository) {
				repo.EXPECT().FindOrCreate(mock.Anything, "123").Return(nil, assert.AnError)
			},
			expectedError: "failed to get conversation: assert.AnError general error for testing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockConversationRepository(t)
			if tt.setupMocks != nil {
				tt.setupMocks(mockRepo)
			}

			service := &AIService{
				repo: mockRepo,
			}

			response, err := service.ProcessEditProfile(context.Background(), tt.request)

			if tt.expectedError != "" {
				assert.EqualError(t, err, tt.expectedError)
				assert.Nil(t, response)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, tt.expectedText, response.Message)
		})
	}
}

func TestProcessProfileAnswer(t *testing.T) {
	tests := []struct {
		name          string
		request       *message.UserMessage
		conv          *conversation.Conversation
		setupMocks    func(*MockConversationRepository, *MockPetProfileRepository)
		expectedText  string
		expectedError string
	}{
		{
			name: "successful answer and complete profile",
			request: &message.UserMessage{
				ChatID: "123",
				UserID: "user1",
				Text:   "Homemade food, no allergies", // The last answer (weight) that completes the profile
			},
			conv: func() *conversation.Conversation {
				conv := conversation.NewConversation("123")
				if err := conv.StartProfileQuestions(context.Background()); err != nil {
					panic(err)
				}

				// Fill in previous answers to simulate near completion
				_, _ = conv.AddQuestionAnswer("Rex")        // name
				_, _ = conv.AddQuestionAnswer("Dog")        // species
				_, _ = conv.AddQuestionAnswer("Labrador")   // breed
				_, _ = conv.AddQuestionAnswer("2020-01-01") // dob
				_, _ = conv.AddQuestionAnswer("Male")       // gender
				_, _ = conv.AddQuestionAnswer("10 kg")      // weight
				_, _ = conv.AddQuestionAnswer("yes")        // neutered
				_, _ = conv.AddQuestionAnswer("high")       // activity level
				_, _ = conv.AddQuestionAnswer("None")       // diet
				// The last answer is the one that will be provided in the test
				return conv
			}(),
			setupMocks: func(repo *MockConversationRepository, profileRepo *MockPetProfileRepository) {
				repo.EXPECT().Save(mock.Anything, mock.Anything).Return(nil)

				profileRepo.EXPECT().SaveProfile(mock.Anything, "user1", mock.MatchedBy(func(p *pet.Profile) bool {
					return p.Name == "Rex" && p.Species == "Dog" && p.Breed == "Labrador" &&
						p.DateOfBirth == "2020-01-01" && p.Gender == "Male" && p.Weight == "10 kg" &&
						p.Neutered == "yes" && p.Activity == "high" && p.FoodPreferences == "Homemade food, no allergies" &&
						p.ChronicDiseases == "None"
				})).Return(nil)
			},
			expectedText: "Pet profile saved successfully",
		},
		{
			name: "continue questioning",
			request: &message.UserMessage{
				ChatID: "123",
				Text:   "Rex", // First answer (name)
			},
			conv: func() *conversation.Conversation {
				conv := conversation.NewConversation("123")
				if err := conv.StartProfileQuestions(context.Background()); err != nil {
					panic(err)
				}
				return conv
			}(),
			setupMocks: func(repo *MockConversationRepository, profileRepo *MockPetProfileRepository) {
				// Expect save after adding the answer
				repo.EXPECT().Save(mock.Anything, mock.MatchedBy(func(conv *conversation.Conversation) bool {
					return conv.State == conversation.StatePetProfileQuestioning
				})).Return(nil)
			},
			expectedText: "What type of pet do you have?", // Second question after name
		},
		{
			name: "answer too long",
			request: &message.UserMessage{
				ChatID: "123",
				Text: "This is a very long answer that is longer than the maximum allowed length. " +
					"This is a very long answer that is longer than the maximum allowed length. " +
					"This is a very long answer that is longer than the maximum allowed length. ",
			},
			conv: func() *conversation.Conversation {
				conv := conversation.NewConversation("123")
				if err := conv.StartProfileQuestions(context.Background()); err != nil {
					panic(err)
				}
				return conv

			}(),
			expectedText: "I apologize, but your message is too long for me to process. Please try to make it shorter and more concise.",
		},
		{
			name: "answer with future date of birth",
			request: &message.UserMessage{
				ChatID: "123",
				Text:   time.Now().Add(48 * time.Hour).Format("2006-01-02"),
			},
			conv: func() *conversation.Conversation {
				conv := conversation.NewConversation("123")
				if err := conv.StartProfileQuestions(context.Background()); err != nil {
					panic(err)
				}

				// Fill in previous answers to go to the date of birth question
				_, _ = conv.AddQuestionAnswer("Rex")      // name
				_, _ = conv.AddQuestionAnswer("Dog")      // species
				_, _ = conv.AddQuestionAnswer("Labrador") // breed

				return conv
			}(),
			expectedText: "Provided date cannot be in the future. Please provide a valid date.",
		},
		{
			name: "answer with invalid date of birth",
			request: &message.UserMessage{
				ChatID: "123",
				Text:   "2020-13-01",
			},
			conv: func() *conversation.Conversation {
				conv := conversation.NewConversation("123")
				if err := conv.StartProfileQuestions(context.Background()); err != nil {
					panic(err)
				}

				// Fill in previous answers to go to the date of birth question
				_, _ = conv.AddQuestionAnswer("Rex")      // name
				_, _ = conv.AddQuestionAnswer("Dog")      // species
				_, _ = conv.AddQuestionAnswer("Labrador") // breed

				return conv
			}(),
			expectedText: "Please provide a date in the valid format YYYY-MM-DD (e.g., 2023-12-31)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockConversationRepository(t)
			mockProfileRepo := NewMockPetProfileRepository(t)

			if tt.setupMocks != nil {
				tt.setupMocks(mockRepo, mockProfileRepo)
			}

			service := &AIService{
				repo:        mockRepo,
				profileRepo: mockProfileRepo,
			}

			response, err := service.ProcessProfileAnswer(context.Background(), tt.conv, tt.request)

			if tt.expectedError != "" {
				assert.EqualError(t, err, tt.expectedError)
				assert.Nil(t, response)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, tt.expectedText, response.Message)
		})
	}
}
