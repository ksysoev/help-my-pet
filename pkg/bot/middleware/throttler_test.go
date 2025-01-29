package middleware

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
)

func TestWithThrottlerLimitsConcurrentProcessing(t *testing.T) {
	var (
		mu           sync.Mutex
		currentCount int
		maxCount     int
		wg           sync.WaitGroup
	)

	// Create a handler that tracks concurrent executions
	handler := HandlerFunc(func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
		mu.Lock()
		currentCount++
		if currentCount > maxCount {
			maxCount = currentCount
		}
		mu.Unlock()

		// Simulate work
		time.Sleep(50 * time.Millisecond)

		mu.Lock()
		currentCount--
		mu.Unlock()

		return tgbotapi.MessageConfig{}, nil
	})

	// Create throttled handler with limit of 5 for testing
	throttled := WithThrottler(5)(handler)

	// Send 10 concurrent requests
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = throttled.Handle(context.Background(), &tgbotapi.Message{})
		}()
	}

	wg.Wait()

	assert.LessOrEqual(t, maxCount, 5, "concurrent processing exceeded limit")
}

func TestWithThrottlerHandlesContextCancellation(t *testing.T) {
	// Create a handler that blocks until explicitly unblocked
	blockCh := make(chan struct{})
	handler := HandlerFunc(func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
		<-blockCh // Block until channel is closed
		return tgbotapi.MessageConfig{}, nil
	})

	// Create throttled handler with limit of 1
	throttled := WithThrottler(1)(handler)

	// Fill up the throttler
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = throttled.Handle(context.Background(), &tgbotapi.Message{})
	}()

	// Wait a bit to ensure the first request has acquired the slot
	time.Sleep(50 * time.Millisecond)

	// Try another request with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := throttled.Handle(ctx, &tgbotapi.Message{})
	assert.Error(t, err, "should return error when context is cancelled")
	assert.Contains(t, err.Error(), "context cancelled")

	// Cleanup: unblock the first handler
	close(blockCh)
	wg.Wait()
}

func TestWithThrottlerNilMessage(t *testing.T) {
	handler := HandlerFunc(func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
		return tgbotapi.MessageConfig{}, nil
	})

	throttled := WithThrottler(1)(handler)
	_, err := throttled.Handle(context.Background(), nil)

	assert.Error(t, err)
	assert.Equal(t, "message is nil", err.Error(), "should handle nil message")
}

func TestWithThrottlerHandlerError(t *testing.T) {
	expectedErr := errors.New("handler error")
	handler := HandlerFunc(func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
		return tgbotapi.MessageConfig{}, expectedErr
	})

	throttled := WithThrottler(1)(handler)
	_, err := throttled.Handle(context.Background(), &tgbotapi.Message{})

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err, "should propagate handler error")
}

func TestWithThrottlerReleasesSlots(t *testing.T) {
	handler := HandlerFunc(func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
		return tgbotapi.MessageConfig{}, nil
	})

	throttled := WithThrottler(1)(handler)

	// First call should succeed
	_, err1 := throttled.Handle(context.Background(), &tgbotapi.Message{})
	assert.NoError(t, err1, "first call should succeed")

	// Second call should also succeed because slot was released
	_, err2 := throttled.Handle(context.Background(), &tgbotapi.Message{})
	assert.NoError(t, err2, "second call should succeed after slot is released")
}
