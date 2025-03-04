package core

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/ksysoev/help-my-pet/pkg/core/conversation"
	"github.com/ksysoev/help-my-pet/pkg/core/message"
)

func (s *AIService) ProcessMessage(ctx context.Context, request *message.UserMessage) (*message.Response, error) {
	slog.DebugContext(ctx, "getting pet advice", "input", request.Text)

	conv, err := s.repo.FindOrCreate(ctx, request.ChatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	switch conv.GetState() {
	case conversation.StateNormal:
		return s.handleNewQuestion(ctx, conv, request)
	case conversation.StatePetProfileQuestioning:
		return s.ProcessProfileAnswer(ctx, conv, request)
	case conversation.StateFollowUpQuestioning:
		return s.ProcessFollowUpAnswer(ctx, conv, request)
	default:
		return nil, fmt.Errorf("unknown conversation state: %s", conv.GetState())
	}
}

// CancelQuestionnaire cancels the active questionnaire for the specified chat ID.
// It retrieves or initializes the conversation, updates its state, and persists the changes to the repository.
// Returns error if retrieving or saving the conversation fails.
func (s *AIService) CancelQuestionnaire(ctx context.Context, chatID string) error {
	conv, err := s.repo.FindOrCreate(ctx, chatID)
	if err != nil {
		return fmt.Errorf("failed to get conversation: %w", err)
	}

	conv.CancelQuestionnaire()

	if err := s.repo.Save(ctx, conv); err != nil {
		return fmt.Errorf("failed to save conversation: %w", err)
	}

	return nil
}

// handleNewQuestion processes a new question from the user
func (s *AIService) handleNewQuestion(ctx context.Context, conv Conversation, request *message.UserMessage) (*message.Response, error) {
	// Check rate limit for new questions
	if s.rateLimiter != nil {
		allowed, err := s.rateLimiter.IsNewQuestionAllowed(ctx, request.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to check rate limit: %w", err)
		}
		if !allowed {
			return nil, fmt.Errorf("rate limit exceeded for user %s", request.UserID)
		}

		if err := s.rateLimiter.RecordNewQuestion(ctx, request.UserID); err != nil {
			return nil, fmt.Errorf("failed to record rate limit: %w", err)
		}
	}

	// Add user's question to conv
	conv.AddMessage("user", request.Text)

	// Save conv immediately after adding user's message
	if err := s.repo.Save(ctx, conv); err != nil {
		return nil, fmt.Errorf("failed to save conversation: %w", err)
	}

	// Build prompt with pet profiles and conv context
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

	prompt += fmt.Sprintf("%s\nCurrent question: %s", conv.History(1), request.Text)

	response, err := s.llm.Analyze(ctx, prompt, request.Images)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI response: %w", err)
	}

	if response.Media != "" {
		conv.AddMessage("media_description", response.Media)
	}

	// Add AI's response to conv
	conv.AddMessage("assistant", response.Text)

	// Handle follow-up questions if any
	if len(response.Questions) > 0 {
		// Initialize questionnaire
		if err := conv.StartFollowUpQuestions(response.Text, response.Questions); err != nil {
			return nil, fmt.Errorf("failed to start follow-up questions: %w", err)
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

		// Return response with the first question
		return message.NewResponse(
			response.Text+"\n\n"+currentQuestion.Text,
			currentQuestion.Answers,
		), nil
	}

	// Save conv state
	if err := s.repo.Save(ctx, conv); err != nil {
		return nil, fmt.Errorf("failed to save conversation: %w", err)
	}

	return message.NewResponse(response.Text, []string{}), nil
}

// ResetUserConversation removes all user profiles and deletes the specified conversation.
// It deletes user-specific profiles using the userID and removes the conversation identified by chatID.
// Returns error if profile removal or conversation deletion fails.
func (s *AIService) ResetUserConversation(ctx context.Context, userID, chatID string) error {
	err := s.profileRepo.RemoveUserProfiles(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to remove user profiles: %w", err)
	}

	err = s.repo.Delete(ctx, chatID)
	if err != nil {
		return fmt.Errorf("failed to remove conversation: %w", err)
	}

	return nil
}
