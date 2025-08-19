package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/zombor/gen/cmd/gen/config"
	"github.com/zombor/gen/cmd/gen/llm"
	"github.com/zombor/gen/cmd/gen/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/generative-ai-go/genai"
	"github.com/ollama/ollama/api"
	"google.golang.org/api/option"
)

func getShell() string {
	shellPath := os.Getenv("SHELL")
	if shellPath == "" {
		return "sh"
	}
	return filepath.Base(shellPath)
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	var provider llm.LLMProvider

	switch cfg.Provider {
	case "gemini":
		ctx := context.Background()
		client, err := genai.NewClient(ctx, option.WithAPIKey(cfg.APIKey))
		if err != nil {
			log.Fatal(err)
		}
		defer client.Close()
		model := client.GenerativeModel(cfg.GeminiModel)
		provider = &llm.GeminiProvider{Model: model}
	case "ollama":
		hostURL, err := url.Parse(cfg.Ollama.Host)
		if err != nil {
			log.Fatal(err)
		}
		client := api.NewClient(hostURL, nil)
		provider = &llm.OllamaProvider{Client: client, Model: cfg.Ollama.Model}
	default:
		fmt.Printf("Unknown provider: %s\n", cfg.Provider)
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		fmt.Println("Usage: gen <prompt>")
		os.Exit(1)
	}

	prompt := strings.Join(os.Args[1:], " ")

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
