package tools_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	payment_requests "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	payment_responses "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_self_tools_mocks/tools_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenExecutingCreatePaymentRuleToolTestingSuite struct {
	sut *tools.CreatePaymentRuleTool

	mockCreatePaymentRuleExecutor *interfaces_mocks.MockCreatePaymentRuleExecutorInterface
}

func WhenExecutingCreatePaymentRuleToolBeforeEach(t *testing.T) *WhenExecutingCreatePaymentRuleToolTestingSuite {
	mockController := gomock.NewController(t)

	mockCreatePaymentRuleExecutor := interfaces_mocks.NewMockCreatePaymentRuleExecutorInterface(mockController)

	return &WhenExecutingCreatePaymentRuleToolTestingSuite{
		sut:                           tools.NewCreatePaymentRuleTool(mockCreatePaymentRuleExecutor),
		mockCreatePaymentRuleExecutor: mockCreatePaymentRuleExecutor,
	}
}

func TestWhenExecutingCreatePaymentRuleTool(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ruleArgs := func() map[string]any {
		return map[string]any{
			"name":   "premium",
			"price":  "1000",
			"config": map[string]any{"network": "eip155:80002"},
		}
	}

	t.Run("Given name, price and config arguments", func(t *testing.T) {
		t.Parallel()

		t.Run("Should forward the encoded request to the use case", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingCreatePaymentRuleToolBeforeEach(t)
			suite.mockCreatePaymentRuleExecutor.EXPECT().
				Execute(ctx, gomock.Any()).
				DoAndReturn(func(_ context.Context, request *payment_requests.SavePaymentRuleRequest) (*payment_responses.PaymentRuleResponse, error) {
					assert.Equal(t, "premium", request.Name)
					assert.Equal(t, "1000", request.Price)
					assert.Empty(t, request.Method)
					assert.Equal(t, json.RawMessage(`{"network":"eip155:80002"}`), request.Config)
					return &payment_responses.PaymentRuleResponse{}, nil
				})

			out, err := suite.sut.Execute(ctx, ruleArgs())

			assert.NoError(t, err)
			assert.NotNil(t, out)
		})
	})

	t.Run("Given the config argument is missing", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail without calling the use case", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingCreatePaymentRuleToolBeforeEach(t)

			args := ruleArgs()
			delete(args, "config")
			out, err := suite.sut.Execute(ctx, args)

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "config")
		})
	})

	t.Run("Given the use case fails", func(t *testing.T) {
		t.Parallel()

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingCreatePaymentRuleToolBeforeEach(t)
			suite.mockCreatePaymentRuleExecutor.EXPECT().
				Execute(ctx, gomock.Any()).
				Return(nil, errors.New("invalid price"))

			out, err := suite.sut.Execute(ctx, ruleArgs())

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "invalid price")
		})
	})
}
