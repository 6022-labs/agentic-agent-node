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

type WhenExecutingDeletePaymentRuleToolTestingSuite struct {
	sut *tools.DeletePaymentRuleTool

	mockDeletePaymentRuleExecutor *interfaces_mocks.MockDeletePaymentRuleExecutorInterface
}

func WhenExecutingDeletePaymentRuleToolBeforeEach(t *testing.T) *WhenExecutingDeletePaymentRuleToolTestingSuite {
	mockController := gomock.NewController(t)

	mockDeletePaymentRuleExecutor := interfaces_mocks.NewMockDeletePaymentRuleExecutorInterface(mockController)

	return &WhenExecutingDeletePaymentRuleToolTestingSuite{
		sut:                           tools.NewDeletePaymentRuleTool(mockDeletePaymentRuleExecutor),
		mockDeletePaymentRuleExecutor: mockDeletePaymentRuleExecutor,
	}
}

func TestWhenExecutingDeletePaymentRuleTool(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given a valid id argument", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()

		t.Run("Should delete and return a success payload", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingDeletePaymentRuleToolBeforeEach(t)
			suite.mockDeletePaymentRuleExecutor.EXPECT().
				Execute(ctx, &payment_requests.DeletePaymentRuleRequest{Id: id}).
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

			suite := WhenExecutingDeletePaymentRuleToolBeforeEach(t)

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

			suite := WhenExecutingDeletePaymentRuleToolBeforeEach(t)
			suite.mockDeletePaymentRuleExecutor.EXPECT().
				Execute(ctx, &payment_requests.DeletePaymentRuleRequest{Id: id}).
				Return(errors.New("not found"))

			out, err := suite.sut.Execute(ctx, map[string]any{"id": id.String()})

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "not found")
		})
	})
}
