package anthropic

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/ksysoev/help-my-pet/pkg/core/message"
)

// Model defines the interface for LLM interactions
type Model interface {
	// Call sends a request to the LLM with a user question and format instructions
	// formatInstructions contain system-level prompting and formatting guidelines
	// question is the user's actual querys
	Call(ctx context.Context, formatInstructions string, question string, imgs []*message.Image) (string, error)
	DescribeImages(ctx context.Context, imgs []*message.Image) (string, error)
}

const systemPrompt = `You are a helpful veterinary AI assistant. Follow these language rules strictly:

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

3. Core Guidelines:
  - Information Gathering:
    - Ask specific follow-up questions to gather necessary details
    - For health issues, ask about symptoms, duration, pet's age, breed, and relevant history
    - For behavior questions, ask about the context, frequency, and environmental factors
	- For diet questions, ask about the pet's age, weight, activity level, and any health conditions
  - Topic Boundaries:
    - If a question is not related to the allowed topics, politely decline and explain your limitations
    - Stay focused on pet care within your defined scope
  - Health Safety Protocol:
    - If symptoms could indicate a serious health issue, recommend veterinary care
    - When discussing health topics, recomend professional veterinary consultation
    - Do not attempt to diagnose without sufficient informatio

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
  - Follow-up information - this section contains the assistant's follow-up questions and user's responses. You should analyze last user's question and last assistant's response from the previous conversation section and based on follow-up information section, you provide final response. You SHOULD NOT ask additional question at this point. 
  - Current question - this section contains the user's current question. You should analyze this question and if information is not enough, you can ask additional questions to get more details for you dianosis. You can use previous conversation section to get more context.

Please provide accurate, helpful, and compassionate advice while following these guidelines strictly.
`

// anthropicModel adapts the Anthropic API to the Model interface
type anthropicModel struct {
	client    *anthropic.Client
	modelID   string
	maxTokens int
}

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

func (m *anthropicModel) Call(ctx context.Context, formatInstructions string, question string, imgs []*message.Image) (string, error) {
	slog.DebugContext(ctx, "Anthropic LLM call", slog.String("format_instructions", formatInstructions), slog.String("question", question))

	blocks := []anthropic.ContentBlockParamUnion{anthropic.NewTextBlock(question)}

	for _, img := range imgs {
		blocks = append(blocks, anthropic.NewImageBlockBase64(img.MIME, img.Data))
	}

	msg, err := m.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.F(m.modelID),
		MaxTokens: anthropic.F(int64(m.maxTokens)),
		System: anthropic.F([]anthropic.TextBlockParam{
			anthropic.NewTextBlock(systemPrompt),
			anthropic.NewTextBlock(formatInstructions),
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
