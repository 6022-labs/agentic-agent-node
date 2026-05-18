package use_cases_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/common/usecase_result"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/responses"
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
			suite.mockClient.EXPECT().List(gomock.Any()).Return(&runtime_models.ExternalMCPConnectionListPayload{
				Connections: []runtime_models.ExternalMCPConnectionPayload{
					{Id: "a", Name: "alpha", Url: "https://a", HeaderAuthKeys: []string{"Authorization"}},
					{Id: "b", Name: "beta", Url: "https://b", HeaderAuthKeys: []string{}},
				},
			}, nil)
		}

		t.Run("Should return Ok with the rows on the wire", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingMcpConnectionsBeforeEach(t)
			initSuite(suite)

			res := suite.sut.Execute(context.Background())

			assert.Equal(t, usecase_result.OutcomeOk, res.Outcome)
			body, ok := res.Body.(*responses.ListMcpConnectionsResponse)
			assert.True(t, ok)
			assert.Len(t, body.Connections, 2)
			assert.Equal(t, "alpha", body.Connections[0].Name)
			assert.Equal(t, []string{"Authorization"}, body.Connections[0].HeaderAuthKeys)
			assert.Empty(t, body.Connections[1].HeaderAuthKeys)
		})
	})
}
