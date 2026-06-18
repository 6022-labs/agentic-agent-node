package use_cases_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/reasoning_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenListingMcpConnectionsTestingSuite struct {
	sut *use_cases.ListMcpConnections

	mockClient *interfaces_mocks.MockMCPConfigClientInterface
}

func WhenListingMcpConnectionsBeforeEach(t *testing.T) *WhenListingMcpConnectionsTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := interfaces_mocks.NewMockMCPConfigClientInterface(mockController)

	sut := use_cases.NewListMcpConnections(mockClient)

	return &WhenListingMcpConnectionsTestingSuite{
		sut: sut,

		mockClient: mockClient,
	}
}

func TestWhenListingMcpConnections(t *testing.T) {
	t.Parallel()

	t.Run("Given the client returns two connections", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingMcpConnectionsTestingSuite) {
			suite.mockClient.EXPECT().List(gomock.Any()).Return(&models.ExternalMCPConnectionListPayload{
				Connections: []models.ExternalMCPConnectionPayload{
					{Id: "a", Name: "alpha", Url: "https://a", HeaderAuthKeys: []string{"Authorization"}},
					{Id: "b", Name: "beta", Url: "https://b", HeaderAuthKeys: []string{}},
				},
			}, nil)
		}

		t.Run("Should return Ok with the rows on the wire", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingMcpConnectionsBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			assert.Len(t, resp.Connections, 2)
			assert.Equal(t, "alpha", resp.Connections[0].Name)
			assert.Equal(t, []string{"Authorization"}, resp.Connections[0].HeaderAuthKeys)
			assert.Empty(t, resp.Connections[1].HeaderAuthKeys)
		})
	})
}
