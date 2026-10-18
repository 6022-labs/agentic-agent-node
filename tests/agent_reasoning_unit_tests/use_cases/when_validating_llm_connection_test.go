package use_cases_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	reasoning_enums "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/services_mocks/interfaces_mocks"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenValidatingLLMConnectionTestingSuite struct {
	sut *use_cases.ValidateLLMConnection

	mockLlmConnectionValidator *interfaces_mocks.MockLLMConnectionValidatorInterface
}

func WhenValidatingLLMConnectionBeforeEach(t *testing.T) *WhenValidatingLLMConnectionTestingSuite {
	mockController := gomock.NewController(t)

	mockLlmConnectionValidator := interfaces_mocks.NewMockLLMConnectionValidatorInterface(mockController)

	sut := use_cases.NewValidateLLMConnection(mockLlmConnectionValidator)

	return &WhenValidatingLLMConnectionTestingSuite{
		sut: sut,

		mockLlmConnectionValidator: mockLlmConnectionValidator,
	}
}

func TestWhenValidatingLLMConnection(t *testing.T) {
	t.Parallel()

	t.Run("Given an invalid request", func(t *testing.T) {
		t.Parallel()

		request := &requests.SetupLLMConnectionRequest{}

		t.Run("Should return a ValidationError without touching dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingLLMConnectionBeforeEach(t)

			err := suite.sut.Execute(context.Background(), request)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})

	t.Run("Given a fully populated request", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenValidatingLLMConnectionTestingSuite) {
			suite.mockLlmConnectionValidator.EXPECT().Validate(gomock.Any(), gomock.Any(), "sk-secret").
				DoAndReturn(func(_ context.Context, c *models.LLMConnection, _ string) error {
					assert.Equal(t, reasoning_enums.LLMConnectionPlatformOllama, c.Platform)
					assert.Equal(t, "llama3", c.Model)
					assert.Equal(t, "http://ollama:11434", lo.FromPtr(c.BaseUrl))
					return nil
				})
		}

		t.Run("Should probe the transient connection built from the request", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingLLMConnectionBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background(), &requests.SetupLLMConnectionRequest{
				Platform: reasoning_enums.LLMConnectionPlatformOllama,
				Model:    "llama3",
				BaseUrl:  lo.ToPtr("http://ollama:11434"),
				ApiKey:   "sk-secret",
			})

			assert.NoError(t, err)
		})
	})

	t.Run("Given the probe rejects the connection", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenValidatingLLMConnectionTestingSuite) {
			suite.mockLlmConnectionValidator.EXPECT().Validate(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(apperrors.NewValidationError("connection", "model not found"))
		}

		t.Run("Should return a ValidationError", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingLLMConnectionBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background(), &requests.SetupLLMConnectionRequest{
				Platform: reasoning_enums.LLMConnectionPlatformOpenAi,
				Model:    "gpt-4o",
			})

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})
}
