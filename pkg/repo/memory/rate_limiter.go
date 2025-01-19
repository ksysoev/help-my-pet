package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ksysoev/help-my-pet/pkg/core"
)

// UserRequests stores request timestamps for a user
type UserRequests struct {
	Timestamps []time.Time
}

var _ core.RateLimiter = (*RateLimiter)(nil)

// RateLimiter implements core.RateLimiter interface using in-memory storage
type RateLimiter struct {
	requests  map[string]*UserRequests
	config    *RateLimitConfig
	whitelist map[string]struct{}
	mu        sync.RWMutex
}

// NewRateLimiter creates a new RateLimiter with the given configuration
func NewRateLimiter(cfg *RateLimitConfig) *RateLimiter {
	whitelist := make(map[string]struct{})
	for _, id := range cfg.WhitelistIDs {
		whitelist[fmt.Sprintf("%d", id)] = struct{}{}
	}

	return &RateLimiter{
		requests:  make(map[string]*UserRequests),
		config:    cfg,
		whitelist: whitelist,
	}
}

// GetUserRequests gets the number of requests a user has made within the given time period
func (r *RateLimiter) GetUserRequests(ctx context.Context, userID string, since time.Time) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	userReqs, exists := r.requests[userID]
	if !exists {
		return 0, nil
	}

	count := 0
	for _, ts := range userReqs.Timestamps {
		if ts.After(since) {
			count++
		}
	}

	return count, nil
}

// GetGlobalRequests gets the total number of requests made by all users since the given time
func (r *RateLimiter) GetGlobalRequests(ctx context.Context, since time.Time) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, userReqs := range r.requests {
		for _, ts := range userReqs.Timestamps {
			if ts.After(since) {
				count++
			}
		}
	}

	return count, nil
}

// AddUserRequest records a new request for a user
func (r *RateLimiter) AddUserRequest(ctx context.Context, userID string, timestamp time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	userReqs, exists := r.requests[userID]
	if !exists {
		userReqs = &UserRequests{
			Timestamps: make([]time.Time, 0),
		}
		r.requests[userID] = userReqs
	}

	// Clean up old timestamps while adding new one
	newTimestamps := make([]time.Time, 0)
	hourAgo := time.Now().Add(-time.Hour)

	for _, ts := range userReqs.Timestamps {
		if ts.After(hourAgo) {
			newTimestamps = append(newTimestamps, ts)
		}
	}

	newTimestamps = append(newTimestamps, timestamp)
	userReqs.Timestamps = newTimestamps

	return nil
}

// IsNewQuestionAllowed checks if a user is allowed to ask a new question
func (r *RateLimiter) IsNewQuestionAllowed(ctx context.Context, userID string) (bool, error) {
	if r.IsWhitelisted(ctx, userID) {
		return true, nil
	}

	// Check hourly limit
	hourAgo := time.Now().Add(-time.Hour)
	hourlyCount, err := r.GetUserRequests(ctx, userID, hourAgo)
	if err != nil {
		return false, fmt.Errorf("failed to get user hourly requests: %w", err)
	}
	if hourlyCount >= r.config.UserHourlyLimit {
		return false, core.ErrRateLimit
	}

	// Check user daily limit
	dayStart := time.Now().Truncate(24 * time.Hour)
	userDailyCount, err := r.GetUserRequests(ctx, userID, dayStart)
	if err != nil {
		return false, fmt.Errorf("failed to get user daily requests: %w", err)
	}
	if userDailyCount >= r.config.UserDailyLimit {
		return false, core.ErrRateLimit
	}

	// Check global daily limit
	globalCount, err := r.GetGlobalRequests(ctx, dayStart)
	if err != nil {
		return false, fmt.Errorf("failed to get global requests: %w", err)
	}
	if globalCount >= r.config.GlobalDailyLimit {
		return false, core.ErrGlobalLimit
	}

	return true, nil
}

// RecordNewQuestion records that a user has asked a new question
func (r *RateLimiter) RecordNewQuestion(ctx context.Context, userID string) error {
	return r.AddUserRequest(ctx, userID, time.Now())
}

// IsWhitelisted checks if a user is whitelisted
func (r *RateLimiter) IsWhitelisted(ctx context.Context, userID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.whitelist[userID]
	return exists
}
