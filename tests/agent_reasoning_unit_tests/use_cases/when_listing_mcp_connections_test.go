package use_cases_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases"
	repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/repositories_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenListingMCPConnectionsTestingSuite struct {
	sut *use_cases.ListMCPConnections

	mockRepository *repositories_mocks.MockMCPConnectionRepositoryInterface
}

func WhenListingMCPConnectionsBeforeEach(t *testing.T) *WhenListingMCPConnectionsTestingSuite {
	mockController := gomock.NewController(t)

	mockRepository := repositories_mocks.NewMockMCPConnectionRepositoryInterface(mockController)

	sut := use_cases.NewListMCPConnections(mockRepository)

	return &WhenListingMCPConnectionsTestingSuite{
		sut: sut,

		mockRepository: mockRepository,
	}
}

func TestWhenListingMCPConnections(t *testing.T) {
	t.Parallel()

	t.Run("Given the repository returns rows", func(t *testing.T) {
		t.Parallel()

		expected := []models.MCPConnection{
			*model_builders.NewMCPConnectionBuilder().WithName("alpha").Build(),
			*model_builders.NewMCPConnectionBuilder().WithName("beta").Build(),
		}
		initSuite := func(suite *WhenListingMCPConnectionsTestingSuite) {
			suite.mockRepository.EXPECT().List().Return(expected, nil)
		}

		t.Run("Should pass them through without hydrating header values", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingMCPConnectionsBeforeEach(t)
			initSuite(suite)

			got, err := suite.sut.Execute()

			assert.NoError(t, err)
			assert.Equal(t, expected, got)
		})
	})
}
