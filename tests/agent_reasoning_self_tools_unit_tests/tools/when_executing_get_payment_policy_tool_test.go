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

type WhenExecutingGetPaymentPolicyToolTestingSuite struct {
	sut *tools.GetPaymentPolicyTool

	mockGetPaymentPolicyExecutor *interfaces_mocks.MockGetPaymentPolicyExecutorInterface
}

func WhenExecutingGetPaymentPolicyToolBeforeEach(t *testing.T) *WhenExecutingGetPaymentPolicyToolTestingSuite {
	mockController := gomock.NewController(t)

	mockGetPaymentPolicyExecutor := interfaces_mocks.NewMockGetPaymentPolicyExecutorInterface(mockController)

	return &WhenExecutingGetPaymentPolicyToolTestingSuite{
		sut:                          tools.NewGetPaymentPolicyTool(mockGetPaymentPolicyExecutor),
		mockGetPaymentPolicyExecutor: mockGetPaymentPolicyExecutor,
	}
}

func TestWhenExecutingGetPaymentPolicyTool(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given the use case fails", func(t *testing.T) {
		t.Parallel()

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingGetPaymentPolicyToolBeforeEach(t)
			suite.mockGetPaymentPolicyExecutor.EXPECT().Execute(ctx).Return(nil, errors.New("boom"))

			out, err := suite.sut.Execute(ctx, nil)

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "boom")
		})
	})

	t.Run("Given the use case succeeds", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return the encoded response", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingGetPaymentPolicyToolBeforeEach(t)
			suite.mockGetPaymentPolicyExecutor.EXPECT().
				Execute(ctx).
				Return(&payment_responses.PaymentPolicyResponse{}, nil)

			out, err := suite.sut.Execute(ctx, nil)

			assert.NoError(t, err)
			assert.NotNil(t, out)
		})
	})
}
