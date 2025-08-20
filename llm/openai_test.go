package llm_test

import (
	"context"
	"errors"
	"io/ioutil"
	"log/slog"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	openai "github.com/sashabaranov/go-openai"

	"github.com/zombor/gen/llm"
)

var _ = Describe("OpenAIProvider", func() {
	var (
		provider                 *llm.OpenAIProvider
		mockCreateChatCompletion func(context.Context, openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error)
		logger                   *slog.Logger
	)

	BeforeEach(func() {
		mockCreateChatCompletion = func(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
			return openai.ChatCompletionResponse{
					Choices: []openai.ChatCompletionChoice{
						{
							Message: openai.ChatCompletionMessage{
								Content: "ls -l",
							},
						},
					},
				},
				nil
		}
		logger = slog.New(slog.NewJSONHandler(ioutil.Discard, nil))
	})

	JustBeforeEach(func() {
		provider = &llm.OpenAIProvider{
			CreateChatCompletion: mockCreateChatCompletion,
			Model:                "gpt-3.5-turbo",
		}
	})

	Describe("GenerateCommand", func() {
		Context("when the OpenAI API call is successful", func() {
			It("returns the generated command", func() {
				command, err := provider.GenerateCommand(context.Background(), logger, "list files", "bash")
				Expect(command).To(Equal("ls -l"))
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when the OpenAI API call returns an error", func() {
			BeforeEach(func() {
				mockCreateChatCompletion = func(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
					return openai.ChatCompletionResponse{}, errors.New("API error")
				}
			})

			It("returns an error", func() {
				command, err := provider.GenerateCommand(context.Background(), logger, "list files", "bash")
				Expect(command).To(BeEmpty())
				Expect(err).To(MatchError("API error"))
			})
		})

		Context("when no command is generated", func() {
			BeforeEach(func() {
				mockCreateChatCompletion = func(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
					return openai.ChatCompletionResponse{},
						nil
				}
			})

			It("returns an error", func() {
				command, err := provider.GenerateCommand(context.Background(), logger, "list files", "bash")
				Expect(command).To(BeEmpty())
				Expect(err).To(MatchError("no command generated"))
			})
		})
	})
})
