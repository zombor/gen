package llm

import (
	"context"
	"fmt"
	"os"

	"github.com/ollama/ollama/api"
)

// OllamaProvider is an implementation of LLMProvider for Ollama.
type OllamaProvider struct {
	Client *api.Client
	Model  string
}

// GenerateCommand generates a command using the Ollama LLM.
func (p *OllamaProvider) GenerateCommand(prompt, shell string) (string, error) {
	ctx := context.Background()
	req := &api.GenerateRequest{
		Model:  p.Model,
		Prompt: fmt.Sprintf("Given the following prompt, generate a single shell command. The command should be able to be executed on a %s machine in a %s shell. The command should be reasonable and not destructive. Return only the command, with no explanation or other text.\n\nPrompt: %s", os.Getenv("GOOS"), shell, prompt),
	}

	var response string
	respCh := make(chan *api.GenerateResponse)
	go func() {
		defer close(respCh)
		err := p.Client.Generate(ctx, req, func(r api.GenerateResponse) error {
			respCh <- &r
			return nil
		})
		if err != nil {
			// TODO: Handle error
		}
	}()

	for resp := range respCh {
		response += resp.Response
	}

	return response, nil
}