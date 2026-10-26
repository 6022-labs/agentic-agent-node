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

type WhenExecutingCreatePaymentReceiveRuleToolTestingSuite struct {
	sut *tools.CreatePaymentReceiveRuleTool

	mockCreatePaymentReceiveRuleExecutor *interfaces_mocks.MockCreatePaymentReceiveRuleExecutorInterface
}

func WhenExecutingCreatePaymentReceiveRuleToolBeforeEach(
	t *testing.T,
) *WhenExecutingCreatePaymentReceiveRuleToolTestingSuite {
	mockController := gomock.NewController(t)

	mockCreatePaymentReceiveRuleExecutor := interfaces_mocks.NewMockCreatePaymentReceiveRuleExecutorInterface(
		mockController,
	)

	return &WhenExecutingCreatePaymentReceiveRuleToolTestingSuite{
		sut: tools.NewCreatePaymentReceiveRuleTool(
			mockCreatePaymentReceiveRuleExecutor,
		),
		mockCreatePaymentReceiveRuleExecutor: mockCreatePaymentReceiveRuleExecutor,
	}
}

func TestWhenExecutingCreatePaymentReceiveRuleTool(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ruleArgs := func() map[string]any {
		return map[string]any{
			"name":   "premium",
			"price":  "1000",
			"config": map[string]any{"network": "eip155:80002"},
		}
	}

	t.Run("Given the config argument is missing", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail without calling the use case", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingCreatePaymentReceiveRuleToolBeforeEach(t)

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

			suite := WhenExecutingCreatePaymentReceiveRuleToolBeforeEach(t)
			suite.mockCreatePaymentReceiveRuleExecutor.EXPECT().
				Execute(ctx, gomock.Any()).
				Return(nil, errors.New("invalid price"))

			out, err := suite.sut.Execute(ctx, ruleArgs())

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "invalid price")
		})
	})

	t.Run("Given name, price and config arguments", func(t *testing.T) {
		t.Parallel()

		t.Run("Should forward the encoded request to the use case", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingCreatePaymentReceiveRuleToolBeforeEach(t)
			suite.mockCreatePaymentReceiveRuleExecutor.EXPECT().
				Execute(ctx, gomock.Any()).
				DoAndReturn(func(_ context.Context, request *payment_requests.SavePaymentReceiveRuleRequest) (*payment_responses.PaymentReceiveRuleResponse, error) {
					assert.Equal(t, "premium", request.Name)
					assert.Equal(t, "1000", request.Price)
					assert.Empty(t, request.Method)
					assert.Equal(t, json.RawMessage(`{"network":"eip155:80002"}`), request.Config)
					return &payment_responses.PaymentReceiveRuleResponse{}, nil
				})

			out, err := suite.sut.Execute(ctx, ruleArgs())

			assert.NoError(t, err)
			assert.NotNil(t, out)
		})
	})
}
