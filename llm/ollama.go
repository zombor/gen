package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ollama/ollama/api"
)

// OllamaProvider is an implementation of LLMProvider for Ollama.
type OllamaProvider struct {
	Generate func(context.Context, *api.GenerateRequest, api.GenerateResponseFunc) error
	Model    string
}

func NewOllamaProvider(client *api.Client, model string) *OllamaProvider {
	return &OllamaProvider{
		Generate: client.Generate,
		Model:    model,
	}
}

// GenerateCommand generates a command using the Ollama LLM.
func (p *OllamaProvider) GenerateCommand(ctx context.Context, prompt, shell string) (string, error) {
	req := &api.GenerateRequest{
		Model:  p.Model,
		Format: json.RawMessage(`"json"`),
		Prompt: fmt.Sprintf(`Given the following prompt, generate a single shell command. The command should be able to be executed on a %s machine in a %s shell. The command should be reasonable and not destructive. Return the command in a json object with a single key "command".\n\nPrompt: %s`, os.Getenv("GOOS"), shell, prompt),
	}

	var response string
	respCh := make(chan *api.GenerateResponse)
	errCh := make(chan error, 1)
	go func() {
		defer close(respCh)
		defer close(errCh)
		err := p.Generate(ctx, req, func(r api.GenerateResponse) error {
			respCh <- &r
			return nil
		})
		if err != nil {
			errCh <- err
		}
	}()

	for resp := range respCh {
		response += resp.Response
	}

	if err := <-errCh; err != nil {
		return "", err
	}

	var command struct {
		Command string `json:"command"`
	}
	if err := json.Unmarshal([]byte(response), &command); err != nil {
		return "", fmt.Errorf("failed to unmarshal response from ollama: %w", err)
	}

	return command.Command, nil
}
