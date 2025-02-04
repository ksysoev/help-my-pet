package conversation

import "github.com/ksysoev/help-my-pet/pkg/core/message"

// PetProfileStateImpl implements QuestionnaireState
type PetProfileStateImpl struct {
	QAPairs      []QuestionAnswer `json:"qa_pairs"`
	CurrentIndex int              `json:"current_index"`
}

// NewPetProfileQuestionnaireState creates questions specific to pet profile creation with validation
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

	return NewPetProfileStateImpl(questions)
}

func NewPetProfileStateImpl(questions []QuestionAnswer) *PetProfileStateImpl {
	return &PetProfileStateImpl{
		QAPairs:      questions,
		CurrentIndex: 0,
	}
}

func (s *PetProfileStateImpl) GetCurrentQuestion() (*message.Question, error) {
	if s.CurrentIndex >= len(s.QAPairs) {
		return nil, ErrNoMoreQuestions
	}
	return &s.QAPairs[s.CurrentIndex].Question, nil
}

func (s *PetProfileStateImpl) ProcessAnswer(answer string) (bool, error) {
	if s.CurrentIndex >= len(s.QAPairs) {
		return false, ErrNoMoreQuestions
	}

	s.QAPairs[s.CurrentIndex].Answer = answer
	s.CurrentIndex++

	return s.CurrentIndex >= len(s.QAPairs), nil
}

func (s *PetProfileStateImpl) GetResults() ([]QuestionAnswer, error) {
	if s.CurrentIndex < len(s.QAPairs) {
		return nil, ErrQuestionnaireIncomplete
	}

	return s.QAPairs, nil
}
