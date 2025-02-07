// Code generated by mockery v2.50.4. DO NOT EDIT.

//go:build !compile

package core

import (
	conversation "github.com/ksysoev/help-my-pet/pkg/core/conversation"
	message "github.com/ksysoev/help-my-pet/pkg/core/message"

	mock "github.com/stretchr/testify/mock"
)

// MockConversation is an autogenerated mock type for the Conversation type
type MockConversation struct {
	mock.Mock
}

type MockConversation_Expecter struct {
	mock *mock.Mock
}

func (_m *MockConversation) EXPECT() *MockConversation_Expecter {
	return &MockConversation_Expecter{mock: &_m.Mock}
}

// AddMessage provides a mock function with given fields: role, content
func (_m *MockConversation) AddMessage(role string, content string) {
	_m.Called(role, content)
}

// MockConversation_AddMessage_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AddMessage'
type MockConversation_AddMessage_Call struct {
	*mock.Call
}

// AddMessage is a helper method to define mock.On call
//   - role string
//   - content string
func (_e *MockConversation_Expecter) AddMessage(role interface{}, content interface{}) *MockConversation_AddMessage_Call {
	return &MockConversation_AddMessage_Call{Call: _e.mock.On("AddMessage", role, content)}
}

func (_c *MockConversation_AddMessage_Call) Run(run func(role string, content string)) *MockConversation_AddMessage_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string))
	})
	return _c
}

func (_c *MockConversation_AddMessage_Call) Return() *MockConversation_AddMessage_Call {
	_c.Call.Return()
	return _c
}

func (_c *MockConversation_AddMessage_Call) RunAndReturn(run func(string, string)) *MockConversation_AddMessage_Call {
	_c.Run(run)
	return _c
}

// AddQuestionAnswer provides a mock function with given fields: answer
func (_m *MockConversation) AddQuestionAnswer(answer string) (bool, error) {
	ret := _m.Called(answer)

	if len(ret) == 0 {
		panic("no return value specified for AddQuestionAnswer")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (bool, error)); ok {
		return rf(answer)
	}
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(answer)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(answer)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockConversation_AddQuestionAnswer_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AddQuestionAnswer'
type MockConversation_AddQuestionAnswer_Call struct {
	*mock.Call
}

// AddQuestionAnswer is a helper method to define mock.On call
//   - answer string
func (_e *MockConversation_Expecter) AddQuestionAnswer(answer interface{}) *MockConversation_AddQuestionAnswer_Call {
	return &MockConversation_AddQuestionAnswer_Call{Call: _e.mock.On("AddQuestionAnswer", answer)}
}

func (_c *MockConversation_AddQuestionAnswer_Call) Run(run func(answer string)) *MockConversation_AddQuestionAnswer_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *MockConversation_AddQuestionAnswer_Call) Return(_a0 bool, _a1 error) *MockConversation_AddQuestionAnswer_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockConversation_AddQuestionAnswer_Call) RunAndReturn(run func(string) (bool, error)) *MockConversation_AddQuestionAnswer_Call {
	_c.Call.Return(run)
	return _c
}

// GetCurrentQuestion provides a mock function with no fields
func (_m *MockConversation) GetCurrentQuestion() (*message.Question, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetCurrentQuestion")
	}

	var r0 *message.Question
	var r1 error
	if rf, ok := ret.Get(0).(func() (*message.Question, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() *message.Question); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*message.Question)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockConversation_GetCurrentQuestion_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetCurrentQuestion'
type MockConversation_GetCurrentQuestion_Call struct {
	*mock.Call
}

// GetCurrentQuestion is a helper method to define mock.On call
func (_e *MockConversation_Expecter) GetCurrentQuestion() *MockConversation_GetCurrentQuestion_Call {
	return &MockConversation_GetCurrentQuestion_Call{Call: _e.mock.On("GetCurrentQuestion")}
}

func (_c *MockConversation_GetCurrentQuestion_Call) Run(run func()) *MockConversation_GetCurrentQuestion_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockConversation_GetCurrentQuestion_Call) Return(_a0 *message.Question, _a1 error) *MockConversation_GetCurrentQuestion_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockConversation_GetCurrentQuestion_Call) RunAndReturn(run func() (*message.Question, error)) *MockConversation_GetCurrentQuestion_Call {
	_c.Call.Return(run)
	return _c
}

// GetID provides a mock function with no fields
func (_m *MockConversation) GetID() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetID")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockConversation_GetID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetID'
type MockConversation_GetID_Call struct {
	*mock.Call
}

// GetID is a helper method to define mock.On call
func (_e *MockConversation_Expecter) GetID() *MockConversation_GetID_Call {
	return &MockConversation_GetID_Call{Call: _e.mock.On("GetID")}
}

func (_c *MockConversation_GetID_Call) Run(run func()) *MockConversation_GetID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockConversation_GetID_Call) Return(_a0 string) *MockConversation_GetID_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockConversation_GetID_Call) RunAndReturn(run func() string) *MockConversation_GetID_Call {
	_c.Call.Return(run)
	return _c
}

// GetQuestionnaireResult provides a mock function with no fields
func (_m *MockConversation) GetQuestionnaireResult() ([]conversation.QuestionAnswer, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetQuestionnaireResult")
	}

	var r0 []conversation.QuestionAnswer
	var r1 error
	if rf, ok := ret.Get(0).(func() ([]conversation.QuestionAnswer, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() []conversation.QuestionAnswer); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]conversation.QuestionAnswer)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockConversation_GetQuestionnaireResult_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetQuestionnaireResult'
type MockConversation_GetQuestionnaireResult_Call struct {
	*mock.Call
}

// GetQuestionnaireResult is a helper method to define mock.On call
func (_e *MockConversation_Expecter) GetQuestionnaireResult() *MockConversation_GetQuestionnaireResult_Call {
	return &MockConversation_GetQuestionnaireResult_Call{Call: _e.mock.On("GetQuestionnaireResult")}
}

func (_c *MockConversation_GetQuestionnaireResult_Call) Run(run func()) *MockConversation_GetQuestionnaireResult_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockConversation_GetQuestionnaireResult_Call) Return(_a0 []conversation.QuestionAnswer, _a1 error) *MockConversation_GetQuestionnaireResult_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockConversation_GetQuestionnaireResult_Call) RunAndReturn(run func() ([]conversation.QuestionAnswer, error)) *MockConversation_GetQuestionnaireResult_Call {
	_c.Call.Return(run)
	return _c
}

// GetState provides a mock function with no fields
func (_m *MockConversation) GetState() conversation.ConversationState {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetState")
	}

	var r0 conversation.ConversationState
	if rf, ok := ret.Get(0).(func() conversation.ConversationState); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(conversation.ConversationState)
	}

	return r0
}

// MockConversation_GetState_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetState'
type MockConversation_GetState_Call struct {
	*mock.Call
}

// GetState is a helper method to define mock.On call
func (_e *MockConversation_Expecter) GetState() *MockConversation_GetState_Call {
	return &MockConversation_GetState_Call{Call: _e.mock.On("GetState")}
}

func (_c *MockConversation_GetState_Call) Run(run func()) *MockConversation_GetState_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockConversation_GetState_Call) Return(_a0 conversation.ConversationState) *MockConversation_GetState_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockConversation_GetState_Call) RunAndReturn(run func() conversation.ConversationState) *MockConversation_GetState_Call {
	_c.Call.Return(run)
	return _c
}

// History provides a mock function with given fields: skip
func (_m *MockConversation) History(skip int) string {
	ret := _m.Called(skip)

	if len(ret) == 0 {
		panic("no return value specified for History")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func(int) string); ok {
		r0 = rf(skip)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockConversation_History_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'History'
type MockConversation_History_Call struct {
	*mock.Call
}

// History is a helper method to define mock.On call
//   - skip int
func (_e *MockConversation_Expecter) History(skip interface{}) *MockConversation_History_Call {
	return &MockConversation_History_Call{Call: _e.mock.On("History", skip)}
}

func (_c *MockConversation_History_Call) Run(run func(skip int)) *MockConversation_History_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(int))
	})
	return _c
}

func (_c *MockConversation_History_Call) Return(_a0 string) *MockConversation_History_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockConversation_History_Call) RunAndReturn(run func(int) string) *MockConversation_History_Call {
	_c.Call.Return(run)
	return _c
}

// StartFollowUpQuestions provides a mock function with given fields: initialPrompt, questions
func (_m *MockConversation) StartFollowUpQuestions(initialPrompt string, questions []message.Question) error {
	ret := _m.Called(initialPrompt, questions)

	if len(ret) == 0 {
		panic("no return value specified for StartFollowUpQuestions")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, []message.Question) error); ok {
		r0 = rf(initialPrompt, questions)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockConversation_StartFollowUpQuestions_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'StartFollowUpQuestions'
type MockConversation_StartFollowUpQuestions_Call struct {
	*mock.Call
}

// StartFollowUpQuestions is a helper method to define mock.On call
//   - initialPrompt string
//   - questions []message.Question
func (_e *MockConversation_Expecter) StartFollowUpQuestions(initialPrompt interface{}, questions interface{}) *MockConversation_StartFollowUpQuestions_Call {
	return &MockConversation_StartFollowUpQuestions_Call{Call: _e.mock.On("StartFollowUpQuestions", initialPrompt, questions)}
}

func (_c *MockConversation_StartFollowUpQuestions_Call) Run(run func(initialPrompt string, questions []message.Question)) *MockConversation_StartFollowUpQuestions_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].([]message.Question))
	})
	return _c
}

func (_c *MockConversation_StartFollowUpQuestions_Call) Return(_a0 error) *MockConversation_StartFollowUpQuestions_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockConversation_StartFollowUpQuestions_Call) RunAndReturn(run func(string, []message.Question) error) *MockConversation_StartFollowUpQuestions_Call {
	_c.Call.Return(run)
	return _c
}

// StartProfileQuestions provides a mock function with no fields
func (_m *MockConversation) StartProfileQuestions() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for StartProfileQuestions")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockConversation_StartProfileQuestions_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'StartProfileQuestions'
type MockConversation_StartProfileQuestions_Call struct {
	*mock.Call
}

// StartProfileQuestions is a helper method to define mock.On call
func (_e *MockConversation_Expecter) StartProfileQuestions() *MockConversation_StartProfileQuestions_Call {
	return &MockConversation_StartProfileQuestions_Call{Call: _e.mock.On("StartProfileQuestions")}
}

func (_c *MockConversation_StartProfileQuestions_Call) Run(run func()) *MockConversation_StartProfileQuestions_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockConversation_StartProfileQuestions_Call) Return(_a0 error) *MockConversation_StartProfileQuestions_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockConversation_StartProfileQuestions_Call) RunAndReturn(run func() error) *MockConversation_StartProfileQuestions_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockConversation creates a new instance of MockConversation. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockConversation(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockConversation {
	mock := &MockConversation{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
