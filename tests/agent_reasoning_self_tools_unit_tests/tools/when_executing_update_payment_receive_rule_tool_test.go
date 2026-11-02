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

type WhenExecutingUpdatePaymentReceiveRuleToolTestingSuite struct {
	sut *tools.UpdatePaymentReceiveRuleTool

	mockUpdatePaymentReceiveRuleExecutor *interfaces_mocks.MockUpdatePaymentReceiveRuleExecutorInterface
}

func WhenExecutingUpdatePaymentReceiveRuleToolBeforeEach(
	t *testing.T,
) *WhenExecutingUpdatePaymentReceiveRuleToolTestingSuite {
	mockController := gomock.NewController(t)

	mockUpdatePaymentReceiveRuleExecutor := interfaces_mocks.NewMockUpdatePaymentReceiveRuleExecutorInterface(
		mockController,
	)

	return &WhenExecutingUpdatePaymentReceiveRuleToolTestingSuite{
		sut: tools.NewUpdatePaymentReceiveRuleTool(
			mockUpdatePaymentReceiveRuleExecutor,
		),
		mockUpdatePaymentReceiveRuleExecutor: mockUpdatePaymentReceiveRuleExecutor,
	}
}

func TestWhenExecutingUpdatePaymentReceiveRuleTool(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	id := uuid.New()
	ruleArgs := func() map[string]any {
		return map[string]any{
			"id":   id.String(),
			"name": "premium",
			"prices": []any{
				map[string]any{"price": "0", "addressPatterns": []any{"0xowner"}},
				map[string]any{"price": "1000", "addressPatterns": []any{".*"}},
			},
			"method": "x402",
			"config": map[string]any{"network": "eip155:80002"},
		}
	}

	t.Run("Given the prices argument is missing", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail without calling the use case", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingUpdatePaymentReceiveRuleToolBeforeEach(t)

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

			suite := WhenExecutingUpdatePaymentReceiveRuleToolBeforeEach(t)

			args := ruleArgs()
			args["prices"] = "1000"
			out, err := suite.sut.Execute(ctx, args)

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "must be a json array")
		})
	})

	t.Run("Given the use case fails", func(t *testing.T) {
		t.Parallel()

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingUpdatePaymentReceiveRuleToolBeforeEach(t)
			suite.mockUpdatePaymentReceiveRuleExecutor.EXPECT().
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

			suite := WhenExecutingUpdatePaymentReceiveRuleToolBeforeEach(t)
			suite.mockUpdatePaymentReceiveRuleExecutor.EXPECT().
				Execute(ctx, gomock.Any()).
				DoAndReturn(func(_ context.Context, request *payment_requests.UpdatePaymentReceiveRuleRequest) (*payment_responses.PaymentReceiveRuleResponse, error) {
					assert.Equal(t, id, request.Id)
					assert.Equal(t, "premium", request.Name)
					assert.Equal(t, []payment_requests.SavePaymentReceiveRulePriceRequest{
						{Price: "0", AddressPatterns: []string{"0xowner"}},
						{Price: "1000", AddressPatterns: []string{".*"}},
					}, request.Prices)
					assert.Equal(t, "x402", request.Method)
					return &payment_responses.PaymentReceiveRuleResponse{}, nil
				})

			out, err := suite.sut.Execute(ctx, ruleArgs())

			assert.NoError(t, err)
			assert.NotNil(t, out)
		})
	})
}
