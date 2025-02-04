package message

// Response represents the structured response from the AI service
type Response struct {
	Message string   `json:"message"` // Main response message
	Answers []string `json:"answers"` // Possible answers for the follow-up question
}

// NewResponse creates a new Response
func NewResponse(message string, answers []string) *Response {
	return &Response{
		Message: message,
		Answers: answers,
	}
}
