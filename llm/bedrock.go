package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"

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
func NewBedrock(ctx context.Context, model, region string) (BedrockModel, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := bedrockruntime.NewFromConfig(cfg)

	switch model {
	case "amazon.nova-lite-v1:0":
		return &NovaLiteModel{
			InvokeModel: client.InvokeModel,
			Model:       model,
		}, nil
	case "amazon.titan-text-lite-v1":
		return &TitanLiteModel{
			InvokeModel: client.InvokeModel,
			Model:       model,
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