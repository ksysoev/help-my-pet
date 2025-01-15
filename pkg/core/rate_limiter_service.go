package core

import (
	"context"
	"time"
)

// RateLimiterService implements RateLimiter interface using a repository
type RateLimiterService struct {
	repo   RateLimitRepository
	config *RateLimitConfig
}

// NewRateLimiterService creates a new RateLimiterService
func NewRateLimiterService(repo RateLimitRepository, config *RateLimitConfig) *RateLimiterService {
	return &RateLimiterService{
		repo:   repo,
		config: config,
	}
}

// IsNewQuestionAllowed checks if a user is allowed to ask a new question
func (s *RateLimiterService) IsNewQuestionAllowed(ctx context.Context, userID string) (bool, error) {
	if s.repo.IsWhitelisted(ctx, userID) {
		return true, nil
	}

	hourAgo := time.Now().Add(-time.Hour)
	count, err := s.repo.GetUserRequests(ctx, userID, hourAgo)
	if err != nil {
		return false, err
	}

	return count < s.config.HourlyLimit, nil
}

// RecordNewQuestion records that a user has asked a new question
func (s *RateLimiterService) RecordNewQuestion(ctx context.Context, userID string) error {
	return s.repo.AddUserRequest(ctx, userID, time.Now())
}
