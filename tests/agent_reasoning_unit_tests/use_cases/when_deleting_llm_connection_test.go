package use_cases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/common/usecase_result"
	repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/repositories_mocks"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenDeletingLLMConnectionTestingSuite struct {
	sut *use_cases.DeleteLLMConnection

	mockRepository    *repositories_mocks.MockLLMConnectionRepositoryInterface
	mockSecretStorage *interfaces_mocks.MockLLMSecretStorageInterface
}

func WhenDeletingLLMConnectionBeforeEach(t *testing.T) *WhenDeletingLLMConnectionTestingSuite {
	mockController := gomock.NewController(t)

	mockRepository := repositories_mocks.NewMockLLMConnectionRepositoryInterface(mockController)
	mockSecretStorage := interfaces_mocks.NewMockLLMSecretStorageInterface(mockController)

	sut := use_cases.NewDeleteLLMConnection(mockRepository, mockSecretStorage)

	return &WhenDeletingLLMConnectionTestingSuite{
		sut: sut,

		mockRepository:    mockRepository,
		mockSecretStorage: mockSecretStorage,
	}
}

func TestWhenDeletingLLMConnection(t *testing.T) {
	t.Parallel()

	t.Run("Given a configured connection", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingLLMConnectionTestingSuite) {
			gomock.InOrder(
				suite.mockRepository.EXPECT().Delete(gomock.Any()).Return(nil),
				suite.mockSecretStorage.EXPECT().Delete(gomock.Any()).Return(nil),
			)
		}

		t.Run("Should delete the row first then the secret and return NoContent", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingLLMConnectionBeforeEach(t)
			initSuite(suite)

			res := suite.sut.Execute(context.Background())

			assert.Equal(t, usecase_result.OutcomeNoContent, res.Outcome)
		})
	})

	t.Run("Given the repository delete fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingLLMConnectionTestingSuite) {
			suite.mockRepository.EXPECT().Delete(gomock.Any()).Return(errors.New("db down"))
		}

		t.Run("Should not touch the secret storage and return InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingLLMConnectionBeforeEach(t)
			initSuite(suite)

			res := suite.sut.Execute(context.Background())

			assert.Equal(t, usecase_result.OutcomeInternalError, res.Outcome)
			assert.Contains(t, res.Message, "db down")
		})
	})
}
