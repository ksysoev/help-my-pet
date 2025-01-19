package memory_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ksysoev/help-my-pet/pkg/repo/memory"
)

func TestRateLimiter(t *testing.T) {
	tests := []struct {
		cfg    *memory.RateLimitConfig
		testFn func(t *testing.T, rl *memory.RateLimiter)
		name   string
	}{
		{
			name: "NewRateLimiter initialization",
			cfg: &memory.RateLimitConfig{
				WhitelistIDs:     []int64{123, 456},
				UserHourlyLimit:  10,
				UserDailyLimit:   15,
				GlobalDailyLimit: 1000,
			},
			testFn: func(t *testing.T, rl *memory.RateLimiter) {
				assert.NotNil(t, rl)
				// Test whitelist initialization
				assert.True(t, rl.IsWhitelisted(context.Background(), "123"))
				assert.True(t, rl.IsWhitelisted(context.Background(), "456"))
				assert.False(t, rl.IsWhitelisted(context.Background(), "789"))
			},
		},
		{
			name: "GetUserRequests with no requests",
			cfg: &memory.RateLimitConfig{
				UserHourlyLimit:  10,
				UserDailyLimit:   15,
				GlobalDailyLimit: 1000,
			},
			testFn: func(t *testing.T, rl *memory.RateLimiter) {
				userID := "user1"
				now := time.Now()

				count, err := rl.GetUserRequests(context.Background(), userID, now.Add(-time.Hour))
				assert.NoError(t, err)
				assert.Equal(t, 0, count)
			},
		},
		{
			name: "GetUserRequests with requests",
			cfg: &memory.RateLimitConfig{
				UserHourlyLimit:  10,
				UserDailyLimit:   15,
				GlobalDailyLimit: 1000,
			},
			testFn: func(t *testing.T, rl *memory.RateLimiter) {
				userID := "user1"
				now := time.Now()

				err := rl.AddUserRequest(context.Background(), userID, now)
				assert.NoError(t, err)

				count, err := rl.GetUserRequests(context.Background(), userID, now.Add(-time.Hour))
				assert.NoError(t, err)
				assert.Equal(t, 1, count)
			},
		},
		{
			name: "GetGlobalRequests with requests",
			cfg: &memory.RateLimitConfig{
				UserHourlyLimit:  10,
				UserDailyLimit:   15,
				GlobalDailyLimit: 1000,
			},
			testFn: func(t *testing.T, rl *memory.RateLimiter) {
				now := time.Now()

				err := rl.AddUserRequest(context.Background(), "user1", now)
				assert.NoError(t, err)
				err = rl.AddUserRequest(context.Background(), "user2", now)
				assert.NoError(t, err)

				count, err := rl.GetGlobalRequests(context.Background(), now.Add(-24*time.Hour))
				assert.NoError(t, err)
				assert.Equal(t, 2, count)
			},
		},
		{
			name: "AddUserRequest with cleanup",
			cfg: &memory.RateLimitConfig{
				UserHourlyLimit:  10,
				UserDailyLimit:   15,
				GlobalDailyLimit: 1000,
			},
			testFn: func(t *testing.T, rl *memory.RateLimiter) {
				userID := "user1"
				now := time.Now()

				// Add one recent and one old request
				err := rl.AddUserRequest(context.Background(), userID, now)
				assert.NoError(t, err)
				err = rl.AddUserRequest(context.Background(), userID, now.Add(-2*time.Hour))
				assert.NoError(t, err)

				// Only recent requests should be counted
				count, err := rl.GetUserRequests(context.Background(), userID, now.Add(-time.Hour))
				assert.NoError(t, err)
				assert.Equal(t, 1, count)
			},
		},
		{
			name: "IsNewQuestionAllowed with hourly limit",
			cfg: &memory.RateLimitConfig{
				UserHourlyLimit:  2,
				UserDailyLimit:   15,
				GlobalDailyLimit: 1000,
				WhitelistIDs:     []int64{999},
			},
			testFn: func(t *testing.T, rl *memory.RateLimiter) {
				userID := "user1"
				whitelistedID := "999"
				now := time.Now()

				// Test initial state
				allowed, err := rl.IsNewQuestionAllowed(context.Background(), userID)
				assert.NoError(t, err)
				assert.True(t, allowed)

				// Add requests up to hourly limit
				err = rl.AddUserRequest(context.Background(), userID, now)
				assert.NoError(t, err)
				err = rl.AddUserRequest(context.Background(), userID, now)
				assert.NoError(t, err)

				// Test at hourly limit
				allowed, err = rl.IsNewQuestionAllowed(context.Background(), userID)
				assert.Error(t, err)
				assert.False(t, allowed)

				// Test whitelisted user
				allowed, err = rl.IsNewQuestionAllowed(context.Background(), whitelistedID)
				assert.NoError(t, err)
				assert.True(t, allowed)
			},
		},
		{
			name: "IsNewQuestionAllowed with user daily limit",
			cfg: &memory.RateLimitConfig{
				UserHourlyLimit:  20,
				UserDailyLimit:   3,
				GlobalDailyLimit: 1000,
				WhitelistIDs:     []int64{999},
			},
			testFn: func(t *testing.T, rl *memory.RateLimiter) {
				userID := "user1"
				now := time.Now()

				// Add requests up to daily limit
				for i := 0; i < 3; i++ {
					err := rl.AddUserRequest(context.Background(), userID, now)
					assert.NoError(t, err)
				}

				// Test at daily limit
				allowed, err := rl.IsNewQuestionAllowed(context.Background(), userID)
				assert.Error(t, err)
				assert.False(t, allowed)

				// Test whitelisted user still allowed
				allowed, err = rl.IsNewQuestionAllowed(context.Background(), "999")
				assert.NoError(t, err)
				assert.True(t, allowed)
			},
		},
		{
			name: "IsNewQuestionAllowed with global daily limit",
			cfg: &memory.RateLimitConfig{
				UserHourlyLimit:  20,
				UserDailyLimit:   20,
				GlobalDailyLimit: 2,
				WhitelistIDs:     []int64{999},
			},
			testFn: func(t *testing.T, rl *memory.RateLimiter) {
				now := time.Now()

				// Add requests up to global limit
				err := rl.AddUserRequest(context.Background(), "user1", now)
				assert.NoError(t, err)
				err = rl.AddUserRequest(context.Background(), "user2", now)
				assert.NoError(t, err)

				// Test at global limit for a new user
				allowed, err := rl.IsNewQuestionAllowed(context.Background(), "user3")
				assert.Error(t, err)
				assert.False(t, allowed)

				// Test whitelisted user still allowed
				allowed, err = rl.IsNewQuestionAllowed(context.Background(), "999")
				assert.NoError(t, err)
				assert.True(t, allowed)
			},
		},
		{
			name: "RecordNewQuestion",
			cfg: &memory.RateLimitConfig{
				UserHourlyLimit:  10,
				UserDailyLimit:   15,
				GlobalDailyLimit: 1000,
			},
			testFn: func(t *testing.T, rl *memory.RateLimiter) {
				userID := "user1"

				// Record a question
				err := rl.RecordNewQuestion(context.Background(), userID)
				assert.NoError(t, err)

				// Verify it was recorded
				count, err := rl.GetUserRequests(context.Background(), userID, time.Now().Add(-time.Hour))
				assert.NoError(t, err)
				assert.Equal(t, 1, count)
			},
		},
		{
			name: "IsWhitelisted",
			cfg: &memory.RateLimitConfig{
				WhitelistIDs:     []int64{123, 456},
				UserHourlyLimit:  10,
				UserDailyLimit:   15,
				GlobalDailyLimit: 1000,
			},
			testFn: func(t *testing.T, rl *memory.RateLimiter) {
				// Test whitelisted users
				assert.True(t, rl.IsWhitelisted(context.Background(), "123"))
				assert.True(t, rl.IsWhitelisted(context.Background(), "456"))

				// Test non-whitelisted user
				assert.False(t, rl.IsWhitelisted(context.Background(), "789"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rl := memory.NewRateLimiter(tt.cfg)
			tt.testFn(t, rl)
		})
	}
}
