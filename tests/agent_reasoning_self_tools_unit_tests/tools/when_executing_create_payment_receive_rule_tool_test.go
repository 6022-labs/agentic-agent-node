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
			"name": "premium",
			"prices": []any{
				map[string]any{"price": "0", "addressPatterns": []any{"0xowner"}},
				map[string]any{"price": "1000", "addressPatterns": []any{".*"}},
			},
			"config": map[string]any{"network": "eip155:80002"},
		}
	}

	t.Run("Given the prices argument is missing", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail without calling the use case", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingCreatePaymentReceiveRuleToolBeforeEach(t)

			args := ruleArgs()
			delete(args, "prices")
			out, err := suite.sut.Execute(ctx, args)

			assert.Nil(t, out)
			assert.ErrorContains(t, err, `argument "prices" is required`)
		})
	})

	t.Run("Given the prices argument is not an array", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail without calling the use case", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingCreatePaymentReceiveRuleToolBeforeEach(t)

			args := ruleArgs()
			args["prices"] = "1000"
			out, err := suite.sut.Execute(ctx, args)

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "must be a json array")
		})
	})

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

	t.Run("Given valid arguments", func(t *testing.T) {
		t.Parallel()

		t.Run("Should forward the encoded request to the use case", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingCreatePaymentReceiveRuleToolBeforeEach(t)
			suite.mockCreatePaymentReceiveRuleExecutor.EXPECT().
				Execute(ctx, gomock.Any()).
				DoAndReturn(func(_ context.Context, request *payment_requests.SavePaymentReceiveRuleRequest) (*payment_responses.PaymentReceiveRuleResponse, error) {
					assert.Equal(t, "premium", request.Name)
					assert.Equal(t, []payment_requests.SavePaymentReceiveRulePriceRequest{
						{Price: "0", AddressPatterns: []string{"0xowner"}},
						{Price: "1000", AddressPatterns: []string{".*"}},
					}, request.Prices)
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
