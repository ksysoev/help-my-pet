package conversation

import "github.com/ksysoev/help-my-pet/pkg/core/message"

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
				Answers: []string{"dog", "cat", "bird", "fish", "other"},
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
				Text: "When was your pet born? (YYYY-MM-DD)",
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
				Text: "How much does your pet weigh? (in kg)",
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

	s.QAPairs[s.CurrentIndex].Answer = answer
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
