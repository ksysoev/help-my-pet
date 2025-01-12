// Code generated by mockery v2.50.2. DO NOT EDIT.

//go:build !compile

package bot

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// MockRateLimiter is an autogenerated mock type for the RateLimiter type
type MockRateLimiter struct {
	mock.Mock
}

type MockRateLimiter_Expecter struct {
	mock *mock.Mock
}

func (_m *MockRateLimiter) EXPECT() *MockRateLimiter_Expecter {
	return &MockRateLimiter_Expecter{mock: &_m.Mock}
}

// IsAllowed provides a mock function with given fields: ctx, userID
func (_m *MockRateLimiter) IsAllowed(ctx context.Context, userID int64) (bool, error) {
	ret := _m.Called(ctx, userID)

	if len(ret) == 0 {
		panic("no return value specified for IsAllowed")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) (bool, error)); ok {
		return rf(ctx, userID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64) bool); ok {
		r0 = rf(ctx, userID)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64) error); ok {
		r1 = rf(ctx, userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockRateLimiter_IsAllowed_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'IsAllowed'
type MockRateLimiter_IsAllowed_Call struct {
	*mock.Call
}

// IsAllowed is a helper method to define mock.On call
//   - ctx context.Context
//   - userID int64
func (_e *MockRateLimiter_Expecter) IsAllowed(ctx interface{}, userID interface{}) *MockRateLimiter_IsAllowed_Call {
	return &MockRateLimiter_IsAllowed_Call{Call: _e.mock.On("IsAllowed", ctx, userID)}
}

func (_c *MockRateLimiter_IsAllowed_Call) Run(run func(ctx context.Context, userID int64)) *MockRateLimiter_IsAllowed_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(int64))
	})
	return _c
}

func (_c *MockRateLimiter_IsAllowed_Call) Return(_a0 bool, _a1 error) *MockRateLimiter_IsAllowed_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockRateLimiter_IsAllowed_Call) RunAndReturn(run func(context.Context, int64) (bool, error)) *MockRateLimiter_IsAllowed_Call {
	_c.Call.Return(run)
	return _c
}

// RecordAccess provides a mock function with given fields: ctx, userID
func (_m *MockRateLimiter) RecordAccess(ctx context.Context, userID int64) error {
	ret := _m.Called(ctx, userID)

	if len(ret) == 0 {
		panic("no return value specified for RecordAccess")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) error); ok {
		r0 = rf(ctx, userID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockRateLimiter_RecordAccess_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RecordAccess'
type MockRateLimiter_RecordAccess_Call struct {
	*mock.Call
}

// RecordAccess is a helper method to define mock.On call
//   - ctx context.Context
//   - userID int64
func (_e *MockRateLimiter_Expecter) RecordAccess(ctx interface{}, userID interface{}) *MockRateLimiter_RecordAccess_Call {
	return &MockRateLimiter_RecordAccess_Call{Call: _e.mock.On("RecordAccess", ctx, userID)}
}

func (_c *MockRateLimiter_RecordAccess_Call) Run(run func(ctx context.Context, userID int64)) *MockRateLimiter_RecordAccess_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(int64))
	})
	return _c
}

func (_c *MockRateLimiter_RecordAccess_Call) Return(_a0 error) *MockRateLimiter_RecordAccess_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockRateLimiter_RecordAccess_Call) RunAndReturn(run func(context.Context, int64) error) *MockRateLimiter_RecordAccess_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockRateLimiter creates a new instance of MockRateLimiter. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockRateLimiter(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockRateLimiter {
	mock := &MockRateLimiter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
