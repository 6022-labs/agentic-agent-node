package use_cases_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/usecase_result"
	repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/repositories_mocks"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenGettingMCPConnectionTestingSuite struct {
	sut *use_cases.GetMCPConnection

	mockRepository    *repositories_mocks.MockMCPConnectionRepositoryInterface
	mockSecretStorage *interfaces_mocks.MockMCPHeaderAuthSecretStorageInterface
}

func WhenGettingMCPConnectionBeforeEach(t *testing.T) *WhenGettingMCPConnectionTestingSuite {
	mockController := gomock.NewController(t)

	mockRepository := repositories_mocks.NewMockMCPConnectionRepositoryInterface(mockController)
	mockSecretStorage := interfaces_mocks.NewMockMCPHeaderAuthSecretStorageInterface(mockController)

	sut := use_cases.NewGetMCPConnection(mockRepository, mockSecretStorage)

	return &WhenGettingMCPConnectionTestingSuite{
		sut: sut,

		mockRepository:    mockRepository,
		mockSecretStorage: mockSecretStorage,
	}
}

func TestWhenGettingMCPConnection(t *testing.T) {
	t.Parallel()

	t.Run("Given an unknown id", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()
		initSuite := func(suite *WhenGettingMCPConnectionTestingSuite) {
			suite.mockRepository.EXPECT().GetById(id).Return(nil, nil)
		}

		t.Run("Should return NotFound without touching the secret storage", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingMCPConnectionBeforeEach(t)
			initSuite(suite)

			res := suite.sut.Execute(context.Background(), id)

			assert.Equal(t, usecase_result.OutcomeNotFound, res.Outcome)
			assert.Equal(t, "mcp_connection_not_found", res.Code)
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
			suite.mockRepository.EXPECT().GetById(id).Return(row, nil)
			suite.mockSecretStorage.EXPECT().GetValue(gomock.Any(), id, "Authorization").Return("Bearer abc", true, nil)
			suite.mockSecretStorage.EXPECT().GetValue(gomock.Any(), id, "X-Trace").Return("trace-1", true, nil)
		}

		t.Run("Should return Ok with the wire response (HeaderAuth is not on the wire)", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingMCPConnectionBeforeEach(t)
			initSuite(suite)

			res := suite.sut.Execute(context.Background(), id)

			assert.Equal(t, usecase_result.OutcomeOk, res.Outcome)
			body, ok := res.Body.(*responses.MCPConnectionResponse)
			assert.True(t, ok)
			assert.Equal(t, id, body.Id)
			assert.ElementsMatch(t, []string{"Authorization", "X-Trace"}, body.HeaderAuthKeys)
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
			suite.mockRepository.EXPECT().GetById(id).Return(row, nil)
			suite.mockSecretStorage.EXPECT().GetValue(gomock.Any(), id, "Authorization").Return("", false, nil)
		}

		t.Run("Should still return Ok with the wire response", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingMCPConnectionBeforeEach(t)
			initSuite(suite)

			res := suite.sut.Execute(context.Background(), id)

			assert.Equal(t, usecase_result.OutcomeOk, res.Outcome)
			body, ok := res.Body.(*responses.MCPConnectionResponse)
			assert.True(t, ok)
			assert.Equal(t, id, body.Id)
		})
	})
}
