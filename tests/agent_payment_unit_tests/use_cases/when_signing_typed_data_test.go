package use_cases_test

import (
	"context"
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	identity_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type WhenSigningTypedDataTestingSuite struct {
	sut *use_cases.SignTypedData

	mockTypedDataSigner                 *identity_mocks.MockTypedDataSignerInterface
	mockTypedDataSpendQuoter            *interfaces_mocks.MockTypedDataSpendQuoterInterface
	mockX402PaymentSpendPolicyEvaluator *interfaces_mocks.MockX402PaymentSpendPolicyEvaluatorInterface
}

func WhenSigningTypedDataBeforeEach(t *testing.T) *WhenSigningTypedDataTestingSuite {
	mockController := gomock.NewController(t)

	mockTypedDataSigner := identity_mocks.NewMockTypedDataSignerInterface(mockController)
	mockTypedDataSpendQuoter := interfaces_mocks.NewMockTypedDataSpendQuoterInterface(mockController)
	mockX402PaymentSpendPolicyEvaluator := interfaces_mocks.NewMockX402PaymentSpendPolicyEvaluatorInterface(
		mockController,
	)

	return &WhenSigningTypedDataTestingSuite{
		sut: use_cases.NewSignTypedData(
			mockTypedDataSigner,
			mockTypedDataSpendQuoter,
			mockX402PaymentSpendPolicyEvaluator,
			zap.NewNop(),
		),
		mockTypedDataSigner:                 mockTypedDataSigner,
		mockTypedDataSpendQuoter:            mockTypedDataSpendQuoter,
		mockX402PaymentSpendPolicyEvaluator: mockX402PaymentSpendPolicyEvaluator,
	}
}

func TestWhenSigningTypedData(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	typedData := &apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain":              {{Name: "name", Type: "string"}},
			"TransferWithAuthorization": {{Name: "to", Type: "address"}},
		},
		PrimaryType: "TransferWithAuthorization",
		Domain:      apitypes.TypedDataDomain{Name: "Test"},
		Message:     apitypes.TypedDataMessage{"to": common_tests.RandomEthAddress().Hex()},
	}
	quote := &payment_interfaces.X402PaymentQuote{
		Network:             "eip155:8453",
		Asset:               common_tests.RandomEthAddress().Hex(),
		PayTo:               common_tests.RandomEthAddress().Hex(),
		Amount:              "10000",
		AssetTransferMethod: enums.AssetTransferMethodEip3009,
	}
	amount := big.NewInt(10000)
	walletAddress := common_tests.RandomEthAddress()
	reservation := &payment_interfaces.X402PaymentSpendReservation{
		RuleId:       uuid.New(),
		Amount:       amount,
		WindowStarts: map[uuid.UUID]time.Time{uuid.New(): time.Now().Truncate(time.Second)},
	}

	t.Run("Given the request is invalid", func(t *testing.T) {
		t.Parallel()

		request := &requests.SignTypedDataRequest{}

		t.Run("Should return a validation error without touching dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningTypedDataBeforeEach(t)

			response, err := suite.sut.Execute(ctx, request)

			assert.Nil(t, response)
			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "typedData", validationErr.Field)
		})
	})

	t.Run("Given the payload cannot be quoted", func(t *testing.T) {
		t.Parallel()

		request := &requests.SignTypedDataRequest{TypedData: typedData}
		initSuite := func(suite *WhenSigningTypedDataTestingSuite) {
			suite.mockTypedDataSpendQuoter.EXPECT().
				Quote(typedData).
				Return(nil, nil, apperrors.NewValidationError("typedData.primaryType", "batch")).
				Times(1)
		}

		// No signer expectation: a payload the policy cannot bound is never signed.
		t.Run("Should return the validation error without signing", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningTypedDataBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(ctx, request)

			assert.Nil(t, response)
			var validationErr *apperrors.ValidationError
			if assert.ErrorAs(t, err, &validationErr) {
				assert.Equal(t, "typedData.primaryType", validationErr.Field)
			}
		})
	})

	t.Run("Given the wallet address is unavailable", func(t *testing.T) {
		t.Parallel()

		request := &requests.SignTypedDataRequest{TypedData: typedData}
		initSuite := func(suite *WhenSigningTypedDataTestingSuite) {
			suite.mockTypedDataSpendQuoter.EXPECT().Quote(typedData).Return(quote, amount, nil).Times(1)
			suite.mockTypedDataSigner.EXPECT().
				WalletAddress(ctx).
				Return(common.Address{}, errors.New("node wallet not bootstrapped")).
				Times(1)
		}

		// No Authorize expectation: the address is read first, so no spend is left to give back.
		t.Run("Should return an InternalError without recording a spend", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningTypedDataBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(ctx, request)

			assert.Nil(t, response)
			var internalErr *apperrors.InternalError
			if assert.ErrorAs(t, err, &internalErr) {
				assert.Equal(t, "wallet_address_unavailable", internalErr.Code)
			}
		})
	})

	t.Run("Given authorizing the spend fails", func(t *testing.T) {
		t.Parallel()

		request := &requests.SignTypedDataRequest{TypedData: typedData}
		initSuite := func(suite *WhenSigningTypedDataTestingSuite) {
			suite.mockTypedDataSpendQuoter.EXPECT().Quote(typedData).Return(quote, amount, nil).Times(1)
			suite.mockTypedDataSigner.EXPECT().WalletAddress(ctx).Return(walletAddress, nil).Times(1)
			suite.mockX402PaymentSpendPolicyEvaluator.EXPECT().
				Authorize(ctx, quote, amount).
				Return(nil, apperrors.NewInternalError("spend_policy_failed", "")).
				Times(1)
		}

		t.Run("Should return the error without signing", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningTypedDataBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(ctx, request)

			assert.Nil(t, response)
			var internalErr *apperrors.InternalError
			if assert.ErrorAs(t, err, &internalErr) {
				assert.Equal(t, "spend_policy_failed", internalErr.Code)
			}
		})
	})

	t.Run("Given the spend policy refuses the payload", func(t *testing.T) {
		t.Parallel()

		request := &requests.SignTypedDataRequest{TypedData: typedData}
		initSuite := func(suite *WhenSigningTypedDataTestingSuite) {
			suite.mockTypedDataSpendQuoter.EXPECT().Quote(typedData).Return(quote, amount, nil).Times(1)
			suite.mockTypedDataSigner.EXPECT().WalletAddress(ctx).Return(walletAddress, nil).Times(1)
			suite.mockX402PaymentSpendPolicyEvaluator.EXPECT().
				Authorize(ctx, quote, amount).
				Return(nil, apperrors.NewUnauthorizedError("spend_cap_exceeded", "")).
				Times(1)
		}

		t.Run("Should return the refusal without signing", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningTypedDataBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(ctx, request)

			assert.Nil(t, response)
			var unauthorizedErr *apperrors.UnauthorizedError
			if assert.ErrorAs(t, err, &unauthorizedErr) {
				assert.Equal(t, "spend_cap_exceeded", unauthorizedErr.Code)
			}
		})
	})

	t.Run("Given the signer fails after the spend was recorded", func(t *testing.T) {
		t.Parallel()

		request := &requests.SignTypedDataRequest{TypedData: typedData}
		initSuite := func(suite *WhenSigningTypedDataTestingSuite) {
			suite.mockTypedDataSpendQuoter.EXPECT().Quote(typedData).Return(quote, amount, nil).Times(1)
			suite.mockTypedDataSigner.EXPECT().WalletAddress(ctx).Return(walletAddress, nil).Times(1)
			suite.mockX402PaymentSpendPolicyEvaluator.EXPECT().
				Authorize(ctx, quote, amount).
				Return(reservation, nil).
				Times(1)
			suite.mockTypedDataSigner.EXPECT().
				SignTypedData(ctx, typedData).
				Return("", errors.New("node wallet not bootstrapped")).
				Times(1)
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningTypedDataBeforeEach(t)
			initSuite(suite)
			suite.mockX402PaymentSpendPolicyEvaluator.EXPECT().
				Release(gomock.Any(), reservation).
				Return(nil).
				Times(1)

			response, err := suite.sut.Execute(ctx, request)

			assert.Nil(t, response)
			var internalErr *apperrors.InternalError
			if assert.ErrorAs(t, err, &internalErr) {
				assert.Equal(t, "typed_data_signing_failed", internalErr.Code)
			}
		})

		t.Run("Should give the spend back", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningTypedDataBeforeEach(t)
			initSuite(suite)
			suite.mockX402PaymentSpendPolicyEvaluator.EXPECT().
				Release(gomock.Any(), reservation).
				Return(nil).
				Times(1)

			_, err := suite.sut.Execute(ctx, request)

			var internalErr *apperrors.InternalError
			if assert.ErrorAs(t, err, &internalErr) {
				assert.Equal(t, "typed_data_signing_failed", internalErr.Code)
			}
		})
	})

	t.Run("Given giving the spend back fails too", func(t *testing.T) {
		t.Parallel()

		request := &requests.SignTypedDataRequest{TypedData: typedData}
		initSuite := func(suite *WhenSigningTypedDataTestingSuite) {
			suite.mockTypedDataSpendQuoter.EXPECT().Quote(typedData).Return(quote, amount, nil).Times(1)
			suite.mockTypedDataSigner.EXPECT().WalletAddress(ctx).Return(walletAddress, nil).Times(1)
			suite.mockX402PaymentSpendPolicyEvaluator.EXPECT().
				Authorize(ctx, quote, amount).
				Return(reservation, nil).
				Times(1)
			suite.mockTypedDataSigner.EXPECT().
				SignTypedData(ctx, typedData).
				Return("", errors.New("node wallet not bootstrapped")).
				Times(1)
			suite.mockX402PaymentSpendPolicyEvaluator.EXPECT().
				Release(gomock.Any(), reservation).
				Return(apperrors.NewInternalError("spend_policy_failed", "")).
				Times(1)
		}

		// The signing failure is what the caller acts on; the failed refund is only logged.
		t.Run("Should return the signing error", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningTypedDataBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(ctx, request)

			assert.Nil(t, response)
			var internalErr *apperrors.InternalError
			if assert.ErrorAs(t, err, &internalErr) {
				assert.Equal(t, "typed_data_signing_failed", internalErr.Code)
			}
		})
	})

	t.Run("Given signing fails for a payload that moves no funds", func(t *testing.T) {
		t.Parallel()

		request := &requests.SignTypedDataRequest{TypedData: typedData}
		initSuite := func(suite *WhenSigningTypedDataTestingSuite) {
			suite.mockTypedDataSpendQuoter.EXPECT().Quote(typedData).Return(nil, nil, nil).Times(1)
			suite.mockTypedDataSigner.EXPECT().WalletAddress(ctx).Return(walletAddress, nil).Times(1)
			suite.mockTypedDataSigner.EXPECT().
				SignTypedData(ctx, typedData).
				Return("", errors.New("node wallet not bootstrapped")).
				Times(1)
		}

		// No Release expectation: nothing was charged, so there is nothing to give back.
		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningTypedDataBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(ctx, request)

			assert.Nil(t, response)
			var internalErr *apperrors.InternalError
			if assert.ErrorAs(t, err, &internalErr) {
				assert.Equal(t, "typed_data_signing_failed", internalErr.Code)
			}
		})
	})

	t.Run("Given a payload that moves no funds", func(t *testing.T) {
		t.Parallel()

		request := &requests.SignTypedDataRequest{TypedData: typedData}
		address := common_tests.RandomEthAddress()
		initSuite := func(suite *WhenSigningTypedDataTestingSuite) {
			suite.mockTypedDataSpendQuoter.EXPECT().Quote(typedData).Return(nil, nil, nil).Times(1)
			suite.mockTypedDataSigner.EXPECT().SignTypedData(ctx, typedData).Return("0xdead", nil).Times(1)
			suite.mockTypedDataSigner.EXPECT().WalletAddress(ctx).Return(address, nil).Times(1)
		}

		// No Authorize expectation: only token payloads are charged against the policy.
		t.Run("Should sign without consulting the spend policy", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningTypedDataBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(ctx, request)

			assert.NoError(t, err)
			if assert.NotNil(t, response) {
				assert.Equal(t, "0xdead", response.Signature)
			}
		})
	})

	t.Run("Given a token payload the policy covers", func(t *testing.T) {
		t.Parallel()

		request := &requests.SignTypedDataRequest{TypedData: typedData}
		address := common_tests.RandomEthAddress()
		initSuite := func(suite *WhenSigningTypedDataTestingSuite) {
			suite.mockTypedDataSpendQuoter.EXPECT().Quote(typedData).Return(quote, amount, nil).Times(1)
			suite.mockX402PaymentSpendPolicyEvaluator.EXPECT().
				Authorize(ctx, quote, amount).
				Return(reservation, nil).
				Times(1)
			suite.mockTypedDataSigner.EXPECT().SignTypedData(ctx, typedData).Return("0xdead", nil).Times(1)
			suite.mockTypedDataSigner.EXPECT().WalletAddress(ctx).Return(address, nil).Times(1)
		}

		t.Run("Should return the signature and the signing address", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningTypedDataBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(ctx, request)

			assert.NoError(t, err)
			if assert.NotNil(t, response) {
				assert.Equal(t, "0xdead", response.Signature)
				assert.Equal(t, address.Hex(), response.Address)
			}
		})
	})
}
