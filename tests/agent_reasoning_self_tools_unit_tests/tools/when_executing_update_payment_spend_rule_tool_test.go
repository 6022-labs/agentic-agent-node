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

type WhenExecutingUpdatePaymentSpendRuleToolTestingSuite struct {
	sut *tools.UpdatePaymentSpendRuleTool

	mockUpdatePaymentSpendRuleExecutor *interfaces_mocks.MockUpdatePaymentSpendRuleExecutorInterface
}

func WhenExecutingUpdatePaymentSpendRuleToolBeforeEach(
	t *testing.T,
) *WhenExecutingUpdatePaymentSpendRuleToolTestingSuite {
	mockController := gomock.NewController(t)

	mockUpdatePaymentSpendRuleExecutor := interfaces_mocks.NewMockUpdatePaymentSpendRuleExecutorInterface(
		mockController,
	)

	return &WhenExecutingUpdatePaymentSpendRuleToolTestingSuite{
		sut:                                tools.NewUpdatePaymentSpendRuleTool(mockUpdatePaymentSpendRuleExecutor),
		mockUpdatePaymentSpendRuleExecutor: mockUpdatePaymentSpendRuleExecutor,
	}
}

func TestWhenExecutingUpdatePaymentSpendRuleTool(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	id := uuid.New()
	ruleArgs := func() map[string]any {
		return map[string]any{
			"id":        id.String(),
			"name":      "Only base, small",
			"maxAmount": "1000",
			"config":    map[string]any{"networkPatterns": []any{"eip155:8453"}},
		}
	}

	// Raising your own cap is the reason this one is opt-in.
	t.Run("Given the tool is registered", func(t *testing.T) {
		t.Parallel()

		t.Run("Should be disabled by default", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingUpdatePaymentSpendRuleToolBeforeEach(t)

			assert.False(t, suite.sut.IsEnabledByDefault())
		})
	})

	t.Run("Given valid arguments", func(t *testing.T) {
		t.Parallel()

		t.Run("Should pair the route id with the rule body", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingUpdatePaymentSpendRuleToolBeforeEach(t)
			suite.mockUpdatePaymentSpendRuleExecutor.EXPECT().
				Execute(ctx, gomock.Any()).
				DoAndReturn(func(
					_ context.Context,
					request *payment_requests.UpdatePaymentSpendRuleRequest,
				) (*payment_responses.PaymentSpendRuleResponse, error) {
					assert.Equal(t, id, request.Id)
					assert.Equal(t, "Only base, small", request.Name)
					assert.Equal(t, "1000", request.MaxAmount)
					return &payment_responses.PaymentSpendRuleResponse{}, nil
				})

			out, err := suite.sut.Execute(ctx, ruleArgs())

			assert.NoError(t, err)
			assert.NotNil(t, out)
		})
	})

	t.Run("Given the id argument is not a uuid", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail without calling the use case", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingUpdatePaymentSpendRuleToolBeforeEach(t)

			args := ruleArgs()
			args["id"] = "nope"
			out, err := suite.sut.Execute(ctx, args)

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "uuid")
		})
	})

	t.Run("Given a required argument is missing", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail without calling the use case", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingUpdatePaymentSpendRuleToolBeforeEach(t)

			args := ruleArgs()
			delete(args, "config")
			out, err := suite.sut.Execute(ctx, args)

			assert.Nil(t, out)
			assert.Error(t, err)
		})
	})

	t.Run("Given the use case fails", func(t *testing.T) {
		t.Parallel()

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingUpdatePaymentSpendRuleToolBeforeEach(t)
			suite.mockUpdatePaymentSpendRuleExecutor.EXPECT().
				Execute(ctx, gomock.Any()).
				Return(nil, errors.New("not found"))

			out, err := suite.sut.Execute(ctx, ruleArgs())

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "not found")
		})
	})
}
