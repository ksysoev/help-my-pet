package conversation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPetProfileQuestionnaireState(t *testing.T) {
	state := NewPetProfileQuestionnaireState()
	assert.NotNil(t, state)
	assert.Equal(t, 6, len(state.QAPairs))
	assert.Equal(t, 0, state.CurrentIndex)
}

func TestGetCurrentQuestion(t *testing.T) {
	state := NewPetProfileQuestionnaireState()

	question, err := state.GetCurrentQuestion()
	assert.NoError(t, err)
	assert.Equal(t, "What is your pet's name?", question.Text)
}

func TestGetCurrentQuestion_NoMoreQuestions(t *testing.T) {
	state := NewPetProfileQuestionnaireState()
	state.CurrentIndex = len(state.QAPairs)

	question, err := state.GetCurrentQuestion()
	assert.Nil(t, question)
	assert.Equal(t, ErrNoMoreQuestions, err)
}

func TestProcessAnswer(t *testing.T) {
	state := NewPetProfileQuestionnaireState()

	done, err := state.ProcessAnswer("Buddy")
	assert.NoError(t, err)
	assert.False(t, done)
	assert.Equal(t, "Buddy", state.QAPairs[0].Answer)
	assert.Equal(t, 1, state.CurrentIndex)
}

func TestProcessAnswer_Complete(t *testing.T) {
	state := NewPetProfileQuestionnaireState()
	for _ = range state.QAPairs[:len(state.QAPairs)-1] {
		done, _ := state.ProcessAnswer("Answer")
		assert.False(t, done)
	}

	done, err := state.ProcessAnswer("Final Answer")
	assert.NoError(t, err)
	assert.True(t, done)
}

func TestProcessAnswer_NoMoreQuestions(t *testing.T) {
	state := NewPetProfileQuestionnaireState()
	state.CurrentIndex = len(state.QAPairs)

	done, err := state.ProcessAnswer("Answer")
	assert.False(t, done)
	assert.Equal(t, ErrNoMoreQuestions, err)
}

func TestGetResults(t *testing.T) {
	state := NewPetProfileQuestionnaireState()
	for _ = range state.QAPairs {
		state.ProcessAnswer("Answer")
	}

	results, err := state.GetResults()
	assert.NoError(t, err)
	assert.Equal(t, len(state.QAPairs), len(results))
}

func TestGetResults_Incomplete(t *testing.T) {
	state := NewPetProfileQuestionnaireState()

	results, err := state.GetResults()
	assert.Nil(t, results)
	assert.Equal(t, ErrQuestionnaireIncomplete, err)
}
