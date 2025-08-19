package llm

// LLMProvider defines the interface for a language model provider.
type LLMProvider interface {
	GenerateCommand(prompt, shell string) (string, error)
}
