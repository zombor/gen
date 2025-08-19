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
	Provider    string
	APIKey      string
	GeminiModel string
	Ollama      OllamaConfig
}

// OllamaConfig holds the configuration for the Ollama provider.
type OllamaConfig struct {
	Host  string
	Model string
}

// Load loads the configuration from a file, environment variables, and flags.
func Load(version, commit, date string) (*Config, error) {
	fs := flag.NewFlagSet("gen", flag.ExitOnError)
	var (
		provider    = fs.String("provider", "gemini", "LLM provider to use (gemini or ollama)")
		apiKey      = fs.String("api-key", "", "Gemini API key")
		geminiModel = fs.String("gemini-model", "gemini-1.5-flash", "Gemini model to use")
		ollamaHost  = fs.String("ollama-host", "http://localhost:11434", "Ollama host")
		ollamaModel = fs.String("ollama-model", "llama2", "Ollama model")
		configPath  = fs.String("config", "", "path to config file")
		showVersion = fs.Bool("version", false, "show version")
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
	cfg.APIKey = *apiKey
	cfg.GeminiModel = *geminiModel
	cfg.Ollama.Host = *ollamaHost
	cfg.Ollama.Model = *ollamaModel

	if *showVersion {
		fmt.Printf("gen version %s (commit: %s, built at: %s)\n", version, commit, date)
		os.Exit(0)
	}

	return cfg, nil
}
