package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/zombor/gen/cmd/gen/config"
	"github.com/zombor/gen/cmd/gen/tui"
	"github.com/zombor/gen/llm"

	"github.com/google/generative-ai-go/genai"
	"github.com/liushuangls/go-anthropic"
	"github.com/ollama/ollama/api"
	openai "github.com/sashabaranov/go-openai"
	opts "google.golang.org/api/option"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func getShell() string {
	shellPath := os.Getenv("SHELL")
	if shellPath == "" {
		return "sh"
	}
	return filepath.Base(shellPath)
}

func main() {
	cfg, args, err := config.Load(version, commit, date)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(ioutil.Discard, nil))
	if cfg.Debug {
		logger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	}
	slog.SetDefault(logger)

	ctx := context.Background()
	var provider llm.LLMProvider

	switch cfg.Provider {
	case "gemini":
		client, err := genai.NewClient(ctx, opts.WithAPIKey(cfg.Gemini.APIKey))
		if err != nil {
			log.Fatal(err)
		}
		defer client.Close()
		model := client.GenerativeModel(cfg.Gemini.Model)
		provider = &llm.GeminiProvider{GenerateContent: model.GenerateContent}
	case "openai":
		client := openai.NewClient(cfg.OpenAI.APIKey)
		provider = &llm.OpenAIProvider{CreateChatCompletion: client.CreateChatCompletion, Model: cfg.OpenAI.Model}
	case "ollama":
		hostURL, err := url.Parse(cfg.Ollama.Host)
		if err != nil {
			log.Fatal(err)
		}
		client := api.NewClient(hostURL, &http.Client{})
		provider = llm.NewOllamaProvider(client, cfg.Ollama.Model)
	case "anthropic":
		client := anthropic.NewClient(cfg.Anthropic.APIKey)
		provider = &llm.AnthropicProvider{CreateMessages: client.CreateMessages, Model: cfg.Anthropic.Model}
	case "bedrock":
		bedrockClient, err := llm.NewBedrock(ctx, cfg.Bedrock.Model, cfg.Bedrock.Region, cfg.Bedrock.InferenceProfile)
		if err != nil {
			log.Fatal(err)
		}
		provider = bedrockClient
	default:
		fmt.Printf("Unknown provider: %s\n", cfg.Provider)
		os.Exit(1)
	}

	prompt := strings.Join(args, " ")

	if cfg.TUI {
		model := tui.NewModel(prompt, provider)
		finalModel, err := tui.Run(model)
		if err != nil {
			fmt.Printf("Error running tui: %v\n", err)
			os.Exit(1)
		}
		m := finalModel.(tui.Model)
		if m.Accepted() {
			runCommand(m.Command())
		}
	} else {
		if prompt == "" {
			fmt.Println("Usage: gen <prompt>")
			os.Exit(1)
		}

		shell := getShell()
		command, err := provider.GenerateCommand(ctx, logger, prompt, shell)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		// The model sometimes returns the command wrapped in backticks, so we remove them.
		command = strings.Trim(command, "`")

		fmt.Printf("Generated command: \n\n%s\n\n", command)
		fmt.Print("Execute? (y/N) ")

		var response string
		fmt.Scanln(&response)

		if strings.ToLower(response) == "y" {
			runCommand(command)
		} else {
			fmt.Println("Command execution aborted.")
		}
	}
}

func runCommand(command string) {
	shell := getShell()
	cmd := exec.Command(shell, "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error executing command: %v\n", err)
		os.Exit(1)
	}
}
