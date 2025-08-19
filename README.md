# gen - AI-Powered Shell Command Generator

gen is a command-line application that leverages Large Language Models (LLMs) to generate shell commands from natural language prompts. It supports multiple LLM providers, starting with Google Gemini and Ollama.

## Features

- Generate shell commands from natural language.
- Support for Google Gemini and Ollama LLM providers.
- Interactive confirmation and command editing before execution.
- Syntax highlighting for generated commands.
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
provider gemini
api-key YOUR_GEMINI_API_KEY
gemini-model gemini-2.5-flash # Optional, defaults to gemini-2.5-flash

# For Ollama:
# provider ollama
# ollama-host http://localhost:11434
# ollama-model llama2
```

### Environment Variables

All configuration options can be set using environment variables prefixed with `GEN_`.

Example:

```bash
export GEN_API_KEY="YOUR_GEMINI_API_KEY"
export GEN_GEMINI_MODEL="gemini-pro"
export GEN_PROVIDER="ollama"
export GEN_OLLAMA_HOST="http://localhost:11434"
export GEN_OLLAMA_MODEL="llama2"
```

### Command-Line Flags

All configuration options can also be set using command-line flags.

Example:

```bash
gen --provider gemini --api-key YOUR_GEMINI_API_KEY --gemini-model gemini-pro "list all files"
```

Available flags:

- `--provider`: LLM provider to use (e.g., `gemini`, `ollama`). Default: `gemini`.
- `--api-key`: Gemini API key. Required if `provider` is `gemini`.
- `--gemini-model`: Gemini model to use. Default: `gemini-2.5-flash`.
- `--ollama-host`: Ollama host URL. Default: `http://localhost:11434`.
- `--ollama-model`: Ollama model to use. Default: `llama2`.
- `--config`: Path to the configuration file. Default: `~/.gen/config`.
- `--version`: Show the version and exit.

## Usage

To generate a shell command, run `gen` followed by your natural language prompt:

```bash
./gen "list all files in the current directory"
```

The application will:

1.  Display a spinner while generating the command.
2.  Show the generated command with syntax highlighting.
3.  Allow you to edit the command.
4.  Ask for confirmation before executing the command.

### Example

```bash
./gen "create a new directory called my_project"
```

## Contributing

Contributions are welcome! Please feel free to open issues or submit pull requests.

## License

This project is licensed under the MIT License.
