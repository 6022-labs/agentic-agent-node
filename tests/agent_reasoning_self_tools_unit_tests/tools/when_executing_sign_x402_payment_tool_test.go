package tools_test

import (
	"context"
	"errors"
	"testing"

	"github.com/6022-labs/agentic-agent-node/tests/common_tests"

	payment_requests "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	payment_responses "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
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
	asset := common_tests.RandomEthAddress().Hex()
	payTo := common_tests.RandomEthAddress().Hex()

	t.Run("Given a quote entry forwarded verbatim from a 402", func(t *testing.T) {
		t.Parallel()

		t.Run("Should map it onto the use case request, extra included", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingSignX402PaymentToolBeforeEach(t)
			suite.mockSignX402PaymentExecutor.EXPECT().
				Execute(ctx, &payment_requests.SignX402PaymentRequest{
					Network: "eip155:80002",
					Asset:   asset,
					PayTo:   payTo,
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
					"asset":   asset,
					"payTo":   payTo,
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

	t.Run("Given an accept entry that does not decode", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError on accept without calling the use case", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingSignX402PaymentToolBeforeEach(t)

			out, err := suite.sut.Execute(ctx, map[string]any{"accept": map[string]any{"amount": 1000}})

			assert.Nil(t, out)
			var validationErr *apperrors.ValidationError
			if assert.ErrorAs(t, err, &validationErr) {
				assert.Equal(t, "accept", validationErr.Field)
			}
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
