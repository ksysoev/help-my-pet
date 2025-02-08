package conversation

import (
	"errors"
	"time"

	"github.com/ksysoev/help-my-pet/pkg/core/message"
)

var (
	ErrInvalidDates = errors.New("invalid date format")
	ErrFutureDate   = errors.New("date is in the future")
)

// PetProfileStateImpl implements QuestionnaireState
type PetProfileStateImpl struct {
	QAPairs      []QuestionAnswer `json:"qa_pairs"`
	CurrentIndex int              `json:"current_index"`
}

// NewPetProfileQuestionnaireState initializes a new pet profile questionnaire state with predefined questions and indexes.
// It creates a list of questions about the pet's profile, such as name, species, breed, birthdate, gender, and weight.
// Returns a pointer to a PetProfileStateImpl instance with the questions and initial index set to 0.
func NewPetProfileQuestionnaireState() *PetProfileStateImpl {
	questions := []QuestionAnswer{
		{
			Question: message.Question{
				Text: "What is your pet's name?",
			},
			Field: "name",
		},
		{
			Question: message.Question{
				Text:    "What type of pet do you have?",
				Answers: []string{"dog", "cat"},
			},
			Field: "species",
		},
		{
			Question: message.Question{
				Text: "What breed is your pet?",
			},
			Field: "breed",
		},
		{
			Question: message.Question{
				Text: "When was your pet born? Please enter the date in the format YYYY-MM-DD (e.g., 2010-12-31).",
			},
			Field: "dob",
		},
		{
			Question: message.Question{
				Text:    "What is your pet's gender?",
				Answers: []string{"male", "female"},
			},
			Field: "gender",
		},
		{
			Question: message.Question{
				Text: "What is your pet's weight? Please specify the weight followed by the unit, e.g., 5 kg",
			},
			Field: "weight",
		},
	}

	return &PetProfileStateImpl{
		QAPairs:      questions,
		CurrentIndex: 0,
	}
}

// GetCurrentQuestion retrieves the current question from the questionnaire.
// It returns an error if no more questions are available to answer.
// Returns the current question or nil if an error occurs.
func (s *PetProfileStateImpl) GetCurrentQuestion() (*message.Question, error) {
	if s.CurrentIndex >= len(s.QAPairs) {
		return nil, ErrNoMoreQuestions
	}
	return &s.QAPairs[s.CurrentIndex].Question, nil
}

// ProcessAnswer stores the provided answer for the current question and advances to the next question.
// It returns true if all questions have been answered, and false otherwise.
// Returns an error if there are no more questions to answer.
func (s *PetProfileStateImpl) ProcessAnswer(answer string) (bool, error) {
	if s.CurrentIndex >= len(s.QAPairs) {
		return false, ErrNoMoreQuestions
	}

	qa := &s.QAPairs[s.CurrentIndex]

	if err := validate(qa.Field, answer); err != nil {
		return false, err
	}

	qa.Answer = answer
	s.CurrentIndex++

	return s.CurrentIndex >= len(s.QAPairs), nil
}

// GetResults retrieves all question and answer pairs from the questionnaire.
// It returns an error if not all questions have been answered.
func (s *PetProfileStateImpl) GetResults() ([]QuestionAnswer, error) {
	if s.CurrentIndex < len(s.QAPairs) {
		return nil, ErrQuestionnaireIncomplete
	}

	return s.QAPairs, nil
}

// validate checks the validity of the provided answer for the specified field.
// It applies field-specific validation such as date format for "dob" and length restrictions for others.
// Returns error if the answer is invalid or does not meet the field's requirements.
func validate(field string, answer string) error {
	switch field {
	case "dob":
		return validateDOB(answer)
	default:
		return validateLength(answer, 100)
	}
}

// validateLength checks if the length of the input string exceeds the specified maximum length.
// It returns an error if the string is too long.
func validateLength(answer string, maxLength int) error {
	if len(answer) > maxLength {
		return message.ErrTextTooLong
	}
	return nil
}

// validateDOB validates whether the given date string is in the "YYYY-MM-DD" format and represents a valid date.
// It returns ErrInvalidDates if the format is incorrect or ErrFutureDate if the date is in the future.
func validateDOB(answer string) error {
	date, err := time.Parse("2006-01-02", answer)

	if err != nil {
		return ErrInvalidDates
	}

	if date.After(time.Now()) {
		return ErrFutureDate
	}

	return nil
}
