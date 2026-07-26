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
		Id:                  id,
		Name:                "github",
		Url:                 "https://mcp.github.example",
		AllowedToolPatterns: []string{"search_issues"},
		HeaderAuth:          map[string]string{"Authorization": ""},
	}
}

func TestWhenUpdatingMcpConnection(t *testing.T) {
	t.Parallel()

	t.Run("Given a request that fails validation", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError without calling the runtime", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingMcpConnectionBeforeEach(t)

			_, err := suite.sut.Execute(context.Background(), &requests.UpdateMcpConnectionRequest{Id: uuid.New()})

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})

	t.Run("Given a header sent without a value", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()
		initSuite := func(suite *WhenUpdatingMcpConnectionTestingSuite) {
			suite.mockClient.EXPECT().Update(gomock.Any(), id, gomock.Any()).DoAndReturn(func(_ context.Context, _ uuid.UUID, body *models.ExternalUpdateMCPConnectionPayload) (*models.ExternalMCPConnectionPayload, error) {
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

	t.Run("Given no header at all", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()
		initSuite := func(suite *WhenUpdatingMcpConnectionTestingSuite) {
			suite.mockClient.EXPECT().Update(gomock.Any(), id, gomock.Any()).DoAndReturn(func(_ context.Context, _ uuid.UUID, body *models.ExternalUpdateMCPConnectionPayload) (*models.ExternalMCPConnectionPayload, error) {
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

	t.Run("Given the runtime does not know the connection", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()
		initSuite := func(suite *WhenUpdatingMcpConnectionTestingSuite) {
			suite.mockClient.EXPECT().Update(gomock.Any(), id, gomock.Any()).Return(nil, gateway_interfaces.ErrMCPConnectionNotFound)
		}

		t.Run("Should surface the client error", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingMcpConnectionBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), generateValidUpdateMcpConnectionRequest(id))

			assert.ErrorIs(t, err, gateway_interfaces.ErrMCPConnectionNotFound)
		})
	})

	t.Run("Given a header with an empty name", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError rather than silently dropping the value", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingMcpConnectionBeforeEach(t)

			request := generateValidUpdateMcpConnectionRequest(uuid.New())
			request.HeaderAuth = map[string]string{"  ": "Bearer xyz"}

			_, err := suite.sut.Execute(context.Background(), request)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "headerAuth", validationErr.Field)
		})
	})
}
