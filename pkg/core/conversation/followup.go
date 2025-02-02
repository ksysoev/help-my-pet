package conversation

// FollowUpQuestionnaireState represents the state for follow-up questions from LLM
type FollowUpQuestionnaireState struct {
	InitialPrompt string           `json:"initial_prompt"`
	QAPairs       []QuestionAnswer `json:"qa_pairs"`
	CurrentIndex  int              `json:"current_index"`
}

func (f *FollowUpQuestionnaireState) GetCurrentQuestion() (*Question, error) {
	if f.CurrentIndex >= len(f.QAPairs) {
		return nil, ErrNoMoreQuestions
	}
	return &f.QAPairs[f.CurrentIndex].Question, nil
}

func (f *FollowUpQuestionnaireState) ProcessAnswer(answer string) (bool, error) {
	if f.CurrentIndex >= len(f.QAPairs) {
		return false, ErrNoMoreQuestions
	}

	f.QAPairs[f.CurrentIndex].Answer = answer
	f.CurrentIndex++

	return f.CurrentIndex >= len(f.QAPairs), nil
}

func (f *FollowUpQuestionnaireState) GetResults() ([]QuestionAnswer, error) {
	if f.CurrentIndex < len(f.QAPairs) {
		return nil, ErrQuestionnaireIncomplete
	}

	return f.QAPairs, nil
}
