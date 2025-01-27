package core

import (
	"fmt"
	"time"
)

// ConversationState represents the current state of the conversation
type ConversationState string

const (
	MaxMessageHistory = 5 // Maximum number of messages to keep in history

	StateNormal      ConversationState = "normal"
	StateQuestioning ConversationState = "questioning"
)

// QuestionAnswer pairs a question with its corresponding answer
type QuestionAnswer struct {
	Answer   string   `json:"answer"`
	Question Question `json:"question"`
}

// QuestionnaireState tracks the state of follow-up questions
type QuestionnaireState struct {
	InitialPrompt string           `json:"initial_prompt"`
	QAPairs       []QuestionAnswer `json:"qa_pairs"`
	CurrentIndex  int              `json:"current_index"`
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

	// Keep only the last N messages
	if len(c.Messages) > MaxMessageHistory {
		c.Messages = c.Messages[len(c.Messages)-MaxMessageHistory:]
	}
}

// GetContext returns all messages in the conversation as context.
func (c *Conversation) GetContext() []Message {
	return c.Messages
}

// StartQuestionnaire initializes the questioning state with follow-up questions
func (c *Conversation) StartQuestionnaire(initialPrompt string, questions []Question) {
	qaPairs := make([]QuestionAnswer, len(questions))
	for i, q := range questions {
		qaPairs[i] = QuestionAnswer{
			Question: q,
			Answer:   "",
		}
	}

	c.State = StateQuestioning
	c.Questionnaire = &QuestionnaireState{
		QAPairs:       qaPairs,
		CurrentIndex:  0,
		InitialPrompt: initialPrompt,
	}
}

// GetCurrentQuestion returns the current question in the questionnaire
func (c *Conversation) GetCurrentQuestion() (*Question, error) {
	if c.State != StateQuestioning || c.Questionnaire == nil {
		return nil, fmt.Errorf("conversation is not in questioning state")
	}

	if c.Questionnaire.CurrentIndex >= len(c.Questionnaire.QAPairs) {
		return nil, fmt.Errorf("no more questions available")
	}

	return &c.Questionnaire.QAPairs[c.Questionnaire.CurrentIndex].Question, nil
}

// AddQuestionAnswer adds an answer to the current question and moves to the next one
func (c *Conversation) AddQuestionAnswer(answer string) (bool, error) {
	if c.State != StateQuestioning || c.Questionnaire == nil {
		return false, fmt.Errorf("conversation is not in questioning state")
	}

	if c.Questionnaire.CurrentIndex >= len(c.Questionnaire.QAPairs) {
		return false, fmt.Errorf("no more questions to answer")
	}

	// Store the answer
	c.Questionnaire.QAPairs[c.Questionnaire.CurrentIndex].Answer = answer
	c.Questionnaire.CurrentIndex++

	// Check if we've collected all answers
	isComplete := c.Questionnaire.CurrentIndex >= len(c.Questionnaire.QAPairs)
	if isComplete {
		// Combine all questions and answers into a single message
		var combinedContent string
		for _, qa := range c.Questionnaire.QAPairs {
			combinedContent += fmt.Sprintf("Q: %s\nA: %s\n\n", qa.Question.Text, qa.Answer)
		}

		// Add combined message to conversation history
		c.AddMessage("questionnaire", combinedContent)

		// Reset state
		c.State = StateNormal
	}

	return isComplete, nil
}

// GetQuestionnaireResult returns the initial prompt and all question-answer pairs
func (c *Conversation) GetQuestionnaireResult() ([]QuestionAnswer, error) {
	if c.Questionnaire == nil {
		return nil, fmt.Errorf("no questionnaire data available")
	}

	return c.Questionnaire.QAPairs, nil
}
