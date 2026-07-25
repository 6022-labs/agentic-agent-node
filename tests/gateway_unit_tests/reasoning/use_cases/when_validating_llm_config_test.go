package use_cases_test

import (
	"context"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/requests"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/reasoning_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenValidatingLlmConfigTestingSuite struct {
	sut *use_cases.ValidateLlmConfig

	mockClient *interfaces_mocks.MockLLMConfigClientInterface
}

func WhenValidatingLlmConfigBeforeEach(t *testing.T) *WhenValidatingLlmConfigTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := interfaces_mocks.NewMockLLMConfigClientInterface(mockController)

	sut := use_cases.NewValidateLlmConfig(mockClient)

	return &WhenValidatingLlmConfigTestingSuite{
		sut: sut,

		mockClient: mockClient,
	}
}

func TestWhenValidatingLlmConfig(t *testing.T) {
	t.Parallel()

	t.Run("Given a request that fails validation", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError without calling the runtime", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingLlmConfigBeforeEach(t)

			err := suite.sut.Execute(context.Background(), &requests.SetLlmConfigRequest{})

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})

	t.Run("Given a valid request", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenValidatingLlmConfigTestingSuite) {
			suite.mockClient.EXPECT().Validate(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, body *models.ExternalSetupLLMConnectionPayload) error {
				assert.Equal(t, "openai", body.Platform)
				assert.Equal(t, "gpt-4o", body.Model)
				assert.Equal(t, "sk-secret", body.ApiKey)
				assert.Nil(t, body.BaseUrl)
				return nil
			})
		}

		t.Run("Should map Provider to Platform and forward a blank BaseUrl as nil", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingLlmConfigBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background(), &requests.SetLlmConfigRequest{
				Provider: "openai",
				Model:    "gpt-4o",
				ApiKey:   "sk-secret",
				BaseUrl:  "  ",
			})

			assert.NoError(t, err)
		})
	})

	t.Run("Given the runtime rejects the connection", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenValidatingLlmConfigTestingSuite) {
			suite.mockClient.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(apperrors.NewValidationError("connection", "unauthorized"))
		}

		t.Run("Should surface the runtime error as-is", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingLlmConfigBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background(), &requests.SetLlmConfigRequest{Provider: "openai", Model: "gpt-4o"})

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})
}
