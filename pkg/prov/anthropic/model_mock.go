// Code generated by mockery v2.50.4. DO NOT EDIT.

//go:build !compile

package anthropic

import (
	context "context"

	message "github.com/ksysoev/help-my-pet/pkg/core/message"
	mock "github.com/stretchr/testify/mock"
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

// Analyze provides a mock function with given fields: ctx, request, imgs
func (_m *MockModel) Analyze(ctx context.Context, request string, imgs []*message.Image) (*message.LLMResult, error) {
	ret := _m.Called(ctx, request, imgs)

	if len(ret) == 0 {
		panic("no return value specified for Analyze")
	}

	var r0 *message.LLMResult
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, []*message.Image) (*message.LLMResult, error)); ok {
		return rf(ctx, request, imgs)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, []*message.Image) *message.LLMResult); ok {
		r0 = rf(ctx, request, imgs)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*message.LLMResult)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, []*message.Image) error); ok {
		r1 = rf(ctx, request, imgs)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockModel_Analyze_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Analyze'
type MockModel_Analyze_Call struct {
	*mock.Call
}

// Analyze is a helper method to define mock.On call
//   - ctx context.Context
//   - request string
//   - imgs []*message.Image
func (_e *MockModel_Expecter) Analyze(ctx interface{}, request interface{}, imgs interface{}) *MockModel_Analyze_Call {
	return &MockModel_Analyze_Call{Call: _e.mock.On("Analyze", ctx, request, imgs)}
}

func (_c *MockModel_Analyze_Call) Run(run func(ctx context.Context, request string, imgs []*message.Image)) *MockModel_Analyze_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].([]*message.Image))
	})
	return _c
}

func (_c *MockModel_Analyze_Call) Return(_a0 *message.LLMResult, _a1 error) *MockModel_Analyze_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockModel_Analyze_Call) RunAndReturn(run func(context.Context, string, []*message.Image) (*message.LLMResult, error)) *MockModel_Analyze_Call {
	_c.Call.Return(run)
	return _c
}

// Report provides a mock function with given fields: ctx, request
func (_m *MockModel) Report(ctx context.Context, request string) (*message.LLMResult, error) {
	ret := _m.Called(ctx, request)

	if len(ret) == 0 {
		panic("no return value specified for Report")
	}

	var r0 *message.LLMResult
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*message.LLMResult, error)); ok {
		return rf(ctx, request)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *message.LLMResult); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*message.LLMResult)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockModel_Report_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Report'
type MockModel_Report_Call struct {
	*mock.Call
}

// Report is a helper method to define mock.On call
//   - ctx context.Context
//   - request string
func (_e *MockModel_Expecter) Report(ctx interface{}, request interface{}) *MockModel_Report_Call {
	return &MockModel_Report_Call{Call: _e.mock.On("Report", ctx, request)}
}

func (_c *MockModel_Report_Call) Run(run func(ctx context.Context, request string)) *MockModel_Report_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockModel_Report_Call) Return(_a0 *message.LLMResult, _a1 error) *MockModel_Report_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockModel_Report_Call) RunAndReturn(run func(context.Context, string) (*message.LLMResult, error)) *MockModel_Report_Call {
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
