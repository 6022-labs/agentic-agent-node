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

type WhenExecutingUpdatePaymentSpendPolicyToolTestingSuite struct {
	sut *tools.UpdatePaymentSpendPolicyTool

	mockUpdatePaymentSpendPolicyExecutor *interfaces_mocks.MockUpdatePaymentSpendPolicyExecutorInterface
}

func WhenExecutingUpdatePaymentSpendPolicyToolBeforeEach(
	t *testing.T,
) *WhenExecutingUpdatePaymentSpendPolicyToolTestingSuite {
	mockController := gomock.NewController(t)

	mockUpdatePaymentSpendPolicyExecutor := interfaces_mocks.NewMockUpdatePaymentSpendPolicyExecutorInterface(
		mockController,
	)

	return &WhenExecutingUpdatePaymentSpendPolicyToolTestingSuite{
		sut: tools.NewUpdatePaymentSpendPolicyTool(
			mockUpdatePaymentSpendPolicyExecutor,
		),
		mockUpdatePaymentSpendPolicyExecutor: mockUpdatePaymentSpendPolicyExecutor,
	}
}

func TestWhenExecutingUpdatePaymentSpendPolicyTool(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	validArgs := func() map[string]any {
		return map[string]any{"payToPatterns": []any{"0xabc.*"}}
	}

	t.Run("Given the payToPatterns argument is missing", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail without calling the use case", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingUpdatePaymentSpendPolicyToolBeforeEach(t)

			out, err := suite.sut.Execute(ctx, map[string]any{})

			assert.Nil(t, out)
			assert.ErrorContains(t, err, `argument "payToPatterns" is required`)
		})
	})

	t.Run("Given the payToPatterns argument is not an array", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail without calling the use case", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingUpdatePaymentSpendPolicyToolBeforeEach(t)

			out, err := suite.sut.Execute(ctx, map[string]any{"payToPatterns": ".*"})

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "must be a json array")
		})
	})

	t.Run("Given the use case fails", func(t *testing.T) {
		t.Parallel()

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingUpdatePaymentSpendPolicyToolBeforeEach(t)
			suite.mockUpdatePaymentSpendPolicyExecutor.EXPECT().
				Execute(ctx, gomock.Any()).
				Return(nil, errors.New("boom"))

			out, err := suite.sut.Execute(ctx, validArgs())

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "boom")
		})
	})

	t.Run("Given valid arguments", func(t *testing.T) {
		t.Parallel()

		t.Run("Should map them onto the update request", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingUpdatePaymentSpendPolicyToolBeforeEach(t)
			suite.mockUpdatePaymentSpendPolicyExecutor.EXPECT().
				Execute(ctx, &payment_requests.UpdatePaymentSpendPolicyRequest{PayToPatterns: []string{"0xabc.*"}}).
				Return(&payment_responses.PaymentSpendPolicyResponse{}, nil).
				Times(1)

			out, err := suite.sut.Execute(ctx, validArgs())

			assert.NoError(t, err)
			assert.NotNil(t, out)
		})

	})
}
