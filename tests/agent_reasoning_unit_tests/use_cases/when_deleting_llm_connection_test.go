package use_cases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases"
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
				suite.mockRepository.EXPECT().Delete().Return(nil),
				suite.mockSecretStorage.EXPECT().Delete(gomock.Any()).Return(nil),
			)
		}

		t.Run("Should delete the row first then the secret", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingLLMConnectionBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
		})
	})

	t.Run("Given the repository delete fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingLLMConnectionTestingSuite) {
			suite.mockRepository.EXPECT().Delete().Return(errors.New("db down"))
		}

		t.Run("Should not touch the secret storage", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingLLMConnectionBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background())

			assert.ErrorContains(t, err, "db down")
		})
	})
}
