package tools_test

import (
	"context"
	"errors"
	"testing"

	payment_requests "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	payment_responses "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_self_tools_mocks/tools_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenExecutingSignX402PaymentToolTestingSuite struct {
	sut *tools.SignX402PaymentTool

	mockSignX402PaymentExecutor *interfaces_mocks.MockSignX402PaymentExecutorInterface
}

func WhenExecutingSignX402PaymentToolBeforeEach(t *testing.T) *WhenExecutingSignX402PaymentToolTestingSuite {
	mockController := gomock.NewController(t)

	mockSignX402PaymentExecutor := interfaces_mocks.NewMockSignX402PaymentExecutorInterface(mockController)

	return &WhenExecutingSignX402PaymentToolTestingSuite{
		sut:                         tools.NewSignX402PaymentTool(mockSignX402PaymentExecutor),
		mockSignX402PaymentExecutor: mockSignX402PaymentExecutor,
	}
}

func TestWhenExecutingSignX402PaymentTool(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given a quote entry forwarded verbatim from a 402", func(t *testing.T) {
		t.Parallel()

		t.Run("Should map it onto the use case request, extra included", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingSignX402PaymentToolBeforeEach(t)
			suite.mockSignX402PaymentExecutor.EXPECT().
				Execute(ctx, &payment_requests.SignX402PaymentRequest{
					Network: "eip155:80002",
					Asset:   "0x41E94Eb019C0762f9Bfcf9Fb1E58725BfB0e7582",
					PayTo:   "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
					Amount:  "10000",
					Extra: map[string]any{
						"assetTransferMethod": "eip3009",
						"name":                "USDC",
						"version":             "2",
					},
				}).
				Return(&payment_responses.X402PaymentSignatureResponse{PaymentSignature: "eyJ4NDAy"}, nil)

			out, err := suite.sut.Execute(ctx, map[string]any{
				"accept": map[string]any{
					"network": "eip155:80002",
					"asset":   "0x41E94Eb019C0762f9Bfcf9Fb1E58725BfB0e7582",
					"payTo":   "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
					"amount":  "10000",
					"extra": map[string]any{
						"assetTransferMethod": "eip3009",
						"name":                "USDC",
						"version":             "2",
					},
				},
			})

			assert.NoError(t, err)
			assert.Equal(t, "eyJ4NDAy", out["paymentSignature"])
		})
	})

	t.Run("Given the accept argument is missing", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail without calling the use case", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingSignX402PaymentToolBeforeEach(t)

			out, err := suite.sut.Execute(ctx, map[string]any{})

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "accept")
		})
	})

	t.Run("Given the use case fails", func(t *testing.T) {
		t.Parallel()

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingSignX402PaymentToolBeforeEach(t)
			suite.mockSignX402PaymentExecutor.EXPECT().
				Execute(ctx, gomock.Any()).
				Return(nil, errors.New("amount exceeds the configured cap"))

			out, err := suite.sut.Execute(ctx, map[string]any{"accept": map[string]any{"network": "eip155:80002"}})

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "exceeds the configured cap")
		})
	})
}
