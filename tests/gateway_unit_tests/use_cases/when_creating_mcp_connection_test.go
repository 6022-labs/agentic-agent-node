package use_cases_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/common/usecase_result"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/responses"
	runtime_models "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/models"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenCreatingMcpConnectionTestingSuite struct {
	sut *use_cases.CreateMcpConnection

	mockClient *interfaces_mocks.MockMCPConfigClientInterface
}

func WhenCreatingMcpConnectionBeforeEach(t *testing.T) *WhenCreatingMcpConnectionTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := interfaces_mocks.NewMockMCPConfigClientInterface(mockController)

	sut := use_cases.NewCreateMcpConnection(mockClient)

	return &WhenCreatingMcpConnectionTestingSuite{
		sut: sut,

		mockClient: mockClient,
	}
}

func generateValidGatewayCreateMcpConnectionRequest() *requests.CreateMcpConnectionRequest {
	return &requests.CreateMcpConnectionRequest{
		Name:            "github",
		Url:             "https://mcp.github.example",
		AuthHeaderName:  "Authorization",
		AuthHeaderValue: "Bearer abc",
	}
}

func TestWhenCreatingMcpConnection(t *testing.T) {
	t.Parallel()

	t.Run("Given a request with one auth header", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingMcpConnectionTestingSuite) {
			suite.mockClient.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, body *runtime_models.ExternalCreateMCPConnectionPayload) (*runtime_models.ExternalMCPConnectionPayload, error) {
				assert.Equal(t, "github", body.Name)
				assert.Equal(t, "https://mcp.github.example", body.Url)
				assert.Equal(t, map[string]string{"Authorization": "Bearer abc"}, body.HeaderAuth)
				return &runtime_models.ExternalMCPConnectionPayload{Id: "new", Name: body.Name, Url: body.Url, HeaderAuthKeys: []string{"Authorization"}}, nil
			})
		}

		t.Run("Should return Created packing the header into a single-entry HeaderAuth map", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingMcpConnectionBeforeEach(t)
			request := generateValidGatewayCreateMcpConnectionRequest()
			initSuite(suite)

			res := suite.sut.Execute(context.Background(), request)

			assert.Equal(t, usecase_result.OutcomeCreated, res.Outcome)
			body, ok := res.Body.(*responses.McpConnectionResponse)
			assert.True(t, ok)
			assert.Equal(t, []string{"Authorization"}, body.HeaderAuthKeys)
		})
	})

	t.Run("Given a request with no auth headers", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingMcpConnectionTestingSuite) {
			suite.mockClient.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, body *runtime_models.ExternalCreateMCPConnectionPayload) (*runtime_models.ExternalMCPConnectionPayload, error) {
				assert.Nil(t, body.HeaderAuth)
				return &runtime_models.ExternalMCPConnectionPayload{Id: "new", Name: body.Name, Url: body.Url}, nil
			})
		}

		t.Run("Should pass nil for HeaderAuth", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingMcpConnectionBeforeEach(t)
			request := generateValidGatewayCreateMcpConnectionRequest()
			request.AuthHeaderName = ""
			request.AuthHeaderValue = ""
			initSuite(suite)

			res := suite.sut.Execute(context.Background(), request)

			assert.Equal(t, usecase_result.OutcomeCreated, res.Outcome)
		})
	})
}
