package conversation

// BaseQuestionnaireState represents the interface that all questionnaire states must implement
type BaseQuestionnaireState interface {
	// GetCurrentQuestion returns the current question to be asked
	GetCurrentQuestion() (*Question, error)

	// ProcessAnswer processes the answer for the current question and returns true if questionnaire is complete
	ProcessAnswer(answer string) (bool, error)

	// GetResults returns the questionnaire results when completed
	GetResults() ([]QuestionAnswer, error)

	// ValidateAnswer validates an answer for the current question
	ValidateAnswer(answer string) error
}

// Error types
var (
	ErrNoMoreQuestions         = NewQuestionnaireError("no more questions available")
	ErrInvalidAnswer           = NewQuestionnaireError("invalid answer for question")
	ErrQuestionnaireIncomplete = NewQuestionnaireError("questionnaire is not complete")
)

// QuestionnaireError represents a questionnaire-specific error
type QuestionnaireError struct {
	message string
}

func NewQuestionnaireError(msg string) *QuestionnaireError {
	return &QuestionnaireError{message: msg}
}

func (e *QuestionnaireError) Error() string {
	return e.message
}
