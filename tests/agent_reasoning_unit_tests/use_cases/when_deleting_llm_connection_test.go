package use_cases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	identity_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/repositories_mocks"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenDeletingLLMConnectionTestingSuite struct {
	sut *use_cases.DeleteLLMConnection

	mockRepository      *repositories_mocks.MockLLMConnectionRepositoryInterface
	mockSecretStorage   *interfaces_mocks.MockLLMSecretStorageInterface
	mockAgentRepository *identity_repositories_mocks.MockAgentRepositoryInterface
}

func WhenDeletingLLMConnectionBeforeEach(t *testing.T) *WhenDeletingLLMConnectionTestingSuite {
	mockController := gomock.NewController(t)

	mockRepository := repositories_mocks.NewMockLLMConnectionRepositoryInterface(mockController)
	mockSecretStorage := interfaces_mocks.NewMockLLMSecretStorageInterface(mockController)
	mockAgentRepository := identity_repositories_mocks.NewMockAgentRepositoryInterface(mockController)

	sut := use_cases.NewDeleteLLMConnection(mockRepository, mockSecretStorage, mockAgentRepository)

	return &WhenDeletingLLMConnectionTestingSuite{
		sut: sut,

		mockRepository:      mockRepository,
		mockSecretStorage:   mockSecretStorage,
		mockAgentRepository: mockAgentRepository,
	}
}

func TestWhenDeletingLLMConnection(t *testing.T) {
	t.Parallel()

	t.Run("Given deleting the LLM connection fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingLLMConnectionTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockRepository.EXPECT().DeleteByAgentId(gomock.Any(), gomock.Any()).Return(errors.New("db down"))
		}

		t.Run("Should return an InternalError without touching the secret storage", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingLLMConnectionBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background())

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.ErrorContains(t, internalErr.Cause, "db down")
		})
	})

	t.Run("Given an LLM connection exists", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingLLMConnectionTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			gomock.InOrder(
				suite.mockRepository.EXPECT().DeleteByAgentId(gomock.Any(), gomock.Any()).Return(nil),
				suite.mockSecretStorage.EXPECT().Delete(gomock.Any()).Return(nil),
			)
		}

		t.Run("Should delete the row before the secret", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingLLMConnectionBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
		})
	})
}
