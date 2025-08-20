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

// BedrockClient represents a client for AWS Bedrock.
type BedrockClient struct {
	InvokeModel func(ctx context.Context, params *bedrockruntime.InvokeModelInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.InvokeModelOutput, error)
	Model       string
}

// NewBedrock creates a new BedrockClient.
func NewBedrock(ctx context.Context, model string) (*BedrockClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := bedrockruntime.NewFromConfig(cfg)

	return &BedrockClient{
		InvokeModel: client.InvokeModel,
		Model:       model,
	}, nil
}

// GenerateCommand implements the LLMProvider interface for BedrockClient.
func (c *BedrockClient) GenerateCommand(ctx context.Context, logger *slog.Logger, prompt, shell string) (string, error) {
	fullPrompt := fmt.Sprintf(`Given the following prompt, generate a single shell command. The command should be able to be executed on a %s machine in a %s shell. The command should be reasonable and not destructive. Return only the command, with no explanation or other text.

Prompt: %s`, os.Getenv("GOOS"), shell, prompt)
	logger.Debug("bedrock prompt", "prompt", fullPrompt)

	body, err := json.Marshal(map[string]any{
		"prompt": fullPrompt,
		"max_tokens_to_sample": 200,
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

	var response map[string]string
	if err := json.Unmarshal(output.Body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal Bedrock response: %w", err)
	}

	text, ok := response["completion"]
	if !ok {
		return "", fmt.Errorf("bedrock response did not contain 'completion' field")
	}

	logger.Debug("bedrock response", "response", text)
	return text, nil
}
