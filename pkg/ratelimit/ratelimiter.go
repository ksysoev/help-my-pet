package ratelimit

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Config holds rate limiting configuration
type Config struct {
	WhitelistIDs []int64 `mapstructure:"whitelist_ids"`
	HourlyLimit  int     `mapstructure:"hourly_limit"`
	DailyLimit   int     `mapstructure:"daily_limit"`
}

// UserLimits tracks rate limits for a specific user
type UserLimits struct {
	HourlyReset time.Time
	DailyReset  time.Time
	HourlyCount int
	DailyCount  int
}

// RateLimiter defines the interface for rate limiting functionality
type RateLimiter interface {
	IsAllowed(ctx context.Context, userID int64) (bool, error)
	RecordAccess(ctx context.Context, userID int64) error
}

// InMemoryRateLimiter implements RateLimiter using in-memory storage
type InMemoryRateLimiter struct {
	limits    map[int64]*UserLimits
	config    *Config
	whitelist map[int64]struct{}
	mu        sync.RWMutex
}

// NewRateLimiter creates a new InMemoryRateLimiter with the given configuration
func NewRateLimiter(cfg *Config) *InMemoryRateLimiter {
	whitelist := make(map[int64]struct{})
	for _, id := range cfg.WhitelistIDs {
		whitelist[id] = struct{}{}
	}

	return &InMemoryRateLimiter{
		limits:    make(map[int64]*UserLimits),
		config:    cfg,
		whitelist: whitelist,
	}
}

// IsAllowed checks if a user is allowed to make a request
func (r *InMemoryRateLimiter) IsAllowed(ctx context.Context, userID int64) (bool, error) {
	// Check whitelist
	if _, ok := r.whitelist[userID]; ok {
		return true, nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	limits, ok := r.limits[userID]
	if !ok {
		return true, nil
	}

	now := time.Now()

	// Reset counters if needed
	if now.After(limits.DailyReset) {
		return true, nil
	}
	if now.After(limits.HourlyReset) {
		if limits.DailyCount < r.config.DailyLimit {
			return true, nil
		}
	}

	// Check daily limit first
	if limits.DailyCount >= r.config.DailyLimit {
		return false, fmt.Errorf("daily limit exceeded for user %d", userID)
	}

	// Then check hourly limit
	if !now.After(limits.HourlyReset) && limits.HourlyCount >= r.config.HourlyLimit {
		return false, fmt.Errorf("hourly limit exceeded for user %d", userID)
	}

	return true, nil
}

// RecordAccess records a user's access and updates their limits
func (r *InMemoryRateLimiter) RecordAccess(ctx context.Context, userID int64) error {
	// Skip recording for whitelisted users
	if _, ok := r.whitelist[userID]; ok {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	limits, ok := r.limits[userID]

	if !ok || now.After(limits.DailyReset) {
		// Initialize new limits
		limits = &UserLimits{
			HourlyCount: 0,
			DailyCount:  0,
			HourlyReset: now.Add(time.Hour),
			DailyReset:  now.Add(24 * time.Hour),
		}
		r.limits[userID] = limits
	} else if now.After(limits.HourlyReset) {
		// Reset hourly count
		limits.HourlyCount = 0
		limits.HourlyReset = now.Add(time.Hour)
	}

	limits.HourlyCount++
	limits.DailyCount++

	return nil
}
