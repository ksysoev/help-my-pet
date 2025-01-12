package ratelimit

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestRateLimiter(t *testing.T) {
	cfg := &Config{
		HourlyLimit:  5,
		DailyLimit:   15,
		WhitelistIDs: []int64{999},
	}

	limiter := NewRateLimiter(cfg)
	ctx := context.Background()

	t.Run("allows whitelisted users", func(t *testing.T) {
		allowed, err := limiter.IsAllowed(ctx, 999)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !allowed {
			t.Error("whitelisted user should be allowed")
		}

		// Record multiple accesses for whitelisted user
		for i := 0; i < cfg.HourlyLimit*2; i++ {
			if err := limiter.RecordAccess(ctx, 999); err != nil {
				t.Errorf("unexpected error recording access: %v", err)
			}
		}

		// Should still be allowed after exceeding limits
		allowed, err = limiter.IsAllowed(ctx, 999)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !allowed {
			t.Error("whitelisted user should always be allowed")
		}
	})

	t.Run("respects hourly limit", func(t *testing.T) {
		userID := int64(1)

		// Use up hourly limit
		for i := 0; i < cfg.HourlyLimit; i++ {
			allowed, err := limiter.IsAllowed(ctx, userID)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !allowed {
				t.Errorf("request %d should be allowed", i+1)
			}
			if err := limiter.RecordAccess(ctx, userID); err != nil {
				t.Errorf("unexpected error recording access: %v", err)
			}
		}

		// Next request should be denied
		allowed, err := limiter.IsAllowed(ctx, userID)
		if err == nil {
			t.Error("expected error for exceeded limit")
		}
		if allowed {
			t.Error("request should be denied after reaching hourly limit")
		}
	})

	t.Run("respects daily limit", func(t *testing.T) {
		userID := int64(2)
		limiter := NewRateLimiter(cfg) // Fresh limiter for this test

		// Simulate multiple hours of usage
		now := time.Now()
		for hour := 0; hour < 3; hour++ {
			// Set current hour's reset time
			hourReset := now.Add(time.Duration(hour+1) * time.Hour)
			limiter.limits[userID] = &UserLimits{
				HourlyReset: hourReset,
				DailyReset:  now.Add(24 * time.Hour),
				HourlyCount: 0,
				DailyCount:  hour * cfg.HourlyLimit,
			}

			// Use up hourly limit for this hour
			for i := 0; i < cfg.HourlyLimit; i++ {
				allowed, err := limiter.IsAllowed(ctx, userID)
				if err != nil {
					t.Errorf("hour %d, request %d: unexpected error: %v", hour+1, i+1, err)
				}
				if !allowed {
					t.Errorf("hour %d, request %d should be allowed", hour+1, i+1)
				}
				if err := limiter.RecordAccess(ctx, userID); err != nil {
					t.Errorf("hour %d, request %d: unexpected error recording access: %v", hour+1, i+1, err)
				}
			}
		}

		// Next request should be denied due to daily limit
		allowed, err := limiter.IsAllowed(ctx, userID)
		if err == nil {
			t.Error("expected error for exceeded daily limit")
		}
		if allowed {
			t.Error("request should be denied after reaching daily limit")
		}
		if err != nil && err.Error() != fmt.Sprintf("daily limit exceeded for user %d", userID) {
			t.Errorf("expected daily limit error, got: %v", err)
		}
	})

	t.Run("first access for new user", func(t *testing.T) {
		userID := int64(4)
		limiter := NewRateLimiter(cfg)

		// First access should be allowed
		allowed, err := limiter.IsAllowed(ctx, userID)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !allowed {
			t.Error("first access should be allowed")
		}

		// Record the access
		if err := limiter.RecordAccess(ctx, userID); err != nil {
			t.Errorf("unexpected error recording access: %v", err)
		}
	})

	t.Run("daily reset in IsAllowed", func(t *testing.T) {
		userID := int64(5)
		limiter := NewRateLimiter(cfg)

		// Set up expired daily limits
		limiter.limits[userID] = &UserLimits{
			HourlyReset: time.Now().Add(time.Hour),
			DailyReset:  time.Now().Add(-time.Minute), // Expired
			HourlyCount: cfg.HourlyLimit - 1,
			DailyCount:  cfg.DailyLimit,
		}

		// Should be allowed due to daily reset
		allowed, err := limiter.IsAllowed(ctx, userID)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !allowed {
			t.Error("request should be allowed after daily reset")
		}
	})

	t.Run("hourly reset with daily limit not exceeded", func(t *testing.T) {
		userID := int64(6)
		limiter := NewRateLimiter(cfg)

		// Set up expired hourly limits but not daily
		limiter.limits[userID] = &UserLimits{
			HourlyReset: time.Now().Add(-time.Minute), // Expired
			DailyReset:  time.Now().Add(time.Hour),
			HourlyCount: cfg.HourlyLimit,
			DailyCount:  cfg.DailyLimit - 1, // Not exceeded
		}

		// Should be allowed due to hourly reset and daily limit not exceeded
		allowed, err := limiter.IsAllowed(ctx, userID)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !allowed {
			t.Error("request should be allowed after hourly reset with daily limit not exceeded")
		}
	})

	t.Run("record access with expired limits", func(t *testing.T) {
		userID := int64(7)
		limiter := NewRateLimiter(cfg)

		// Set up expired limits
		limiter.limits[userID] = &UserLimits{
			HourlyReset: time.Now().Add(-time.Hour),      // Expired
			DailyReset:  time.Now().Add(-24 * time.Hour), // Expired
			HourlyCount: cfg.HourlyLimit,
			DailyCount:  cfg.DailyLimit,
		}

		// Record access should reset both limits
		if err := limiter.RecordAccess(ctx, userID); err != nil {
			t.Errorf("unexpected error recording access: %v", err)
		}

		// Verify counts were reset
		limits := limiter.limits[userID]
		if limits.HourlyCount != 1 {
			t.Errorf("expected hourly count to be 1, got %d", limits.HourlyCount)
		}
		if limits.DailyCount != 1 {
			t.Errorf("expected daily count to be 1, got %d", limits.DailyCount)
		}
		if !limits.HourlyReset.After(time.Now()) {
			t.Error("hourly reset time should be in the future")
		}
		if !limits.DailyReset.After(time.Now()) {
			t.Error("daily reset time should be in the future")
		}
	})

	t.Run("resets limits after time period", func(t *testing.T) {
		userID := int64(3)
		limiter := NewRateLimiter(cfg)

		// Use up hourly limit
		for i := 0; i < cfg.HourlyLimit; i++ {
			if err := limiter.RecordAccess(ctx, userID); err != nil {
				t.Errorf("unexpected error recording access: %v", err)
			}
		}

		// Simulate time passing
		limiter.limits[userID].HourlyReset = time.Now().Add(-time.Minute)

		// Should be allowed again
		allowed, err := limiter.IsAllowed(ctx, userID)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !allowed {
			t.Error("request should be allowed after reset period")
		}
	})
}
