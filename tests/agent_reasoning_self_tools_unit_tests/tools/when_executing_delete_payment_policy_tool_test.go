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

type WhenExecutingDeletePaymentPolicyToolTestingSuite struct {
	sut *tools.DeletePaymentPolicyTool

	mockDeletePaymentPolicyExecutor *interfaces_mocks.MockDeletePaymentPolicyExecutorInterface
}

func WhenExecutingDeletePaymentPolicyToolBeforeEach(t *testing.T) *WhenExecutingDeletePaymentPolicyToolTestingSuite {
	mockController := gomock.NewController(t)

	mockDeletePaymentPolicyExecutor := interfaces_mocks.NewMockDeletePaymentPolicyExecutorInterface(mockController)

	return &WhenExecutingDeletePaymentPolicyToolTestingSuite{
		sut:                             tools.NewDeletePaymentPolicyTool(mockDeletePaymentPolicyExecutor),
		mockDeletePaymentPolicyExecutor: mockDeletePaymentPolicyExecutor,
	}
}

func TestWhenExecutingDeletePaymentPolicyTool(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given the use case succeeds", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a success payload", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingDeletePaymentPolicyToolBeforeEach(t)
			suite.mockDeletePaymentPolicyExecutor.EXPECT().Execute(ctx).Return(nil)

			out, err := suite.sut.Execute(ctx, nil)

			assert.NoError(t, err)
			assert.Equal(t, map[string]any{"success": true}, out)
		})
	})

	t.Run("Given the use case fails", func(t *testing.T) {
		t.Parallel()

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingDeletePaymentPolicyToolBeforeEach(t)
			suite.mockDeletePaymentPolicyExecutor.EXPECT().Execute(ctx).Return(errors.New("db down"))

			out, err := suite.sut.Execute(ctx, nil)

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "db down")
		})
	})
}
