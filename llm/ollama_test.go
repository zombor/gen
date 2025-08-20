package llm_test

import (
	"context"
	"errors"

	"github.com/ollama/ollama/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/zombor/gen/llm"
)

var _ = Describe("OllamaProvider", func() {
	var (
		provider     *llm.OllamaProvider
		generateFunc func(context.Context, *api.GenerateRequest, api.GenerateResponseFunc) error
		mockResponse string
		mockError    error
	)

	BeforeEach(func() {
		// Default mock values for successful command generation
		mockResponse = `{"command": "echo hello"}`
		mockError = nil

		generateFunc = func(ctx context.Context, req *api.GenerateRequest, fn api.GenerateResponseFunc) error {
			_ = fn(api.GenerateResponse{Response: mockResponse})
			return mockError
		}
	})

	JustBeforeEach(func() {
		provider = &llm.OllamaProvider{
			Generate: generateFunc,
			Model:    "test-model",
		}
	})

	Context("GenerateCommand", func() {
		When("command generation is successful", func() {
			It("should return the generated command and no error", func() {
				command, err := provider.GenerateCommand(context.Background(), "say hello", "bash")
				Expect([]interface{}{command, err}).To(ConsistOf("echo hello", nil))
			})
		})

		When("generate fails", func() {
			BeforeEach(func() {
				mockResponse = ""
				mockError = errors.New("ollama error")
			})

			It("should return an ollama error and empty command", func() {
				command, err := provider.GenerateCommand(context.Background(), "say hello", "bash")
				Expect([]interface{}{command, err}).To(ConsistOf("", errors.New("ollama error")))
			})
		})

		When("unmarshaling the response fails", func() {
			BeforeEach(func() {
				mockResponse = "invalid json"
			})

			It("should return an unmarshal error and empty command", func() {
				command, err := provider.GenerateCommand(context.Background(), "say hello", "bash")
				Expect(command).To(BeEmpty())
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
