package tools_test

import (
	"context"
	"errors"
	"testing"

	payment_requests "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	payment_responses "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_self_tools_mocks/tools_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenExecutingCreatePaymentSpendRuleToolTestingSuite struct {
	sut *tools.CreatePaymentSpendRuleTool

	mockCreatePaymentSpendRuleExecutor *interfaces_mocks.MockCreatePaymentSpendRuleExecutorInterface
}

func WhenExecutingCreatePaymentSpendRuleToolBeforeEach(
	t *testing.T,
) *WhenExecutingCreatePaymentSpendRuleToolTestingSuite {
	mockController := gomock.NewController(t)

	mockCreatePaymentSpendRuleExecutor := interfaces_mocks.NewMockCreatePaymentSpendRuleExecutorInterface(
		mockController,
	)

	return &WhenExecutingCreatePaymentSpendRuleToolTestingSuite{
		sut:                                tools.NewCreatePaymentSpendRuleTool(mockCreatePaymentSpendRuleExecutor),
		mockCreatePaymentSpendRuleExecutor: mockCreatePaymentSpendRuleExecutor,
	}
}

func TestWhenExecutingCreatePaymentSpendRuleTool(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	validArgs := func() map[string]any {
		return map[string]any{
			"name":                "Only base, small",
			"maxAmountPerPayment": "1000",
			"config":              map[string]any{"networkPatterns": []any{"eip155:8453"}},
		}
	}

	// Enabling this tool lets the agent widen its own cap, so the owner must opt in.
	t.Run("Given the tool is registered", func(t *testing.T) {
		t.Parallel()

		t.Run("Should be disabled by default", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingCreatePaymentSpendRuleToolBeforeEach(t)

			assert.False(t, suite.sut.IsEnabledByDefault())
			assert.Equal(t, "create_payment_spend_rule", suite.sut.Definition().Name)
		})
	})

	t.Run("Given valid arguments", func(t *testing.T) {
		t.Parallel()

		t.Run("Should map them onto the save request", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingCreatePaymentSpendRuleToolBeforeEach(t)
			suite.mockCreatePaymentSpendRuleExecutor.EXPECT().
				Execute(ctx, gomock.Any()).
				DoAndReturn(func(
					_ context.Context,
					request *payment_requests.SavePaymentSpendRuleRequest,
				) (*payment_responses.PaymentSpendRuleResponse, error) {
					assert.Equal(t, "Only base, small", request.Name)
					assert.Equal(t, "1000", request.MaxAmountPerPayment)
					assert.Contains(t, string(request.Config), "eip155:8453")
					return &payment_responses.PaymentSpendRuleResponse{}, nil
				})

			out, err := suite.sut.Execute(ctx, validArgs())

			assert.NoError(t, err)
			assert.NotNil(t, out)
		})

		t.Run("Should read every argument the tool asks for", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingCreatePaymentSpendRuleToolBeforeEach(t)

			for _, name := range suite.sut.Definition().InputScheme.Required {
				_, ok := validArgs()[name]
				assert.True(t, ok, "the schema requires %q but Execute never reads it", name)
			}
		})
	})

	t.Run("Given a required argument is missing", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail without calling the use case", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingCreatePaymentSpendRuleToolBeforeEach(t)

			args := validArgs()
			delete(args, "maxAmountPerPayment")
			out, err := suite.sut.Execute(ctx, args)

			assert.Nil(t, out)
			assert.Error(t, err)
		})
	})

	t.Run("Given the use case fails", func(t *testing.T) {
		t.Parallel()

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingCreatePaymentSpendRuleToolBeforeEach(t)
			suite.mockCreatePaymentSpendRuleExecutor.EXPECT().
				Execute(ctx, gomock.Any()).
				Return(nil, errors.New("boom"))

			out, err := suite.sut.Execute(ctx, validArgs())

			assert.Nil(t, out)
			assert.Error(t, err)
		})
	})
}
