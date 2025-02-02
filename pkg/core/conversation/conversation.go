package conversation

import (
	"fmt"
	"time"
)

// ConversationState represents the current state of the conversation
type ConversationState string

const (
	MaxMessageHistory = 5 // Maximum number of messages to keep in history

	StateNormal                ConversationState = "normal"
	StateFollowUpQuestioning   ConversationState = "questioning" // Used for LLM questionnaire (backward compatibility)
	StatePetProfileQuestioning ConversationState = "pet_profile_questioning"
)

// QuestionnaireState represents the interface that all questionnaire states must implement
type QuestionnaireState interface {
	// GetCurrentQuestion returns the current question to be asked
	GetCurrentQuestion() (*Question, error)

	// ProcessAnswer processes the answer for the current question and returns true if questionnaire is complete
	ProcessAnswer(answer string) (bool, error)

	// GetResults returns the questionnaire results when completed
	GetResults() ([]QuestionAnswer, error)
}

// Question represents a follow-up question with optional predefined answers
type Question struct {
	Text    string   `json:"text"`
	Answers []string `json:"answers,omitempty"`
}

// QuestionAnswer pairs a question with its corresponding answer
type QuestionAnswer struct {
	Answer   string   `json:"answer"`
	Question Question `json:"question"`
}

// Conversation represents a chat conversation with its context and messages.
type Conversation struct {
	ID            string
	State         ConversationState
	Messages      []Message
	Questionnaire QuestionnaireState `json:"questionnaire"`
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

// StartFollowUpQuestionnaire initializes the follow-up questioning state (backward compatible name)
func (c *Conversation) StartFollowUpQuestionnaire(initialPrompt string, questions []Question) error {
	if c.State != StateNormal {
		return fmt.Errorf("conversation is not in normal state %s", c.State)
	}

	if len(questions) == 0 {
		return fmt.Errorf("no follow-up questions provided")
	}

	qaPairs := make([]QuestionAnswer, len(questions))
	for i, q := range questions {
		qaPairs[i] = QuestionAnswer{
			Question: q,
			Answer:   "",
		}
	}

	c.State = StateFollowUpQuestioning
	c.Questionnaire = &FollowUpQuestionnaireState{
		QAPairs:       qaPairs,
		CurrentIndex:  0,
		InitialPrompt: initialPrompt,
	}

	return nil
}

// StartPetProfileQuestionnaire initializes the pet profile questionnaire
func (c *Conversation) StartProfileQuestionnaire() error {
	if c.State != StateNormal {
		return fmt.Errorf("conversation is not in normal state %s", c.State)
	}

	c.State = StatePetProfileQuestioning
	c.Questionnaire = NewPetProfileQuestionnaireState()

	return nil
}

// GetCurrentQuestion returns the current question in the active questionnaire
func (c *Conversation) GetCurrentQuestion() (*Question, error) {
	switch c.State {
	case StateFollowUpQuestioning, StatePetProfileQuestioning: // LLM questionnaire
		if c.Questionnaire == nil {
			return nil, fmt.Errorf("questionnaire not initialized")
		}

		return c.Questionnaire.GetCurrentQuestion()
	default:
		return nil, fmt.Errorf("conversation is not in a questioning state")
	}
}

// AddQuestionAnswer adds an answer to the current question and moves to the next one
func (c *Conversation) AddQuestionAnswer(answer string) (bool, error) {
	switch c.State {
	case StateFollowUpQuestioning, StatePetProfileQuestioning:
		if c.Questionnaire == nil {
			return false, fmt.Errorf("pet profile questionnaire not initialized")
		}

		isComplete, err := c.Questionnaire.ProcessAnswer(answer)
		if err != nil {
			return false, fmt.Errorf("failed to process answer: %w", err)
		}

		return isComplete, nil

	default:
		return false, fmt.Errorf("conversation is not in a questioning state")
	}
}

// GetQuestionnaireResult returns all question-answer pairs from the active questionnaire
func (c *Conversation) GetQuestionnaireResult() ([]QuestionAnswer, error) {
	switch c.State {
	case StateFollowUpQuestioning, StatePetProfileQuestioning: // LLM questionnair
		if c.Questionnaire == nil {
			return nil, fmt.Errorf("questionnaire not initialized")
		}
		return c.Questionnaire.GetResults()

	default:
		return nil, fmt.Errorf("conversation is not in a questioning state")
	}
}
