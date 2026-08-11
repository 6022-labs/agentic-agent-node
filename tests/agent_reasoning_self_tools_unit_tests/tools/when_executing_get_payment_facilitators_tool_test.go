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

type WhenExecutingGetPaymentFacilitatorsToolTestingSuite struct {
	sut *tools.GetPaymentFacilitatorsTool

	mockGetPaymentFacilitatorsExecutor *interfaces_mocks.MockGetPaymentFacilitatorsExecutorInterface
}

func WhenExecutingGetPaymentFacilitatorsToolBeforeEach(
	t *testing.T,
) *WhenExecutingGetPaymentFacilitatorsToolTestingSuite {
	mockController := gomock.NewController(t)

	mockGetPaymentFacilitatorsExecutor := interfaces_mocks.NewMockGetPaymentFacilitatorsExecutorInterface(
		mockController,
	)

	return &WhenExecutingGetPaymentFacilitatorsToolTestingSuite{
		sut:                                tools.NewGetPaymentFacilitatorsTool(mockGetPaymentFacilitatorsExecutor),
		mockGetPaymentFacilitatorsExecutor: mockGetPaymentFacilitatorsExecutor,
	}
}

func TestWhenExecutingGetPaymentFacilitatorsTool(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given the use case succeeds", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return the encoded response", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingGetPaymentFacilitatorsToolBeforeEach(t)
			suite.mockGetPaymentFacilitatorsExecutor.EXPECT().
				Execute(ctx).
				Return(&payment_responses.PaymentFacilitatorsResponse{}, nil)

			out, err := suite.sut.Execute(ctx, nil)

			assert.NoError(t, err)
			assert.NotNil(t, out)
		})
	})

	t.Run("Given the use case fails", func(t *testing.T) {
		t.Parallel()

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingGetPaymentFacilitatorsToolBeforeEach(t)
			suite.mockGetPaymentFacilitatorsExecutor.EXPECT().Execute(ctx).Return(nil, errors.New("boom"))

			out, err := suite.sut.Execute(ctx, nil)

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "boom")
		})
	})
}
