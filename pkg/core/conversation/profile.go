package conversation

import (
	"cmp"
	"context"
	"time"
	"unicode/utf8"

	"github.com/ksysoev/help-my-pet/pkg/core/message"
	"github.com/ksysoev/help-my-pet/pkg/i18n"
)

const defaultMaxLength = 100

var (
	fieldMaxLength = map[string]int{
		"name":             20,
		"species":          20,
		"breed":            30,
		"dob":              20,
		"gender":           20,
		"weight":           20,
		"neutered":         20,
		"activity":         20,
		"chronic_diseases": 200,
		"food_preferences": 200,
	}
)

// PetProfileStateImpl implements QuestionnaireState
type PetProfileStateImpl struct {
	QAPairs      []QuestionAnswer `json:"qa_pairs"`
	CurrentIndex int              `json:"current_index"`
}

// NewPetProfileQuestionnaireState initializes a new pet profile questionnaire state with predefined questions and indexes.
// It creates a list of questions about the pet's profile, such as name, species, breed, birthdate, gender, and weight.
// Returns a pointer to a PetProfileStateImpl instance with the questions and initial index set to 0.
func NewPetProfileQuestionnaireState(ctx context.Context) *PetProfileStateImpl {
	questions := []QuestionAnswer{
		{
			Question: message.Question{
				Text: i18n.GetLocale(ctx).Sprintf("What is your pet's name?"),
			},
			Field: "name",
		},
		{
			Question: message.Question{
				Text:    i18n.GetLocale(ctx).Sprintf("What type of pet do you have?"),
				Answers: []string{i18n.GetLocale(ctx).Sprintf("dog"), i18n.GetLocale(ctx).Sprintf("cat")},
			},
			Field: "species",
		},
		{
			Question: message.Question{
				Text: i18n.GetLocale(ctx).Sprintf("What breed is your pet?"),
			},
			Field: "breed",
		},
		{
			Question: message.Question{
				Text: i18n.GetLocale(ctx).Sprintf("When was your pet born? Please enter the date in the format YYYY-MM-DD (e.g., 2010-12-31)."),
			},
			Field: "dob",
		},
		{
			Question: message.Question{
				Text:    i18n.GetLocale(ctx).Sprintf("What is your pet's gender?"),
				Answers: []string{i18n.GetLocale(ctx).Sprintf("male"), i18n.GetLocale(ctx).Sprintf("female")},
			},
			Field: "gender",
		},
		{
			Question: message.Question{
				Text: i18n.GetLocale(ctx).Sprintf("What is your pet's weight? Please specify the weight followed by the unit, e.g., 5 kg"),
			},
			Field: "weight",
		},
		{
			Question: message.Question{
				Text:    i18n.GetLocale(ctx).Sprintf("Is your pet spayed or neutered?"),
				Answers: []string{i18n.GetLocale(ctx).Sprintf("yes"), i18n.GetLocale(ctx).Sprintf("no")},
			},
			Field: "neutered",
		},
		{
			Question: message.Question{
				Text:    i18n.GetLocale(ctx).Sprintf("How active is your pet?"),
				Answers: []string{i18n.GetLocale(ctx).Sprintf("low"), i18n.GetLocale(ctx).Sprintf("medium"), i18n.GetLocale(ctx).Sprintf("high")},
			},
			Field: "activity",
		},
		{
			Question: message.Question{
				Text: i18n.GetLocale(ctx).Sprintf("Does your pet have any chronic diseases?"),
			},
			Field: "chronic_diseases",
		},
		{
			Question: message.Question{
				Text: i18n.GetLocale(ctx).Sprintf("What are your pet's food preferences or dietary restrictions?"),
			},
			Field: "food_preferences",
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
		return validateLength(answer, cmp.Or(fieldMaxLength[field], defaultMaxLength))
	}
}

// validateLength checks if the length of the input string exceeds the specified maximum length.
// It returns an error if the string is too long.
func validateLength(answer string, maxLength int) error {
	if utf8.RuneCountInString(answer) > maxLength {
		return message.ErrTextTooLong
	}
	return nil
}

// validateDOB validates whether the given date string is in the "YYYY-MM-DD" format and represents a valid date.
// It returns ErrInvalidDates if the format is incorrect or ErrFutureDate if the date is in the future.
func validateDOB(answer string) error {
	date, err := time.Parse("2006-01-02", answer)

	if err != nil {
		return message.ErrInvalidDates
	}

	if date.After(time.Now()) {
		return message.ErrFutureDate
	}

	return nil
}
