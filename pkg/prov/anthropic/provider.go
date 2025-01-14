package anthropic

import (
	"context"
	"fmt"

	"github.com/ksysoev/help-my-pet/pkg/core"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
)

const systemPrompt = `You are a helpful veterinary AI assistant. You are only allowed to answer questions related to the following topics:

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
   - If symptoms could indicate a serious health issue, immediately recommend veterinary care
   - When discussing health topics, emphasize the importance of professional veterinary consultation
   - Do not attempt to diagnose without sufficient information

4. Response Format:
   Your response must be a valid JSON object with the following structure:
   {
     "text": "Your main response text here",
     "questions": [
       {
         "text": "Follow-up question text",
         "answers": [  // Optional predefined answers
           {
             "text": "Display text for the answer",
             "value": "Internal value for the answer"
           }
         ]
       }
     ]
   }

Please provide accurate, helpful, and compassionate advice while following these guidelines strictly.`

type Config struct {
	APIKey    string `mapstructure:"api_key"`
	Model     string `mapstructure:"model"`
	MaxTokens int    `mapstructure:"max_tokens"`
}

type Provider struct {
	caller LLMCaller
	model  string
	config Config
}

func New(cfg Config) (*Provider, error) {
	llm, err := anthropic.New(anthropic.WithToken(cfg.APIKey))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Anthropic LLM: %w", err)
	}

	return &Provider{
		caller: NewLLMAdapter(llm),
		model:  cfg.Model,
		config: cfg,
	}, nil
}

func (p *Provider) Call(ctx context.Context, prompt string, options ...llms.CallOption) (*core.Response, error) {
	defaultOptions := []llms.CallOption{
		llms.WithModel(p.model),
		llms.WithMaxTokens(p.config.MaxTokens),
	}
	options = append(defaultOptions, options...)

	fullPrompt := fmt.Sprintf("%s\n\nQuestion: %s", systemPrompt, prompt)
	response, err := p.caller.Call(ctx, fullPrompt, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to call Anthropic LLM: %w", err)
	}

	structuredResponse, err := core.ParseResponse(response)
	if err != nil {
		// If parsing fails, create a simple response with just the text
		return &core.Response{
			Text:      response,
			Questions: []core.Question{},
		}, nil
	}

	return structuredResponse, nil
}
