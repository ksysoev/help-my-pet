// Code generated by mockery v2.50.4. DO NOT EDIT.

//go:build !compile

package bot

import (
	context "context"

	core "github.com/ksysoev/help-my-pet/pkg/core"
	mock "github.com/stretchr/testify/mock"
)

// MockAIProvider is an autogenerated mock type for the AIProvider type
type MockAIProvider struct {
	mock.Mock
}

type MockAIProvider_Expecter struct {
	mock *mock.Mock
}

func (_m *MockAIProvider) EXPECT() *MockAIProvider_Expecter {
	return &MockAIProvider_Expecter{mock: &_m.Mock}
}

// GetPetAdvice provides a mock function with given fields: ctx, request
func (_m *MockAIProvider) GetPetAdvice(ctx context.Context, request *core.UserMessage) (*core.PetAdviceResponse, error) {
	ret := _m.Called(ctx, request)

	if len(ret) == 0 {
		panic("no return value specified for GetPetAdvice")
	}

	var r0 *core.PetAdviceResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *core.UserMessage) (*core.PetAdviceResponse, error)); ok {
		return rf(ctx, request)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *core.UserMessage) *core.PetAdviceResponse); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*core.PetAdviceResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *core.UserMessage) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockAIProvider_GetPetAdvice_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetPetAdvice'
type MockAIProvider_GetPetAdvice_Call struct {
	*mock.Call
}

// GetPetAdvice is a helper method to define mock.On call
//   - ctx context.Context
//   - request *core.UserMessage
func (_e *MockAIProvider_Expecter) GetPetAdvice(ctx interface{}, request interface{}) *MockAIProvider_GetPetAdvice_Call {
	return &MockAIProvider_GetPetAdvice_Call{Call: _e.mock.On("GetPetAdvice", ctx, request)}
}

func (_c *MockAIProvider_GetPetAdvice_Call) Run(run func(ctx context.Context, request *core.UserMessage)) *MockAIProvider_GetPetAdvice_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*core.UserMessage))
	})
	return _c
}

func (_c *MockAIProvider_GetPetAdvice_Call) Return(_a0 *core.PetAdviceResponse, _a1 error) *MockAIProvider_GetPetAdvice_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockAIProvider_GetPetAdvice_Call) RunAndReturn(run func(context.Context, *core.UserMessage) (*core.PetAdviceResponse, error)) *MockAIProvider_GetPetAdvice_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockAIProvider creates a new instance of MockAIProvider. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockAIProvider(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockAIProvider {
	mock := &MockAIProvider{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
