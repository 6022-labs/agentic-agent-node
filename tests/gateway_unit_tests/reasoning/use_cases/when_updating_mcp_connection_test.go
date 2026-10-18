package use_cases_test

import (
	"context"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/models"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/requests"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/reasoning_mocks/services_mocks/interfaces_mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenUpdatingMcpConnectionTestingSuite struct {
	sut *use_cases.UpdateMcpConnection

	mockClient *interfaces_mocks.MockMCPConfigClientInterface
}

func WhenUpdatingMcpConnectionBeforeEach(t *testing.T) *WhenUpdatingMcpConnectionTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := interfaces_mocks.NewMockMCPConfigClientInterface(mockController)

	sut := use_cases.NewUpdateMcpConnection(mockClient)

	return &WhenUpdatingMcpConnectionTestingSuite{
		sut: sut,

		mockClient: mockClient,
	}
}

func generateValidUpdateMcpConnectionRequest(id uuid.UUID) *requests.UpdateMcpConnectionRequest {
	return &requests.UpdateMcpConnectionRequest{
		Id: id,
		McpConnectionRequest: requests.McpConnectionRequest{
			Name:                "github",
			Url:                 "https://mcp.github.example",
			AllowedToolPatterns: []string{"search_issues"},
		},
		McpHeaderAuthRequest: requests.McpHeaderAuthRequest{HeaderAuth: map[string]string{"Authorization": ""}},
	}
}

func TestWhenUpdatingMcpConnection(t *testing.T) {
	t.Parallel()

	t.Run("Given an invalid request", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error without touching dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingMcpConnectionBeforeEach(t)

			_, err := suite.sut.Execute(context.Background(), &requests.UpdateMcpConnectionRequest{Id: uuid.New()})

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()
		initSuite := func(suite *WhenUpdatingMcpConnectionTestingSuite) {
			suite.mockClient.EXPECT().Update(gomock.Any(), id, gomock.Any()).
				Return(nil, apperrors.NewUnavailableError("runtime_unavailable", "runtime down"))
		}

		t.Run("Should surface the UnavailableError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingMcpConnectionBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), generateValidUpdateMcpConnectionRequest(id))

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
			assert.Equal(t, "runtime_unavailable", unavailableErr.Code)
		})
	})

	t.Run("Given the runtime answers 400", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()
		initSuite := func(suite *WhenUpdatingMcpConnectionTestingSuite) {
			suite.mockClient.EXPECT().Update(gomock.Any(), id, gomock.Any()).
				Return(nil, apperrors.NewValidationError("name", "name is invalid"))
		}

		t.Run("Should surface the ValidationError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingMcpConnectionBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), generateValidUpdateMcpConnectionRequest(id))

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "name", validationErr.Field)
		})
	})

	t.Run("Given the connection does not exist", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()
		initSuite := func(suite *WhenUpdatingMcpConnectionTestingSuite) {
			suite.mockClient.EXPECT().
				Update(gomock.Any(), id, gomock.Any()).
				Return(nil, gateway_interfaces.ErrMCPConnectionNotFound)
		}

		t.Run("Should surface the client error", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingMcpConnectionBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), generateValidUpdateMcpConnectionRequest(id))

			assert.ErrorIs(t, err, gateway_interfaces.ErrMCPConnectionNotFound)
		})
	})

	t.Run("Given the runtime answers 409", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()
		initSuite := func(suite *WhenUpdatingMcpConnectionTestingSuite) {
			suite.mockClient.EXPECT().Update(gomock.Any(), id, gomock.Any()).
				Return(nil, apperrors.NewConflictError("connection_conflict", "connection already exists"))
		}

		t.Run("Should surface the ConflictError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingMcpConnectionBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), generateValidUpdateMcpConnectionRequest(id))

			var conflictErr *apperrors.ConflictError
			assert.ErrorAs(t, err, &conflictErr)
			assert.Equal(t, "connection_conflict", conflictErr.Code)
		})
	})

	t.Run("Given the runtime answers 500", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()
		initSuite := func(suite *WhenUpdatingMcpConnectionTestingSuite) {
			suite.mockClient.EXPECT().Update(gomock.Any(), id, gomock.Any()).
				Return(nil, apperrors.NewUnavailableError("runtime_unavailable", "runtime down"))
		}

		t.Run("Should surface the UnavailableError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingMcpConnectionBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), generateValidUpdateMcpConnectionRequest(id))

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
			assert.Equal(t, "runtime_unavailable", unavailableErr.Code)
		})
	})

	t.Run("Given a blank header value", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()
		initSuite := func(suite *WhenUpdatingMcpConnectionTestingSuite) {
			suite.mockClient.EXPECT().
				Update(gomock.Any(), id, gomock.Any()).
				DoAndReturn(func(_ context.Context, _ uuid.UUID, body *models.ExternalUpdateMCPConnectionPayload) (*models.ExternalMCPConnectionPayload, error) {
					// The blank value is forwarded on purpose: the runtime reads it as "keep the stored one".
					assert.Equal(t, map[string]string{"Authorization": ""}, body.HeaderAuth)
					assert.Equal(t, []string{"search_issues"}, body.AllowedToolPatterns)
					return &models.ExternalMCPConnectionPayload{Id: id.String(), Name: body.Name, Url: body.Url}, nil
				})
		}

		t.Run("Should forward the blank value so the runtime keeps the stored header", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingMcpConnectionBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background(), generateValidUpdateMcpConnectionRequest(id))

			assert.NoError(t, err)
			assert.Equal(t, id.String(), resp.Id)
		})
	})

	t.Run("Given no header auth", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()
		initSuite := func(suite *WhenUpdatingMcpConnectionTestingSuite) {
			suite.mockClient.EXPECT().
				Update(gomock.Any(), id, gomock.Any()).
				DoAndReturn(func(_ context.Context, _ uuid.UUID, body *models.ExternalUpdateMCPConnectionPayload) (*models.ExternalMCPConnectionPayload, error) {
					assert.Nil(t, body.HeaderAuth)
					return &models.ExternalMCPConnectionPayload{Id: id.String()}, nil
				})
		}

		t.Run("Should pass nil for HeaderAuth so the runtime drops every stored header", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingMcpConnectionBeforeEach(t)
			initSuite(suite)

			request := generateValidUpdateMcpConnectionRequest(id)
			request.HeaderAuth = nil
			_, err := suite.sut.Execute(context.Background(), request)

			assert.NoError(t, err)
		})
	})
}
