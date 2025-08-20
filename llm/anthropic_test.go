package llm_test

import (
	"context"
	"errors"
	"os"

	anthropic "github.com/liushuangls/go-anthropic"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/zombor/gen/llm"
)

var _ = Describe("AnthropicProvider", func() {
	var (
		provider           *llm.AnthropicProvider
		mockCreateMessages func(context.Context, anthropic.MessagesRequest) (anthropic.MessagesResponse, error)
		model              string
	)

	BeforeEach(func() {
		model = "claude-test"
		provider = &llm.AnthropicProvider{
			CreateMessages: func(ctx context.Context, req anthropic.MessagesRequest) (anthropic.MessagesResponse, error) {
				return mockCreateMessages(ctx, req)
			},
			Model: model,
		}
	})

	Describe("GenerateCommand", func() {
		var (
			prompt  string
			shell   string
			command string
			err     error
		)

		BeforeEach(func() {
			prompt = "list files"
			shell = "bash"
		})

		JustBeforeEach(func() {
			command, err = provider.GenerateCommand(context.Background(), prompt, shell)
		})

		Context("when the API call is successful", func() {
			BeforeEach(func() {
				mockCreateMessages = func(ctx context.Context, req anthropic.MessagesRequest) (anthropic.MessagesResponse, error) {
					Expect(*req.Messages[0].Content[0].Text).To(ContainSubstring(prompt))
					Expect(*req.Messages[0].Content[0].Text).To(ContainSubstring(os.Getenv("GOOS")))
					Expect(*req.Messages[0].Content[0].Text).To(ContainSubstring(shell))
					text := "ls -l"
					return anthropic.MessagesResponse{
						Content: []anthropic.MessagesContent{
							{Type: "text", Text: text},
						},
					}, nil
				}
			})

			It("returns the generated command", func() {
				Expect(command).To(Equal("ls -l"))
			})

			It("does not return an error", func() {
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when the API call returns an error", func() {
			BeforeEach(func() {
				mockCreateMessages = func(ctx context.Context, req anthropic.MessagesRequest) (anthropic.MessagesResponse, error) {
					return anthropic.MessagesResponse{}, errors.New("anthropic API error")
				}
			})

			It("returns an error", func() {
				Expect(err).To(MatchError(ContainSubstring("anthropic API error")))
			})

			It("returns an empty command", func() {
				Expect(command).To(BeEmpty())
			})
		})

		Context("when no command is generated", func() {
			BeforeEach(func() {
				mockCreateMessages = func(ctx context.Context, req anthropic.MessagesRequest) (anthropic.MessagesResponse, error) {
					return anthropic.MessagesResponse{},
						nil
				}
			})

			It("returns an error", func() {
				Expect(err).To(MatchError(ContainSubstring("no command generated")))
			})

			It("returns an empty command", func() {
				Expect(command).To(BeEmpty())
			})
		})

		Context("when the content type is not text", func() {
			BeforeEach(func() {
				mockCreateMessages = func(ctx context.Context, req anthropic.MessagesRequest) (anthropic.MessagesResponse, error) {
					return anthropic.MessagesResponse{
						Content: []anthropic.MessagesContent{
							{Type: "image"},
						},
					}, nil
				}
			})

			It("returns an error", func() {
				Expect(err).To(MatchError(ContainSubstring("no command generated")))
			})

			It("returns an empty command", func() {
				Expect(command).To(BeEmpty())
			})
		})
	})
})
