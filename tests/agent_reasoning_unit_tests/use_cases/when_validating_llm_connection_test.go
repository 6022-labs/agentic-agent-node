package use_cases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	reasoning_enums "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	identity_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/repositories_mocks"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenValidatingLLMConnectionTestingSuite struct {
	sut *use_cases.ValidateLLMConnection

	mockRepository             *repositories_mocks.MockLLMConnectionRepositoryInterface
	mockLlmConnectionValidator *interfaces_mocks.MockLLMConnectionValidatorInterface
	mockAgentRepository        *identity_repositories_mocks.MockAgentRepositoryInterface
}

func WhenValidatingLLMConnectionBeforeEach(t *testing.T) *WhenValidatingLLMConnectionTestingSuite {
	mockController := gomock.NewController(t)

	mockRepository := repositories_mocks.NewMockLLMConnectionRepositoryInterface(mockController)
	mockLlmConnectionValidator := interfaces_mocks.NewMockLLMConnectionValidatorInterface(mockController)
	mockAgentRepository := identity_repositories_mocks.NewMockAgentRepositoryInterface(mockController)

	sut := use_cases.NewValidateLLMConnection(
		mockRepository,
		mockLlmConnectionValidator,
		mockAgentRepository,
	)

	return &WhenValidatingLLMConnectionTestingSuite{
		sut: sut,

		mockRepository:             mockRepository,
		mockLlmConnectionValidator: mockLlmConnectionValidator,
		mockAgentRepository:        mockAgentRepository,
	}
}

func (s *WhenValidatingLLMConnectionTestingSuite) givenCurrentConnection(current *models.LLMConnection) {
	s.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
	s.mockRepository.EXPECT().GetByAgentId(gomock.Any(), gomock.Any()).Return(current, nil)
}

func TestWhenValidatingLLMConnection(t *testing.T) {
	t.Parallel()

	openAi := reasoning_enums.LLMConnectionPlatformOpenAi
	ollama := reasoning_enums.LLMConnectionPlatformOllama

	openAiConnection := model_builders.NewLLMConnectionBuilder().WithPlatform(openAi).Build()
	ollamaConnection := model_builders.NewLLMConnectionBuilder().WithPlatform(ollama).Build()

	t.Run("Given an invalid request", func(t *testing.T) {
		t.Parallel()

		request := &requests.SetupLLMConnectionRequest{Platform: openAi, ApiKey: "sk-secret"}

		t.Run("Should return a ValidationError without touching dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingLLMConnectionBeforeEach(t)

			err := suite.sut.Execute(context.Background(), request)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "model", validationErr.Field)
		})
	})

	t.Run("Given the agent lookup fails", func(t *testing.T) {
		t.Parallel()

		request := &requests.SetupLLMConnectionRequest{Platform: openAi, Model: "gpt-4o", ApiKey: "sk-secret"}

		initSuite := func(suite *WhenValidatingLLMConnectionTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(nil, errors.New("db down"))
		}

		t.Run("Should return an InternalError without touching other dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingLLMConnectionBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background(), request)

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "agent_lookup_failed", internalErr.Code)
		})
	})

	t.Run("Given the LLM connection lookup fails", func(t *testing.T) {
		t.Parallel()

		request := &requests.SetupLLMConnectionRequest{Platform: openAi, Model: "gpt-4o", ApiKey: "sk-secret"}

		initSuite := func(suite *WhenValidatingLLMConnectionTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockRepository.EXPECT().GetByAgentId(gomock.Any(), gomock.Any()).Return(nil, errors.New("db down"))
		}

		t.Run("Should return an InternalError without touching other dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingLLMConnectionBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background(), request)

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "validate_failed", internalErr.Code)
		})
	})

	t.Run("Given an api key for a local provider", func(t *testing.T) {
		t.Parallel()

		request := &requests.SetupLLMConnectionRequest{Platform: ollama, Model: "llama3", ApiKey: "sk-secret"}

		initSuite := func(suite *WhenValidatingLLMConnectionTestingSuite) {
			suite.givenCurrentConnection(openAiConnection)
		}

		t.Run("Should return a ValidationError on apiKey without touching the validator", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingLLMConnectionBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background(), request)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "apiKey", validationErr.Field)
		})
	})

	t.Run("Given a blank api key for a new cloud provider", func(t *testing.T) {
		t.Parallel()

		request := &requests.SetupLLMConnectionRequest{Platform: openAi, Model: "gpt-4o"}

		initSuite := func(suite *WhenValidatingLLMConnectionTestingSuite) {
			suite.givenCurrentConnection(ollamaConnection)
		}

		t.Run("Should return a ValidationError on apiKey without touching the validator", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingLLMConnectionBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background(), request)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "apiKey", validationErr.Field)
		})
	})

	t.Run("Given a blank api key for a cloud provider and no connection exists", func(t *testing.T) {
		t.Parallel()

		request := &requests.SetupLLMConnectionRequest{Platform: openAi, Model: "gpt-4o"}

		initSuite := func(suite *WhenValidatingLLMConnectionTestingSuite) {
			suite.givenCurrentConnection(nil)
		}

		t.Run("Should return a ValidationError on apiKey without touching the validator", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingLLMConnectionBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background(), request)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "apiKey", validationErr.Field)
		})
	})

	t.Run("Given the probe rejects the connection", func(t *testing.T) {
		t.Parallel()

		request := &requests.SetupLLMConnectionRequest{Platform: openAi, Model: "gpt-4o", ApiKey: "sk-secret"}

		initSuite := func(suite *WhenValidatingLLMConnectionTestingSuite) {
			suite.givenCurrentConnection(openAiConnection)
			suite.mockLlmConnectionValidator.EXPECT().Validate(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(apperrors.NewValidationError("connection", "unauthorized"))
		}

		t.Run("Should return a ValidationError", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingLLMConnectionBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background(), request)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})

	t.Run("Given a blank api key for the configured cloud provider", func(t *testing.T) {
		t.Parallel()

		request := &requests.SetupLLMConnectionRequest{Platform: openAi, Model: "gpt-4o"}

		initSuite := func(suite *WhenValidatingLLMConnectionTestingSuite) {
			suite.givenCurrentConnection(openAiConnection)
			suite.mockLlmConnectionValidator.EXPECT().Validate(gomock.Any(), gomock.Any(), "").Return(nil)
		}

		t.Run("Should probe without an api key", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingLLMConnectionBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background(), request)

			assert.NoError(t, err)
		})
	})

	t.Run("Given a blank api key for a local provider", func(t *testing.T) {
		t.Parallel()

		request := &requests.SetupLLMConnectionRequest{Platform: ollama, Model: "llama3"}

		initSuite := func(suite *WhenValidatingLLMConnectionTestingSuite) {
			suite.givenCurrentConnection(openAiConnection)
			suite.mockLlmConnectionValidator.EXPECT().Validate(gomock.Any(), gomock.Any(), "").Return(nil)
		}

		t.Run("Should probe without an api key", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingLLMConnectionBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background(), request)

			assert.NoError(t, err)
		})
	})

	t.Run("Given a fully populated request", func(t *testing.T) {
		t.Parallel()

		request := &requests.SetupLLMConnectionRequest{
			Platform: openAi,
			Model:    "gpt-4o",
			BaseUrl:  lo.ToPtr("https://proxy.example.com/v1"),
			ApiKey:   "sk-secret",
		}

		initSuite := func(suite *WhenValidatingLLMConnectionTestingSuite) {
			suite.givenCurrentConnection(nil)
			suite.mockLlmConnectionValidator.EXPECT().Validate(gomock.Any(), gomock.Any(), "sk-secret").
				DoAndReturn(func(_ context.Context, c *models.LLMConnection, _ string) error {
					assert.Equal(t, reasoning_enums.LLMConnectionPlatformOpenAi, c.Platform)
					assert.Equal(t, "gpt-4o", c.Model)
					assert.Equal(t, "https://proxy.example.com/v1", lo.FromPtr(c.BaseUrl))
					return nil
				})
		}

		t.Run("Should probe the transient connection built from the request", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingLLMConnectionBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background(), request)

			assert.NoError(t, err)
		})
	})
}
