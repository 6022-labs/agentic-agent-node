package tools_test

import (
	"context"
	"errors"
	"testing"

	payment_responses "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_self_tools_mocks/tools_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenExecutingListPaymentSpendRulesToolTestingSuite struct {
	sut *tools.ListPaymentSpendRulesTool

	mockListPaymentSpendRulesExecutor *interfaces_mocks.MockListPaymentSpendRulesExecutorInterface
}

func WhenExecutingListPaymentSpendRulesToolBeforeEach(
	t *testing.T,
) *WhenExecutingListPaymentSpendRulesToolTestingSuite {
	mockController := gomock.NewController(t)

	mockListPaymentSpendRulesExecutor := interfaces_mocks.NewMockListPaymentSpendRulesExecutorInterface(mockController)

	return &WhenExecutingListPaymentSpendRulesToolTestingSuite{
		sut:                               tools.NewListPaymentSpendRulesTool(mockListPaymentSpendRulesExecutor),
		mockListPaymentSpendRulesExecutor: mockListPaymentSpendRulesExecutor,
	}
}

func TestWhenExecutingListPaymentSpendRulesTool(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Even reading the policy is opt-in, matching every other payment self tool.
	t.Run("Given the tool is registered", func(t *testing.T) {
		t.Parallel()

		t.Run("Should be disabled by default", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingListPaymentSpendRulesToolBeforeEach(t)

			assert.False(t, suite.sut.IsEnabledByDefault())
		})
	})

	t.Run("Given the use case succeeds", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return the encoded response", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingListPaymentSpendRulesToolBeforeEach(t)
			suite.mockListPaymentSpendRulesExecutor.EXPECT().
				Execute(ctx).
				Return(&payment_responses.ListPaymentSpendRulesResponse{}, nil)

			out, err := suite.sut.Execute(ctx, nil)

			assert.NoError(t, err)
			assert.NotNil(t, out)
		})
	})

	t.Run("Given the use case fails", func(t *testing.T) {
		t.Parallel()

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingListPaymentSpendRulesToolBeforeEach(t)
			suite.mockListPaymentSpendRulesExecutor.EXPECT().Execute(ctx).Return(nil, errors.New("boom"))

			out, err := suite.sut.Execute(ctx, nil)

			assert.Nil(t, out)
			assert.Error(t, err)
		})
	})
}
