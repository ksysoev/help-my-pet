package conversation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRequestManager_StartRequest(t *testing.T) {
	t.Run("starts new request", func(t *testing.T) {
		rm := NewRequestManager()
		chatID := int64(123)

		cancel := rm.StartRequest(chatID)
		assert.NotNil(t, cancel)

		// Verify request is tracked
		rm.mu.RLock()
		state, exists := rm.requests[chatID]
		assert.True(t, exists)
		assert.NotNil(t, state)
		assert.NotNil(t, state.ctx)
		assert.NotNil(t, state.cancel)
		rm.mu.RUnlock()

		// Cleanup
		cancel()

		// Wait for cleanup goroutine
		time.Sleep(10 * time.Millisecond)

		// Verify request was removed
		rm.mu.RLock()
		state, exists = rm.requests[chatID]
		assert.False(t, exists)
		assert.Nil(t, state)
		rm.mu.RUnlock()
	})

	t.Run("cancels previous request when starting new one", func(t *testing.T) {
		rm := NewRequestManager()
		chatID := int64(123)

		// Start first request
		cancel1 := rm.StartRequest(chatID)
		assert.NotNil(t, cancel1)

		// Start second request
		cancel2 := rm.StartRequest(chatID)
		assert.NotNil(t, cancel2)

		// Wait for cleanup goroutine
		time.Sleep(10 * time.Millisecond)

		// Call first cancel, which should be a no-op since request was already cancelled
		cancel1()

		// Verify only second request exists
		rm.mu.RLock()
		state, exists := rm.requests[chatID]
		assert.True(t, exists)
		assert.NotNil(t, state)
		assert.NotNil(t, state.ctx)
		assert.NotNil(t, state.cancel)
		rm.mu.RUnlock()

		// Cleanup
		cancel2()
	})
}

func TestRequestManager_CancelPreviousRequest(t *testing.T) {
	t.Run("cancels existing request", func(t *testing.T) {
		rm := NewRequestManager()
		chatID := int64(123)

		// Start request
		cancel := rm.StartRequest(chatID)
		assert.NotNil(t, cancel)

		// Verify request exists
		rm.mu.RLock()
		state, exists := rm.requests[chatID]
		assert.True(t, exists)
		assert.NotNil(t, state)
		assert.NotNil(t, state.ctx)
		assert.NotNil(t, state.cancel)
		rm.mu.RUnlock()

		// Cancel request
		rm.CancelPreviousRequest(chatID)

		// Wait for cleanup goroutine
		time.Sleep(10 * time.Millisecond)

		// Verify request was removed
		rm.mu.RLock()
		state, exists = rm.requests[chatID]
		assert.False(t, exists)
		assert.Nil(t, state)
		rm.mu.RUnlock()
	})

	t.Run("handles non-existent chat ID", func(t *testing.T) {
		rm := NewRequestManager()
		chatID := int64(123)

		// Should not panic
		rm.CancelPreviousRequest(chatID)

		rm.mu.RLock()
		assert.Len(t, rm.requests, 0)
		rm.mu.RUnlock()
	})
}
