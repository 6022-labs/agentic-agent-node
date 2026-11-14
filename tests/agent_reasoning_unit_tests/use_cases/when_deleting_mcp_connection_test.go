package use_cases_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	identity_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/repositories_mocks"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenDeletingMCPConnectionTestingSuite struct {
	sut *use_cases.DeleteMCPConnection

	mockRepository    *repositories_mocks.MockMCPConnectionRepositoryInterface
	mockSecretStorage *interfaces_mocks.MockMCPHeaderAuthSecretStorageInterface
	mockAgentResolver *identity_interfaces_mocks.MockAgentResolverInterface
}

func WhenDeletingMCPConnectionBeforeEach(t *testing.T) *WhenDeletingMCPConnectionTestingSuite {
	mockController := gomock.NewController(t)

	mockRepository := repositories_mocks.NewMockMCPConnectionRepositoryInterface(mockController)
	mockSecretStorage := interfaces_mocks.NewMockMCPHeaderAuthSecretStorageInterface(mockController)
	mockAgentResolver := identity_interfaces_mocks.NewMockAgentResolverInterface(mockController)

	sut := use_cases.NewDeleteMCPConnection(mockRepository, mockSecretStorage, mockAgentResolver)

	return &WhenDeletingMCPConnectionTestingSuite{
		sut: sut,

		mockRepository:    mockRepository,
		mockSecretStorage: mockSecretStorage,
		mockAgentResolver: mockAgentResolver,
	}
}

func TestWhenDeletingMCPConnection(t *testing.T) {
	t.Parallel()

	t.Run("Given the connection does not exist", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()
		initSuite := func(suite *WhenDeletingMCPConnectionTestingSuite) {
			suite.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockRepository.EXPECT().GetById(gomock.Any(), gomock.Any(), id).Return(nil, nil)
		}

		t.Run("Should be a no-op returning NoContent", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingMCPConnectionBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background(), &requests.DeleteMCPConnectionRequest{Id: id})

			assert.NoError(t, err)
		})
	})

	t.Run("Given an existing connection with two headers", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()
		initSuite := func(suite *WhenDeletingMCPConnectionTestingSuite) {
			row := model_builders.NewMCPConnectionBuilder().
				WithId(id).
				WithHeaderAuthKeys([]string{"Authorization", "X-Trace"}).
				Build()
			suite.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockRepository.EXPECT().GetById(gomock.Any(), gomock.Any(), id).Return(row, nil)
			suite.mockSecretStorage.EXPECT().DeleteValue(gomock.Any(), id, "Authorization").Return(nil)
			suite.mockSecretStorage.EXPECT().DeleteValue(gomock.Any(), id, "X-Trace").Return(nil)
			suite.mockRepository.EXPECT().DeleteById(gomock.Any(), gomock.Any(), id).Return(nil)
		}

		t.Run("Should clear every header from vault before dropping the row", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingMCPConnectionBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background(), &requests.DeleteMCPConnectionRequest{Id: id})

			assert.NoError(t, err)
		})
	})
}
