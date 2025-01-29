package middleware

import (
	"context"
	"errors"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
)

type testHandler struct {
	err      error
	response tgbotapi.MessageConfig
}

func (h *testHandler) Handle(_ context.Context, _ *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	return h.response, h.err
}

func TestUse(t *testing.T) {
	type testCase struct {
		handler         Handler
		expectedErr     error
		message         *tgbotapi.Message
		name            string
		middlewares     []Middleware
		expectedMessage tgbotapi.MessageConfig
	}

	tests := []testCase{
		{
			name:    "no_middlewares",
			handler: &testHandler{response: tgbotapi.MessageConfig{}, err: nil},
			message: &tgbotapi.Message{Text: "test"},
			expectedMessage: tgbotapi.MessageConfig{
				Text: "",
			},
			expectedErr: nil,
		},
		{
			name: "single_middleware_modifies_response",
			handler: &testHandler{
				response: tgbotapi.MessageConfig{Text: "hello"},
				err:      nil,
			},
			middlewares: []Middleware{
				func(next Handler) Handler {
					return HandlerFunc(func(ctx context.Context, message *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
						res, err := next.Handle(ctx, message)
						res.Text = res.Text + " world"
						return res, err
					})
				},
			},
			message: &tgbotapi.Message{Text: "test"},
			expectedMessage: tgbotapi.MessageConfig{
				Text: "hello world",
			},
			expectedErr: nil,
		},
		{
			name: "multiple_middlewares_applied_in_order",
			handler: &testHandler{
				response: tgbotapi.MessageConfig{Text: "start"},
				err:      nil,
			},
			middlewares: []Middleware{
				func(next Handler) Handler {
					return HandlerFunc(func(ctx context.Context, message *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
						res, err := next.Handle(ctx, message)
						res.Text = res.Text + " middle"
						return res, err
					})
				},
				func(next Handler) Handler {
					return HandlerFunc(func(ctx context.Context, message *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
						res, err := next.Handle(ctx, message)
						res.Text = res.Text + " end"
						return res, err
					})
				},
			},
			message: &tgbotapi.Message{Text: "test"},
			expectedMessage: tgbotapi.MessageConfig{
				Text: "start middle end",
			},
			expectedErr: nil,
		},
		{
			name: "middleware_returns_error",
			handler: &testHandler{
				response: tgbotapi.MessageConfig{Text: ""},
				err:      nil,
			},
			middlewares: []Middleware{
				func(next Handler) Handler {
					return HandlerFunc(func(ctx context.Context, message *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
						return tgbotapi.MessageConfig{}, errors.New("middleware error")
					})
				},
			},
			message: &tgbotapi.Message{Text: "test"},
			expectedMessage: tgbotapi.MessageConfig{
				Text: "",
			},
			expectedErr: errors.New("middleware error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := Use(tc.handler, tc.middlewares...)
			res, err := handler.Handle(context.Background(), tc.message)
			assert.Equal(t, tc.expectedMessage, res)
			assert.Equal(t, tc.expectedErr, err)
		})
	}
}
