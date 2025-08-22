package llm_test

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/zombor/gen/llm"
)

var _ = Describe("Bedrock Models", func() {
	var (
		mockInvokeModel func(ctx context.Context, params *bedrockruntime.InvokeModelInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.InvokeModelOutput, error)
		logger          *slog.Logger
	)

	BeforeEach(func() {
		logger = slog.New(slog.NewJSONHandler(ioutil.Discard, nil))
	})

	Describe("NovaLiteModel", func() {
		var (
			model            *llm.NovaLiteModel
			prompt           string
			shell            string
			expectedResponse string
			response         string
			err              error
		)

		BeforeEach(func() {
			prompt = "Hello, Nova!"
			shell = "bash"
			expectedResponse = "Nova says hi!"
		})

		JustBeforeEach(func() {
			model = &llm.NovaLiteModel{
				InvokeModel: mockInvokeModel,
				Model:       "amazon.nova-lite-v1:0",
			}
			response, err = model.GenerateCommand(context.Background(), logger, prompt, shell)
		})

		Context("when the API call is successful", func() {
			BeforeEach(func() {
				mockInvokeModel = func(ctx context.Context, params *bedrockruntime.InvokeModelInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.InvokeModelOutput, error) {
					var reqBody map[string]any
					err := json.Unmarshal(params.Body, &reqBody)
					Expect(err).ToNot(HaveOccurred())
					Expect(reqBody["schemaVersion"]).To(Equal("messages-v1"))

					respBody, _ := json.Marshal(map[string]any{
						"output": map[string]any{
							"message": map[string]any{
								"content": []any{
									map[string]any{"text": expectedResponse},
								},
							},
						},
					})
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
	})

	Describe("TitanLiteModel", func() {
		var (
			model            *llm.TitanLiteModel
			prompt           string
			shell            string
			expectedResponse string
			response         string
			err              error
		)

		BeforeEach(func() {
			prompt = "Hello, Titan!"
			shell = "zsh"
			expectedResponse = "Titan says hi!"
		})

		JustBeforeEach(func() {
			model = &llm.TitanLiteModel{
				InvokeModel: mockInvokeModel,
				Model:       "amazon.titan-text-lite-v1",
			}
			response, err = model.GenerateCommand(context.Background(), logger, prompt, shell)
		})

		Context("when the API call is successful", func() {
			BeforeEach(func() {
				mockInvokeModel = func(ctx context.Context, params *bedrockruntime.InvokeModelInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.InvokeModelOutput, error) {
					var reqBody map[string]any
					err := json.Unmarshal(params.Body, &reqBody)
					Expect(err).ToNot(HaveOccurred())
					Expect(reqBody["inputText"]).ToNot(BeEmpty())

					respBody, _ := json.Marshal(map[string][]map[string]string{
						"results": {{"outputText": expectedResponse}},
					})
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
	})

	Describe("NewBedrock", func() {
		Context("with a supported model", func() {
			It("returns a NovaLiteModel for amazon.nova-lite-v1:0", func() {
				model, err := llm.NewBedrock(context.Background(), "amazon.nova-lite-v1:0", "us-east-1", "")
				Expect(err).ToNot(HaveOccurred())
				Expect(model).To(BeAssignableToTypeOf(&llm.NovaLiteModel{}))
			})

			It("returns a TitanLiteModel for amazon.titan-text-lite-v1", func() {
				model, err := llm.NewBedrock(context.Background(), "amazon.titan-text-lite-v1", "us-east-1", "")
				Expect(err).ToNot(HaveOccurred())
				Expect(model).To(BeAssignableToTypeOf(&llm.TitanLiteModel{}))
			})
		})

		Context("with an unsupported model", func() {
			It("returns an error", func() {
				_, err := llm.NewBedrock(context.Background(), "unsupported-model", "us-east-1", "")
				Expect(err).To(MatchError("unsupported Bedrock model: unsupported-model"))
			})
		})
	})
})
