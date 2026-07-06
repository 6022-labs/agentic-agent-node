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

type WhenExecutingListPaymentRulesToolTestingSuite struct {
	sut *tools.ListPaymentRulesTool

	mockListPaymentRulesExecutor *interfaces_mocks.MockListPaymentRulesExecutorInterface
}

func WhenExecutingListPaymentRulesToolBeforeEach(t *testing.T) *WhenExecutingListPaymentRulesToolTestingSuite {
	mockController := gomock.NewController(t)

	mockListPaymentRulesExecutor := interfaces_mocks.NewMockListPaymentRulesExecutorInterface(mockController)

	return &WhenExecutingListPaymentRulesToolTestingSuite{
		sut:                          tools.NewListPaymentRulesTool(mockListPaymentRulesExecutor),
		mockListPaymentRulesExecutor: mockListPaymentRulesExecutor,
	}
}

func TestWhenExecutingListPaymentRulesTool(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given the use case succeeds", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return the encoded response", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingListPaymentRulesToolBeforeEach(t)
			suite.mockListPaymentRulesExecutor.EXPECT().Execute(ctx).Return(&payment_responses.ListPaymentRulesResponse{}, nil)

			out, err := suite.sut.Execute(ctx, nil)

			assert.NoError(t, err)
			assert.NotNil(t, out)
		})
	})

	t.Run("Given the use case fails", func(t *testing.T) {
		t.Parallel()

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingListPaymentRulesToolBeforeEach(t)
			suite.mockListPaymentRulesExecutor.EXPECT().Execute(ctx).Return(nil, errors.New("boom"))

			out, err := suite.sut.Execute(ctx, nil)

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "boom")
		})
	})
}
