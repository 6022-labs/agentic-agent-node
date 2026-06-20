package use_cases_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	identity_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/repositories_mocks"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenGettingMCPConnectionTestingSuite struct {
	sut *use_cases.GetMCPConnection

	mockRepository      *repositories_mocks.MockMCPConnectionRepositoryInterface
	mockSecretStorage   *interfaces_mocks.MockMCPHeaderAuthSecretStorageInterface
	mockAgentRepository *identity_repositories_mocks.MockAgentRepositoryInterface
}

func WhenGettingMCPConnectionBeforeEach(t *testing.T) *WhenGettingMCPConnectionTestingSuite {
	mockController := gomock.NewController(t)

	mockRepository := repositories_mocks.NewMockMCPConnectionRepositoryInterface(mockController)
	mockSecretStorage := interfaces_mocks.NewMockMCPHeaderAuthSecretStorageInterface(mockController)
	mockAgentRepository := identity_repositories_mocks.NewMockAgentRepositoryInterface(mockController)

	sut := use_cases.NewGetMCPConnection(mockRepository, mockSecretStorage, mockAgentRepository)

	return &WhenGettingMCPConnectionTestingSuite{
		sut: sut,

		mockRepository:      mockRepository,
		mockSecretStorage:   mockSecretStorage,
		mockAgentRepository: mockAgentRepository,
	}
}

func TestWhenGettingMCPConnection(t *testing.T) {
	t.Parallel()

	t.Run("Given an unknown id", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()
		initSuite := func(suite *WhenGettingMCPConnectionTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockRepository.EXPECT().GetById(gomock.Any(), gomock.Any(), id).Return(nil, nil)
		}

		t.Run("Should return NotFound without touching the secret storage", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingMCPConnectionBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), id)

			var notFoundErr *apperrors.NotFoundError
			assert.ErrorAs(t, err, &notFoundErr)
			assert.Equal(t, "mcp_connection_not_found", notFoundErr.Code)
		})
	})

	t.Run("Given a connection with two header keys", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()
		row := model_builders.NewMCPConnectionBuilder().
			WithId(id).
			WithName("github").
			WithUrl("https://mcp.github.example").
			WithHeaderAuthKeys([]string{"Authorization", "X-Trace"}).
			Build()
		initSuite := func(suite *WhenGettingMCPConnectionTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockRepository.EXPECT().GetById(gomock.Any(), gomock.Any(), id).Return(row, nil)
			suite.mockSecretStorage.EXPECT().GetValue(gomock.Any(), id, "Authorization").Return("Bearer abc", true, nil)
			suite.mockSecretStorage.EXPECT().GetValue(gomock.Any(), id, "X-Trace").Return("trace-1", true, nil)
		}

		t.Run("Should return Ok with the wire response (HeaderAuth is not on the wire)", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingMCPConnectionBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background(), id)

			assert.NoError(t, err)
			assert.Equal(t, id, resp.Id)
			assert.ElementsMatch(t, []string{"Authorization", "X-Trace"}, resp.HeaderAuthKeys)
		})
	})

	t.Run("Given a vault key is missing for one header", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()
		row := model_builders.NewMCPConnectionBuilder().
			WithId(id).
			WithHeaderAuthKeys([]string{"Authorization"}).
			Build()
		initSuite := func(suite *WhenGettingMCPConnectionTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockRepository.EXPECT().GetById(gomock.Any(), gomock.Any(), id).Return(row, nil)
			suite.mockSecretStorage.EXPECT().GetValue(gomock.Any(), id, "Authorization").Return("", false, nil)
		}

		t.Run("Should still return Ok with the wire response", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingMCPConnectionBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background(), id)

			assert.NoError(t, err)
			assert.Equal(t, id, resp.Id)
		})
	})
}
