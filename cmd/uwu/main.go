package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/zombor/uwu/cmd/uwu/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// LLMProvider defines the interface for a language model provider.
type LLMProvider interface {
	GenerateCommand(prompt, shell string) (string, error)
}

// GeminiProvider is an implementation of LLMProvider for Google's Gemini.
type GeminiProvider struct {
	Model *genai.GenerativeModel
}

// GenerateCommand generates a command using the Gemini LLM.
func (p *GeminiProvider) GenerateCommand(prompt, shell string) (string, error) {
	ctx := context.Background()
	resp, err := p.Model.GenerateContent(ctx, genai.Text(fmt.Sprintf("Given the following prompt, generate a single shell command. The command should be able to be executed on a %s machine in a %s shell. The command should be reasonable and not destructive. Return only the command, with no explanation or other text.\n\nPrompt: %s", os.Getenv("GOOS"), shell, prompt)))
	if err != nil {
		return "", err
	}

	if len(resp.Candidates) > 0 {
		for _, part := range resp.Candidates[0].Content.Parts {
			if txt, ok := part.(genai.Text); ok {
				return string(txt), nil
			}
		}
	}

	return "", fmt.Errorf("no command generated")
}

func getShell() string {
	shellPath := os.Getenv("SHELL")
	if shellPath == "" {
		return "sh"
	}
	return filepath.Base(shellPath)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: uwu <prompt>")
		os.Exit(1)
	}

	prompt := strings.Join(os.Args[1:], " ")

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: GEMINI_API_KEY environment variable not set.")
		os.Exit(1)
	}

	command, confirmed := tui.Run(func(send func(tea.Msg)) {
		ctx := context.Background()
		client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
		if err != nil {
			log.Fatal(err)
		}
		defer client.Close()

		model := client.GenerativeModel("gemini-2.5-flash")

		provider := &GeminiProvider{
			Model: model,
		}

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