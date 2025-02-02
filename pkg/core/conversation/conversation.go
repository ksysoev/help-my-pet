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
	StateQuestioning           ConversationState = "questioning" // Used for LLM questionnaire (backward compatibility)
	StatePetProfileQuestioning ConversationState = "pet_profile_questioning"
)

// QuestionAnswer pairs a question with its corresponding answer
type QuestionAnswer struct {
	Answer   string   `json:"answer"`
	Question Question `json:"question"`
}

// FollowUpQuestionnaireState represents the state for follow-up questions from LLM
type FollowUpQuestionnaireState struct {
	InitialPrompt string           `json:"initial_prompt"`
	QAPairs       []QuestionAnswer `json:"qa_pairs"`
	CurrentIndex  int              `json:"current_index"`
}

// Conversation represents a chat conversation with its context and messages.
type Conversation struct {
	ID               string
	State            ConversationState
	Messages         []Message
	Questionnaire    *FollowUpQuestionnaireState `json:"questionnaire"` // For backward compatibility with Redis
	petQuestionnaire BaseQuestionnaireState      // For pet profile questions
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

// StartQuestionnaire initializes the follow-up questioning state (backward compatible name)
func (c *Conversation) StartQuestionnaire(initialPrompt string, questions []Question) {
	qaPairs := make([]QuestionAnswer, len(questions))
	for i, q := range questions {
		qaPairs[i] = QuestionAnswer{
			Question: q,
			Answer:   "",
		}
	}

	c.State = StateQuestioning
	c.Questionnaire = &FollowUpQuestionnaireState{
		QAPairs:       qaPairs,
		CurrentIndex:  0,
		InitialPrompt: initialPrompt,
	}
}

// StartPetProfileQuestionnaire initializes the pet profile questionnaire
func (c *Conversation) StartPetProfileQuestionnaire() {
	c.State = StatePetProfileQuestioning
	c.petQuestionnaire = NewPetProfileQuestionnaireState()
}

// GetCurrentQuestion returns the current question in the active questionnaire
func (c *Conversation) GetCurrentQuestion() (*Question, error) {
	switch c.State {
	case StateQuestioning: // LLM questionnaire
		if c.Questionnaire == nil {
			return nil, fmt.Errorf("llm questionnaire not initialized")
		}
		if c.Questionnaire.CurrentIndex >= len(c.Questionnaire.QAPairs) {
			return nil, ErrNoMoreQuestions
		}
		return &c.Questionnaire.QAPairs[c.Questionnaire.CurrentIndex].Question, nil

	case StatePetProfileQuestioning:
		if c.petQuestionnaire == nil {
			return nil, fmt.Errorf("pet profile questionnaire not initialized")
		}
		return c.petQuestionnaire.GetCurrentQuestion()

	default:
		return nil, fmt.Errorf("conversation is not in a questioning state")
	}
}

// AddQuestionAnswer adds an answer to the current question and moves to the next one
func (c *Conversation) AddQuestionAnswer(answer string) (bool, error) {
	switch c.State {
	case StateQuestioning: // LLM questionnaire
		if c.Questionnaire == nil {
			return false, fmt.Errorf("llm questionnaire not initialized")
		}
		if c.Questionnaire.CurrentIndex >= len(c.Questionnaire.QAPairs) {
			return false, ErrNoMoreQuestions
		}

		c.Questionnaire.QAPairs[c.Questionnaire.CurrentIndex].Answer = answer
		c.Questionnaire.CurrentIndex++

		isComplete := c.Questionnaire.CurrentIndex >= len(c.Questionnaire.QAPairs)
		if isComplete {
			var combinedContent string
			for _, qa := range c.Questionnaire.QAPairs {
				combinedContent += fmt.Sprintf("Q: %s\nA: %s\n\n", qa.Question.Text, qa.Answer)
			}
			c.AddMessage("questionnaire", combinedContent)
			c.State = StateNormal
		}
		return isComplete, nil

	case StatePetProfileQuestioning:
		if c.petQuestionnaire == nil {
			return false, fmt.Errorf("pet profile questionnaire not initialized")
		}

		isComplete, err := c.petQuestionnaire.ProcessAnswer(answer)
		if err != nil {
			return false, fmt.Errorf("failed to process answer: %w", err)
		}

		if isComplete {
			results, err := c.petQuestionnaire.GetResults()
			if err != nil {
				return true, fmt.Errorf("failed to get questionnaire results: %w", err)
			}

			var combinedContent string
			for _, qa := range results {
				combinedContent += fmt.Sprintf("Q: %s\nA: %s\n\n", qa.Question.Text, qa.Answer)
			}
			c.AddMessage("pet-profile", combinedContent)
			c.State = StateNormal
		}
		return isComplete, nil

	default:
		return false, fmt.Errorf("conversation is not in a questioning state")
	}
}

// GetQuestionnaireResult returns all question-answer pairs from the active questionnaire
func (c *Conversation) GetQuestionnaireResult() ([]QuestionAnswer, error) {
	switch c.State {
	case StateQuestioning: // LLM questionnaire
		if c.Questionnaire == nil {
			return nil, fmt.Errorf("llm questionnaire not initialized")
		}
		return c.Questionnaire.QAPairs, nil

	case StatePetProfileQuestioning:
		if c.petQuestionnaire == nil {
			return nil, fmt.Errorf("pet profile questionnaire not initialized")
		}
		return c.petQuestionnaire.GetResults()

	default:
		return nil, fmt.Errorf("conversation is not in a questioning state")
	}
}

// GetQuestionnaireType returns the current type of questionnaire being used
func (c *Conversation) GetQuestionnaireType() string {
	switch c.State {
	case StateQuestioning:
		return "llm"
	case StatePetProfileQuestioning:
		return "pet_profile"
	default:
		return "none"
	}
}

// Question represents a follow-up question with optional predefined answers
type Question struct {
	Text    string   `json:"text"`
	Answers []string `json:"answers,omitempty"`
}
