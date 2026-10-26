package use_cases_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/x402"
	payment_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenAuthorizingPaymentReceiveAccessTestingSuite struct {
	sut *use_cases.AuthorizePaymentReceiveAccess

	mockEvaluator *payment_interfaces_mocks.MockX402PaymentReceivePolicyEvaluatorInterface
}

func WhenAuthorizingPaymentReceiveAccessBeforeEach(
	t *testing.T,
) *WhenAuthorizingPaymentReceiveAccessTestingSuite {
	mockController := gomock.NewController(t)

	mockEvaluator := payment_interfaces_mocks.NewMockX402PaymentReceivePolicyEvaluatorInterface(mockController)

	return &WhenAuthorizingPaymentReceiveAccessTestingSuite{
		sut:           use_cases.NewAuthorizePaymentReceiveAccess(mockEvaluator),
		mockEvaluator: mockEvaluator,
	}
}

func TestWhenAuthorizingPaymentReceiveAccess(t *testing.T) {
	t.Parallel()

	t.Run("Given the address is not an address", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error without evaluating", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingPaymentReceiveAccessBeforeEach(t)

			response, err := suite.sut.Execute(context.Background(), &requests.AuthorizePaymentReceiveAccessRequest{
				Address: "not-an-address",
			})

			assert.Nil(t, response)
			var validationError *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationError)
			assert.Equal(t, "address", validationError.Field)
		})
	})

	t.Run("Given the evaluator fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAuthorizingPaymentReceiveAccessTestingSuite) {
			suite.mockEvaluator.EXPECT().
				Evaluate(gomock.Any(), gomock.Any()).
				Return(nil, apperrors.NewInternalError("access_evaluation_failed", "")).
				Times(1)
		}

		t.Run("Should propagate the error rather than refuse", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingPaymentReceiveAccessBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), &requests.AuthorizePaymentReceiveAccessRequest{})

			assert.Nil(t, response)
			var internalError *apperrors.InternalError
			assert.ErrorAs(t, err, &internalError)
		})
	})

	t.Run("Given the evaluator refuses", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAuthorizingPaymentReceiveAccessTestingSuite) {
			suite.mockEvaluator.EXPECT().
				Evaluate(gomock.Any(), gomock.Any()).
				Return(&payment_interfaces.PaymentReceiveAccessDecision{
					Outcome:   enums.PaymentReceiveAccessOutcomePaymentRequired,
					ErrorCode: "insufficient_amount",
				}, nil).
				Times(1)
		}

		t.Run("Should carry no settlement", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingPaymentReceiveAccessBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), &requests.AuthorizePaymentReceiveAccessRequest{})

			assert.NoError(t, err)
			assert.Equal(t, enums.PaymentReceiveAccessOutcomePaymentRequired, response.Outcome)
			assert.Equal(t, "insufficient_amount", response.ErrorCode)
			assert.Nil(t, response.Settlement)
		})
	})

	t.Run("Given the evaluator authorizes a payable request", func(t *testing.T) {
		t.Parallel()

		payer := common_tests.RandomEthAddress().Hex()

		initSuite := func(suite *WhenAuthorizingPaymentReceiveAccessTestingSuite) {
			// The credentials the evaluator receives are the use case's only real work.
			suite.mockEvaluator.EXPECT().
				Evaluate(gomock.Any(), &payment_interfaces.PaymentReceiveAccessCredentials{
					Address:          payer,
					PaymentSignature: "a-header",
				}).
				Return(&payment_interfaces.PaymentReceiveAccessDecision{
					Outcome: enums.PaymentReceiveAccessOutcomePayable,
					Payer:   payer,
					Accepts: []x402.PaymentRequirements{{Network: "eip155:80002", Amount: "10000"}},
					Settlement: &payment_interfaces.PaymentReceiveSettlement{
						Requirements:       x402.PaymentRequirements{Amount: "10000"},
						FacilitatorBaseUrl: "https://facilitator.test",
						Network:            "eip155:80002",
					},
				}, nil).
				Times(1)
		}

		t.Run("Should map the settlement the gateway needs", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingPaymentReceiveAccessBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), &requests.AuthorizePaymentReceiveAccessRequest{
				Address:          "  " + payer + "  ",
				PaymentSignature: "  a-header  ",
			})

			assert.NoError(t, err)
			assert.Equal(t, enums.PaymentReceiveAccessOutcomePayable, response.Outcome)
			assert.NotNil(t, response.Settlement)
			assert.Equal(t, "https://facilitator.test", response.Settlement.FacilitatorBaseUrl)
			assert.Equal(t, "eip155:80002", response.Settlement.Network)
			// Payer becomes the caller's identity downstream, so it must survive the mapping.
			assert.Equal(t, payer, response.Payer)
			assert.Len(t, response.Accepts, 1)
		})
	})
}
