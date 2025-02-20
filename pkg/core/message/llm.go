package message

// LLMResult represents a structured response from the LLM
type LLMResult struct {
	Text      string     `json:"text"`
	Questions []Question `json:"questions"`
	Media     string     `json:"media,omitempty"`
	Thoughts  string     `json:"thoughts"`
}

// Question represents a follow-up question with optional predefined answers
type Question struct {
	Text    string   `json:"text"`
	Answers []string `json:"answers,omitempty"`
}
