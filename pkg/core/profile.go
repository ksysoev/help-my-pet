package core

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/ksysoev/help-my-pet/pkg/core/conversation"
)

func (s *AIService) ProcessEditProfile(ctx context.Context, request *UserMessage) (*Response, error) {
	slog.DebugContext(ctx, "managing pet profile", "input", request.Text)

	conv, err := s.repo.FindOrCreate(ctx, request.ChatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	// Check if user is in the middle of a conversation
	if conv.State != conversation.StateNormal {
		return nil, errors.New("cannot manage profile during a conversation")
	}

	// Start pet profile questionnaire
	if err := conv.StartProfileQuestions(); err != nil {
		return nil, fmt.Errorf("failed to start profile questions: %w", err)
	}

	// Get the first question
	currentQuestion, err := conv.GetCurrentQuestion()
	if err != nil {
		return nil, fmt.Errorf("failed to get first question: %w", err)
	}

	// Save conv state
	if err := s.repo.Save(ctx, conv); err != nil {
		return nil, fmt.Errorf("failed to save conversation: %w", err)
	}

	// Return response with first question
	var answers []string
	if len(currentQuestion.Answers) > 0 {
		answers = currentQuestion.Answers
	}
	return NewResponse(currentQuestion.Text, answers), nil
}

func (s *AIService) ProcessProfileAnswer(ctx context.Context, conv *conversation.Conversation, request *UserMessage) (*Response, error) {
	slog.DebugContext(ctx, "managing pet profile", "input", request.Text)

	conv, err := s.repo.FindOrCreate(ctx, request.ChatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	// Check if user is in the middle of a conversation
	if conv.State != conversation.StatePetProfileQuestioning {
		return nil, errors.New("cannot manage profile during a conversation")
	}

	// Add answer to the current question
	isComplete, err := conv.AddQuestionAnswer(request.Text)
	if err != nil {
		return nil, fmt.Errorf("failed to add answer: %w", err)
	}

	// If questionnaire is complete, return success message
	if isComplete {
		result, err := conv.GetQuestionnaireResult()
		if err != nil {
			return nil, fmt.Errorf("failed to get questionnaire result: %w", err)
		}

		var profile PetProfile

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
				return nil, fmt.Errorf("unknown field %s", qa.Field)
			}
		}

		// Save conv state
		if err := s.repo.Save(ctx, conv); err != nil {
			return nil, fmt.Errorf("failed to save conversation: %w", err)
		}

		if err = s.profileRepo.SaveProfile(ctx, request.UserID, &profile); err != nil {
			return nil, fmt.Errorf("failed to save profile: %w", err)
		}

		return NewResponse("Pet profile saved successfully", []string{}), nil
	}

	// Get the next question
	currentQuestion, err := conv.GetCurrentQuestion()
	if err != nil {
		return nil, fmt.Errorf("failed to get next question: %w", err)
	}

	// Return response with next question
	var answers []string
	if len(currentQuestion.Answers) > 0 {
		answers = currentQuestion.Answers
	}
	return NewResponse(currentQuestion.Text, answers), nil
}
