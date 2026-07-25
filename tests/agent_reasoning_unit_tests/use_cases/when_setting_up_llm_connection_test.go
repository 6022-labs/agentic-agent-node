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
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenSettingUpLLMConnectionTestingSuite struct {
	sut *use_cases.SetupLLMConnection

	mockRepository             *repositories_mocks.MockLLMConnectionRepositoryInterface
	mockSecretStorage          *interfaces_mocks.MockLLMSecretStorageInterface
	mockLlmConnectionValidator *interfaces_mocks.MockLLMConnectionValidatorInterface
	mockAgentRepository        *identity_repositories_mocks.MockAgentRepositoryInterface
}

func WhenSettingUpLLMConnectionBeforeEach(t *testing.T) *WhenSettingUpLLMConnectionTestingSuite {
	mockController := gomock.NewController(t)

	mockRepository := repositories_mocks.NewMockLLMConnectionRepositoryInterface(mockController)
	mockSecretStorage := interfaces_mocks.NewMockLLMSecretStorageInterface(mockController)
	mockLlmConnectionValidator := interfaces_mocks.NewMockLLMConnectionValidatorInterface(mockController)
	mockAgentRepository := identity_repositories_mocks.NewMockAgentRepositoryInterface(mockController)

	sut := use_cases.NewSetupLLMConnection(mockRepository, mockSecretStorage, mockLlmConnectionValidator, mockAgentRepository)

	return &WhenSettingUpLLMConnectionTestingSuite{
		sut: sut,

		mockRepository:             mockRepository,
		mockSecretStorage:          mockSecretStorage,
		mockLlmConnectionValidator: mockLlmConnectionValidator,
		mockAgentRepository:        mockAgentRepository,
	}
}

func generateValidSetupLLMConnectionRequest() *requests.SetupLLMConnectionRequest {
	return &requests.SetupLLMConnectionRequest{
		Platform: reasoning_enums.LLMConnectionPlatformOpenAi,
		Model:    "gpt-4o",
		ApiKey:   "sk-secret",
	}
}

func TestWhenSettingUpLLMConnection(t *testing.T) {
	t.Parallel()

	t.Run("Given the request is invalid", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error without touching dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettingUpLLMConnectionBeforeEach(t)

			request := generateValidSetupLLMConnectionRequest()
			request.Model = ""

			_, err := suite.sut.Execute(context.Background(), request)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})

	t.Run("Given a valid request", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSettingUpLLMConnectionTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			gomock.InOrder(
				suite.mockLlmConnectionValidator.EXPECT().Validate(gomock.Any(), gomock.Any(), "sk-secret").DoAndReturn(func(_ context.Context, c *models.LLMConnection, _ string) error {
					assert.Equal(t, reasoning_enums.LLMConnectionPlatformOpenAi, c.Platform)
					assert.Equal(t, "gpt-4o", c.Model)
					return nil
				}),
				suite.mockSecretStorage.EXPECT().Save(gomock.Any(), "sk-secret").Return(nil),
				suite.mockRepository.EXPECT().Upsert(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, c *models.LLMConnection) (*models.LLMConnection, error) {
					assert.Equal(t, reasoning_enums.LLMConnectionPlatformOpenAi, c.Platform)
					assert.Equal(t, "gpt-4o", c.Model)
					assert.Equal(t, reasoning_enums.LLMConnectionSetupStatusEstablished, c.SetupStatus)
					return c, nil
				}),
			)
		}

		t.Run("Should validate the connection, save the api key in vault, then upsert the row with established status", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettingUpLLMConnectionBeforeEach(t)
			request := generateValidSetupLLMConnectionRequest()
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background(), request)

			assert.NoError(t, err)
			assert.Equal(t, reasoning_enums.LLMConnectionPlatformOpenAi, resp.Platform)
		})
	})

	t.Run("Given the connection fails validation", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSettingUpLLMConnectionTestingSuite) {
			suite.mockLlmConnectionValidator.EXPECT().Validate(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(apperrors.NewValidationError("connection", "unauthorized"))
		}

		t.Run("Should return the validation error without touching vault or the repository", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettingUpLLMConnectionBeforeEach(t)
			request := generateValidSetupLLMConnectionRequest()
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})

	t.Run("Given the api key is empty and one is already stored", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSettingUpLLMConnectionTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			gomock.InOrder(
				suite.mockLlmConnectionValidator.EXPECT().Validate(gomock.Any(), gomock.Any(), "").Return(nil),
				suite.mockSecretStorage.EXPECT().Get(gomock.Any()).Return("sk-existing", true, nil),
				suite.mockRepository.EXPECT().Upsert(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, c *models.LLMConnection) (*models.LLMConnection, error) {
					return c, nil
				}),
			)
		}

		t.Run("Should keep the stored key and upsert the row without saving", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettingUpLLMConnectionBeforeEach(t)
			request := generateValidSetupLLMConnectionRequest()
			request.ApiKey = ""
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request)

			assert.NoError(t, err)
		})
	})

	t.Run("Given the api key is empty and none is stored yet", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSettingUpLLMConnectionTestingSuite) {
			suite.mockLlmConnectionValidator.EXPECT().Validate(gomock.Any(), gomock.Any(), "").Return(nil)
			suite.mockSecretStorage.EXPECT().Get(gomock.Any()).Return("", false, nil)
		}

		t.Run("Should return a validation error without touching the repository", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettingUpLLMConnectionBeforeEach(t)
			request := generateValidSetupLLMConnectionRequest()
			request.ApiKey = ""
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})

	t.Run("Given vault save fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSettingUpLLMConnectionTestingSuite) {
			suite.mockLlmConnectionValidator.EXPECT().Validate(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			suite.mockSecretStorage.EXPECT().Save(gomock.Any(), gomock.Any()).Return(errors.New("vault down"))
		}

		t.Run("Should not touch the repository and return InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettingUpLLMConnectionBeforeEach(t)
			request := generateValidSetupLLMConnectionRequest()
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request)

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Contains(t, internalErr.Message, "vault down")
		})
	})
}
