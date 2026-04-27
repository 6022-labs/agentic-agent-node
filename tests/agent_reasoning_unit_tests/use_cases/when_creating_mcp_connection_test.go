package use_cases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/repositories_mocks"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/services_mocks/interfaces_mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenCreatingMCPConnectionTestingSuite struct {
	sut *use_cases.CreateMCPConnection

	mockRepository    *repositories_mocks.MockMCPConnectionRepositoryInterface
	mockSecretStorage *interfaces_mocks.MockMCPHeaderAuthSecretStorageInterface
}

func WhenCreatingMCPConnectionBeforeEach(t *testing.T) *WhenCreatingMCPConnectionTestingSuite {
	mockController := gomock.NewController(t)

	mockRepository := repositories_mocks.NewMockMCPConnectionRepositoryInterface(mockController)
	mockSecretStorage := interfaces_mocks.NewMockMCPHeaderAuthSecretStorageInterface(mockController)

	sut := use_cases.NewCreateMCPConnection(mockRepository, mockSecretStorage)

	return &WhenCreatingMCPConnectionTestingSuite{
		sut: sut,

		mockRepository:    mockRepository,
		mockSecretStorage: mockSecretStorage,
	}
}

func TestWhenCreatingMCPConnection(t *testing.T) {
	t.Parallel()

	t.Run("Given a request with one header", func(t *testing.T) {
		t.Parallel()

		generatedId := uuid.New()
		initSuite := func(suite *WhenCreatingMCPConnectionTestingSuite) {
			suite.mockRepository.EXPECT().Create(gomock.Any()).DoAndReturn(func(c *models.MCPConnection) (*models.MCPConnection, error) {
				assert.Equal(t, "github", c.Name)
				assert.Equal(t, "https://mcp.github.example", c.Url)
				assert.Len(t, c.HeaderAuthKeys, 1)
				assert.Equal(t, "Authorization", string(c.HeaderAuthKeys[0]))
				c.Id = generatedId
				return c, nil
			})
			suite.mockSecretStorage.EXPECT().SaveValue(gomock.Any(), generatedId, "Authorization", "Bearer abc").Return(nil)
		}

		t.Run("Should persist the row and the header value under the generated id", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingMCPConnectionBeforeEach(t)
			request := &requests.CreateMCPConnectionRequest{
				Name:       "github",
				Url:        "https://mcp.github.example",
				HeaderAuth: map[string]string{"Authorization": "Bearer abc"},
			}
			initSuite(suite)

			res, err := suite.sut.Execute(context.Background(), request)

			assert.NoError(t, err)
			assert.Equal(t, generatedId, res.Id)
			assert.Equal(t, "Bearer abc", res.HeaderAuth["Authorization"])
		})
	})

	t.Run("Given a request with no headers", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingMCPConnectionTestingSuite) {
			suite.mockRepository.EXPECT().Create(gomock.Any()).DoAndReturn(func(c *models.MCPConnection) (*models.MCPConnection, error) {
				assert.Empty(t, c.HeaderAuthKeys)
				c.Id = uuid.New()
				return c, nil
			})
		}

		t.Run("Should not call SaveValue", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingMCPConnectionBeforeEach(t)
			request := &requests.CreateMCPConnectionRequest{Name: "anon", Url: "https://x.example"}
			initSuite(suite)

			res, err := suite.sut.Execute(context.Background(), request)

			assert.NoError(t, err)
			assert.NotNil(t, res)
		})
	})

	t.Run("Given the repository create fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingMCPConnectionTestingSuite) {
			suite.mockRepository.EXPECT().Create(gomock.Any()).Return(nil, errors.New("db down"))
		}

		t.Run("Should not touch the secret storage", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingMCPConnectionBeforeEach(t)
			request := &requests.CreateMCPConnectionRequest{
				Name: "x", Url: "https://x.example",
				HeaderAuth: map[string]string{"K": "V"},
			}
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request)

			assert.ErrorContains(t, err, "db down")
		})
	})
}
