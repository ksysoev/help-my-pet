package conversation

import (
	"context"
	"sync"
)

// requestState represents the state of a chat request
type requestState struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// RequestManager handles tracking and cancellation of ongoing requests per chat
type RequestManager struct {
	requests map[int64]*requestState
	mu       sync.RWMutex
}

// NewRequestManager creates a new RequestManager instance
func NewRequestManager() *RequestManager {
	return &RequestManager{
		requests: make(map[int64]*requestState),
	}
}

// StartRequest starts tracking a new request for a chat and returns a cancel function
func (m *RequestManager) StartRequest(chatID int64) context.CancelFunc {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Cancel any existing request for this chat
	m.CancelPreviousRequest(chatID)

	// Create new context and cancel function
	ctx, cancel := context.WithCancel(context.Background())
	m.requests[chatID] = &requestState{
		ctx:    ctx,
		cancel: cancel,
	}

	// Cleanup when context is done
	go func() {
		<-ctx.Done()
		m.mu.Lock()
		if state, exists := m.requests[chatID]; exists && state.ctx == ctx {
			delete(m.requests, chatID)
		}
		m.mu.Unlock()
	}()

	return cancel
}

// CancelPreviousRequest cancels any existing request for the given chat
func (m *RequestManager) CancelPreviousRequest(chatID int64) {
	if state, exists := m.requests[chatID]; exists {
		state.cancel()
		delete(m.requests, chatID)
	}
}
