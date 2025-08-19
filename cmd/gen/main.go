package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/zombor/gen/cmd/gen/config"
	"github.com/zombor/gen/cmd/gen/tui"
	"github.com/zombor/gen/llm"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/generative-ai-go/genai"
	"github.com/ollama/ollama/api"
	opts "google.golang.org/api/option"
	openai "github.com/sashabaranov/go-openai"
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
	cfg, err := config.Load(version, commit, date)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	var provider llm.LLMProvider

	switch cfg.Provider {
	case "gemini":
		ctx := context.Background()
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
		client := api.NewClient(hostURL, nil)
		provider = llm.NewOllamaProvider(client, cfg.Ollama.Model)
	default:
		fmt.Printf("Unknown provider: %s\n", cfg.Provider)
		os.Exit(1)
	}

	prompt := strings.Join(flag.Args(), " ")

	if prompt == "" {
		fmt.Println("Usage: gen <prompt>")
		os.Exit(1)
	}

	command, confirmed := tui.Run(func(send func(tea.Msg)) {
		shell := getShell()
		command, err := provider.GenerateCommand(prompt, shell)
		if err != nil {
			fmt.Printf("Error generating command: %v\n", err)
			os.Exit(1)
		}

		// The model sometimes returns the command wrapped in backticks, so we remove them.
		command = strings.Trim(command, "`")

		send(tui.CommandGeneratedMsg(command))
	})

	if confirmed {
		shell := getShell()
		cmd := exec.Command(shell, "-c", command)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			fmt.Printf("Error executing command: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println("Command execution aborted.")
	}
}
