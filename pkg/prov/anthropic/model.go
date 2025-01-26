package anthropic

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// Model defines the interface for LLM interactions
type Model interface {
	// Call sends a request to the LLM with a user question and format instructions
	// formatInstructions contain system-level prompting and formatting guidelines
	// question is the user's actual querys
	Call(ctx context.Context, formatInstructions string, question string) (string, error)
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
4. Information Gathering:
   - Ask specific follow-up questions to gather necessary details
   - For health issues, ask about symptoms, duration, pet's age, breed, and relevant history
   - For behavior questions, ask about the context, frequency, and environmental factors
   - For diet questions, ask about the pet's age, weight, activity level, and any health conditions

5. Topic Boundaries:
   - If a question is not related to the allowed topics, politely decline and explain your limitations
   - Stay focused on pet care within your defined scope

6. Health Safety Protocol:
   - If symptoms could indicate a serious health issue, recommend veterinary care
   - When discussing health topics, recomend professional veterinary consultation
   - Do not attempt to diagnose without sufficient informatio

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

func (m *anthropicModel) Call(ctx context.Context, formatInstructions string, question string) (string, error) {
	slog.Debug("Anthropic LLM call", slog.String("format_instructions", formatInstructions), slog.String("question", question))

	message, err := m.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.F(m.modelID),
		MaxTokens: anthropic.F(int64(m.maxTokens)),
		System: anthropic.F([]anthropic.TextBlockParam{
			anthropic.NewTextBlock(systemPrompt),
			anthropic.NewTextBlock(formatInstructions),
		}),
		Messages: anthropic.F([]anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(question)),
		}),
	})

	if err != nil {
		return "", fmt.Errorf("failed to call Anthropic API: %w", err)
	}

	if len(message.Content) == 0 {
		return "", fmt.Errorf("empty response from Anthropic API")
	}

	return message.Content[0].Text, nil
}
