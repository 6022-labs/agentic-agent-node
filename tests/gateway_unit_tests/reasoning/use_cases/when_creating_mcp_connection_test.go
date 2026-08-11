package use_cases_test

import (
	"context"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"

	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/requests"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/reasoning_mocks/services_mocks/interfaces_mocks"
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
		Name:                "github",
		Url:                 "https://mcp.github.example",
		AllowedToolPatterns: []string{"search_issues"},
		HeaderAuth:          map[string]string{"Authorization": "Bearer abc"},
	}
}

func TestWhenCreatingMcpConnection(t *testing.T) {
	t.Parallel()

	t.Run("Given a request that fails validation", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError without calling the runtime", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingMcpConnectionBeforeEach(t)

			_, err := suite.sut.Execute(context.Background(), &requests.CreateMcpConnectionRequest{})

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})

	t.Run("Given a request with two auth headers", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingMcpConnectionTestingSuite) {
			suite.mockClient.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, body *models.ExternalCreateMCPConnectionPayload) (*models.ExternalMCPConnectionPayload, error) {
					assert.Equal(t, "github", body.Name)
					assert.Equal(t, "https://mcp.github.example", body.Url)
					assert.Equal(t, map[string]string{"Authorization": "Bearer abc", "X-Api-Key": "k"}, body.HeaderAuth)
					return &models.ExternalMCPConnectionPayload{
						Id:             "new",
						Name:           body.Name,
						Url:            body.Url,
						HeaderAuthKeys: []string{"Authorization", "X-Api-Key"},
					}, nil
				})
		}

		t.Run("Should forward every header and return Created", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingMcpConnectionBeforeEach(t)
			request := generateValidGatewayCreateMcpConnectionRequest()
			request.HeaderAuth["X-Api-Key"] = "k"
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background(), request)

			assert.NoError(t, err)
			assert.Equal(t, []string{"Authorization", "X-Api-Key"}, resp.HeaderAuthKeys)
		})
	})

	t.Run("Given a request with no auth headers", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingMcpConnectionTestingSuite) {
			suite.mockClient.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, body *models.ExternalCreateMCPConnectionPayload) (*models.ExternalMCPConnectionPayload, error) {
					assert.Nil(t, body.HeaderAuth)
					return &models.ExternalMCPConnectionPayload{Id: "new", Name: body.Name, Url: body.Url}, nil
				})
		}

		t.Run("Should pass nil for HeaderAuth", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingMcpConnectionBeforeEach(t)
			request := generateValidGatewayCreateMcpConnectionRequest()
			request.HeaderAuth = nil
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request)

			assert.NoError(t, err)
		})
	})
}
