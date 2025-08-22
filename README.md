# gen - AI-Powered Shell Command Generator

gen is a command-line application that leverages Large Language Models (LLMs) to generate shell commands from natural language prompts. It supports multiple LLM providers: Google Gemini, OpenAI, Anthropic, Ollama, and Amazon Bedrock.

## Features

- Generate shell commands from natural language.
- Multiple providers: Gemini, OpenAI, Anthropic, Ollama, Bedrock.
- Optional TUI to review/edit and confirm before executing.
- Debug logging option.
- Configuration via file, environment variables, or command-line flags.

## Installation

To build `gen`, make sure you have Go installed (Go 1.21 or later is recommended).

```bash
git clone https://github.com/zombor/gen.git
cd gen
go build -o gen ./cmd/gen
```

This will create an executable named `gen` in the project root directory.

## Configuration

gen can be configured using a configuration file, environment variables, or command-line flags. The order of precedence is: command-line flags > environment variables > configuration file.

### Configuration File

The default configuration file is located at `~/.gen/config`. This file uses a plain key-value format.

Example `~/.gen/config`:

```
# Provider selection
provider gemini

# Gemini
gemini-api-key YOUR_GEMINI_API_KEY
gemini-model gemini-1.5-flash # default: gemini-1.5-flash

# OpenAI
# provider openai
# openai-api-key YOUR_OPENAI_API_KEY
# openai-model gpt-4o

# Anthropic
# provider anthropic
# anthropic-api-key YOUR_ANTHROPIC_API_KEY
# anthropic-model claude-3-opus-20240229

# Ollama
# provider ollama
# ollama-host http://localhost:11434
# ollama-model llama2

# Bedrock
# provider bedrock
# bedrock-model amazon.nova-lite-v1:0
# bedrock-region us-east-1
# bedrock-inference-profile arn:aws:bedrock:us-east-1:123456789012:inference-profile/your-profile (optional)

# App
# debug true
# tui true
```

### Environment Variables

All configuration options can be set using environment variables prefixed with `GEN_`.

Examples:

```bash
# Common
export GEN_PROVIDER="gemini"
export GEN_DEBUG="false"
export GEN_TUI="true"

# Gemini
export GEN_GEMINI_API_KEY="YOUR_GEMINI_API_KEY"
export GEN_GEMINI_MODEL="gemini-1.5-flash"

# OpenAI
export GEN_OPENAI_API_KEY="YOUR_OPENAI_API_KEY"
export GEN_OPENAI_MODEL="gpt-4o"

# Anthropic
export GEN_ANTHROPIC_API_KEY="YOUR_ANTHROPIC_API_KEY"
export GEN_ANTHROPIC_MODEL="claude-3-opus-20240229"

# Ollama
export GEN_OLLAMA_HOST="http://localhost:11434"
export GEN_OLLAMA_MODEL="llama2"

# Bedrock
export GEN_BEDROCK_MODEL="amazon.nova-lite-v1:0"
export GEN_BEDROCK_REGION="us-east-1"
export GEN_BEDROCK_INFERENCE_PROFILE="arn:aws:bedrock:us-east-1:123456789012:inference-profile/your-profile"
```

### Command-Line Flags

All configuration options can also be set using command-line flags.

Examples:

```bash
# Gemini
./gen --provider gemini --gemini-api-key $GEN_GEMINI_API_KEY \
  --gemini-model gemini-1.5-flash "list all files"

# OpenAI
./gen --provider openai --openai-api-key $GEN_OPENAI_API_KEY \
  --openai-model gpt-4o "list all files"

# Anthropic
./gen --provider anthropic --anthropic-api-key $GEN_ANTHROPIC_API_KEY \
  --anthropic-model claude-3-opus-20240229 "list all files"

# Ollama
./gen --provider ollama --ollama-host http://localhost:11434 \
  --ollama-model llama2 "list all files"

# Bedrock (direct model ID)
./gen --provider bedrock --bedrock-model amazon.nova-lite-v1:0 \
  --bedrock-region us-east-1 "list all files"

# Bedrock (with inference profile)
./gen --provider bedrock --bedrock-model anthropic.claude-sonnet-4-20250514-v1:0 \
  --bedrock-inference-profile arn:aws:bedrock:us-east-1:123456789012:inference-profile/your-profile \
  --bedrock-region us-east-1 "list all files"
```

Available flags (partial):

- `--provider`: LLM provider to use (`gemini`, `openai`, `ollama`, `anthropic`, `bedrock`). Default: `gemini`.
- `--debug`: enable debug logging. Default: `false`.
- `--tui`: enable TUI confirmation/edit flow. Default: `true`.
- `--config`: Path to the configuration file. Default: `~/.gen/config`.
- `--version`: Show the version and exit.
- Gemini: `--gemini-api-key`, `--gemini-model`.
- OpenAI: `--openai-api-key`, `--openai-model`.
- Anthropic: `--anthropic-api-key`, `--anthropic-model`.
- Ollama: `--ollama-host`, `--ollama-model`.
- Bedrock: `--bedrock-model`, `--bedrock-region`, `--bedrock-inference-profile` (optional).

## Provider notes

- Gemini, OpenAI, Anthropic: standard chat APIs.
- Ollama: local models via the Ollama daemon.
- Bedrock:
  - Supported models for shaping: `amazon.nova-lite-v1:0`, `amazon.titan-text-lite-v1`, `openai.gpt-oss-120b-1:0`, `anthropic.claude-sonnet-4-20250514-v1:0`.
  - Some Bedrock models (e.g., Anthropic Sonnet 4) must be invoked via an inference profile. Use `--bedrock-inference-profile` to pass the profile ID/ARN. The program uses the explicit model name to select the correct request/response schema and uses the inference profile (if provided) as the actual `ModelId` for invocation.

## Usage

To generate a shell command, run `gen` followed by your prompt:

```bash
./gen "list all files in the current directory"
```

The application will:

1. Show the generated command.
2. Allow you to edit it (in TUI mode).
3. Ask for confirmation before execution.

### Example

```bash
./gen "create a new directory called my_project"
```

## Contributing

Contributions are welcome! Please feel free to open issues or submit pull requests.

## License

This project is licensed under the MIT License.
