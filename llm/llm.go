package llm

import (
	"context"
)

// LLMProvider defines the interface for a language model provider.
type LLMProvider interface {
	GenerateCommand(ctx context.Context, prompt, shell string) (string, error)
}
