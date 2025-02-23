package anthropic

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/ksysoev/help-my-pet/pkg/core/message"
)

// Model defines the interface for LLM interactions
type Model interface {
	Call(ctx context.Context, systemPrompts, request string, imgs []*message.Image) (string, error)
}

const coreGuidelines = `Core Guidelines strictly:
1. Language Detection and Response:
  - ALWAYS analyze the language of the user's input first
  - MUST respond in the EXACT SAME language as the user's question
  - If unable to confidently detect the language, default to English
  - Your internal thinking process should ALWAYS be in English regardless of response language
2. Topics and Scope:
  You are only allowed to answer questions related to these topics:
    • Pet health and behavior questions
    • Diet and nutrition advice
    • Training tips and techniques
    • General pet care guidance
3. Guidelines:
  - Topic Boundaries:
    - If a question is not related to the allowed topics, politely decline and explain your limitations
    - Stay focused on pet care within your defined scope
  - Health Safety Protocol:
    - If symptoms could indicate a serious health issue, recommend veterinary care
    - When discussing health topics, recommend professional veterinary consultation
    - Do not attempt to diagnose without sufficient information
4. Language and Tone:
  - Use clear, simple, and professional language
  - Avoid jargon, slang, or overly technical terms
  - Be empathetic, supportive, and non-judgmental in your responses
5. Request structure:
  - System Information: This section contains system-specific information that may help provide accurate advice or context.
  - Pet Profile: This section contains the user's pet profile information, you can use this information to provide more accurate advice
  - Previous conversation - this section contains previous messages of current conversation, this section may contain 3 types of message:
    - user: user's message or question
	- assistant: assistant's response
	- questionnaire: user's responses to the assistant's questions
    - media_description: detailed description of media content provided by user
  - Follow-up information - this section contains the assistant's follow-up questions and user's responses. You should analyze last user's question and last assistant's response from the previous conversation section and based on follow-up information section, you provide final response. You SHOULD NOT ask additional question at this point. 
  - Current question - this section contains the user's current question. You should analyze this question and if information is not enough, you can ask additional questions to get more details for you diagnosis. You can use previous conversation section to get more context.
`

// anthropicModel adapts the Anthropic API to the Model interface
type anthropicModel struct {
	client    *anthropic.Client
	modelID   string
	maxTokens int
}

// newAnthropicModel initializes a new instance of anthropicModel configured with the provided parameters.
// It sets up the Anthropic client using the API key, associates the specified model ID, and configures the maximum tokens.
// apiKey is the authentication key used to interact with the Anthropic API.
// modelID identifies the target model to use, such as "claude-2".
// maxTokens determines the maximum tokens allowed per request, enforcing output length constraints.
// Returns a pointer to an anthropicModel instance for interacting with Anthropic API and an error if client initialization fails.
func newAnthropicModel(apiKey string, modelID string, maxTokens int) (*anthropicModel, error) {
	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	if client == nil {
		return nil, fmt.Errorf("failed to create Anthropic client")
	}

	return &anthropicModel{
		client:    client,
		modelID:   modelID,
		maxTokens: maxTokens,
	}, nil
}

// Call sends a request to the Anthropic API with system prompts, user inputs, and optional images.
// It constructs the message payload using the provided request string, images, and context system settings.
// ctx is the execution context for the request, systemPrompts is the contextual system instruction,
// request is the user's input, and imgs is a slice of images to include in the request.
// Returns the response text from the API and an error if the request fails or the API response is invalid.
func (m *anthropicModel) Call(ctx context.Context, systemPrompts, request string, imgs []*message.Image) (string, error) {
	blocks := []anthropic.ContentBlockParamUnion{anthropic.NewTextBlock(request)}

	for _, img := range imgs {
		blocks = append(blocks, anthropic.NewImageBlockBase64(img.MIME, img.Data))
	}

	msg, err := m.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.F(m.modelID),
		MaxTokens: anthropic.F(int64(m.maxTokens)),
		System: anthropic.F([]anthropic.TextBlockParam{
			anthropic.NewTextBlock(coreGuidelines),
			anthropic.NewTextBlock(systemPrompts),
		}),
		Messages: anthropic.F([]anthropic.MessageParam{anthropic.NewUserMessage(blocks...)}),
	})

	if err != nil {
		return "", fmt.Errorf("failed to call Anthropic API: %w", err)
	}

	if len(msg.Content) == 0 {
		return "", fmt.Errorf("empty response from Anthropic API")
	}

	return msg.Content[0].Text, nil
}
