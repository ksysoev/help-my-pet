package conversation

// PetProfileStateImpl implements QuestionnaireState
type PetProfileStateImpl struct {
	QAPairs      []QuestionAnswer `json:"qa_pairs"`
	CurrentIndex int              `json:"current_index"`
}

// NewPetProfileQuestionnaireState creates questions specific to pet profile creation with validation
func NewPetProfileQuestionnaireState() *PetProfileStateImpl {
	questions := []QuestionAnswer{
		{
			Question: Question{
				Text: "What is your pet's name?",
			},
			Field: "name",
		},
		{
			Question: Question{
				Text:    "What type of pet do you have?",
				Answers: []string{"dog", "cat", "bird", "fish", "other"},
			},
			Field: "species",
		},
		{
			Question: Question{
				Text: "What breed is your pet?",
			},
			Field: "breed",
		},
		{
			Question: Question{
				Text: "When was your pet born? (YYYY-MM-DD)",
			},
			Field: "dob",
		},
		{
			Question: Question{
				Text:    "What is your pet's gender?",
				Answers: []string{"male", "female"},
			},
			Field: "gender",
		},
		{
			Question: Question{
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

func (s *PetProfileStateImpl) GetCurrentQuestion() (*Question, error) {
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

// // Validation functions
// func validatePetName(name string) error {
// 	if len(name) < 1 {
// 		return NewQuestionnaireError("pet name cannot be empty")
// 	}
// 	if len(name) > 50 {
// 		return NewQuestionnaireError("pet name is too long (max 50 characters)")
// 	}
// 	return nil
// }
//
// func validateSpecies(species string) error {
// 	validSpecies := map[string]bool{
// 		"dog": true,
// 		"cat": true,
// 	}
// 	if !validSpecies[species] {
// 		return NewQuestionnaireError("invalid species type")
// 	}
// 	return nil
// }
//
// func validateBreed(breed string) error {
// 	if len(breed) < 1 {
// 		return NewQuestionnaireError("breed cannot be empty")
// 	}
// 	if len(breed) > 50 {
// 		return NewQuestionnaireError("breed name is too long (max 50 characters)")
// 	}
// 	return nil
// }
//
// func validateDateOfBirth(dob string) error {
// 	_, err := time.Parse("2006-01-02", dob)
// 	if err != nil {
// 		return NewQuestionnaireError("invalid date format, use DD-MM-YYYY")
// 	}
// 	return nil
// }
//
// func validateGender(gender string) error {
// 	if gender != "male" && gender != "female" {
// 		return NewQuestionnaireError("gender must be either 'male' or 'female'")
// 	}
// 	return nil
// }
//
// func validateWeight(weight string) error {
// 	var w float64
// 	_, err := fmt.Sscanf(weight, "%f", &w)
// 	if err != nil {
// 		return NewQuestionnaireError("weight must be a valid number")
// 	}
// 	if w <= 0 || w > 500 {
// 		return NewQuestionnaireError("weight must be between 0 and 500 kg")
// 	}
// 	return nil
// }
