package core

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/ksysoev/help-my-pet/pkg/core/conversation"
)

func (s *AIService) ProcessManageProfile(ctx context.Context, request *UserMessage) (*Response, error) {
	slog.DebugContext(ctx, "managing pet profile", "input", request.Text)

	conv, err := s.repo.FindOrCreate(ctx, request.ChatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conv: %w", err)
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
		return nil, fmt.Errorf("failed to save conv: %w", err)
	}

	// Return response with first question
	answers := []string{}
	if len(currentQuestion.Answers) > 0 {
		answers = currentQuestion.Answers
	}
	return NewPetAdviceResponse(currentQuestion.Text, answers), nil
}

func (s *AIService) ProcessManageProfileAnswer(ctx context.Context, request *UserMessage) (*Response, error) {
	slog.DebugContext(ctx, "managing pet profile", "input", request.Text)

	conv, err := s.repo.FindOrCreate(ctx, request.ChatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conv: %w", err)
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

	// Save conv state
	if err := s.repo.Save(ctx, conv); err != nil {
		return nil, fmt.Errorf("failed to save conv: %w", err)
	}

	// If questionnaire is complete, return success message
	if isComplete {

	}

	// Get the next question
	currentQuestion, err := conv.GetCurrentQuestion()
	if err != nil {
		return nil, fmt.Errorf("failed to get next question: %w", err)
	}

	// Return response with next question
	answers := []string{}
	if len(currentQuestion.Answers) > 0 {
		answers = currentQuestion.Answers
	}
	return NewPetAdviceResponse(currentQuestion.Text, answers), nil
}
