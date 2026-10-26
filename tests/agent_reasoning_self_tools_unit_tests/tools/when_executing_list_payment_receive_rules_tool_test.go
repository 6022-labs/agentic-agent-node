package tools_test

import (
	"context"
	"errors"
	"testing"

	payment_responses "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_self_tools_mocks/tools_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenExecutingListPaymentReceiveRulesToolTestingSuite struct {
	sut *tools.ListPaymentReceiveRulesTool

	mockListPaymentReceiveRulesExecutor *interfaces_mocks.MockListPaymentReceiveRulesExecutorInterface
}

func WhenExecutingListPaymentReceiveRulesToolBeforeEach(
	t *testing.T,
) *WhenExecutingListPaymentReceiveRulesToolTestingSuite {
	mockController := gomock.NewController(t)

	mockListPaymentReceiveRulesExecutor := interfaces_mocks.NewMockListPaymentReceiveRulesExecutorInterface(
		mockController,
	)

	return &WhenExecutingListPaymentReceiveRulesToolTestingSuite{
		sut:                                 tools.NewListPaymentReceiveRulesTool(mockListPaymentReceiveRulesExecutor),
		mockListPaymentReceiveRulesExecutor: mockListPaymentReceiveRulesExecutor,
	}
}

func TestWhenExecutingListPaymentReceiveRulesTool(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given the use case fails", func(t *testing.T) {
		t.Parallel()

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingListPaymentReceiveRulesToolBeforeEach(t)
			suite.mockListPaymentReceiveRulesExecutor.EXPECT().Execute(ctx).Return(nil, errors.New("boom"))

			out, err := suite.sut.Execute(ctx, nil)

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "boom")
		})
	})

	t.Run("Given the use case succeeds", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return the encoded response", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingListPaymentReceiveRulesToolBeforeEach(t)
			suite.mockListPaymentReceiveRulesExecutor.EXPECT().
				Execute(ctx).
				Return(&payment_responses.ListPaymentReceiveRulesResponse{}, nil)

			out, err := suite.sut.Execute(ctx, nil)

			assert.NoError(t, err)
			assert.NotNil(t, out)
		})
	})
}
