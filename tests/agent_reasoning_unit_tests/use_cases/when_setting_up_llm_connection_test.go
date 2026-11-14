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
	identity_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/repositories_mocks"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenSettingUpLLMConnectionTestingSuite struct {
	sut *use_cases.SetupLLMConnection

	mockRepository             *repositories_mocks.MockLLMConnectionRepositoryInterface
	mockSecretStorage          *interfaces_mocks.MockLLMSecretStorageInterface
	mockLlmConnectionValidator *interfaces_mocks.MockLLMConnectionValidatorInterface
	mockAgentResolver          *identity_interfaces_mocks.MockAgentResolverInterface
}

func WhenSettingUpLLMConnectionBeforeEach(t *testing.T) *WhenSettingUpLLMConnectionTestingSuite {
	mockController := gomock.NewController(t)

	mockRepository := repositories_mocks.NewMockLLMConnectionRepositoryInterface(mockController)
	mockSecretStorage := interfaces_mocks.NewMockLLMSecretStorageInterface(mockController)
	mockLlmConnectionValidator := interfaces_mocks.NewMockLLMConnectionValidatorInterface(mockController)
	mockAgentResolver := identity_interfaces_mocks.NewMockAgentResolverInterface(mockController)

	sut := use_cases.NewSetupLLMConnection(
		mockRepository,
		mockSecretStorage,
		mockLlmConnectionValidator,
		mockAgentResolver,
	)

	return &WhenSettingUpLLMConnectionTestingSuite{
		sut: sut,

		mockRepository:             mockRepository,
		mockSecretStorage:          mockSecretStorage,
		mockLlmConnectionValidator: mockLlmConnectionValidator,
		mockAgentResolver:          mockAgentResolver,
	}
}

func (s *WhenSettingUpLLMConnectionTestingSuite) givenCurrentConnection(current *models.LLMConnection) {
	s.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
	s.mockRepository.EXPECT().GetByAgentId(gomock.Any(), gomock.Any()).Return(current, nil)
}

func (s *WhenSettingUpLLMConnectionTestingSuite) expectUpsert() {
	s.mockRepository.EXPECT().
		Upsert(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, c *models.LLMConnection) (*models.LLMConnection, error) {
			return c, nil
		})
}

func TestWhenSettingUpLLMConnection(t *testing.T) {
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

			suite := WhenSettingUpLLMConnectionBeforeEach(t)

			_, err := suite.sut.Execute(context.Background(), request)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "model", validationErr.Field)
		})
	})

	t.Run("Given the agent lookup fails", func(t *testing.T) {
		t.Parallel()

		request := &requests.SetupLLMConnectionRequest{Platform: openAi, Model: "gpt-4o", ApiKey: "sk-secret"}

		initSuite := func(suite *WhenSettingUpLLMConnectionTestingSuite) {
			suite.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(nil, errors.New("db down"))
		}

		t.Run("Should return an InternalError without touching other dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettingUpLLMConnectionBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request)

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "agent_lookup_failed", internalErr.Code)
		})
	})

	t.Run("Given the LLM connection lookup fails", func(t *testing.T) {
		t.Parallel()

		request := &requests.SetupLLMConnectionRequest{Platform: openAi, Model: "gpt-4o", ApiKey: "sk-secret"}

		initSuite := func(suite *WhenSettingUpLLMConnectionTestingSuite) {
			suite.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockRepository.EXPECT().GetByAgentId(gomock.Any(), gomock.Any()).Return(nil, errors.New("db down"))
		}

		t.Run("Should return an InternalError without touching other dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettingUpLLMConnectionBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request)

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "setup_failed", internalErr.Code)
		})
	})

	t.Run("Given an api key for a local provider", func(t *testing.T) {
		t.Parallel()

		request := &requests.SetupLLMConnectionRequest{Platform: ollama, Model: "llama3", ApiKey: "sk-secret"}

		initSuite := func(suite *WhenSettingUpLLMConnectionTestingSuite) {
			suite.givenCurrentConnection(openAiConnection)
		}

		t.Run("Should return a ValidationError on apiKey without touching the validator", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettingUpLLMConnectionBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "apiKey", validationErr.Field)
		})
	})

	t.Run("Given a blank api key for a new cloud provider", func(t *testing.T) {
		t.Parallel()

		request := &requests.SetupLLMConnectionRequest{Platform: openAi, Model: "gpt-4o"}

		initSuite := func(suite *WhenSettingUpLLMConnectionTestingSuite) {
			suite.givenCurrentConnection(ollamaConnection)
		}

		t.Run("Should return a ValidationError on apiKey without touching the validator", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettingUpLLMConnectionBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "apiKey", validationErr.Field)
		})
	})

	t.Run("Given a blank api key for a cloud provider and no connection exists", func(t *testing.T) {
		t.Parallel()

		request := &requests.SetupLLMConnectionRequest{Platform: openAi, Model: "gpt-4o"}

		initSuite := func(suite *WhenSettingUpLLMConnectionTestingSuite) {
			suite.givenCurrentConnection(nil)
		}

		t.Run("Should return a ValidationError on apiKey without touching the validator", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettingUpLLMConnectionBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "apiKey", validationErr.Field)
		})
	})

	t.Run("Given the probe rejects the connection", func(t *testing.T) {
		t.Parallel()

		request := &requests.SetupLLMConnectionRequest{Platform: openAi, Model: "gpt-4o", ApiKey: "sk-secret"}

		initSuite := func(suite *WhenSettingUpLLMConnectionTestingSuite) {
			suite.givenCurrentConnection(openAiConnection)
			suite.mockLlmConnectionValidator.EXPECT().Validate(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(apperrors.NewValidationError("connection", "unauthorized"))
		}

		t.Run("Should return a ValidationError without touching vault or the repository", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettingUpLLMConnectionBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})

	t.Run("Given saving the api key fails", func(t *testing.T) {
		t.Parallel()

		request := &requests.SetupLLMConnectionRequest{Platform: openAi, Model: "gpt-4o", ApiKey: "sk-secret"}

		initSuite := func(suite *WhenSettingUpLLMConnectionTestingSuite) {
			suite.givenCurrentConnection(nil)
			suite.mockLlmConnectionValidator.EXPECT().Validate(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			suite.mockSecretStorage.EXPECT().Save(gomock.Any(), gomock.Any()).Return(errors.New("vault down"))
		}

		t.Run("Should return an InternalError without touching the repository", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettingUpLLMConnectionBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request)

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "setup_failed", internalErr.Code)
			assert.ErrorContains(t, internalErr.Cause, "vault down")
		})
	})

	t.Run("Given reading the stored api key fails", func(t *testing.T) {
		t.Parallel()

		request := &requests.SetupLLMConnectionRequest{Platform: openAi, Model: "gpt-4o"}

		initSuite := func(suite *WhenSettingUpLLMConnectionTestingSuite) {
			suite.givenCurrentConnection(openAiConnection)
			suite.mockLlmConnectionValidator.EXPECT().Validate(gomock.Any(), gomock.Any(), "").Return(nil)
			suite.mockSecretStorage.EXPECT().Get(gomock.Any()).Return("", false, errors.New("vault down"))
		}

		t.Run("Should return an InternalError without touching the repository", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettingUpLLMConnectionBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request)

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "setup_failed", internalErr.Code)
		})
	})

	t.Run("Given a blank api key for the configured cloud provider and none is stored", func(t *testing.T) {
		t.Parallel()

		request := &requests.SetupLLMConnectionRequest{Platform: openAi, Model: "gpt-4o"}

		initSuite := func(suite *WhenSettingUpLLMConnectionTestingSuite) {
			suite.givenCurrentConnection(openAiConnection)
			suite.mockLlmConnectionValidator.EXPECT().Validate(gomock.Any(), gomock.Any(), "").Return(nil)
			suite.mockSecretStorage.EXPECT().Get(gomock.Any()).Return("", false, nil)
		}

		t.Run("Should return a ValidationError on apiKey without touching the repository", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettingUpLLMConnectionBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "apiKey", validationErr.Field)
		})
	})

	t.Run("Given upserting the connection fails", func(t *testing.T) {
		t.Parallel()

		request := &requests.SetupLLMConnectionRequest{
			Platform: ollama,
			Model:    "llama3",
			BaseUrl:  lo.ToPtr("http://host.docker.internal:11434"),
		}

		initSuite := func(suite *WhenSettingUpLLMConnectionTestingSuite) {
			suite.givenCurrentConnection(ollamaConnection)
			suite.mockLlmConnectionValidator.EXPECT().Validate(gomock.Any(), gomock.Any(), "").Return(nil)
			suite.mockRepository.EXPECT().Upsert(gomock.Any(), gomock.Any()).Return(nil, errors.New("db down"))
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettingUpLLMConnectionBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request)

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "setup_failed", internalErr.Code)
			assert.ErrorContains(t, internalErr.Cause, "db down")
		})
	})

	t.Run("Given deleting the stored api key fails", func(t *testing.T) {
		t.Parallel()

		request := &requests.SetupLLMConnectionRequest{
			Platform: ollama,
			Model:    "llama3",
			BaseUrl:  lo.ToPtr("http://host.docker.internal:11434"),
		}

		initSuite := func(suite *WhenSettingUpLLMConnectionTestingSuite) {
			suite.givenCurrentConnection(openAiConnection)
			suite.mockLlmConnectionValidator.EXPECT().Validate(gomock.Any(), gomock.Any(), "").Return(nil)
			suite.expectUpsert()
			suite.mockSecretStorage.EXPECT().Delete(gomock.Any()).Return(errors.New("vault down"))
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettingUpLLMConnectionBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request)

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "setup_failed", internalErr.Code)
			assert.ErrorContains(t, internalErr.Cause, "vault down")
		})
	})

	t.Run("Given a blank api key for the configured cloud provider and one is stored", func(t *testing.T) {
		t.Parallel()

		request := &requests.SetupLLMConnectionRequest{Platform: openAi, Model: "gpt-4o"}

		initSuite := func(suite *WhenSettingUpLLMConnectionTestingSuite) {
			suite.givenCurrentConnection(openAiConnection)
			gomock.InOrder(
				suite.mockLlmConnectionValidator.EXPECT().Validate(gomock.Any(), gomock.Any(), "").Return(nil),
				suite.mockSecretStorage.EXPECT().Get(gomock.Any()).Return("sk-existing", true, nil),
			)
			suite.expectUpsert()
		}

		t.Run("Should return the connection", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettingUpLLMConnectionBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background(), request)

			assert.NoError(t, err)
			assert.Equal(t, reasoning_enums.LLMConnectionPlatformOpenAi, resp.Platform)
		})
	})

	t.Run("Given a blank api key for a local provider", func(t *testing.T) {
		t.Parallel()

		request := &requests.SetupLLMConnectionRequest{
			Platform: ollama,
			Model:    "llama3",
			BaseUrl:  lo.ToPtr("http://host.docker.internal:11434"),
		}

		deleted := false

		initSuite := func(suite *WhenSettingUpLLMConnectionTestingSuite) {
			suite.givenCurrentConnection(ollamaConnection)
			gomock.InOrder(
				suite.mockLlmConnectionValidator.EXPECT().Validate(gomock.Any(), gomock.Any(), "").Return(nil),
				suite.mockRepository.EXPECT().
					Upsert(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, c *models.LLMConnection) (*models.LLMConnection, error) {
						return c, nil
					}),
				suite.mockSecretStorage.EXPECT().Delete(gomock.Any()).DoAndReturn(func(_ context.Context) error {
					deleted = true
					return nil
				}),
			)
		}

		t.Run("Should delete the stored key after saving the connection", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettingUpLLMConnectionBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request)

			assert.NoError(t, err)
			assert.True(t, deleted)
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

		initSuite := func(suite *WhenSettingUpLLMConnectionTestingSuite) {
			suite.givenCurrentConnection(nil)
			gomock.InOrder(
				suite.mockLlmConnectionValidator.EXPECT().
					Validate(gomock.Any(), gomock.Any(), "sk-secret").
					DoAndReturn(func(_ context.Context, c *models.LLMConnection, _ string) error {
						assert.Equal(t, reasoning_enums.LLMConnectionPlatformOpenAi, c.Platform)
						assert.Equal(t, "gpt-4o", c.Model)
						assert.Equal(t, "https://proxy.example.com/v1", lo.FromPtr(c.BaseUrl))
						return nil
					}),
				suite.mockSecretStorage.EXPECT().Save(gomock.Any(), "sk-secret").Return(nil),
				suite.mockRepository.EXPECT().
					Upsert(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, c *models.LLMConnection) (*models.LLMConnection, error) {
						assert.Equal(t, reasoning_enums.LLMConnectionPlatformOpenAi, c.Platform)
						assert.Equal(t, "gpt-4o", c.Model)
						assert.Equal(t, "https://proxy.example.com/v1", lo.FromPtr(c.BaseUrl))
						assert.Equal(t, reasoning_enums.LLMConnectionSetupStatusEstablished, c.SetupStatus)
						return c, nil
					}),
			)
		}

		t.Run("Should return the connection", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettingUpLLMConnectionBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background(), request)

			assert.NoError(t, err)
			assert.Equal(t, reasoning_enums.LLMConnectionPlatformOpenAi, resp.Platform)
		})
	})
}
