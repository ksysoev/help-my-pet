package middleware

import (
	"context"
	"sync"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
)

func TestWithRequestReducerNilMessage(t *testing.T) {
	handler := HandlerFunc(func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
		return tgbotapi.MessageConfig{}, nil
	})

	middleware := WithRequestReducer()
	wrapped := middleware(handler)

	_, err := wrapped.Handle(context.Background(), nil)
	assert.Error(t, err)
	assert.Equal(t, "message is nil", err.Error())
}

func TestWithRequestReducerCancelsPreviousRequest(t *testing.T) {
	var (
		firstCtxCancelled bool
		wg                sync.WaitGroup
	)

	handler := HandlerFunc(func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
		select {
		case <-ctx.Done():
			firstCtxCancelled = true
		case <-time.After(200 * time.Millisecond):
		}
		return tgbotapi.MessageConfig{}, nil
	})

	middleware := WithRequestReducer()
	wrapped := middleware(handler)

	// Start first request
	wg.Add(1)
	go func() {
		defer wg.Done()
		msg := &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: 123}, MessageID: 1}
		_, _ = wrapped.Handle(context.Background(), msg)
	}()

	// Give time for first request to start
	time.Sleep(50 * time.Millisecond)

	// Send second request from same chat
	msg2 := &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: 123}, MessageID: 2}
	_, _ = wrapped.Handle(context.Background(), msg2)

	wg.Wait()
	assert.True(t, firstCtxCancelled, "first request should have been cancelled")
}

func TestWithRequestReducerAllowsConcurrentRequestsFromDifferentChats(t *testing.T) {
	var (
		completedRequests sync.Map
		wg                sync.WaitGroup
	)

	handler := HandlerFunc(func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
		time.Sleep(100 * time.Millisecond)
		completedRequests.Store(msg.Chat.ID, true)
		return tgbotapi.MessageConfig{}, nil
	})

	middleware := WithRequestReducer()
	wrapped := middleware(handler)

	// Start concurrent requests from different chats
	chatIDs := []int64{123, 456}
	for _, chatID := range chatIDs {
		wg.Add(1)
		go func(id int64) {
			defer wg.Done()
			msg := &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: id}, MessageID: 1}
			_, _ = wrapped.Handle(context.Background(), msg)
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
	handler := HandlerFunc(func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
		return tgbotapi.MessageConfig{}, nil
	})

	middleware := WithRequestReducer()
	wrapped := middleware(handler)

	msg := &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: 123}, MessageID: 1}

	// First request
	_, err1 := wrapped.Handle(context.Background(), msg)
	assert.NoError(t, err1)

	// Wait for cleanup
	time.Sleep(50 * time.Millisecond)

	// Second request should work
	_, err2 := wrapped.Handle(context.Background(), msg)
	assert.NoError(t, err2)
}
