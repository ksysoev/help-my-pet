package message

// LLMResult represents a structured response from the LLM
type LLMResult struct {
	Text         string     `json:"text"`                   // Main response text
	Questions    []Question `json:"questions"`              // Optional follow-up questions
	Observations string     `json:"observations,omitempty"` // Optional observations
}

// Question represents a follow-up question with optional predefined answers
type Question struct {
	Text    string   `json:"text"`
	Answers []string `json:"answers,omitempty"`
}
