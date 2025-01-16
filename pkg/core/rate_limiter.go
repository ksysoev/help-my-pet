package core

import (
	"context"
	"time"
)

// RateLimitConfig holds configuration for rate limiting
type RateLicmitConfig struct {
	Whitelist   []string
	HourlyLimit int
}

// RateLimiter defines the interface for rate limiting functionality
type RateLimiter interface {
	// IsNewQuestionAllowed checks if a user is allowed to ask a new question
	IsNewQuestionAllowed(ctx context.Context, userID string) (bool, error)
	// RecordNewQuestion records that a user has asked a new question
	RecordNewQuestion(ctx context.Context, userID string) error
}

// RateLimitRepository defines the interface for storing rate limit data
type RateLimitRepository interface {
	// GetUserRequests gets the number of requests a user has made within the last hour
	GetUserRequests(ctx context.Context, userID string, since time.Time) (int, error)
	// AddUserRequest records a new request for a user
	AddUserRequest(ctx context.Context, userID string, timestamp time.Time) error
	// IsWhitelisted checks if a user is whitelisted (not subject to rate limiting)
	IsWhitelisted(ctx context.Context, userID string) bool
}
