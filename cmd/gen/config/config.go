package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/peterbourgon/ff/v3"
)

// Config holds the configuration for the application.
type Config struct {
	Provider string
	Gemini   GeminiConfig
	OpenAI   OpenAIConfig
	Ollama   OllamaConfig
	Anthropic AnthropicConfig
}

// GeminiConfig holds the configuration for the Gemini provider.
type GeminiConfig struct {
	APIKey string
	Model  string
}

// OpenAIConfig holds the configuration for the OpenAI provider.
type OpenAIConfig struct {
	APIKey string
	Model  string
}

// OllamaConfig holds the configuration for the Ollama provider.
type OllamaConfig struct {
	Host  string
	Model string
}

// AnthropicConfig holds the configuration for the Anthropic provider.
type AnthropicConfig struct {
	APIKey string
	Model  string
}

// Load loads the configuration from a file, environment variables, and flags.
func Load(version, commit, date string) (*Config, error) {
	fs := flag.NewFlagSet("gen", flag.ExitOnError)
	var (
		provider       = fs.String("provider", "gemini", "LLM provider to use (gemini, openai, ollama, or anthropic)")
		geminiAPIKey   = fs.String("gemini-api-key", "", "Gemini API key")
		geminiModel    = fs.String("gemini-model", "gemini-1.5-flash", "Gemini model to use")
		openaiAPIKey   = fs.String("openai-api-key", "", "OpenAI API key")
		openaiModel    = fs.String("openai-model", "gpt-4o", "OpenAI model to use")
		ollamaHost     = fs.String("ollama-host", "http://localhost:11434", "Ollama host")
		ollamaModel    = fs.String("ollama-model", "llama2", "Ollama model")
		anthropicAPIKey = fs.String("anthropic-api-key", "", "Anthropic API key")
		anthropicModel  = fs.String("anthropic-model", "claude-3-opus-20240229", "Anthropic model to use")
		configPath     = fs.String("config", "", "path to config file")
		showVersion    = fs.Bool("version", false, "show version")
	)

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	// Default config path
	defaultConfigPath := filepath.Join(home, ".gen", "config")

	// If the config flag is not set, use the default path
	if *configPath == "" {
		*configPath = defaultConfigPath
	}

	// Create a new config
	cfg := &Config{}

	// Parse the config file, flags, and env vars.
	err = ff.Parse(fs, os.Args[1:],
		ff.WithEnvVarPrefix("GEN"),
		ff.WithConfigFile(*configPath),
		ff.WithConfigFileParser(ff.PlainParser),
	)
	if err != nil {
		return nil, err
	}

	cfg.Provider = *provider
	cfg.Gemini.APIKey = *geminiAPIKey
	cfg.Gemini.Model = *geminiModel
	cfg.OpenAI.APIKey = *openaiAPIKey
	cfg.OpenAI.Model = *openaiModel
	cfg.Ollama.Host = *ollamaHost
	cfg.Ollama.Model = *ollamaModel
	cfg.Anthropic.APIKey = *anthropicAPIKey
	cfg.Anthropic.Model = *anthropicModel

	if *showVersion {
		fmt.Printf("gen version %s (commit: %s, built at: %s)\n", version, commit, date)
		os.Exit(0)
	}

	return cfg, nil
}
