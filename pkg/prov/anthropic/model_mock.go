// Code generated by mockery v2.50.4. DO NOT EDIT.

//go:build !compile

package anthropic

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
	llms "github.com/tmc/langchaingo/llms"
)

// MockModel is an autogenerated mock type for the Model type
type MockModel struct {
	mock.Mock
}

type MockModel_Expecter struct {
	mock *mock.Mock
}

func (_m *MockModel) EXPECT() *MockModel_Expecter {
	return &MockModel_Expecter{mock: &_m.Mock}
}

// Call provides a mock function with given fields: ctx, prompt, options
func (_m *MockModel) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, prompt)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for Call")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, ...llms.CallOption) (string, error)); ok {
		return rf(ctx, prompt, options...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, ...llms.CallOption) string); ok {
		r0 = rf(ctx, prompt, options...)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, ...llms.CallOption) error); ok {
		r1 = rf(ctx, prompt, options...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockModel_Call_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Call'
type MockModel_Call_Call struct {
	*mock.Call
}

// Call is a helper method to define mock.On call
//   - ctx context.Context
//   - prompt string
//   - options ...llms.CallOption
func (_e *MockModel_Expecter) Call(ctx interface{}, prompt interface{}, options ...interface{}) *MockModel_Call_Call {
	return &MockModel_Call_Call{Call: _e.mock.On("Call",
		append([]interface{}{ctx, prompt}, options...)...)}
}

func (_c *MockModel_Call_Call) Run(run func(ctx context.Context, prompt string, options ...llms.CallOption)) *MockModel_Call_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]llms.CallOption, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(llms.CallOption)
			}
		}
		run(args[0].(context.Context), args[1].(string), variadicArgs...)
	})
	return _c
}

func (_c *MockModel_Call_Call) Return(_a0 string, _a1 error) *MockModel_Call_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockModel_Call_Call) RunAndReturn(run func(context.Context, string, ...llms.CallOption) (string, error)) *MockModel_Call_Call {
	_c.Call.Return(run)
	return _c
}

// GenerateContent provides a mock function with given fields: ctx, messages, options
func (_m *MockModel) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, messages)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for GenerateContent")
	}

	var r0 *llms.ContentResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, []llms.MessageContent, ...llms.CallOption) (*llms.ContentResponse, error)); ok {
		return rf(ctx, messages, options...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, []llms.MessageContent, ...llms.CallOption) *llms.ContentResponse); ok {
		r0 = rf(ctx, messages, options...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*llms.ContentResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, []llms.MessageContent, ...llms.CallOption) error); ok {
		r1 = rf(ctx, messages, options...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockModel_GenerateContent_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GenerateContent'
type MockModel_GenerateContent_Call struct {
	*mock.Call
}

// GenerateContent is a helper method to define mock.On call
//   - ctx context.Context
//   - messages []llms.MessageContent
//   - options ...llms.CallOption
func (_e *MockModel_Expecter) GenerateContent(ctx interface{}, messages interface{}, options ...interface{}) *MockModel_GenerateContent_Call {
	return &MockModel_GenerateContent_Call{Call: _e.mock.On("GenerateContent",
		append([]interface{}{ctx, messages}, options...)...)}
}

func (_c *MockModel_GenerateContent_Call) Run(run func(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption)) *MockModel_GenerateContent_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]llms.CallOption, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(llms.CallOption)
			}
		}
		run(args[0].(context.Context), args[1].([]llms.MessageContent), variadicArgs...)
	})
	return _c
}

func (_c *MockModel_GenerateContent_Call) Return(_a0 *llms.ContentResponse, _a1 error) *MockModel_GenerateContent_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockModel_GenerateContent_Call) RunAndReturn(run func(context.Context, []llms.MessageContent, ...llms.CallOption) (*llms.ContentResponse, error)) *MockModel_GenerateContent_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockModel creates a new instance of MockModel. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockModel(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockModel {
	mock := &MockModel{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
