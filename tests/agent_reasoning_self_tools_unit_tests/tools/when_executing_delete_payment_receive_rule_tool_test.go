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

type WhenExecutingDeletePaymentReceiveRuleToolTestingSuite struct {
	sut *tools.DeletePaymentReceiveRuleTool

	mockDeletePaymentReceiveRuleExecutor *interfaces_mocks.MockDeletePaymentReceiveRuleExecutorInterface
}

func WhenExecutingDeletePaymentReceiveRuleToolBeforeEach(
	t *testing.T,
) *WhenExecutingDeletePaymentReceiveRuleToolTestingSuite {
	mockController := gomock.NewController(t)

	mockDeletePaymentReceiveRuleExecutor := interfaces_mocks.NewMockDeletePaymentReceiveRuleExecutorInterface(
		mockController,
	)

	return &WhenExecutingDeletePaymentReceiveRuleToolTestingSuite{
		sut: tools.NewDeletePaymentReceiveRuleTool(
			mockDeletePaymentReceiveRuleExecutor,
		),
		mockDeletePaymentReceiveRuleExecutor: mockDeletePaymentReceiveRuleExecutor,
	}
}

func TestWhenExecutingDeletePaymentReceiveRuleTool(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given a malformed id argument", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail without calling the use case", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingDeletePaymentReceiveRuleToolBeforeEach(t)

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

			suite := WhenExecutingDeletePaymentReceiveRuleToolBeforeEach(t)
			suite.mockDeletePaymentReceiveRuleExecutor.EXPECT().
				Execute(ctx, &payment_requests.DeletePaymentReceiveRuleRequest{Id: id}).
				Return(errors.New("not found"))

			out, err := suite.sut.Execute(ctx, map[string]any{"id": id.String()})

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "not found")
		})
	})

	t.Run("Given a valid id argument", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()

		t.Run("Should return a success payload", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingDeletePaymentReceiveRuleToolBeforeEach(t)
			suite.mockDeletePaymentReceiveRuleExecutor.EXPECT().
				Execute(ctx, &payment_requests.DeletePaymentReceiveRuleRequest{Id: id}).
				Return(nil)

			out, err := suite.sut.Execute(ctx, map[string]any{"id": id.String()})

			assert.NoError(t, err)
			assert.Equal(t, map[string]any{"success": true}, out)
		})
	})
}
