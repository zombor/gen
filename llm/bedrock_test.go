package llm_test

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log/slog"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/zombor/gen/llm"
)

var _ = Describe("BedrockClient", func() {
	var (
		bedrockClient   *llm.BedrockClient
		mockInvokeModel func(ctx context.Context, params *bedrockruntime.InvokeModelInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.InvokeModelOutput, error)
		modelID         string
		logger          *slog.Logger
	)

	BeforeEach(func() {
		modelID = "test-model"
		logger = slog.New(slog.NewJSONHandler(ioutil.Discard, nil))
	})

	Describe("GenerateCommand", func() {
		var (
			prompt           string
			shell            string
			expectedResponse string
			response         string
			err              error
		)

		BeforeEach(func() {
			prompt = "Hello, Bedrock!"
			shell = "bash"
			expectedResponse = "Bedrock says hi!"
		})

		JustBeforeEach(func() {
			bedrockClient = &llm.BedrockClient{
				InvokeModel: mockInvokeModel,
				Model:       modelID,
			}
			response, err = bedrockClient.GenerateCommand(context.Background(), logger, prompt, shell)
		})

		Context("when the API call is successful", func() {
			BeforeEach(func() {
				mockInvokeModel = func(ctx context.Context, params *bedrockruntime.InvokeModelInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.InvokeModelOutput, error) {
					Expect(params.ModelId).To(Equal(aws.String(modelID)))
					Expect(params.ContentType).To(Equal(aws.String("application/json")))
					Expect(params.Accept).To(Equal(aws.String("application/json")))

					var reqBody map[string]any
					err := json.Unmarshal(params.Body, &reqBody)
					Expect(err).ToNot(HaveOccurred())
					Expect(reqBody["prompt"]).To(ContainSubstring(prompt))
					Expect(reqBody["prompt"]).To(ContainSubstring(shell))
					Expect(reqBody["prompt"]).To(ContainSubstring(os.Getenv("GOOS")))

					respBody, _ := json.Marshal(map[string]string{"completion": expectedResponse})
					return &bedrockruntime.InvokeModelOutput{
						ContentType: aws.String("application/json"),
						Body:        respBody,
					}, nil
				}
			})

			It("returns the expected response", func() {
				Expect(response).To(Equal(expectedResponse))
			})

			It("does not return an error", func() {
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when the API call returns an error", func() {
			BeforeEach(func() {
				mockInvokeModel = func(ctx context.Context, params *bedrockruntime.InvokeModelInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.InvokeModelOutput, error) {
					return nil, errors.New("bedrock API error")
				}
			})

			It("returns an error", func() {
				Expect(err).To(MatchError(ContainSubstring("failed to invoke Bedrock model")))
			})

			It("returns an empty string", func() {
				Expect(response).To(BeEmpty())
			})
		})

		Context("when the API call returns an AccessDeniedException", func() {
			BeforeEach(func() {
				mockInvokeModel = func(ctx context.Context, params *bedrockruntime.InvokeModelInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.InvokeModelOutput, error) {
					return nil, &types.AccessDeniedException{Message: aws.String("Access Denied")}
				}
			})

			It("returns an access denied error", func() {
				Expect(err).To(MatchError(ContainSubstring("access denied to Bedrock API")))
			})

			It("returns an empty string", func() {
				Expect(response).To(BeEmpty())
			})
		})

		Context("when the response body is invalid JSON", func() {
			BeforeEach(func() {
				mockInvokeModel = func(ctx context.Context, params *bedrockruntime.InvokeModelInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.InvokeModelOutput, error) {
					return &bedrockruntime.InvokeModelOutput{
						ContentType: aws.String("application/json"),
						Body:        []byte("invalid json"),
					}, nil
				}
			})

			It("returns an error", func() {
				Expect(err).To(MatchError(ContainSubstring("failed to unmarshal Bedrock response")))
			})

			It("returns an empty string", func() {
				Expect(response).To(BeEmpty())
			})
		})

		Context("when the response body does not contain 'completion' field", func() {
			BeforeEach(func() {
				mockInvokeModel = func(ctx context.Context, params *bedrockruntime.InvokeModelInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.InvokeModelOutput, error) {
					respBody, _ := json.Marshal(map[string]string{"message": "no completion"})
					return &bedrockruntime.InvokeModelOutput{
						ContentType: aws.String("application/json"),
						Body:        respBody,
					}, nil
				}
			})

			It("returns an error", func() {
				Expect(err).To(MatchError(ContainSubstring("bedrock response did not contain 'completion' field")))
			})

			It("returns an empty string", func() {
				Expect(response).To(BeEmpty())
			})
		})
	})
})
