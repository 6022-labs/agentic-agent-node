package tools_test

import (
	"context"
	"errors"
	"testing"

	payment_requests "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_self_tools_mocks/tools_mocks/interfaces_mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenExecutingDeletePaymentSpendRuleToolTestingSuite struct {
	sut *tools.DeletePaymentSpendRuleTool

	mockDeletePaymentSpendRuleExecutor *interfaces_mocks.MockDeletePaymentSpendRuleExecutorInterface
}

func WhenExecutingDeletePaymentSpendRuleToolBeforeEach(
	t *testing.T,
) *WhenExecutingDeletePaymentSpendRuleToolTestingSuite {
	mockController := gomock.NewController(t)

	mockDeletePaymentSpendRuleExecutor := interfaces_mocks.NewMockDeletePaymentSpendRuleExecutorInterface(
		mockController,
	)

	return &WhenExecutingDeletePaymentSpendRuleToolTestingSuite{
		sut:                                tools.NewDeletePaymentSpendRuleTool(mockDeletePaymentSpendRuleExecutor),
		mockDeletePaymentSpendRuleExecutor: mockDeletePaymentSpendRuleExecutor,
	}
}

func TestWhenExecutingDeletePaymentSpendRuleTool(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given the rule exists", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()

		t.Run("Should report success", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingDeletePaymentSpendRuleToolBeforeEach(t)
			suite.mockDeletePaymentSpendRuleExecutor.EXPECT().
				Execute(ctx, &payment_requests.DeletePaymentSpendRuleRequest{Id: id}).
				Return(nil)

			out, err := suite.sut.Execute(ctx, map[string]any{"id": id.String()})

			assert.NoError(t, err)
			assert.Equal(t, map[string]any{"success": true}, out)
		})
	})

	t.Run("Given the id argument is not a uuid", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail without calling the use case", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingDeletePaymentSpendRuleToolBeforeEach(t)

			out, err := suite.sut.Execute(ctx, map[string]any{"id": "nope"})

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "uuid")
		})
	})

	t.Run("Given the use case fails", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingDeletePaymentSpendRuleToolBeforeEach(t)
			suite.mockDeletePaymentSpendRuleExecutor.EXPECT().
				Execute(ctx, &payment_requests.DeletePaymentSpendRuleRequest{Id: id}).
				Return(errors.New("not found"))

			out, err := suite.sut.Execute(ctx, map[string]any{"id": id.String()})

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "not found")
		})
	})
}
