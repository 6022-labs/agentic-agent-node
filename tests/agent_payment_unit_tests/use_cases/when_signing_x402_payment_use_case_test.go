package use_cases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenSigningX402PaymentUseCaseTestingSuite struct {
	sut *use_cases.SignX402Payment

	mockX402PaymentSigner *interfaces_mocks.MockX402PaymentSignerInterface
}

func WhenSigningX402PaymentUseCaseBeforeEach(t *testing.T) *WhenSigningX402PaymentUseCaseTestingSuite {
	mockController := gomock.NewController(t)

	mockX402PaymentSigner := interfaces_mocks.NewMockX402PaymentSignerInterface(mockController)

	return &WhenSigningX402PaymentUseCaseTestingSuite{
		sut:                   use_cases.NewSignX402Payment(mockX402PaymentSigner),
		mockX402PaymentSigner: mockX402PaymentSigner,
	}
}

func TestWhenSigningX402PaymentUseCase(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	network := "eip155:80002"
	asset := common_tests.RandomEthAddress().Hex()
	payTo := common_tests.RandomEthAddress().Hex()

	signedRequest := func() *requests.SignX402PaymentRequest {
		return &requests.SignX402PaymentRequest{
			Network: network,
			Asset:   asset,
			PayTo:   payTo,
			Amount:  "10000",
			Extra:   map[string]any{"assetTransferMethod": "eip3009", "name": "USDC", "version": "2"},
		}
	}

	t.Run("Given the request is invalid", func(t *testing.T) {
		t.Parallel()

		request := signedRequest()
		request.Amount = "-1"

		t.Run("Should return a validation error without touching dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningX402PaymentUseCaseBeforeEach(t)

			response, err := suite.sut.Execute(ctx, request)

			assert.Nil(t, response)
			var validationError *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationError)
			assert.Equal(t, "amount", validationError.Field)
		})
	})

	t.Run("Given the spend policy refuses the quote", func(t *testing.T) {
		t.Parallel()

		request := signedRequest()
		initSuite := func(suite *WhenSigningX402PaymentUseCaseTestingSuite) {
			suite.mockX402PaymentSigner.EXPECT().
				Sign(ctx, gomock.Any()).
				Return(nil, apperrors.NewUnauthorizedError("spend_payee_not_allowed", "")).
				Times(1)
		}

		// A refusal is the caller's answer, not an outage that invites a retry.
		t.Run("Should return the UnauthorizedError naming why", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningX402PaymentUseCaseBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(ctx, request)

			assert.Nil(t, response)
			var unauthorizedError *apperrors.UnauthorizedError
			assert.ErrorAs(t, err, &unauthorizedError)
			assert.Equal(t, "spend_payee_not_allowed", unauthorizedError.Code)
		})
	})

	t.Run("Given the signer rejects the quote", func(t *testing.T) {
		t.Parallel()

		request := signedRequest()
		initSuite := func(suite *WhenSigningX402PaymentUseCaseTestingSuite) {
			suite.mockX402PaymentSigner.EXPECT().
				Sign(ctx, gomock.Any()).
				Return(nil, apperrors.NewValidationError("assetTransferMethod", "must be one of eip3009, permit2")).
				Times(1)
		}

		t.Run("Should return the ValidationError", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningX402PaymentUseCaseBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(ctx, request)

			assert.Nil(t, response)
			var validationError *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationError)
			assert.Equal(t, "assetTransferMethod", validationError.Field)
		})
	})

	t.Run("Given evaluating the spend policy fails", func(t *testing.T) {
		t.Parallel()

		request := signedRequest()
		initSuite := func(suite *WhenSigningX402PaymentUseCaseTestingSuite) {
			suite.mockX402PaymentSigner.EXPECT().
				Sign(ctx, gomock.Any()).
				Return(nil, apperrors.NewInternalError("spend_policy_failed", "")).
				Times(1)
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningX402PaymentUseCaseBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(ctx, request)

			assert.Nil(t, response)
			var internalError *apperrors.InternalError
			assert.ErrorAs(t, err, &internalError)
			assert.Equal(t, "spend_policy_failed", internalError.Code)
		})
	})

	t.Run("Given the signer returns a conflict", func(t *testing.T) {
		t.Parallel()

		request := signedRequest()
		initSuite := func(suite *WhenSigningX402PaymentUseCaseTestingSuite) {
			suite.mockX402PaymentSigner.EXPECT().
				Sign(ctx, gomock.Any()).
				Return(nil, apperrors.NewConflictError("agent_not_minted", "")).
				Times(1)
		}

		t.Run("Should pass the ConflictError through", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningX402PaymentUseCaseBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(ctx, request)

			assert.Nil(t, response)
			var conflictError *apperrors.ConflictError
			if assert.ErrorAs(t, err, &conflictError) {
				assert.Equal(t, "agent_not_minted", conflictError.Code)
			}
		})
	})

	t.Run("Given the signer returns a not found error", func(t *testing.T) {
		t.Parallel()

		request := signedRequest()
		initSuite := func(suite *WhenSigningX402PaymentUseCaseTestingSuite) {
			suite.mockX402PaymentSigner.EXPECT().
				Sign(ctx, gomock.Any()).
				Return(nil, apperrors.NewNotFoundError("wallet_not_found", "")).
				Times(1)
		}

		t.Run("Should pass the NotFoundError through", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningX402PaymentUseCaseBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(ctx, request)

			assert.Nil(t, response)
			var notFoundError *apperrors.NotFoundError
			if assert.ErrorAs(t, err, &notFoundError) {
				assert.Equal(t, "wallet_not_found", notFoundError.Code)
			}
		})
	})

	t.Run("Given the signer fails", func(t *testing.T) {
		t.Parallel()

		request := signedRequest()
		initSuite := func(suite *WhenSigningX402PaymentUseCaseTestingSuite) {
			suite.mockX402PaymentSigner.EXPECT().
				Sign(ctx, gomock.Any()).
				Return(nil, errors.New("node wallet not bootstrapped")).
				Times(1)
		}

		t.Run("Should return an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningX402PaymentUseCaseBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(ctx, request)

			assert.Nil(t, response)
			var unavailableError *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableError)
			assert.Equal(t, "x402_payment_signing_failed", unavailableError.Code)
		})
	})

	t.Run("Given a quote carrying its domain under extra", func(t *testing.T) {
		t.Parallel()

		request := signedRequest()
		initSuite := func(suite *WhenSigningX402PaymentUseCaseTestingSuite) {
			suite.mockX402PaymentSigner.EXPECT().
				Sign(ctx, &payment_interfaces.X402PaymentQuote{
					Network:             network,
					Asset:               asset,
					PayTo:               payTo,
					Amount:              "10000",
					AssetTransferMethod: enums.AssetTransferMethodEip3009,
					Eip712Name:          "USDC",
					Eip712Version:       "2",
				}).
				Return(&payment_interfaces.X402SignedPayment{PaymentSignature: "eyJ4NDAy"}, nil).
				Times(1)
		}

		t.Run("Should sign the quote with its transfer method and domain", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningX402PaymentUseCaseBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(ctx, request)

			assert.NoError(t, err)
			assert.Equal(t, "eyJ4NDAy", response.PaymentSignature)
		})
	})
}
