package conversation

import (
	"context"
	"testing"

	"github.com/ksysoev/help-my-pet/pkg/core/message"
	"github.com/stretchr/testify/assert"
)

func TestValidateLength(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxLength int
		wantErr   error
	}{
		{"valid length", "short text", 100, nil},
		{"exact length", "12345", 5, nil},
		{"exceeds length", "This text is longer than allowed", 10, message.ErrTextTooLong},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLength(tt.input, tt.maxLength)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestValidateDOB(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{"valid date", "2010-12-31", nil},
		{"invalid format", "12-31-2010", message.ErrInvalidDates},
		{"future date", "2030-01-01", message.ErrFutureDate},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDOB(tt.input)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestNewPetProfileQuestionnaireState(t *testing.T) {
	state := NewPetProfileQuestionnaireState(context.Background())
	assert.NotNil(t, state)
	assert.Equal(t, 10, len(state.QAPairs))
	assert.Equal(t, 0, state.CurrentIndex)
}

func TestGetCurrentQuestion(t *testing.T) {
	state := NewPetProfileQuestionnaireState(context.Background())

	question, err := state.GetCurrentQuestion()
	assert.NoError(t, err)
	assert.Equal(t, "What is your pet's name?", question.Text)
}

func TestGetCurrentQuestion_NoMoreQuestions(t *testing.T) {
	state := NewPetProfileQuestionnaireState(context.Background())
	state.CurrentIndex = len(state.QAPairs)

	question, err := state.GetCurrentQuestion()
	assert.Nil(t, question)
	assert.Equal(t, ErrNoMoreQuestions, err)
}

func TestProcessAnswer(t *testing.T) {
	state := NewPetProfileQuestionnaireState(context.Background())

	// Valid answer
	done, err := state.ProcessAnswer("Buddy")
	assert.NoError(t, err)
	assert.False(t, done)
	assert.Equal(t, "Buddy", state.QAPairs[0].Answer)
	assert.Equal(t, 1, state.CurrentIndex)

	// Answer with excessive length
	done, err = state.ProcessAnswer("This answer is intentionally too long to trigger validation error.............................................")
	assert.Equal(t, message.ErrTextTooLong, err)
	assert.False(t, done)

	// Invalid date
	state.CurrentIndex = 3 // Set index to DOB question
	done, err = state.ProcessAnswer("invalid-date")
	assert.Equal(t, message.ErrInvalidDates, err)
	assert.False(t, done)
}

func TestProcessAnswer_Complete(t *testing.T) {
	state := NewPetProfileQuestionnaireState(context.Background())
	for range state.QAPairs[:len(state.QAPairs)-1] {
		done, _ := state.ProcessAnswer("2000-01-01")
		assert.False(t, done)
	}

	done, err := state.ProcessAnswer("Final Answer")
	assert.NoError(t, err)
	assert.True(t, done)
}

func TestProcessAnswer_NoMoreQuestions(t *testing.T) {
	state := NewPetProfileQuestionnaireState(context.Background())
	state.CurrentIndex = len(state.QAPairs)

	done, err := state.ProcessAnswer("Answer")
	assert.False(t, done)
	assert.Equal(t, ErrNoMoreQuestions, err)
}

func TestGetResults(t *testing.T) {
	state := NewPetProfileQuestionnaireState(context.Background())
	for range state.QAPairs {
		_, err := state.ProcessAnswer("2000-01-01")
		assert.NoError(t, err)
	}

	results, err := state.GetResults()
	assert.NoError(t, err)
	assert.Equal(t, len(state.QAPairs), len(results))
}

func TestGetResults_Incomplete(t *testing.T) {
	state := NewPetProfileQuestionnaireState(context.Background())

	results, err := state.GetResults()
	assert.Nil(t, results)
	assert.Equal(t, ErrQuestionnaireIncomplete, err)
}
