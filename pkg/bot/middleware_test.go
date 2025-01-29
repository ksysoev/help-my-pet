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
		expectedError error
		handler       Handler
		getMessage    func(lang string, msgType i18n.Message) string
		message       *tgbotapi.Message
		checkLang     func(t *testing.T, lang string)
		name          string
		expectedMsg   string
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

			handler := WithErrorHandling(wrappedGetMessage, tt.handler)
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
	throttled := WithThrottler(5)(handler)

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
	throttled := WithThrottler(1)(handler)

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

	throttled := WithThrottler(1)(handler)
	_, err := throttled(context.Background(), nil)

	assert.Error(t, err)
	assert.Equal(t, "message is nil", err.Error(), "should handle nil message")
}

func TestWithThrottlerHandlerError(t *testing.T) {
	expectedErr := errors.New("handler error")
	handler := func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
		return tgbotapi.MessageConfig{}, expectedErr
	}

	throttled := WithThrottler(1)(handler)
	_, err := throttled(context.Background(), &tgbotapi.Message{})

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err, "should propagate handler error")
}

func TestWithThrottlerReleasesSlots(t *testing.T) {
	handler := func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
		return tgbotapi.MessageConfig{}, nil
	}

	throttled := WithThrottler(1)(handler)

	// First call should succeed
	_, err1 := throttled(context.Background(), &tgbotapi.Message{})
	assert.NoError(t, err1, "first call should succeed")

	// Second call should also succeed because slot was released
	_, err2 := throttled(context.Background(), &tgbotapi.Message{})
	assert.NoError(t, err2, "second call should succeed after slot is released")
}

func TestWithRequestReducerNilMessage(t *testing.T) {
	handler := func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
		return tgbotapi.MessageConfig{}, nil
	}

	middleware := WithRequestReducer()
	wrapped := middleware(handler)

	_, err := wrapped(context.Background(), nil)
	assert.Error(t, err)
	assert.Equal(t, "message is nil", err.Error())
}

func TestWithRequestReducerCancelsPreviousRequest(t *testing.T) {
	var (
		firstCtxCancelled bool
		wg                sync.WaitGroup
	)

	handler := func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
		select {
		case <-ctx.Done():
			firstCtxCancelled = true
		case <-time.After(200 * time.Millisecond):
		}
		return tgbotapi.MessageConfig{}, nil
	}

	middleware := WithRequestReducer()
	wrapped := middleware(handler)

	// Start first request
	wg.Add(1)
	go func() {
		defer wg.Done()
		msg := &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: 123}, MessageID: 1}
		_, _ = wrapped(context.Background(), msg)
	}()

	// Give time for first request to start
	time.Sleep(50 * time.Millisecond)

	// Send second request from same chat
	msg2 := &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: 123}, MessageID: 2}
	_, _ = wrapped(context.Background(), msg2)

	wg.Wait()
	assert.True(t, firstCtxCancelled, "first request should have been cancelled")
}

func TestWithRequestReducerAllowsConcurrentRequestsFromDifferentChats(t *testing.T) {
	var (
		completedRequests sync.Map
		wg                sync.WaitGroup
	)

	handler := func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
		time.Sleep(100 * time.Millisecond)
		completedRequests.Store(msg.Chat.ID, true)
		return tgbotapi.MessageConfig{}, nil
	}

	middleware := WithRequestReducer()
	wrapped := middleware(handler)

	// Start concurrent requests from different chats
	chatIDs := []int64{123, 456}
	for _, chatID := range chatIDs {
		wg.Add(1)
		go func(id int64) {
			defer wg.Done()
			msg := &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: id}, MessageID: 1}
			_, _ = wrapped(context.Background(), msg)
		}(chatID)
	}

	wg.Wait()

	// Verify both requests completed
	for _, chatID := range chatIDs {
		completed, ok := completedRequests.Load(chatID)
		assert.True(t, ok, "request for chat %d should have completed", chatID)
		assert.True(t, completed.(bool))
	}
}

func TestWithRequestReducerCleansUpAfterCompletion(t *testing.T) {
	handler := func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
		return tgbotapi.MessageConfig{}, nil
	}

	middleware := WithRequestReducer()
	wrapped := middleware(handler)

	msg := &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: 123}, MessageID: 1}

	// First request
	_, err1 := wrapped(context.Background(), msg)
	assert.NoError(t, err1)

	// Wait for cleanup
	time.Sleep(50 * time.Millisecond)

	// Second request should work
	_, err2 := wrapped(context.Background(), msg)
	assert.NoError(t, err2)
}
