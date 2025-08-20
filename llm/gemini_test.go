package llm_test

import (
	"context"
	"errors"
	"io/ioutil"
	"log/slog"

	"github.com/google/generative-ai-go/genai"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/zombor/gen/llm"
)

var _ = Describe("GeminiProvider", func() {
	var (
		generateContentFunc func(context.Context, ...genai.Part) (*genai.GenerateContentResponse, error)
		mockResponse        *genai.GenerateContentResponse
		mockError           error
		logger              *slog.Logger

		command string
		err     error
	)

	BeforeEach(func() {
		// Default mock values for successful command generation
		mockResponse = &genai.GenerateContentResponse{
			Candidates: []*genai.Candidate{
				{
					Content: &genai.Content{
						Parts: []genai.Part{
							genai.Text("ls -l"),
						},
					},
				},
			},
		}
		mockError = nil
		logger = slog.New(slog.NewJSONHandler(ioutil.Discard, nil))

		generateContentFunc = func(context.Context, ...genai.Part) (*genai.GenerateContentResponse, error) {
			return mockResponse, mockError
		}
	})

	JustBeforeEach(func() {
		command, err = (&llm.GeminiProvider{
			GenerateContent: generateContentFunc,
		}).GenerateCommand(context.Background(), logger, "list files", "bash")
	})

	Context("GenerateContent", func() {
	})

	Context("GenerateCommand", func() {
		When("command generation is successful", func() {
			It("should return the generated command and no error", func() {
				Expect(command, err).To(Equal("ls -l"))
			})
		})

		When("content generation fails", func() {
			BeforeEach(func() {
				mockResponse = nil
				mockError = errors.New("API error")
			})

			It("should return an API error and empty command", func() {
				Expect(command, err).Error().To(MatchError(mockError))
			})
		})

		When("no command is generated", func() {
			BeforeEach(func() {
				mockResponse = &genai.GenerateContentResponse{}
				mockError = nil
			})

			It("should return a 'no command generated' error and empty command", func() {
				Expect(command, err).Error().To(Equal(errors.New("no command generated")))
			})
		})
	})
})
