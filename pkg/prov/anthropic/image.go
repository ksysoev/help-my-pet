package anthropic

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/ksysoev/help-my-pet/pkg/core/message"
)

const imageProccesorPrompt = `You are a helpful veterinary AI assistant. 
Your task is to anylyze the images that provided by the user and provide detailed description of the image(s).
Your report will be used by the veterinarian to diagnose the pet's condition and provide appropriate treatment, so it is important to be as accurate as possible.
Explain your thoughts and reasoning in detail, and provide any relevant information that may help the veterinarian make an informed decision.
`

func (m *anthropicModel) DescribeImages(ctx context.Context, imgs []*message.Image) (string, error) {
	if len(imgs) == 0 {
		return "", fmt.Errorf("no images provided")
	}

	var blocks []anthropic.ContentBlockParamUnion
	for _, img := range imgs {
		blocks = append(blocks, anthropic.NewImageBlockBase64(img.MIME, img.Data))
	}

	msg, err := m.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.F("claude-3-5-sonnet-latest"),
		MaxTokens: anthropic.F(int64(m.maxTokens)),
		System: anthropic.F([]anthropic.TextBlockParam{
			anthropic.NewTextBlock(systemPrompt),
		}),
		Messages: anthropic.F([]anthropic.MessageParam{anthropic.NewUserMessage(blocks...)}),
	})

	if err != nil {
		return "", fmt.Errorf("failed to call Anthropic API: %w", err)
	}

	if len(msg.Content) == 0 {
		return "", fmt.Errorf("empty response from Anthropic API")
	}

	slog.Debug("Anthropic image description", slog.String("response", msg.Content[0].Text))

	return msg.Content[0].Text, nil
}
