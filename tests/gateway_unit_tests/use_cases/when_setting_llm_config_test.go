package use_cases_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/requests"
	runtime_models "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/models"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenSettingLlmConfigTestingSuite struct {
	sut *use_cases.SetLlmConfig

	mockClient *interfaces_mocks.MockLLMConfigClientInterface
}

func WhenSettingLlmConfigBeforeEach(t *testing.T) *WhenSettingLlmConfigTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := interfaces_mocks.NewMockLLMConfigClientInterface(mockController)

	sut := use_cases.NewSetLlmConfig(mockClient)

	return &WhenSettingLlmConfigTestingSuite{
		sut: sut,

		mockClient: mockClient,
	}
}

func generateValidSetLlmConfigRequest() *requests.SetLlmConfigRequest {
	return &requests.SetLlmConfigRequest{
		Provider: "open-ai",
		Model:    "gpt-4o",
		ApiKey:   "sk-secret",
		BaseUrl:  "https://api.openai.com/v1",
	}
}

func TestWhenSettingLlmConfig(t *testing.T) {
	t.Parallel()

	t.Run("Given a request with a non-empty BaseUrl", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSettingLlmConfigTestingSuite) {
			suite.mockClient.EXPECT().Setup(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, body *runtime_models.ExternalSetupLLMConnectionPayload) (*runtime_models.ExternalLLMConnectionPayload, error) {
				assert.Equal(t, "open-ai", body.Platform)
				assert.Equal(t, "gpt-4o", body.Model)
				assert.Equal(t, "sk-secret", body.ApiKey)
				assert.NotNil(t, body.BaseUrl)
				assert.Equal(t, "https://api.openai.com/v1", *body.BaseUrl)
				return &runtime_models.ExternalLLMConnectionPayload{Id: "new", Platform: body.Platform, Model: body.Model, BaseUrl: body.BaseUrl, SetupStatus: "established"}, nil
			})
		}

		t.Run("Should return Ok mapping Provider to Platform and forwarding BaseUrl as a non-nil pointer", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettingLlmConfigBeforeEach(t)
			request := generateValidSetLlmConfigRequest()
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background(), request)

			assert.NoError(t, err)
			assert.Equal(t, "open-ai", resp.Provider)
		})
	})

	t.Run("Given a request with a blank BaseUrl", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSettingLlmConfigTestingSuite) {
			suite.mockClient.EXPECT().Setup(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, body *runtime_models.ExternalSetupLLMConnectionPayload) (*runtime_models.ExternalLLMConnectionPayload, error) {
				assert.Nil(t, body.BaseUrl)
				return &runtime_models.ExternalLLMConnectionPayload{Id: "new", Platform: body.Platform, Model: body.Model}, nil
			})
		}

		t.Run("Should pass nil for BaseUrl on the runtime payload", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettingLlmConfigBeforeEach(t)
			request := generateValidSetLlmConfigRequest()
			request.BaseUrl = "   "
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request)

			assert.NoError(t, err)
		})
	})
}
