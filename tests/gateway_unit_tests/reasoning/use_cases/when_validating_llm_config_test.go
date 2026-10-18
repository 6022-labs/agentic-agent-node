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

	t.Run("Given an invalid request", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error without touching dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingLlmConfigBeforeEach(t)

			err := suite.sut.Execute(context.Background(), &requests.SetLlmConfigRequest{})

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenValidatingLlmConfigTestingSuite) {
			suite.mockClient.EXPECT().Validate(gomock.Any(), gomock.Any()).
				Return(apperrors.NewUnavailableError("runtime_unavailable", "runtime down"))
		}

		t.Run("Should surface the UnavailableError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingLlmConfigBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(
				context.Background(),
				&requests.SetLlmConfigRequest{Provider: "openai", Model: "gpt-4o"},
			)

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
			assert.Equal(t, "runtime_unavailable", unavailableErr.Code)
		})
	})

	t.Run("Given the runtime answers 400", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenValidatingLlmConfigTestingSuite) {
			suite.mockClient.EXPECT().
				Validate(gomock.Any(), gomock.Any()).
				Return(apperrors.NewValidationError("connection", "unauthorized"))
		}

		t.Run("Should surface the ValidationError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingLlmConfigBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(
				context.Background(),
				&requests.SetLlmConfigRequest{Provider: "openai", Model: "gpt-4o"},
			)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})

	t.Run("Given the runtime accepts the connection", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenValidatingLlmConfigTestingSuite) {
			suite.mockClient.EXPECT().
				Validate(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, body *models.ExternalSetupLLMConnectionPayload) error {
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
}
