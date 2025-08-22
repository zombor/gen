package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

// BedrockModel is an interface for Bedrock models.
type BedrockModel interface {
	GenerateCommand(ctx context.Context, logger *slog.Logger, prompt, shell string) (string, error)
}

// NewBedrock creates a new BedrockModel.
// If inferenceProfile is provided, it will be used as the ModelId for InvokeModel,
// while the explicit model string is still used to choose the request/response schema.
func NewBedrock(ctx context.Context, model, region, inferenceProfile string) (BedrockModel, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := bedrockruntime.NewFromConfig(cfg)

	targetModelId := model
	if inferenceProfile != "" {
		targetModelId = inferenceProfile
	}

	switch model {
	case "amazon.nova-lite-v1:0":
		return &NovaLiteModel{
			InvokeModel: client.InvokeModel,
			Model:       targetModelId,
		}, nil
	case "amazon.titan-text-lite-v1":
		return &TitanLiteModel{
			InvokeModel: client.InvokeModel,
			Model:       targetModelId,
		}, nil
	case "openai.gpt-oss-120b-1:0":
		return &OpenAIGPTOSSModel{
			InvokeModel: client.InvokeModel,
			Model:       targetModelId,
		}, nil
	case "anthropic.claude-sonnet-4-20250514-v1:0":
		return &AnthropicSonnet4Model{
			InvokeModel: client.InvokeModel,
			Model:       targetModelId,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported Bedrock model: %s", model)
	}
}

// NovaLiteModel represents the amazon.nova-lite-v1:0 model.
type NovaLiteModel struct {
	InvokeModel func(ctx context.Context, params *bedrockruntime.InvokeModelInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.InvokeModelOutput, error)
	Model       string
}

type novaLiteResponse struct {
	Output struct {
		Message struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
		} `json:"message"`
	} `json:"output"`
}

// GenerateCommand implements the BedrockModel interface for NovaLiteModel.
func (c *NovaLiteModel) GenerateCommand(ctx context.Context, logger *slog.Logger, prompt, shell string) (string, error) {
	fullPrompt := fmt.Sprintf(`Given the following prompt, generate a single shell command. The command should be able to be executed on a %s machine in a %s shell. The command should be reasonable and not destructive. Return only the command, with no explanation or other text.

Prompt: %s`, os.Getenv("GOOS"), shell, prompt)
	logger.Debug("bedrock prompt", "prompt", fullPrompt)

	body, err := json.Marshal(map[string]any{
		"schemaVersion": "messages-v1",
		"messages": []any{
			map[string]any{"role": "user", "content": []any{
				map[string]any{"text": fullPrompt},
			}},
		},
		"inferenceConfig": map[string]any{
			"maxTokens": 200,
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal prompt: %w", err)
	}

	output, err := c.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(c.Model),
		ContentType: aws.String("application/json"),
		Body:        body,
		Accept:      aws.String("application/json"),
	})
	if err != nil {
		var ae *types.AccessDeniedException
		if errors.As(err, &ae) {
			return "", fmt.Errorf("access denied to Bedrock API. Please check your AWS credentials and permissions: %w", err)
		}
		return "", fmt.Errorf("failed to invoke Bedrock model: %w", err)
	}

	var response novaLiteResponse
	if err := json.Unmarshal(output.Body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal Bedrock response: %w", err)
	}

	if len(response.Output.Message.Content) == 0 {
		return "", fmt.Errorf("bedrock response did not contain any content")
	}

	text := response.Output.Message.Content[0].Text

	logger.Debug("bedrock response", "response", text)
	return text, nil
}

// TitanLiteModel represents the amazon.titan-text-lite-v1 model.
type TitanLiteModel struct {
	InvokeModel func(ctx context.Context, params *bedrockruntime.InvokeModelInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.InvokeModelOutput, error)
	Model       string
}

type titanLiteResponse struct {
	Results []struct {
		OutputText string `json:"outputText"`
	} `json:"results"`
}

// GenerateCommand implements the BedrockModel interface for TitanLiteModel.
// GenerateCommand implements the BedrockModel interface for TitanLiteModel.
func (c *TitanLiteModel) GenerateCommand(ctx context.Context, logger *slog.Logger, prompt, shell string) (string, error) {
	fullPrompt := fmt.Sprintf(`System: You are a helpful assistant that generates shell commands. The user will provide a prompt and you will generate a single shell command that can be executed on a %s machine in a %s shell. The command should be reasonable and not destructive. Return only the command, with no explanation or other text.

User: list all files in the current directory
Assistant: ls -l

User: %s
Assistant:`, os.Getenv("GOOS"), shell, prompt)
	logger.Debug("bedrock prompt", "prompt", fullPrompt)

	body, err := json.Marshal(map[string]any{
		"inputText": fullPrompt,
		"textGenerationConfig": map[string]any{
			"maxTokenCount": 200,
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal prompt: %w", err)
	}

	output, err := c.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(c.Model),
		ContentType: aws.String("application/json"),
		Body:        body,
		Accept:      aws.String("application/json"),
	})
	if err != nil {
		var ae *types.AccessDeniedException
		if errors.As(err, &ae) {
			return "", fmt.Errorf("access denied to Bedrock API. Please check your AWS credentials and permissions: %w", err)
		}
		return "", fmt.Errorf("failed to invoke Bedrock model: %w", err)
	}

	var response titanLiteResponse
	if err := json.Unmarshal(output.Body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal Bedrock response: %w", err)
	}

	if len(response.Results) == 0 {
		return "", fmt.Errorf("bedrock response did not contain any results")
	}

	text := response.Results[0].OutputText

	logger.Debug("bedrock response", "response", text)
	return text, nil
}

// OpenAIGPTOSSModel represents the openai.gpt-oss-120b-1:0 model.
type OpenAIGPTOSSModel struct {
	InvokeModel func(ctx context.Context, params *bedrockruntime.InvokeModelInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.InvokeModelOutput, error)
	Model       string
}

type openAIChatResponse struct {
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// GenerateCommand implements the BedrockModel interface for OpenAIGPTOSSModel.
func (c *OpenAIGPTOSSModel) GenerateCommand(ctx context.Context, logger *slog.Logger, prompt, shell string) (string, error) {
	fullPrompt := fmt.Sprintf(`Given the following prompt, generate a single shell command. The command should be able to be executed on a %s machine in a %s shell. The command should be reasonable and not destructive. Return only the command, with no explanation or other text.

Prompt: %s`, os.Getenv("GOOS"), shell, prompt)
	logger.Debug("bedrock prompt", "prompt", fullPrompt)

	body, err := json.Marshal(map[string]any{
		"messages": []any{
			map[string]any{
				"role":    "system",
				"content": "You are a shell command generator. Return only the final shell command. Do not include any explanations, chain-of-thought, or tags such as <reasoning>. Do not wrap the command in quotes or backticks.",
			},
			map[string]any{
				"role":    "user",
				"content": "list all files in the current directory",
			},
			map[string]any{
				"role":    "assistant",
				"content": "ls -A",
			},
			map[string]any{
				"role":    "user",
				"content": fullPrompt,
			},
		},
		"temperature":           0,
		"max_completion_tokens": 200,
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal prompt: %w", err)
	}

	output, err := c.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(c.Model),
		ContentType: aws.String("application/json"),
		Body:        body,
		Accept:      aws.String("application/json"),
	})
	if err != nil {
		var ae *types.AccessDeniedException
		if errors.As(err, &ae) {
			return "", fmt.Errorf("access denied to Bedrock API. Please check your AWS credentials and permissions: %w", err)
		}
		return "", fmt.Errorf("failed to invoke Bedrock model: %w", err)
	}

	var response openAIChatResponse
	if err := json.Unmarshal(output.Body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal Bedrock response: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("bedrock response did not contain any choices")
	}

	text := response.Choices[0].Message.Content

	logger.Debug("bedrock response", "response", text)
	return strings.TrimSpace(regexp.MustCompile("(?s)<reasoning>.*?</reasoning>").ReplaceAllString(text, "")), nil
}

// AnthropicSonnet4Model represents the anthropic.claude-sonnet-4-20250514-v1:0 model.
type AnthropicSonnet4Model struct {
	InvokeModel func(ctx context.Context, params *bedrockruntime.InvokeModelInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.InvokeModelOutput, error)
	Model       string
}

type anthropicMessageResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

// GenerateCommand implements the BedrockModel interface for AnthropicSonnet4Model.
func (c *AnthropicSonnet4Model) GenerateCommand(ctx context.Context, logger *slog.Logger, prompt, shell string) (string, error) {
	fullPrompt := fmt.Sprintf(`Given the following prompt, generate a single shell command. The command should be able to be executed on a %s machine in a %s shell. The command should be reasonable and not destructive. Return only the command, with no explanation or other text.

Prompt: %s`, os.Getenv("GOOS"), shell, prompt)
	logger.Debug("bedrock prompt", "prompt", fullPrompt)

	body, err := json.Marshal(map[string]any{
		"anthropic_version": "bedrock-2023-05-31",
		"messages": []any{
			map[string]any{
				"role": "user",
				"content": []any{
					map[string]any{"type": "text", "text": fullPrompt},
				},
			},
		},
		"max_tokens": 200,
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal prompt: %w", err)
	}

	output, err := c.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(c.Model),
		ContentType: aws.String("application/json"),
		Body:        body,
		Accept:      aws.String("application/json"),
	})
	if err != nil {
		var ae *types.AccessDeniedException
		if errors.As(err, &ae) {
			return "", fmt.Errorf("access denied to Bedrock API. Please check your AWS credentials and permissions: %w", err)
		}
		return "", fmt.Errorf("failed to invoke Bedrock model: %w", err)
	}

	var response anthropicMessageResponse
	if err := json.Unmarshal(output.Body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal Bedrock response: %w", err)
	}

	if len(response.Content) == 0 {
		return "", fmt.Errorf("bedrock response did not contain any content")
	}

	text := response.Content[0].Text

	logger.Debug("bedrock response", "response", text)
	return text, nil
}
