package llm

import (
	"context"
	"fmt"
	"os"

	"github.com/google/generative-ai-go/genai"
)

// GeminiProvider is an implementation of LLMProvider for Google's Gemini.
type GeminiProvider struct {
	GenerateContent func(context.Context, ...genai.Part) (*genai.GenerateContentResponse, error)
}

// GenerateCommand generates a command using the Gemini LLM.
func (p *GeminiProvider) GenerateCommand(ctx context.Context, prompt, shell string) (string, error) {
	resp, err := p.GenerateContent(ctx, genai.Text(fmt.Sprintf("Given the following prompt, generate a single shell command. The command should be able to be executed on a %s machine in a %s shell. The command should be reasonable and not destructive. Return only the command, with no explanation or other text.\n\nPrompt: %s", os.Getenv("GOOS"), shell, prompt)))

	if err != nil {
		return "", err
	}

	if len(resp.Candidates) > 0 {
		for _, part := range resp.Candidates[0].Content.Parts {
			if txt, ok := part.(genai.Text); ok {
				return string(txt), nil
			}
		}
	}

	return "", fmt.Errorf("no command generated")
}
