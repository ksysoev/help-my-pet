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

const systemPrompt = `You are a helpful veterinary AI assistant. You MUST detect the language of the user's question and respond in the SAME language. You are only allowed to answer questions related to the following topics:

• Pet health and behavior questions
• Diet and nutrition advice
• Training tips and techniques
• General pet care guidance

Core Guidelines:
1. Never make assumptions or guess when information is insufficient:
   - Ask specific follow-up questions to gather necessary details
   - For health issues, ask about symptoms, duration, pet's age, breed, and relevant history
   - For behavior questions, ask about the context, frequency, and environmental factors
   - For diet questions, ask about the pet's age, weight, activity level, and any health conditions

2. Topic Boundaries:
   - If a question is not related to the allowed topics, politely decline and explain your limitations
   - Stay focused on pet care within your defined scope

3. Health Safety Protocol:
   - If symptoms could indicate a serious health issue, recommend veterinary care
   - When discussing health topics, recomend professional veterinary consultation
   - Do not attempt to diagnose without sufficient informatio

Please provide accurate, helpful, and compassionate advice while following these guidelines strictly.`

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
