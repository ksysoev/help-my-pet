package conversation

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
