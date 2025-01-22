package core

import (
	"fmt"
	"time"
)

// ConversationState represents the current state of the conversation
type ConversationState string

const (
	StateNormal      ConversationState = "normal"
	StateQuestioning ConversationState = "questioning"
)

// QuestionnaireState tracks the state of follow-up questions
type QuestionnaireState struct {
	InitialPrompt string
	Questions     []Question
	Answers       []string
	CurrentIndex  int
}

// Conversation represents a chat conversation with its context and messages.
type Conversation struct {
	Questionnaire *QuestionnaireState
	ID            string
	State         ConversationState
	Messages      []Message
}

// Message represents a single message in a conversation.
type Message struct {
	Timestamp time.Time
	Role      string
	Content   string
}

// NewConversation creates a new conversation with the given ID.
func NewConversation(id string) *Conversation {
	return &Conversation{
		ID:       id,
		Messages: make([]Message, 0),
		State:    StateNormal,
	}
}

// AddMessage adds a new message to the conversation.
func (c *Conversation) AddMessage(role, content string) {
	c.Messages = append(c.Messages, Message{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	})
}

// GetContext returns all messages in the conversation as context.
func (c *Conversation) GetContext() []Message {
	return c.Messages
}

// StartQuestionnaire initializes the questioning state with follow-up questions
func (c *Conversation) StartQuestionnaire(initialPrompt string, questions []Question) {
	c.State = StateQuestioning
	c.Questionnaire = &QuestionnaireState{
		Questions:     questions,
		CurrentIndex:  0,
		Answers:       make([]string, len(questions)),
		InitialPrompt: initialPrompt,
	}
}

// GetCurrentQuestion returns the current question in the questionnaire
func (c *Conversation) GetCurrentQuestion() (*Question, error) {
	if c.State != StateQuestioning || c.Questionnaire == nil {
		return nil, fmt.Errorf("conversation is not in questioning state")
	}

	if c.Questionnaire.CurrentIndex >= len(c.Questionnaire.Questions) {
		return nil, fmt.Errorf("no more questions available")
	}

	return &c.Questionnaire.Questions[c.Questionnaire.CurrentIndex], nil
}

// AddQuestionAnswer adds an answer to the current question and moves to the next one
func (c *Conversation) AddQuestionAnswer(answer string) (bool, error) {
	if c.State != StateQuestioning || c.Questionnaire == nil {
		return false, fmt.Errorf("conversation is not in questioning state")
	}

	if c.Questionnaire.CurrentIndex >= len(c.Questionnaire.Questions) {
		return false, fmt.Errorf("no more questions to answer")
	}

	// Store the answer
	c.Questionnaire.Answers[c.Questionnaire.CurrentIndex] = answer
	c.Questionnaire.CurrentIndex++

	// Check if we've collected all answers
	isComplete := c.Questionnaire.CurrentIndex >= len(c.Questionnaire.Questions)
	if isComplete {
		c.State = StateNormal
	}

	return isComplete, nil
}

// GetQuestionnaireResult returns the initial prompt and all collected answers
func (c *Conversation) GetQuestionnaireResult() (string, []string, error) {
	if c.Questionnaire == nil {
		return "", nil, fmt.Errorf("no questionnaire data available")
	}

	return c.Questionnaire.InitialPrompt, c.Questionnaire.Answers, nil
}
