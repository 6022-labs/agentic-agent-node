package use_cases_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/usecase_result"
	repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/repositories_mocks"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenUpdatingMCPConnectionTestingSuite struct {
	sut *use_cases.UpdateMCPConnection

	mockRepository    *repositories_mocks.MockMCPConnectionRepositoryInterface
	mockSecretStorage *interfaces_mocks.MockMCPHeaderAuthSecretStorageInterface
}

func WhenUpdatingMCPConnectionBeforeEach(t *testing.T) *WhenUpdatingMCPConnectionTestingSuite {
	mockController := gomock.NewController(t)

	mockRepository := repositories_mocks.NewMockMCPConnectionRepositoryInterface(mockController)
	mockSecretStorage := interfaces_mocks.NewMockMCPHeaderAuthSecretStorageInterface(mockController)

	sut := use_cases.NewUpdateMCPConnection(mockRepository, mockSecretStorage)

	return &WhenUpdatingMCPConnectionTestingSuite{
		sut: sut,

		mockRepository:    mockRepository,
		mockSecretStorage: mockSecretStorage,
	}
}

func TestWhenUpdatingMCPConnection(t *testing.T) {
	t.Parallel()

	t.Run("Given an unknown id", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()
		initSuite := func(suite *WhenUpdatingMCPConnectionTestingSuite) {
			suite.mockRepository.EXPECT().GetById(gomock.Any(), id).Return(nil, nil)
		}

		t.Run("Should return NotFound without touching update or vault", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingMCPConnectionBeforeEach(t)
			initSuite(suite)

			res := suite.sut.Execute(context.Background(), id, &requests.UpdateMCPConnectionRequest{Name: "x", Url: "https://x"})

			assert.Equal(t, usecase_result.OutcomeNotFound, res.Outcome)
			assert.Equal(t, "mcp_connection_not_found", res.Code)
		})
	})

	t.Run("Given an existing connection with two headers", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()
		initSuite := func(suite *WhenUpdatingMCPConnectionTestingSuite) {
			row := model_builders.NewMCPConnectionBuilder().
				WithId(id).
				WithName("old-name").
				WithUrl("https://old").
				WithHeaderAuthKeys([]string{"Authorization", "X-Drop"}).
				Build()
			suite.mockRepository.EXPECT().GetById(gomock.Any(), id).Return(row, nil)
			suite.mockSecretStorage.EXPECT().DeleteValue(gomock.Any(), id, "X-Drop").Return(nil)
			suite.mockRepository.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, c *models.MCPConnection) (*models.MCPConnection, error) {
				assert.Equal(t, "new-name", c.Name)
				assert.Equal(t, "https://new", c.Url)
				assert.Len(t, c.HeaderAuthKeys, 1)
				assert.Equal(t, "Authorization", string(c.HeaderAuthKeys[0]))
				return c, nil
			})
			suite.mockSecretStorage.EXPECT().SaveValue(gomock.Any(), id, "Authorization", "Bearer new").Return(nil)
		}

		t.Run("Should drop removed header secrets, update the row, and return Ok", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingMCPConnectionBeforeEach(t)
			initSuite(suite)

			res := suite.sut.Execute(context.Background(), id, &requests.UpdateMCPConnectionRequest{
				Name:       "new-name",
				Url:        "https://new",
				HeaderAuth: map[string]string{"Authorization": "Bearer new"},
			})

			assert.Equal(t, usecase_result.OutcomeOk, res.Outcome)
			body, ok := res.Body.(*responses.MCPConnectionResponse)
			assert.True(t, ok)
			assert.Equal(t, id, body.Id)
		})
	})
}
