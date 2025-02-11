package core

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/ksysoev/help-my-pet/pkg/core/conversation"
	"github.com/ksysoev/help-my-pet/pkg/core/message"
	"github.com/ksysoev/help-my-pet/pkg/core/pet"
	"github.com/ksysoev/help-my-pet/pkg/i18n"
)

// ProcessEditProfile initiates a pet profile questionnaire for a user in a conversation context.
// It retrieves or creates a conversation, starts the questionnaire, and fetches the first question.
// Returns the first question with possible answers or an error if any retrieval, initialization, or save operation fails.
func (s *AIService) ProcessEditProfile(ctx context.Context, request *message.UserMessage) (*message.Response, error) {
	slog.DebugContext(ctx, "managing pet profile", "input", request.Text)

	conv, err := s.repo.FindOrCreate(ctx, request.ChatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	// Start pet profile questionnaire
	if err := conv.StartProfileQuestions(ctx); err != nil {
		return nil, fmt.Errorf("failed to start profile questions: %w", err)
	}

	// Get the first question
	question, err := conv.GetCurrentQuestion()
	if err != nil {
		return nil, fmt.Errorf("failed to get first question: %w", err)
	}

	// Save conv state
	if err := s.repo.Save(ctx, conv); err != nil {
		return nil, fmt.Errorf("failed to save conversation: %w", err)
	}

	return message.NewResponse(question.Text, question.Answers), nil
}

// ProcessProfileAnswer processes a user's response to a pet profile question in an ongoing conversation.
// It adds the response to the current question, updates conversation state, and determines the next step.
// If the questionnaire is complete, it finalizes the profile process; otherwise, it retrieves the next question.
// Returns the next question or a success response upon completion, and an error if any operation fails during processing.
func (s *AIService) ProcessProfileAnswer(ctx context.Context, conv Conversation, request *message.UserMessage) (*message.Response, error) {
	slog.DebugContext(ctx, "managing pet profile", "input", request.Text)

	// Add answer to the current question
	isComplete, err := conv.AddQuestionAnswer(request.Text)
	switch {
	case errors.Is(err, message.ErrTextTooLong):
		return message.NewResponse(i18n.GetLocale(ctx).Sprintf("I apologize, but your message is too long for me to process. Please try to make it shorter and more concise."), nil), nil
	case errors.Is(err, message.ErrFutureDate):
		return message.NewResponse(i18n.GetLocale(ctx).Sprintf("Provided date cannot be in the future. Please provide a valid date."), nil), nil
	case errors.Is(err, message.ErrInvalidDates):
		return message.NewResponse(i18n.GetLocale(ctx).Sprintf("Please provide a date in the valid format YYYY-MM-DD (e.g., 2023-12-31)"), nil), nil
	case err != nil:
		return nil, fmt.Errorf("failed to add question answer: %w", err)
	}

	// If questionnaire is complete, return success message
	if isComplete {
		return s.handleCompletedProfile(ctx, conv, request)
	}

	// Get the next question
	question, err := conv.GetCurrentQuestion()
	if err != nil {
		return nil, fmt.Errorf("failed to get next question: %w", err)
	}

	// Save conversation state after adding answer
	if err := s.repo.Save(ctx, conv); err != nil {
		return nil, fmt.Errorf("failed to save conversation state: %w", err)
	}

	return message.NewResponse(question.Text, question.Answers), nil
}

// handleCompletedProfile finalizes the pet profile questionnaire and saves the profile and conversation state.
// It retrieves the completed questionnaire results, generates a profile, and stores it in the profile repository.
// Returns a success response upon successful save or an error if any retrieval, creation, or save operation fails.
func (s *AIService) handleCompletedProfile(ctx context.Context, conv Conversation, request *message.UserMessage) (*message.Response, error) {
	result, err := conv.GetQuestionnaireResult()
	if err != nil {
		return nil, fmt.Errorf("failed to get questionnaire result: %w", err)
	}

	profile, err := createProfile(result)
	if err != nil {
		return nil, fmt.Errorf("failed to create profile: %w", err)
	}

	// Save conv state
	if err := s.repo.Save(ctx, conv); err != nil {
		return nil, fmt.Errorf("failed to save conversation: %w", err)
	}

	if err := s.profileRepo.SaveProfile(ctx, request.UserID, &profile); err != nil {
		return nil, fmt.Errorf("failed to save profile: %w", err)
	}

	return message.NewResponse("Pet profile saved successfully", []string{}), nil
}

// createProfile generates a pet profile from a slice of QuestionAnswer results.
// It maps specific fields in the QuestionAnswer slice to the corresponding fields in the Profile struct.
// Returns a populated Profile and an error if a field in the input slice is unrecognized.
func createProfile(result []conversation.QuestionAnswer) (pet.Profile, error) {
	var profile pet.Profile

	for _, qa := range result {
		switch qa.Field {
		case "name":
			profile.Name = qa.Answer
		case "species":
			profile.Species = qa.Answer
		case "breed":
			profile.Breed = qa.Answer
		case "dob":
			profile.DateOfBirth = qa.Answer
		case "gender":
			profile.Gender = qa.Answer
		case "weight":
			profile.Weight = qa.Answer
		default:
			return pet.Profile{}, fmt.Errorf("unknown field %s", qa.Field)
		}
	}
	return profile, nil
}
