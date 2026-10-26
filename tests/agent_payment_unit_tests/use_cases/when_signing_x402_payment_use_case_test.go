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
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

const (
	quoteNetwork = "eip155:80002"
	quoteAsset   = "0x41E94Eb019C0762f9Bfcf9Fb1E58725BfB0e7582"
	quotePayTo   = "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
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

	t.Run("Given a quote carrying its domain under extra", func(t *testing.T) {
		t.Parallel()

		t.Run("Should read the transfer method and eip-712 domain out of it", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningX402PaymentUseCaseBeforeEach(t)
			suite.mockX402PaymentSigner.EXPECT().
				Sign(ctx, &payment_interfaces.X402PaymentQuote{
					Network:             quoteNetwork,
					Asset:               quoteAsset,
					PayTo:               quotePayTo,
					Amount:              "10000",
					AssetTransferMethod: enums.AssetTransferMethodEip3009,
					Eip712Name:          "USDC",
					Eip712Version:       "2",
				}).
				Return(&payment_interfaces.X402SignedPayment{PaymentSignature: "eyJ4NDAy"}, nil)

			response, err := suite.sut.Execute(ctx, &requests.SignX402PaymentRequest{
				Network: quoteNetwork,
				Asset:   quoteAsset,
				PayTo:   quotePayTo,
				Amount:  "10000",
				Extra: map[string]any{
					"assetTransferMethod": "eip3009",
					"name":                "USDC",
					"version":             "2",
				},
			})

			assert.NoError(t, err)
			assert.Equal(t, "eyJ4NDAy", response.PaymentSignature)
		})
	})

	t.Run("Given a permit2 quote", func(t *testing.T) {
		t.Parallel()

		t.Run("Should not require an eip-712 domain", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningX402PaymentUseCaseBeforeEach(t)
			suite.mockX402PaymentSigner.EXPECT().
				Sign(ctx, gomock.Any()).
				Return(&payment_interfaces.X402SignedPayment{}, nil)

			_, err := suite.sut.Execute(ctx, &requests.SignX402PaymentRequest{
				Network: quoteNetwork,
				Asset:   quoteAsset,
				PayTo:   quotePayTo,
				Amount:  "0",
				Extra:   map[string]any{"assetTransferMethod": "permit2"},
			})

			assert.NoError(t, err)
		})
	})

	t.Run("Given an eip3009 quote that omits its domain", func(t *testing.T) {
		t.Parallel()

		// A guessed domain signs a digest that recovers to another address, so the payment could never verify.
		t.Run("Should refuse rather than guess one", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningX402PaymentUseCaseBeforeEach(t)

			response, err := suite.sut.Execute(ctx, &requests.SignX402PaymentRequest{
				Network:             quoteNetwork,
				Asset:               quoteAsset,
				PayTo:               quotePayTo,
				Amount:              "10000",
				AssetTransferMethod: "eip3009",
			})

			assert.Nil(t, response)
			var validationError *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationError)
			assert.Equal(t, "eip712Name", validationError.Field)
		})

		t.Run("Should refuse when only the version is missing", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningX402PaymentUseCaseBeforeEach(t)

			response, err := suite.sut.Execute(ctx, &requests.SignX402PaymentRequest{
				Network:             quoteNetwork,
				Asset:               quoteAsset,
				PayTo:               quotePayTo,
				Amount:              "10000",
				AssetTransferMethod: "eip3009",
				Eip712Name:          "USDC",
			})

			assert.Nil(t, response)
			assert.ErrorContains(t, err, "eip712Name")
		})
	})

	t.Run("Given a standard x402 quote, which names no transfer method", func(t *testing.T) {
		t.Parallel()

		// assetTransferMethod is a 6022 addition.
		t.Run("Should infer eip3009 from the domain it does carry", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningX402PaymentUseCaseBeforeEach(t)
			suite.mockX402PaymentSigner.EXPECT().
				Sign(ctx, &payment_interfaces.X402PaymentQuote{
					Network:             quoteNetwork,
					Asset:               quoteAsset,
					PayTo:               quotePayTo,
					Amount:              "10000",
					AssetTransferMethod: enums.AssetTransferMethodEip3009,
					Eip712Name:          "USDC",
					Eip712Version:       "2",
				}).
				Return(&payment_interfaces.X402SignedPayment{}, nil)

			_, err := suite.sut.Execute(ctx, &requests.SignX402PaymentRequest{
				Network: quoteNetwork,
				Asset:   quoteAsset,
				PayTo:   quotePayTo,
				Amount:  "10000",
				Extra:   map[string]any{"name": "USDC", "version": "2"},
			})

			assert.NoError(t, err)
		})

		t.Run("Should refuse when it carries no domain either", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningX402PaymentUseCaseBeforeEach(t)

			response, err := suite.sut.Execute(ctx, &requests.SignX402PaymentRequest{
				Network: quoteNetwork,
				Asset:   quoteAsset,
				PayTo:   quotePayTo,
				Amount:  "10000",
			})

			assert.Nil(t, response)
			assert.ErrorContains(t, err, "eip712Name")
		})
	})

	t.Run("Given a malformed quote", func(t *testing.T) {
		t.Parallel()

		malformedCases := []struct {
			name    string
			request *requests.SignX402PaymentRequest
			field   string
		}{
			{
				name: "a non-caip2 network",
				request: &requests.SignX402PaymentRequest{
					Network: "80002",
					Asset:   quoteAsset,
					PayTo:   quotePayTo,
					Amount:  "1",
				},
				field: "network",
			},
			{
				name: "a non-address asset",
				request: &requests.SignX402PaymentRequest{
					Network: quoteNetwork,
					Asset:   "usdc",
					PayTo:   quotePayTo,
					Amount:  "1",
				},
				field: "asset",
			},
			{
				name: "a negative amount",
				request: &requests.SignX402PaymentRequest{
					Network: quoteNetwork,
					Asset:   quoteAsset,
					PayTo:   quotePayTo,
					Amount:  "-1",
				},
				field: "amount",
			},
			{
				name: "an unknown transfer method",
				request: &requests.SignX402PaymentRequest{
					Network: quoteNetwork, Asset: quoteAsset, PayTo: quotePayTo, Amount: "1",
					AssetTransferMethod: "swap",
				},
				field: "assetTransferMethod",
			},
		}

		for _, malformedCase := range malformedCases {
			t.Run("Should refuse "+malformedCase.name, func(t *testing.T) {
				t.Parallel()

				suite := WhenSigningX402PaymentUseCaseBeforeEach(t)

				response, err := suite.sut.Execute(ctx, malformedCase.request)

				assert.Nil(t, response)
				var validationError *apperrors.ValidationError
				assert.ErrorAs(t, err, &validationError)
				assert.Equal(t, malformedCase.field, validationError.Field)
			})
		}
	})

	t.Run("Given the signer refuses the quote", func(t *testing.T) {
		t.Parallel()

		// A refusal is the caller's problem, not an outage: 502 would claim an upstream broke and invite a pointless retry.
		t.Run("Should keep it a validation error rather than an outage", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningX402PaymentUseCaseBeforeEach(t)
			suite.mockX402PaymentSigner.EXPECT().
				Sign(ctx, gomock.Any()).
				Return(nil, apperrors.NewValidationError("assetTransferMethod", "must be one of eip3009, permit2"))

			response, err := suite.sut.Execute(ctx, &requests.SignX402PaymentRequest{
				Network: quoteNetwork,
				Asset:   quoteAsset,
				PayTo:   quotePayTo,
				Amount:  "1",
				Extra:   map[string]any{"assetTransferMethod": "permit2"},
			})

			assert.Nil(t, response)
			var validationError *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationError)
			var unavailableError *apperrors.UnavailableError
			assert.False(t, errors.As(err, &unavailableError))
		})
	})

	t.Run("Given the signer fails", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail as unavailable", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningX402PaymentUseCaseBeforeEach(t)
			suite.mockX402PaymentSigner.EXPECT().
				Sign(ctx, gomock.Any()).
				Return(nil, errors.New("node wallet not bootstrapped"))

			response, err := suite.sut.Execute(ctx, &requests.SignX402PaymentRequest{
				Network: quoteNetwork,
				Asset:   quoteAsset,
				PayTo:   quotePayTo,
				Amount:  "1",
				Extra:   map[string]any{"assetTransferMethod": "permit2"},
			})

			assert.Nil(t, response)
			var unavailableError *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableError)
		})
	})
}
