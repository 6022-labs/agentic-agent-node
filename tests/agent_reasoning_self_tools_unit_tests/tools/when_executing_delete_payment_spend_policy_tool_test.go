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

type WhenExecutingDeletePaymentSpendPolicyToolTestingSuite struct {
	sut *tools.DeletePaymentSpendPolicyTool

	mockDeletePaymentSpendPolicyExecutor *interfaces_mocks.MockDeletePaymentSpendPolicyExecutorInterface
}

func WhenExecutingDeletePaymentSpendPolicyToolBeforeEach(
	t *testing.T,
) *WhenExecutingDeletePaymentSpendPolicyToolTestingSuite {
	mockController := gomock.NewController(t)

	mockDeletePaymentSpendPolicyExecutor := interfaces_mocks.NewMockDeletePaymentSpendPolicyExecutorInterface(
		mockController,
	)

	return &WhenExecutingDeletePaymentSpendPolicyToolTestingSuite{
		sut: tools.NewDeletePaymentSpendPolicyTool(mockDeletePaymentSpendPolicyExecutor),

		mockDeletePaymentSpendPolicyExecutor: mockDeletePaymentSpendPolicyExecutor,
	}
}

func TestWhenExecutingDeletePaymentSpendPolicyTool(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	t.Run("Given the reset succeeds", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report success", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingDeletePaymentSpendPolicyToolBeforeEach(t)
			suite.mockDeletePaymentSpendPolicyExecutor.EXPECT().Execute(ctx).Return(nil)

			out, err := suite.sut.Execute(ctx, nil)

			assert.NoError(t, err)
			assert.Equal(t, map[string]any{"success": true}, out)
		})
	})

	t.Run("Given the use case fails", func(t *testing.T) {
		t.Parallel()

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingDeletePaymentSpendPolicyToolBeforeEach(t)
			suite.mockDeletePaymentSpendPolicyExecutor.EXPECT().Execute(ctx).Return(errors.New("boom"))

			out, err := suite.sut.Execute(ctx, nil)

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "boom")
		})
	})
}
