package llm

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	anthropic "github.com/liushuangls/go-anthropic"
)

// AnthropicProvider is an implementation of LLMProvider for Anthropic.
type AnthropicProvider struct {
	CreateMessages func(context.Context, anthropic.MessagesRequest) (anthropic.MessagesResponse, error)
	Model          string
}

// GenerateCommand generates a command using the Anthropic LLM.
func (p *AnthropicProvider) GenerateCommand(ctx context.Context, logger *slog.Logger, prompt, shell string) (string, error) {
	fullPrompt := fmt.Sprintf(`Given the following prompt, generate a single shell command. The command should be able to be executed on a %s machine in a %s shell. The command should be reasonable and not destructive. Return only the command, with no explanation or other text.

Prompt: %s`, os.Getenv("GOOS"), shell, prompt)
	logger.Debug("anthropic prompt", "prompt", fullPrompt)

	resp, err := p.CreateMessages(
		ctx,
		anthropic.MessagesRequest{
			Model: p.Model,
			Messages: []anthropic.Message{
				anthropic.NewUserTextMessage(fullPrompt),
			},
			MaxTokens: 1000, // A reasonable default, can be made configurable if needed
		},
	)
	if err != nil {
		return "", err
	}

	if len(resp.Content) > 0 && resp.Content[0].Type == "text" {
		logger.Debug("anthropic response", "response", resp.Content[0].Text)
		return resp.Content[0].Text, nil
	}

	return "", fmt.Errorf("no command generated")
}
