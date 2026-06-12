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
	repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/repositories_mocks"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenSettingUpLLMConnectionTestingSuite struct {
	sut *use_cases.SetupLLMConnection

	mockRepository    *repositories_mocks.MockLLMConnectionRepositoryInterface
	mockSecretStorage *interfaces_mocks.MockLLMSecretStorageInterface
}

func WhenSettingUpLLMConnectionBeforeEach(t *testing.T) *WhenSettingUpLLMConnectionTestingSuite {
	mockController := gomock.NewController(t)

	mockRepository := repositories_mocks.NewMockLLMConnectionRepositoryInterface(mockController)
	mockSecretStorage := interfaces_mocks.NewMockLLMSecretStorageInterface(mockController)

	sut := use_cases.NewSetupLLMConnection(mockRepository, mockSecretStorage)

	return &WhenSettingUpLLMConnectionTestingSuite{
		sut: sut,

		mockRepository:    mockRepository,
		mockSecretStorage: mockSecretStorage,
	}
}

func generateValidSetupLLMConnectionRequest() *requests.SetupLLMConnectionRequest {
	return &requests.SetupLLMConnectionRequest{
		Platform: reasoning_enums.LLMConnectionPlatformOpenAI,
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
			gomock.InOrder(
				suite.mockSecretStorage.EXPECT().Save(gomock.Any(), "sk-secret").Return(nil),
				suite.mockRepository.EXPECT().Upsert(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, c *models.LLMConnection) (*models.LLMConnection, error) {
					assert.Equal(t, reasoning_enums.LLMConnectionPlatformOpenAI, c.Platform)
					assert.Equal(t, "gpt-4o", c.Model)
					assert.Equal(t, reasoning_enums.LLMConnectionSetupStatusEstablished, c.SetupStatus)
					return c, nil
				}),
			)
		}

		t.Run("Should save the api key in vault then upsert the row with established status", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettingUpLLMConnectionBeforeEach(t)
			request := generateValidSetupLLMConnectionRequest()
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background(), request)

			assert.NoError(t, err)
			assert.Equal(t, reasoning_enums.LLMConnectionPlatformOpenAI, resp.Platform)
		})
	})

	t.Run("Given vault save fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSettingUpLLMConnectionTestingSuite) {
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
