package llm_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	anthropic "github.com/liushuangls/go-anthropic"

	"github.com/zombor/gen/llm"
)

var _ = Describe("AnthropicProvider", func() {
	var (
		provider *llm.AnthropicProvider
		mockCreateMessages func(context.Context, anthropic.MessagesRequest) (*anthropic.MessagesResponse, error)
	)

	BeforeEach(func() {
		mockCreateMessages = func(ctx context.Context, req anthropic.MessagesRequest) (*anthropic.MessagesResponse, error) {
			text := "ls -l"
			return &anthropic.MessagesResponse{
				Content: []anthropic.MessagesContent{
					{
						Type: "text",
						Text: text,
					},
				},
			},
			nil
		}
	})

	JustBeforeEach(func() {
		provider = &llm.AnthropicProvider{
			CreateMessages: mockCreateMessages,
			Model:          "claude-3-opus-20240229",
		}
	})

	Describe("GenerateCommand", func() {
		Context("when the Anthropic API call is successful", func() {
			It("returns the generated command", func() {
				command, err := provider.GenerateCommand("list files", "bash")
				Expect(command).To(Equal("ls -l"))
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when the Anthropic API call returns an error", func() {
			BeforeEach(func() {
				mockCreateMessages = func(ctx context.Context, req anthropic.MessagesRequest) (*anthropic.MessagesResponse, error) {
					return nil, errors.New("API error")
				}
			})

			It("returns an error", func() {
				command, err := provider.GenerateCommand("list files", "bash")
				Expect(command).To(BeEmpty())
				Expect(err).To(MatchError("API error"))
			})
		})

		Context("when no command is generated", func() {
			BeforeEach(func() {
				mockCreateMessages = func(ctx context.Context, req anthropic.MessagesRequest) (*anthropic.MessagesResponse, error) {
					return &anthropic.MessagesResponse{},
					nil
				}
			})

			It("returns an error", func() {
				command, err := provider.GenerateCommand("list files", "bash")
				Expect(command).To(BeEmpty())
				Expect(err).To(MatchError("no command generated"))
			})
		})
	})
})
