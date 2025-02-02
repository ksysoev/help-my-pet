package core

import (
	"context"
	"errors"
	"fmt"

	"github.com/ksysoev/help-my-pet/pkg/core/conversation"
)

func (s *AIService) ProcessFollowUpAnswer(ctx context.Context, conv *conversation.Conversation, request *UserMessage) (*Response, error) {
	// Store the answer and check if questionnaire is complete
	isComplete, err := conv.AddQuestionAnswer(request.Text)
	if err != nil {
		return nil, fmt.Errorf("failed to add question answer: %w", err)
	}

	// Save conv after adding answer
	if err := s.repo.Save(ctx, conv); err != nil {
		return nil, fmt.Errorf("failed to save conv: %w", err)
	}

	if isComplete {
		// Get all collected question-answer pairs
		qaPairs, err := conv.GetQuestionnaireResult()
		if err != nil {
			return nil, fmt.Errorf("failed to get questionnaire result: %w", err)
		}

		// Build prompt with conv history and question-answer pairs
		var prompt string

		// Fetch pet profile from repository
		petProfile, err := s.profileRepo.GetCurrentProfile(ctx, request.UserID)
		if errors.Is(err, ErrProfileNotFound) {
			// If no profile found, do not include pet profiles in prompt
		} else if err != nil {
			return nil, fmt.Errorf("failed to fetch pet profiles: %w", err)
		} else {
			// Include pet profiles in prompt
			prompt += fmt.Sprintf("%s\n\n", petProfile.String())
		}

		prompt += "Previous conv:\n"
		history := conv.GetContext()
		for _, msg := range history[:len(history)-1] {
			prompt += fmt.Sprintf("%s: %s\n\n", msg.Role, msg.Content)
		}

		prompt += "\nFollow-up information:\n"
		for _, qa := range qaPairs {
			prompt += fmt.Sprintf("Question: %s\nAnswer: %s\n", qa.Question.Text, qa.Answer)
		}

		// Get final response from LLM
		response, err := s.llm.Call(ctx, prompt)
		if err != nil {
			return nil, fmt.Errorf("failed to get AI response: %w", err)
		}

		// Add AI's response to conv history
		conv.AddMessage("assistant", response.Text)

		// Save conv state
		if err := s.repo.Save(ctx, conv); err != nil {
			return nil, fmt.Errorf("failed to save conv: %w", err)
		}

		return NewResponse(response.Text, []string{}), nil
	}

	// Get next question
	currentQuestion, err := conv.GetCurrentQuestion()
	if err != nil {
		return nil, fmt.Errorf("failed to get next question: %w", err)
	}

	// Save conv state
	if err := s.repo.Save(ctx, conv); err != nil {
		return nil, fmt.Errorf("failed to save conv: %w", err)
	}

	return NewResponse(currentQuestion.Text, currentQuestion.Answers), nil
}
