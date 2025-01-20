package bot

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ksysoev/help-my-pet/pkg/i18n"
	"github.com/stretchr/testify/assert"
)

func TestWithErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		handler       Handler
		getMessage    func(lang string, msgType i18n.Message) string
		message       *tgbotapi.Message
		expectedError error
		expectedMsg   string
		checkLang     func(t *testing.T, lang string)
	}{
		{
			name: "handles error from handler",
			handler: func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
				return tgbotapi.MessageConfig{}, errors.New("handler error")
			},
			getMessage: func(lang string, msgType i18n.Message) string {
				return "error message"
			},
			message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: 123},
				From: &tgbotapi.User{LanguageCode: "en"},
			},
			expectedError: nil,
			expectedMsg:   "error message",
			checkLang: func(t *testing.T, lang string) {
				assert.Equal(t, "en", lang)
			},
		},
		{
			name: "passes through successful response",
			handler: func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
				return tgbotapi.NewMessage(123, "success"), nil
			},
			getMessage: func(lang string, msgType i18n.Message) string {
				return "error message"
			},
			message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: 123},
				From: &tgbotapi.User{LanguageCode: "en"},
			},
			expectedError: nil,
			expectedMsg:   "success",
			checkLang:     func(t *testing.T, lang string) {},
		},
		{
			name: "handles message without From field",
			handler: func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
				return tgbotapi.MessageConfig{}, errors.New("handler error")
			},
			getMessage: func(lang string, msgType i18n.Message) string {
				return "error message"
			},
			message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: 123},
			},
			expectedError: nil,
			expectedMsg:   "error message",
			checkLang: func(t *testing.T, lang string) {
				assert.Equal(t, "", lang)
			},
		},
		{
			name: "handles message with empty language code",
			handler: func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
				return tgbotapi.MessageConfig{}, errors.New("handler error")
			},
			getMessage: func(lang string, msgType i18n.Message) string {
				return "error message"
			},
			message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: 123},
				From: &tgbotapi.User{},
			},
			expectedError: nil,
			expectedMsg:   "error message",
			checkLang: func(t *testing.T, lang string) {
				assert.Equal(t, "", lang)
			},
		},
		{
			name: "handles context cancellation",
			handler: func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
				return tgbotapi.MessageConfig{}, context.Canceled
			},
			getMessage: func(lang string, msgType i18n.Message) string {
				return "error message"
			},
			message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: 123},
				From: &tgbotapi.User{LanguageCode: "en"},
			},
			expectedError: nil,
			expectedMsg:   "error message",
			checkLang: func(t *testing.T, lang string) {
				assert.Equal(t, "en", lang)
			},
		},
		{
			name: "handles nil chat",
			handler: func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
				return tgbotapi.MessageConfig{}, errors.New("handler error")
			},
			getMessage: func(lang string, msgType i18n.Message) string {
				return "error message"
			},
			message:       &tgbotapi.Message{},
			expectedError: nil,
			expectedMsg:   "error message",
			checkLang: func(t *testing.T, lang string) {
				assert.Equal(t, "", lang)
			},
		},
		{
			name: "handles nil message",
			handler: func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
				return tgbotapi.MessageConfig{}, errors.New("handler error")
			},
			getMessage: func(lang string, msgType i18n.Message) string {
				return "error message"
			},
			message:       nil,
			expectedError: errors.New("message is nil"),
			expectedMsg:   "",
			checkLang:     func(t *testing.T, lang string) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedLang string
			wrappedGetMessage := func(lang string, msgType i18n.Message) string {
				capturedLang = lang
				return tt.getMessage(lang, msgType)
			}

			handler := withErrorHandling(wrappedGetMessage, tt.handler)
			msgConfig, err := handler(context.Background(), tt.message)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedMsg, msgConfig.Text)
				tt.checkLang(t, capturedLang)
			}
		})
	}
}

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
	// Create a handler that blocks until explicitly unblocked
	blockCh := make(chan struct{})
	handler := func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
		<-blockCh // Block until channel is closed
		return tgbotapi.MessageConfig{}, nil
	}

	// Create throttled handler with limit of 1
	throttled := withThrottler(1)(handler)

	// Fill up the throttler
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = throttled(context.Background(), &tgbotapi.Message{})
	}()

	// Wait a bit to ensure the first request has acquired the slot
	time.Sleep(50 * time.Millisecond)

	// Try another request with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := throttled(ctx, &tgbotapi.Message{})
	assert.Error(t, err, "should return error when context is cancelled")
	assert.Contains(t, err.Error(), "context cancelled")

	// Cleanup: unblock the first handler
	close(blockCh)
	wg.Wait()
}

func TestWithThrottlerNilMessage(t *testing.T) {
	handler := func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
		return tgbotapi.MessageConfig{}, nil
	}

	throttled := withThrottler(1)(handler)
	_, err := throttled(context.Background(), nil)

	assert.Error(t, err)
	assert.Equal(t, "message is nil", err.Error(), "should handle nil message")
}

func TestWithThrottlerHandlerError(t *testing.T) {
	expectedErr := errors.New("handler error")
	handler := func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
		return tgbotapi.MessageConfig{}, expectedErr
	}

	throttled := withThrottler(1)(handler)
	_, err := throttled(context.Background(), &tgbotapi.Message{})

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err, "should propagate handler error")
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
