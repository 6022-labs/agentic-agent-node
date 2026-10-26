package tools_test

import (
	"context"
	"errors"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_self_tools_mocks/tools_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenExecutingDeletePaymentReceivePolicyToolTestingSuite struct {
	sut *tools.DeletePaymentReceivePolicyTool

	mockDeletePaymentReceivePolicyExecutor *interfaces_mocks.MockDeletePaymentReceivePolicyExecutorInterface
}

func WhenExecutingDeletePaymentReceivePolicyToolBeforeEach(
	t *testing.T,
) *WhenExecutingDeletePaymentReceivePolicyToolTestingSuite {
	mockController := gomock.NewController(t)

	mockDeletePaymentReceivePolicyExecutor := interfaces_mocks.NewMockDeletePaymentReceivePolicyExecutorInterface(
		mockController,
	)

	return &WhenExecutingDeletePaymentReceivePolicyToolTestingSuite{
		sut: tools.NewDeletePaymentReceivePolicyTool(
			mockDeletePaymentReceivePolicyExecutor,
		),
		mockDeletePaymentReceivePolicyExecutor: mockDeletePaymentReceivePolicyExecutor,
	}
}

func TestWhenExecutingDeletePaymentReceivePolicyTool(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given the use case fails", func(t *testing.T) {
		t.Parallel()

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingDeletePaymentReceivePolicyToolBeforeEach(t)
			suite.mockDeletePaymentReceivePolicyExecutor.EXPECT().Execute(ctx).Return(errors.New("db down"))

			out, err := suite.sut.Execute(ctx, nil)

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "db down")
		})
	})

	t.Run("Given the use case succeeds", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a success payload", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingDeletePaymentReceivePolicyToolBeforeEach(t)
			suite.mockDeletePaymentReceivePolicyExecutor.EXPECT().Execute(ctx).Return(nil)

			out, err := suite.sut.Execute(ctx, nil)

			assert.NoError(t, err)
			assert.Equal(t, map[string]any{"success": true}, out)
		})
	})
}
