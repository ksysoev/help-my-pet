package bot

import (
	"context"
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
	handler := func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
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
	}

	// Create throttled handler with limit of 5 for testing
	throttled := withThrottler(5)(handler)

	// Send 10 concurrent requests
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = throttled(context.Background(), &tgbotapi.Message{})
		}()
	}

	wg.Wait()

	assert.LessOrEqual(t, maxCount, 5, "concurrent processing exceeded limit")
}

func TestWithThrottlerHandlesContextCancellation(t *testing.T) {
	handler := func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
		time.Sleep(100 * time.Millisecond)
		return tgbotapi.MessageConfig{}, nil
	}

	throttled := withThrottler(1)(handler)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := throttled(ctx, &tgbotapi.Message{})
	assert.Error(t, err, "should return error when context is cancelled")
}

func TestWithThrottlerReleasesSlots(t *testing.T) {
	handler := func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
		return tgbotapi.MessageConfig{}, nil
	}

	throttled := withThrottler(1)(handler)

	// First call should succeed
	_, err1 := throttled(context.Background(), &tgbotapi.Message{})
	assert.NoError(t, err1, "first call should succeed")

	// Second call should also succeed because slot was released
	_, err2 := throttled(context.Background(), &tgbotapi.Message{})
	assert.NoError(t, err2, "second call should succeed after slot is released")
}
