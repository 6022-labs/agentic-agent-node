package use_cases_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases"
	identity_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/repositories_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenListingMCPConnectionsTestingSuite struct {
	sut *use_cases.ListMCPConnections

	mockRepository      *repositories_mocks.MockMCPConnectionRepositoryInterface
	mockAgentRepository *identity_repositories_mocks.MockAgentRepositoryInterface
}

func WhenListingMCPConnectionsBeforeEach(t *testing.T) *WhenListingMCPConnectionsTestingSuite {
	mockController := gomock.NewController(t)

	mockRepository := repositories_mocks.NewMockMCPConnectionRepositoryInterface(mockController)
	mockAgentRepository := identity_repositories_mocks.NewMockAgentRepositoryInterface(mockController)

	sut := use_cases.NewListMCPConnections(mockRepository, mockAgentRepository)

	return &WhenListingMCPConnectionsTestingSuite{
		sut: sut,

		mockRepository:      mockRepository,
		mockAgentRepository: mockAgentRepository,
	}
}

func TestWhenListingMCPConnections(t *testing.T) {
	t.Parallel()

	t.Run("Given the repository returns rows", func(t *testing.T) {
		t.Parallel()

		stored := []models.MCPConnection{
			*model_builders.NewMCPConnectionBuilder().WithName("alpha").Build(),
			*model_builders.NewMCPConnectionBuilder().WithName("beta").Build(),
		}
		initSuite := func(suite *WhenListingMCPConnectionsTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockRepository.EXPECT().ListByAgentId(gomock.Any(), gomock.Any()).Return(stored, nil)
		}

		t.Run("Should return Ok with the wire response", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingMCPConnectionsBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			assert.Len(t, resp.Connections, 2)
			assert.Equal(t, "alpha", resp.Connections[0].Name)
		})
	})
}
