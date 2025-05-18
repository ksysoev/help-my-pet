package middleware

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func TestWithRequestSequencerLimitsPerUserProcessing(t *testing.T) {
	var (
		mu               sync.Mutex
		activeUserCounts = make(map[int64]int)
		maxUserCounts    = make(map[int64]int)
		wg               sync.WaitGroup
	)

	// Create a handler that tracks concurrent executions per user
	handler := HandlerFunc(func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
		userID := msg.From.ID

		mu.Lock()
		activeUserCounts[userID]++
		if activeUserCounts[userID] > maxUserCounts[userID] {
			maxUserCounts[userID] = activeUserCounts[userID]
		}
		mu.Unlock()

		// Simulate work
		time.Sleep(50 * time.Millisecond)

		mu.Lock()
		activeUserCounts[userID]--
		mu.Unlock()

		return tgbotapi.MessageConfig{}, nil
	})

	// Create sequenced handler
	sequenced := WithRequestSequencer()(handler)

	// Send multiple concurrent requests for different users
	userIDs := []int64{1, 2, 3}
	for _, userID := range userIDs {
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(id int64) {
				defer wg.Done()
				msg := &tgbotapi.Message{
					From: &tgbotapi.User{ID: id},
				}
				_, _ = sequenced.Handle(context.Background(), msg)
			}(userID)
		}
	}

	wg.Wait()

	// Verify that each user had at most 1 concurrent request
	for _, userID := range userIDs {
		if maxUserCounts[userID] > 1 {
			t.Errorf("concurrent processing for user %d exceeded limit of 1: got %d", userID, maxUserCounts[userID])
		}
	}
}

func TestWithRequestSequencerHandlesContextCancellation(t *testing.T) {
	// Create a handler that blocks until explicitly unblocked
	blockCh := make(chan struct{})
	handler := HandlerFunc(func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
		<-blockCh // Block until channel is closed
		return tgbotapi.MessageConfig{}, nil
	})

	// Create sequenced handler
	sequenced := WithRequestSequencer()(handler)

	// User ID for testing
	userID := int64(123)

	// Fill up the sequencer for this user
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		msg := &tgbotapi.Message{
			From: &tgbotapi.User{ID: userID},
		}
		_, _ = sequenced.Handle(context.Background(), msg)
	}()

	// Wait a bit to ensure the first request has acquired the slot
	time.Sleep(50 * time.Millisecond)

	// Try another request for the same user with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	msg := &tgbotapi.Message{
		From: &tgbotapi.User{ID: userID},
	}
	_, err := sequenced.Handle(ctx, msg)
	if err == nil {
		t.Error("expected error when context is cancelled, got nil")
	}
	if err != nil && !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled error, got %v", err)
	}

	// Cleanup: unblock the first handler
	close(blockCh)
	wg.Wait()
}

func TestWithRequestSequencerHandlerError(t *testing.T) {
	expectedErr := errors.New("handler error")
	handler := HandlerFunc(func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
		return tgbotapi.MessageConfig{}, expectedErr
	})

	sequenced := WithRequestSequencer()(handler)
	_, err := sequenced.Handle(context.Background(), &tgbotapi.Message{
		From: &tgbotapi.User{ID: 123},
	})

	if err == nil {
		t.Error("expected error from handler, got nil")
	}
	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}

func TestWithRequestSequencerNilMessage(t *testing.T) {
	handler := HandlerFunc(func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
		return tgbotapi.MessageConfig{}, nil
	})

	sequenced := WithRequestSequencer()(handler)
	_, err := sequenced.Handle(context.Background(), nil)

	if err == nil {
		t.Error("expected error for nil message, got nil")
	}
}

func TestWithRequestSequencerNilUser(t *testing.T) {
	handler := HandlerFunc(func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
		return tgbotapi.MessageConfig{}, nil
	})

	sequenced := WithRequestSequencer()(handler)
	_, err := sequenced.Handle(context.Background(), &tgbotapi.Message{})

	if err == nil {
		t.Error("expected error for nil user, got nil")
	}
}
