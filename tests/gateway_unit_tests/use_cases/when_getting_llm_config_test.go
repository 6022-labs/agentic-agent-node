package use_cases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/common/usecase_result"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/responses"
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
			suite.mockClient.EXPECT().Get(gomock.Any()).Return(&runtime_models.ExternalLLMConnectionPayload{
				Id:          "abc",
				Platform:    "open-ai",
				Model:       "gpt-4o",
				BaseUrl:     &baseUrl,
				SetupStatus: "established",
				CreatedAt:   time.Date(2026, 4, 27, 10, 0, 0, 0, time.UTC),
				UpdatedAt:   time.Date(2026, 4, 27, 11, 0, 0, 0, time.UTC),
			}, nil)
		}

		t.Run("Should return Ok mapping Platform to Provider and dereferencing BaseUrl", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingLlmConfigBeforeEach(t)
			initSuite(suite)

			res := suite.sut.Execute(context.Background())

			assert.Equal(t, usecase_result.OutcomeOk, res.Outcome)
			body, ok := res.Body.(*responses.GetLlmConfigResponse)
			assert.True(t, ok)
			assert.Equal(t, "abc", body.Id)
			assert.Equal(t, "open-ai", body.Provider)
			assert.Equal(t, "gpt-4o", body.Model)
			assert.Equal(t, baseUrl, body.BaseUrl)
			assert.Equal(t, "established", body.SetupStatus)
		})
	})

	t.Run("Given the client returns ErrLLMConnectionNotFound", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingLlmConfigTestingSuite) {
			suite.mockClient.EXPECT().Get(gomock.Any()).Return(nil, gateway_interfaces.ErrLLMConnectionNotFound)
		}

		t.Run("Should return NotFound with not_configured", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingLlmConfigBeforeEach(t)
			initSuite(suite)

			res := suite.sut.Execute(context.Background())

			assert.Equal(t, usecase_result.OutcomeNotFound, res.Outcome)
			assert.Equal(t, "not_configured", res.Code)
		})
	})

	t.Run("Given the client returns a transport error", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingLlmConfigTestingSuite) {
			suite.mockClient.EXPECT().Get(gomock.Any()).Return(nil, errors.New("boom"))
		}

		t.Run("Should return Unavailable with the message", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingLlmConfigBeforeEach(t)
			initSuite(suite)

			res := suite.sut.Execute(context.Background())

			assert.Equal(t, usecase_result.OutcomeUnavailable, res.Outcome)
			assert.Contains(t, res.Message, "boom")
		})
	})

	t.Run("Given the payload has a nil BaseUrl", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingLlmConfigTestingSuite) {
			suite.mockClient.EXPECT().Get(gomock.Any()).Return(&runtime_models.ExternalLLMConnectionPayload{
				Id: "abc", Platform: "ollama", Model: "llama3", SetupStatus: "established",
			}, nil)
		}

		t.Run("Should leave BaseUrl as the empty string", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingLlmConfigBeforeEach(t)
			initSuite(suite)

			res := suite.sut.Execute(context.Background())

			assert.Equal(t, usecase_result.OutcomeOk, res.Outcome)
			body, ok := res.Body.(*responses.GetLlmConfigResponse)
			assert.True(t, ok)
			assert.Equal(t, "", body.BaseUrl)
		})
	})
}
