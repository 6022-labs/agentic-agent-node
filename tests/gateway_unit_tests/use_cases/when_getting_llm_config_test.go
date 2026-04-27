package use_cases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases"
	runtime_models "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/models"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenGettingLlmConfigTestingSuite struct {
	sut *use_cases.GetLlmConfig

	mockClient *interfaces_mocks.MockLLMConfigClientInterface
}

func WhenGettingLlmConfigBeforeEach(t *testing.T) *WhenGettingLlmConfigTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := interfaces_mocks.NewMockLLMConfigClientInterface(mockController)

	sut := use_cases.NewGetLlmConfig(mockClient)

	return &WhenGettingLlmConfigTestingSuite{
		sut: sut,

		mockClient: mockClient,
	}
}

func TestWhenGettingLlmConfig(t *testing.T) {
	t.Parallel()

	t.Run("Given the client returns a payload with a non-nil BaseUrl", func(t *testing.T) {
		t.Parallel()

		baseUrl := "https://api.openai.com/v1"
		initSuite := func(suite *WhenGettingLlmConfigTestingSuite) {
			suite.mockClient.EXPECT().Get(gomock.Any()).Return(&runtime_models.LLMConnectionPayload{
				Id:          "abc",
				Platform:    "open-ai",
				Model:       "gpt-4o",
				BaseUrl:     &baseUrl,
				SetupStatus: "established",
				CreatedAt:   time.Date(2026, 4, 27, 10, 0, 0, 0, time.UTC),
				UpdatedAt:   time.Date(2026, 4, 27, 11, 0, 0, 0, time.UTC),
			}, nil)
		}

		t.Run("Should map Platform to Provider and dereference BaseUrl", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingLlmConfigBeforeEach(t)
			initSuite(suite)

			res, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, "abc", res.Id)
			assert.Equal(t, "open-ai", res.Provider)
			assert.Equal(t, "gpt-4o", res.Model)
			assert.Equal(t, baseUrl, res.BaseUrl)
			assert.Equal(t, "established", res.SetupStatus)
		})
	})

	t.Run("Given the client returns ErrLLMConnectionNotFound", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingLlmConfigTestingSuite) {
			suite.mockClient.EXPECT().Get(gomock.Any()).Return(nil, gateway_interfaces.ErrLLMConnectionNotFound)
		}

		t.Run("Should map it to ErrLLMConnectionNotConfigured", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingLlmConfigBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background())

			assert.True(t, errors.Is(err, use_cases.ErrLLMConnectionNotConfigured))
		})
	})

	t.Run("Given the client returns a transport error", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingLlmConfigTestingSuite) {
			suite.mockClient.EXPECT().Get(gomock.Any()).Return(nil, errors.New("boom"))
		}

		t.Run("Should propagate the error untouched", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingLlmConfigBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background())

			assert.ErrorContains(t, err, "boom")
		})
	})

	t.Run("Given the payload has a nil BaseUrl", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingLlmConfigTestingSuite) {
			suite.mockClient.EXPECT().Get(gomock.Any()).Return(&runtime_models.LLMConnectionPayload{
				Id: "abc", Platform: "ollama", Model: "llama3", SetupStatus: "established",
			}, nil)
		}

		t.Run("Should leave BaseUrl as the empty string", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingLlmConfigBeforeEach(t)
			initSuite(suite)

			res, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, "", res.BaseUrl)
		})
	})
}
