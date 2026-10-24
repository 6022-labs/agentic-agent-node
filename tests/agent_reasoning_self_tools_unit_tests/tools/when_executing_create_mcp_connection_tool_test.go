package tools_test

import (
	"context"
	"errors"
	"testing"

	reasoning_requests "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	reasoning_responses "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_self_tools_mocks/tools_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenExecutingCreateMCPConnectionToolTestingSuite struct {
	sut *tools.CreateMCPConnectionTool

	mockCreateMCPConnectionExecutor *interfaces_mocks.MockCreateMCPConnectionExecutorInterface
}

func WhenExecutingCreateMCPConnectionToolBeforeEach(t *testing.T) *WhenExecutingCreateMCPConnectionToolTestingSuite {
	mockController := gomock.NewController(t)

	mockCreateMCPConnectionExecutor := interfaces_mocks.NewMockCreateMCPConnectionExecutorInterface(mockController)

	return &WhenExecutingCreateMCPConnectionToolTestingSuite{
		sut:                             tools.NewCreateMCPConnectionTool(mockCreateMCPConnectionExecutor),
		mockCreateMCPConnectionExecutor: mockCreateMCPConnectionExecutor,
	}
}

func TestWhenExecutingCreateMCPConnectionTool(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given the url argument is missing", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail without calling the use case", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingCreateMCPConnectionToolBeforeEach(t)

			out, err := suite.sut.Execute(ctx, map[string]any{"name": "weather"})

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "url")
		})
	})

	t.Run("Given the use case fails", func(t *testing.T) {
		t.Parallel()

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingCreateMCPConnectionToolBeforeEach(t)
			suite.mockCreateMCPConnectionExecutor.EXPECT().
				Execute(ctx, gomock.Any()).
				Return(nil, errors.New("name cannot contain underscores"))

			out, err := suite.sut.Execute(ctx, map[string]any{"name": "my_tools", "url": "https://mcp.example.com"})

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "underscores")
		})
	})

	t.Run("Given name and url arguments", func(t *testing.T) {
		t.Parallel()

		t.Run("Should forward a header-less request to the use case", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingCreateMCPConnectionToolBeforeEach(t)
			suite.mockCreateMCPConnectionExecutor.EXPECT().
				Execute(ctx, &reasoning_requests.CreateMCPConnectionRequest{
					MCPConnectionRequest: reasoning_requests.MCPConnectionRequest{
						Name:                "weather",
						Url:                 "https://mcp.example.com",
						AllowedToolPatterns: []string{".*"},
					},
				}).
				Return(&reasoning_responses.MCPConnectionResponse{}, nil)

			out, err := suite.sut.Execute(ctx, map[string]any{"name": "weather", "url": "https://mcp.example.com"})

			assert.NoError(t, err)
			assert.NotNil(t, out)
		})
	})
}
