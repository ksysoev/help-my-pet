package conversation

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ksysoev/help-my-pet/pkg/core/message"
)

// ConversationState represents the current state of the conversation
type ConversationState string

const (
	MaxMessageHistory = 5   // Maximum number of messages to keep in history
	MaxAnswerLength   = 200 // Maximum length of an answer

	StateNormal                ConversationState = "normal"
	StateFollowUpQuestioning   ConversationState = "questioning" // Used for LLM questionnaire (backward compatibility)
	StatePetProfileQuestioning ConversationState = "pet_profile_questioning"
	StateCompleted             ConversationState = "completed"
)

var (
	ErrNoMoreQuestions         = errors.New("no more questions available")
	ErrQuestionnaireIncomplete = errors.New("questionnaire is not complete")
)

// QuestionnaireState represents the interface that all questionnaire states must implement
type QuestionnaireState interface {
	// GetCurrentQuestion returns the current question to be asked
	GetCurrentQuestion() (*message.Question, error)

	// ProcessAnswer processes the answer for the current question and returns true if questionnaire is complete
	ProcessAnswer(answer string) (bool, error)

	// GetResults returns the questionnaire results when completed
	GetResults() ([]QuestionAnswer, error)
}

// QuestionAnswer pairs a question with its corresponding answer
type QuestionAnswer struct {
	Answer   string           `json:"answer"`
	Field    string           `json:"field,omitempty"`
	Question message.Question `json:"question"`
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

// NewConversation creates a new conversation with the given GetID.
func NewConversation(id string) *Conversation {
	return &Conversation{
		ID:       id,
		Messages: make([]Message, 0),
		State:    StateNormal,
	}
}

// GetState retrieves the current state of the conversation.
// It returns the state as a ConversationState type.
func (c *Conversation) GetState() ConversationState {
	return c.State
}

// GetID retrieves the unique identifier of the conversation.
// It returns a string representing the conversation's ID.
func (c *Conversation) GetID() string {
	return c.ID
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

// History retrieves the conversation history up until the specified number of recent messages to skip.
// It includes all messages older than the skipped count in the returned history.
// skip specifies the number of most recent messages to exclude from the history.
// Returns a string representation of the filtered conversation history.
func (c *Conversation) History(skip int) string {
	if len(c.Messages) <= skip {
		return ""
	}

	history := "Previous conversation:\n"
	for _, msg := range c.Messages[:len(c.Messages)-skip] {
		history += fmt.Sprintf("%s: %s\n", msg.Role, msg.Content)
	}

	return history
}

// StartFollowUpQuestions initializes the follow-up questioning state (backward compatible name)
func (c *Conversation) StartFollowUpQuestions(initialPrompt string, questions []message.Question) error {
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
	c.Questionnaire = NewFollowUpQuestionnaireState(initialPrompt, questions)

	return nil
}

// StartPetProfileQuestionnaire initializes the pet profile questionnaire
func (c *Conversation) StartProfileQuestions() error {
	if c.State != StateNormal {
		return fmt.Errorf("conversation is not in normal state %s", c.State)
	}

	c.State = StatePetProfileQuestioning
	c.Questionnaire = NewPetProfileQuestionnaireState()

	return nil
}

// GetCurrentQuestion returns the current question in the active questionnaire
func (c *Conversation) GetCurrentQuestion() (*message.Question, error) {
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

		if isComplete {
			c.State = StateCompleted
		}

		return isComplete, nil

	default:
		return false, fmt.Errorf("conversation is not in a questioning state")
	}
}

// GetQuestionnaireResult returns all question-answer pairs from the active questionnaire
func (c *Conversation) GetQuestionnaireResult() ([]QuestionAnswer, error) {
	switch c.State {
	case StateCompleted: // LLM questionnair
		c.State = StateNormal
		if c.Questionnaire == nil {
			return nil, fmt.Errorf("questionnaire not initialized")
		}

		q := c.Questionnaire
		c.Questionnaire = nil

		return q.GetResults()

	default:
		return nil, fmt.Errorf("conversation is not in a completed state")
	}
}

// Unmarshal parses the JSON-encoded data and returns a new conversation.
func Unmarshal(data []byte) (*Conversation, error) {
	var tmpConv struct {
		ID            string
		State         ConversationState
		Messages      []Message
		Questionnaire json.RawMessage `json:"questionnaire"`
	}

	if err := json.Unmarshal(data, &tmpConv); err != nil {
		return nil, fmt.Errorf("failed to unmarshal conversation: %w", err)
	}

	switch tmpConv.State {
	case StateNormal, StateCompleted:
		// We don't unmarshal completed state as it's not needed, if conversation stuck in completed state,
		// it means we failed to process questionnaire result, and we just need to reset the state to normal

		return &Conversation{
			ID:       tmpConv.ID,
			State:    StateNormal,
			Messages: tmpConv.Messages,
		}, nil
	case StatePetProfileQuestioning:
		var q PetProfileStateImpl
		if err := json.Unmarshal(tmpConv.Questionnaire, &q); err != nil {
			return nil, fmt.Errorf("failed to unmarshal pet profile questionnaire: %w", err)
		}

		return &Conversation{
			ID:            tmpConv.ID,
			State:         StatePetProfileQuestioning,
			Messages:      tmpConv.Messages,
			Questionnaire: &q,
		}, nil
	case StateFollowUpQuestioning:
		var q FollowUpQuestionnaireState
		if err := json.Unmarshal(tmpConv.Questionnaire, &q); err != nil {
			return nil, fmt.Errorf("failed to unmarshal follow-up questionnaire: %w", err)
		}

		return &Conversation{
			ID:            tmpConv.ID,
			State:         StateFollowUpQuestioning,
			Messages:      tmpConv.Messages,
			Questionnaire: &q,
		}, nil
	default:
		return nil, fmt.Errorf("unknown conversation state: %s", tmpConv.State)
	}
}
