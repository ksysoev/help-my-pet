package message

// LLMResult represents a structured response from the LLM
type LLMResult struct {
	Text      string     `json:"text"`
	Questions []Question `json:"questions"`
	Media     string     `json:"media,omitempty"`
	Reasoning string     `json:"reasoning"`
}

// Question represents a follow-up question with optional predefined answers
type Question struct {
	Text    string   `json:"text"`
	Reason  string   `json:"reason,omitempty"`
	Answers []string `json:"answers,omitempty"`
}
