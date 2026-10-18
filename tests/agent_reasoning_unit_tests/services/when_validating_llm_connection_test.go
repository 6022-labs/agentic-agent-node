package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenValidatingLLMConnectionTestingSuite struct {
	sut *services.LLMConnectionValidator

	mockChatCompletionRequester *interfaces_mocks.MockChatCompletionRequesterInterface
}

func WhenValidatingLLMConnectionBeforeEach(t *testing.T) *WhenValidatingLLMConnectionTestingSuite {
	mockController := gomock.NewController(t)

	mockChatCompletionRequester := interfaces_mocks.NewMockChatCompletionRequesterInterface(mockController)

	sut := services.NewLLMConnectionValidator(mockChatCompletionRequester)

	return &WhenValidatingLLMConnectionTestingSuite{
		sut: sut,

		mockChatCompletionRequester: mockChatCompletionRequester,
	}
}

func TestWhenValidatingLLMConnection(t *testing.T) {
	t.Parallel()

	llmConnection := model_builders.NewLLMConnectionBuilder().Build()

	t.Run("Given the probe fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenValidatingLLMConnectionTestingSuite) {
			suite.mockChatCompletionRequester.EXPECT().ProbeConnection(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(errors.New("401 invalid api key"))
		}

		t.Run("Should return a ValidationError carrying the provider message", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingLLMConnectionBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Validate(context.Background(), llmConnection, "")

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "connection", validationErr.Field)
			assert.Contains(t, validationErr.Message, "401 invalid api key")
		})
	})

	t.Run("Given the probe succeeds", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenValidatingLLMConnectionTestingSuite) {
			suite.mockChatCompletionRequester.EXPECT().ProbeConnection(gomock.Any(), gomock.Any(), "sk-secret").
				DoAndReturn(func(_ context.Context, c *models.LLMConnection, _ string) error {
					assert.Equal(t, llmConnection.Model, c.Model)
					return nil
				})
		}

		t.Run("Should return no error", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingLLMConnectionBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Validate(context.Background(), llmConnection, "sk-secret")

			assert.NoError(t, err)
		})
	})
}
