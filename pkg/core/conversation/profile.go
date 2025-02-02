package conversation

import (
	"fmt"
	"time"
)

// PetProfileQuestion represents a question for pet profile creation with validation
type PetProfileQuestion struct {
	Question
	ValidateFn func(string) error
}

// PetProfileStateImpl implements QuestionnaireState
type PetProfileStateImpl struct {
	questions    []PetProfileQuestion
	answers      []string
	currentIndex int
}

// NewPetProfileQuestionnaireState creates questions specific to pet profile creation with validation
func NewPetProfileQuestionnaireState() *PetProfileStateImpl {
	questions := []PetProfileQuestion{
		{
			Question: Question{
				Text: "What is your pet's name?",
			},
			ValidateFn: validatePetName,
		},
		{
			Question: Question{
				Text:    "What type of pet do you have?",
				Answers: []string{"dog", "cat", "bird", "fish", "other"},
			},
			ValidateFn: validateSpecies,
		},
		{
			Question: Question{
				Text: "What breed is your pet?",
			},
			ValidateFn: validateBreed,
		},
		{
			Question: Question{
				Text: "When was your pet born? (YYYY-MM-DD)",
			},
			ValidateFn: validateDateOfBirth,
		},
		{
			Question: Question{
				Text:    "What is your pet's gender?",
				Answers: []string{"male", "female"},
			},
			ValidateFn: validateGender,
		},
		{
			Question: Question{
				Text: "How much does your pet weigh? (in kg)",
			},
			ValidateFn: validateWeight,
		},
	}

	return NewPetProfileStateImpl(questions)
}

func NewPetProfileStateImpl(questions []PetProfileQuestion) *PetProfileStateImpl {
	return &PetProfileStateImpl{
		questions:    questions,
		answers:      make([]string, len(questions)),
		currentIndex: 0,
	}
}

func (s *PetProfileStateImpl) GetCurrentQuestion() (*Question, error) {
	if s.currentIndex >= len(s.questions) {
		return nil, ErrNoMoreQuestions
	}
	return &s.questions[s.currentIndex].Question, nil
}

func (s *PetProfileStateImpl) ProcessAnswer(answer string) (bool, error) {
	if s.currentIndex >= len(s.questions) {
		return false, ErrNoMoreQuestions
	}

	if err := s.ValidateAnswer(answer); err != nil {
		return false, err
	}

	s.answers[s.currentIndex] = answer
	s.currentIndex++

	return s.currentIndex >= len(s.questions), nil
}

func (s *PetProfileStateImpl) ValidateAnswer(answer string) error {
	return s.questions[s.currentIndex].ValidateFn(answer)
}

func (s *PetProfileStateImpl) GetResults() ([]QuestionAnswer, error) {
	if s.currentIndex < len(s.questions) {
		return nil, ErrQuestionnaireIncomplete
	}

	results := make([]QuestionAnswer, len(s.questions))
	for i := range s.questions {
		results[i] = QuestionAnswer{
			Question: s.questions[i].Question,
			Answer:   s.answers[i],
		}
	}
	return results, nil
}

// Validation functions
func validatePetName(name string) error {
	if len(name) < 1 {
		return NewQuestionnaireError("pet name cannot be empty")
	}
	if len(name) > 50 {
		return NewQuestionnaireError("pet name is too long (max 50 characters)")
	}
	return nil
}

func validateSpecies(species string) error {
	validSpecies := map[string]bool{
		"dog": true,
		"cat": true,
	}
	if !validSpecies[species] {
		return NewQuestionnaireError("invalid species type")
	}
	return nil
}

func validateBreed(breed string) error {
	if len(breed) < 1 {
		return NewQuestionnaireError("breed cannot be empty")
	}
	if len(breed) > 50 {
		return NewQuestionnaireError("breed name is too long (max 50 characters)")
	}
	return nil
}

func validateDateOfBirth(dob string) error {
	_, err := time.Parse("2006-01-02", dob)
	if err != nil {
		return NewQuestionnaireError("invalid date format, use DD-MM-YYYY")
	}
	return nil
}

func validateGender(gender string) error {
	if gender != "male" && gender != "female" {
		return NewQuestionnaireError("gender must be either 'male' or 'female'")
	}
	return nil
}

func validateWeight(weight string) error {
	var w float64
	_, err := fmt.Sscanf(weight, "%f", &w)
	if err != nil {
		return NewQuestionnaireError("weight must be a valid number")
	}
	if w <= 0 || w > 500 {
		return NewQuestionnaireError("weight must be between 0 and 500 kg")
	}
	return nil
}
