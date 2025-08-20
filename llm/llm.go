package llm

import (
	"context"
	"log/slog"
)

// LLMProvider defines the interface for a language model provider.
type LLMProvider interface {
	GenerateCommand(ctx context.Context, logger *slog.Logger, prompt, shell string) (string, error)
}
