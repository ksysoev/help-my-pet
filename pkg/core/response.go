package core

// PetAdviceResponse represents the structured response from the AI service
type PetAdviceResponse struct {
	Message string   `json:"message"` // Main response message
	Answers []string `json:"answers"` // Possible answers for the follow-up question
}

// NewPetAdviceResponse creates a new PetAdviceResponse
func NewPetAdviceResponse(message string, answers []string) *PetAdviceResponse {
	return &PetAdviceResponse{
		Message: message,
		Answers: answers,
	}
}
