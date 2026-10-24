package tools_test

import (
	"context"
	"errors"
	"testing"

	payment_requests "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	payment_responses "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_self_tools_mocks/tools_mocks/interfaces_mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenExecutingUpdatePaymentRuleToolTestingSuite struct {
	sut *tools.UpdatePaymentRuleTool

	mockUpdatePaymentRuleExecutor *interfaces_mocks.MockUpdatePaymentRuleExecutorInterface
}

func WhenExecutingUpdatePaymentRuleToolBeforeEach(t *testing.T) *WhenExecutingUpdatePaymentRuleToolTestingSuite {
	mockController := gomock.NewController(t)

	mockUpdatePaymentRuleExecutor := interfaces_mocks.NewMockUpdatePaymentRuleExecutorInterface(mockController)

	return &WhenExecutingUpdatePaymentRuleToolTestingSuite{
		sut:                           tools.NewUpdatePaymentRuleTool(mockUpdatePaymentRuleExecutor),
		mockUpdatePaymentRuleExecutor: mockUpdatePaymentRuleExecutor,
	}
}

func TestWhenExecutingUpdatePaymentRuleTool(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	id := uuid.New()
	ruleArgs := func() map[string]any {
		return map[string]any{
			"id":     id.String(),
			"name":   "premium",
			"price":  "2000",
			"method": "x402",
			"config": map[string]any{"network": "eip155:80002"},
		}
	}

	t.Run("Given the id argument is missing", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail without calling the use case", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingUpdatePaymentRuleToolBeforeEach(t)

			args := ruleArgs()
			delete(args, "id")
			out, err := suite.sut.Execute(ctx, args)

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "id")
		})
	})

	t.Run("Given the use case fails", func(t *testing.T) {
		t.Parallel()

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingUpdatePaymentRuleToolBeforeEach(t)
			suite.mockUpdatePaymentRuleExecutor.EXPECT().
				Execute(ctx, gomock.Any()).
				Return(nil, errors.New("rule not found"))

			out, err := suite.sut.Execute(ctx, ruleArgs())

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "rule not found")
		})
	})

	t.Run("Given a full update payload", func(t *testing.T) {
		t.Parallel()

		t.Run("Should forward the id and the request to the use case", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingUpdatePaymentRuleToolBeforeEach(t)
			suite.mockUpdatePaymentRuleExecutor.EXPECT().
				Execute(ctx, gomock.Any()).
				DoAndReturn(func(_ context.Context, request *payment_requests.UpdatePaymentRuleRequest) (*payment_responses.PaymentRuleResponse, error) {
					assert.Equal(t, id, request.Id)
					assert.Equal(t, "premium", request.Name)
					assert.Equal(t, "2000", request.Price)
					assert.Equal(t, "x402", request.Method)
					return &payment_responses.PaymentRuleResponse{}, nil
				})

			out, err := suite.sut.Execute(ctx, ruleArgs())

			assert.NoError(t, err)
			assert.NotNil(t, out)
		})
	})
}
