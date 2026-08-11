package tools_test

import (
	"context"
	"errors"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools"
	common_responses "github.com/6022-labs/agentic-agent-node/src/runtime_common/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_self_tools_mocks/tools_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenExecutingGetBlockchainConfigToolTestingSuite struct {
	sut *tools.GetBlockchainConfigTool

	mockGetBlockchainConfigExecutor *interfaces_mocks.MockGetBlockchainConfigExecutorInterface
}

func WhenExecutingGetBlockchainConfigToolBeforeEach(t *testing.T) *WhenExecutingGetBlockchainConfigToolTestingSuite {
	mockController := gomock.NewController(t)

	mockGetBlockchainConfigExecutor := interfaces_mocks.NewMockGetBlockchainConfigExecutorInterface(mockController)

	return &WhenExecutingGetBlockchainConfigToolTestingSuite{
		sut:                             tools.NewGetBlockchainConfigTool(mockGetBlockchainConfigExecutor),
		mockGetBlockchainConfigExecutor: mockGetBlockchainConfigExecutor,
	}
}

func TestWhenExecutingGetBlockchainConfigTool(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given the use case succeeds", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return the encoded response", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingGetBlockchainConfigToolBeforeEach(t)
			suite.mockGetBlockchainConfigExecutor.EXPECT().
				Execute(ctx).
				Return(&common_responses.BlockchainConfigResponse{}, nil)

			out, err := suite.sut.Execute(ctx, nil)

			assert.NoError(t, err)
			assert.NotNil(t, out)
		})
	})

	t.Run("Given the use case fails", func(t *testing.T) {
		t.Parallel()

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingGetBlockchainConfigToolBeforeEach(t)
			suite.mockGetBlockchainConfigExecutor.EXPECT().Execute(ctx).Return(nil, errors.New("boom"))

			out, err := suite.sut.Execute(ctx, nil)

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "boom")
		})
	})
}
