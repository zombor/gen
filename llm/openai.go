package llm

import (
	"context"
	"fmt"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

// OpenAIProvider is an implementation of LLMProvider for OpenAI.
type OpenAIProvider struct {
	CreateChatCompletion func(context.Context, openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error)
	Model                string
}

// GenerateCommand generates a command using the OpenAI LLM.
func (p *OpenAIProvider) GenerateCommand(ctx context.Context, prompt, shell string) (string, error) {
	resp, err := p.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: p.Model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role: openai.ChatMessageRoleUser,
					Content: fmt.Sprintf(`Given the following prompt, generate a single shell command. The command should be able to be executed on a %s machine in a %s shell. The command should be reasonable and not destructive. Return only the command, with no explanation or other text.

Prompt: %s`, os.Getenv("GOOS"), shell, prompt),
				},
			},
		},
	)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) > 0 {
		return resp.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no command generated")
}
