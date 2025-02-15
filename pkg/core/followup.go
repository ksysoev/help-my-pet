package core

import (
	"context"
	"errors"
	"fmt"

	"github.com/ksysoev/help-my-pet/pkg/core/message"
	"github.com/ksysoev/help-my-pet/pkg/i18n"
)

// ProcessFollowUpAnswer processes a user's answer to a follow-up question during a conversation.
// It validates and stores the answer, checks if the questionnaire is complete, and either transitions to the next step
// or returns the next question. If the follow-up is complete, it generates a concluding response.
// Returns the response to the current or next question, or an error if validation, state updates, or saving fails.
func (s *AIService) ProcessFollowUpAnswer(ctx context.Context, conv Conversation, request *message.UserMessage) (*message.Response, error) {
	// Store the answer and check if questionnaire is complete
	isComplete, err := conv.AddQuestionAnswer(request.Text)

	switch {
	case errors.Is(err, message.ErrTextTooLong):
		return message.NewResponse(i18n.GetLocale(ctx).Sprintf("I apologize, but your message is too long for me to process. Please try to make it shorter and more concise."), nil), nil
	case err != nil:
		return nil, fmt.Errorf("failed to add question answer: %w", err)
	}

	// Save conv after adding answer
	if err := s.repo.Save(ctx, conv); err != nil {
		return nil, fmt.Errorf("failed to save conversation: %w", err)
	}

	if isComplete {
		return s.handleCompletedFollowUp(ctx, conv, request)
	}

	// Get next question
	currentQuestion, err := conv.GetCurrentQuestion()
	if err != nil {
		return nil, fmt.Errorf("failed to get next question: %w", err)
	}

	// Save conv state
	if err := s.repo.Save(ctx, conv); err != nil {
		return nil, fmt.Errorf("failed to save conversation: %w", err)
	}

	return message.NewResponse(currentQuestion.Text, currentQuestion.Answers), nil
}

// handleCompletedFollowUp finalizes the follow-up process by generating a comprehensive response using conversation history.
// It constructs a prompt incorporating conversation history and Q&A pairs, fetches a response from the AI model, and appends it to the conversation history.
// Saves the updated conversation and returns the generated response.
// Returns an error if prompt preparation, AI response retrieval, or conversation saving fails.
func (s *AIService) handleCompletedFollowUp(ctx context.Context, conv Conversation, request *message.UserMessage) (*message.Response, error) {
	// Build prompt with conv history and question-answer pairs
	prompt, err := s.prepareFollowUpPrompt(ctx, conv, request)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare prompt: %w", err)
	}

	// Get final response from LLM
	response, err := s.llm.Call(ctx, prompt, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI response: %w", err)
	}

	// Add AI's response to conv history
	conv.AddMessage("assistant", response.Text)

	// Save conv state
	if err := s.repo.Save(ctx, conv); err != nil {
		return nil, fmt.Errorf("failed to save conversation: %w", err)
	}

	return message.NewResponse(response.Text, []string{}), nil
}

// prepareFollowUpPrompt constructs a detailed prompt for the AI model based on conversation history, pet profiles, and Q&A data.
// It retrieves any completed Q&A pairs and the relevant pet profile, if available, to enhance the context of the prompt.
// Returns the prepared prompt string and an error if fetching Q&A pairs, pet profiles, or other context data fails.
func (s *AIService) prepareFollowUpPrompt(ctx context.Context, conv Conversation, request *message.UserMessage) (string, error) {
	// Get all collected question-answer pairs
	qaPairs, err := conv.GetQuestionnaireResult()
	if err != nil {
		return "", fmt.Errorf("failed to get questionnaire result: %w", err)
	}

	var prompt string

	// Fetch pet profile from repository
	petProfile, err := s.profileRepo.GetCurrentProfile(ctx, request.UserID)
	if errors.Is(err, ErrProfileNotFound) {
		// If no profile found, do not include pet profiles in prompt
	} else if err != nil {
		return "", fmt.Errorf("failed to fetch pet profiles: %w", err)
	} else {
		// Include pet profiles in prompt
		prompt += fmt.Sprintf("%s\n\n", petProfile.String())
	}

	prompt += fmt.Sprintf("%s\nFollow-up information:\n", conv.History(1))
	for _, qa := range qaPairs {
		prompt += fmt.Sprintf("Question: %s\nAnswer: %s\n", qa.Question.Text, qa.Answer)
	}
	return prompt, nil
}
