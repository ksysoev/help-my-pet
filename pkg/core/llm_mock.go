// Code generated by mockery v2.50.4. DO NOT EDIT.

//go:build !compile

package core

import (
	context "context"

	message "github.com/ksysoev/help-my-pet/pkg/core/message"
	mock "github.com/stretchr/testify/mock"
)

// MockLLM is an autogenerated mock type for the LLM type
type MockLLM struct {
	mock.Mock
}

type MockLLM_Expecter struct {
	mock *mock.Mock
}

func (_m *MockLLM) EXPECT() *MockLLM_Expecter {
	return &MockLLM_Expecter{mock: &_m.Mock}
}

// Call provides a mock function with given fields: ctx, prompt
func (_m *MockLLM) Call(ctx context.Context, prompt string) (*message.LLMResult, error) {
	ret := _m.Called(ctx, prompt)

	if len(ret) == 0 {
		panic("no return value specified for Call")
	}

	var r0 *message.LLMResult
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*message.LLMResult, error)); ok {
		return rf(ctx, prompt)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *message.LLMResult); ok {
		r0 = rf(ctx, prompt)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*message.LLMResult)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, prompt)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockLLM_Call_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Call'
type MockLLM_Call_Call struct {
	*mock.Call
}

// Call is a helper method to define mock.On call
//   - ctx context.Context
//   - prompt string
func (_e *MockLLM_Expecter) Call(ctx interface{}, prompt interface{}) *MockLLM_Call_Call {
	return &MockLLM_Call_Call{Call: _e.mock.On("Call", ctx, prompt)}
}

func (_c *MockLLM_Call_Call) Run(run func(ctx context.Context, prompt string)) *MockLLM_Call_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockLLM_Call_Call) Return(_a0 *message.LLMResult, _a1 error) *MockLLM_Call_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockLLM_Call_Call) RunAndReturn(run func(context.Context, string) (*message.LLMResult, error)) *MockLLM_Call_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockLLM creates a new instance of MockLLM. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockLLM(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockLLM {
	mock := &MockLLM{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
