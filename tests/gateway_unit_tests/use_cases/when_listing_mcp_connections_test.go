package use_cases_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases"
	runtime_models "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/models"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/services_mocks/interfaces_mocks"
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
			suite.mockClient.EXPECT().List(gomock.Any()).Return(&runtime_models.MCPConnectionListPayload{
				Connections: []runtime_models.MCPConnectionPayload{
					{Id: "a", Name: "alpha", Url: "https://a", HeaderAuthKeys: []string{"Authorization"}},
					{Id: "b", Name: "beta", Url: "https://b", HeaderAuthKeys: []string{}},
				},
			}, nil)
		}

		t.Run("Should pass them through with header keys preserved", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingMcpConnectionsBeforeEach(t)
			initSuite(suite)

			res, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			assert.Len(t, res.Connections, 2)
			assert.Equal(t, "alpha", res.Connections[0].Name)
			assert.Equal(t, []string{"Authorization"}, res.Connections[0].HeaderAuthKeys)
			assert.Empty(t, res.Connections[1].HeaderAuthKeys)
		})
	})
}
