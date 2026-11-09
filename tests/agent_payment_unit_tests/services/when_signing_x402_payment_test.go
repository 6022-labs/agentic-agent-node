package services_test

import (
	"context"
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/services"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	payment_settings "github.com/6022-labs/agentic-agent-node/src/agent_payment/settings"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/x402"
	identity_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

const signingNetwork = "eip155:80002"

var (
	signingAsset = common_tests.RandomEthAddress().Hex()
	signingPayTo = common_tests.RandomEthAddress().Hex()
)

type WhenSigningX402PaymentTestingSuite struct {
	sut *services.X402PaymentSigner

	mockTypedDataSigner        *identity_mocks.MockTypedDataSignerInterface
	mockPermit2ApprovalEnsurer *interfaces_mocks.MockPermit2ApprovalEnsurerInterface
	mockSpendPolicyEvaluator   *interfaces_mocks.MockX402PaymentSpendPolicyEvaluatorInterface
	walletKey                  *ecdsa.PrivateKey
	walletAddress              common.Address
}

func WhenSigningX402PaymentBeforeEach(t *testing.T) *WhenSigningX402PaymentTestingSuite {
	mockController := gomock.NewController(t)

	walletKey, err := crypto.GenerateKey()
	assert.NoError(t, err)
	walletAddress := crypto.PubkeyToAddress(walletKey.PublicKey)

	mockTypedDataSigner := identity_mocks.NewMockTypedDataSignerInterface(mockController)
	mockPermit2ApprovalEnsurer := interfaces_mocks.NewMockPermit2ApprovalEnsurerInterface(mockController)
	mockSpendPolicyEvaluator := interfaces_mocks.NewMockX402PaymentSpendPolicyEvaluatorInterface(mockController)

	suite := &WhenSigningX402PaymentTestingSuite{
		sut: services.NewX402PaymentSigner(
			mockTypedDataSigner,
			mockPermit2ApprovalEnsurer,
			mockSpendPolicyEvaluator,
			newSpendSettings(),
			zap.NewNop(),
		),
		mockTypedDataSigner:        mockTypedDataSigner,
		mockPermit2ApprovalEnsurer: mockPermit2ApprovalEnsurer,
		mockSpendPolicyEvaluator:   mockSpendPolicyEvaluator,
		walletKey:                  walletKey,
		walletAddress:              walletAddress,
	}

	return suite
}

// expectWalletSigning lets the wallet sign the real digest, so the gateway verifier can be pointed at the result.
func (s *WhenSigningX402PaymentTestingSuite) expectWalletSigning() {
	s.mockTypedDataSigner.EXPECT().WalletAddress(gomock.Any()).Return(s.walletAddress, nil).AnyTimes()
	s.mockTypedDataSigner.EXPECT().
		SignDigest(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, digest []byte) ([]byte, error) {
			signature, err := crypto.Sign(digest, s.walletKey)
			if err != nil {
				return nil, err
			}
			signature[crypto.RecoveryIDOffset] += 27
			return signature, nil
		}).
		AnyTimes()
}

func signingReservation() *payment_interfaces.X402PaymentSpendReservation {
	return &payment_interfaces.X402PaymentSpendReservation{
		RuleId:       uuid.New(),
		Amount:       big.NewInt(10000),
		WindowStarts: map[uuid.UUID]time.Time{uuid.New(): time.Now().Truncate(time.Second)},
	}
}

// newSpendSettings builds the settings through koanf the way the app parses them.
func newSpendSettings() *payment_settings.PaymentSpendSettings {
	k := koanf.New(".")
	_ = k.Load(confmap.Provider(map[string]any{
		"payment": map[string]any{
			"spend": map[string]any{"validity_seconds": 300, "valid_after_skew_seconds": 60},
		},
	}, "."), nil)
	return payment_settings.NewPaymentSpendSettings(zap.NewNop(), k)
}

func (s *WhenSigningX402PaymentTestingSuite) decodePaymentPayload(t *testing.T, header string) *x402.PaymentPayload {
	t.Helper()

	decoded, err := base64.StdEncoding.DecodeString(header)
	assert.NoError(t, err)

	var payload x402.PaymentPayload
	assert.NoError(t, json.Unmarshal(decoded, &payload))
	return &payload
}

func TestWhenSigningX402Payment(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given the network is not eip155", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on network", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningX402PaymentBeforeEach(t)
			quote := eip3009SigningQuote()
			quote.Network = "base"

			signed, err := suite.sut.Sign(ctx, quote)

			assert.Nil(t, signed)
			var validationError *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationError)
			assert.Equal(t, "network", validationError.Field)
		})
	})

	t.Run("Given the amount is negative", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on amount", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningX402PaymentBeforeEach(t)
			quote := eip3009SigningQuote()
			quote.Amount = "-1"

			signed, err := suite.sut.Sign(ctx, quote)

			assert.Nil(t, signed)
			var validationError *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationError)
			assert.Equal(t, "amount", validationError.Field)
		})
	})

	t.Run("Given the amount is not a number", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on amount", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningX402PaymentBeforeEach(t)
			quote := eip3009SigningQuote()
			quote.Amount = "ten"

			signed, err := suite.sut.Sign(ctx, quote)

			assert.Nil(t, signed)
			var validationError *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationError)
			assert.Equal(t, "amount", validationError.Field)
		})
	})

	t.Run("Given an unsupported transfer method", func(t *testing.T) {
		t.Parallel()

		// No spend expectation: a malformed quote must be rejected before the gate, or it burns budget.
		t.Run("Should refuse to sign without consulting the spend policy", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningX402PaymentBeforeEach(t)

			signed, err := suite.sut.Sign(ctx, &payment_interfaces.X402PaymentQuote{
				Network:             signingNetwork,
				Asset:               signingAsset,
				PayTo:               signingPayTo,
				Amount:              "10000",
				AssetTransferMethod: "swap",
			})

			assert.Nil(t, signed)
			var validationError *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationError)
			assert.Equal(t, "assetTransferMethod", validationError.Field)
		})
	})

	t.Run("Given evaluating the spend policy fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSigningX402PaymentTestingSuite) {
			suite.mockSpendPolicyEvaluator.EXPECT().
				Authorize(ctx, gomock.Any(), big.NewInt(10000)).
				Return(nil, apperrors.NewInternalError("spend_policy_failed", errors.New("db down"))).
				Times(1)
		}

		// No Release expectation: nothing was recorded, so nothing is given back.
		t.Run("Should return the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningX402PaymentBeforeEach(t)
			initSuite(suite)

			signed, err := suite.sut.Sign(ctx, eip3009SigningQuote())

			assert.Nil(t, signed)
			var internalError *apperrors.InternalError
			assert.ErrorAs(t, err, &internalError)
			assert.Equal(t, "spend_policy_failed", internalError.Code)
		})
	})

	t.Run("Given the spend policy refuses the quote", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSigningX402PaymentTestingSuite) {
			suite.mockSpendPolicyEvaluator.EXPECT().
				Authorize(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, apperrors.NewUnauthorizedError("spend_payee_not_allowed", "")).
				Times(1)
		}

		t.Run("Should return the refusal instead of a signature", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningX402PaymentBeforeEach(t)
			initSuite(suite)

			signed, err := suite.sut.Sign(ctx, &payment_interfaces.X402PaymentQuote{
				Network:             signingNetwork,
				Asset:               signingAsset,
				PayTo:               signingPayTo,
				Amount:              "10000",
				AssetTransferMethod: enums.AssetTransferMethodEip3009,
				Eip712Name:          "USDC",
				Eip712Version:       "2",
			})

			assert.Nil(t, signed)
			var unauthorizedError *apperrors.UnauthorizedError
			assert.ErrorAs(t, err, &unauthorizedError)
			assert.Equal(t, "spend_payee_not_allowed", unauthorizedError.Code)
		})

		// The gate runs before EnsureApproval so a refusal never costs gas.
		t.Run("Should not grant permit2 an allowance", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningX402PaymentBeforeEach(t)
			initSuite(suite)

			signed, err := suite.sut.Sign(ctx, &payment_interfaces.X402PaymentQuote{
				Network:             signingNetwork,
				Asset:               signingAsset,
				PayTo:               signingPayTo,
				Amount:              "10000",
				AssetTransferMethod: enums.AssetTransferMethodPermit2,
			})

			assert.Nil(t, signed)
			var unauthorizedError *apperrors.UnauthorizedError
			assert.ErrorAs(t, err, &unauthorizedError)
			assert.Equal(t, "spend_payee_not_allowed", unauthorizedError.Code)
		})
	})

	t.Run("Given the wallet address is unavailable", func(t *testing.T) {
		t.Parallel()

		reservation := signingReservation()

		initSuite := func(suite *WhenSigningX402PaymentTestingSuite) {
			suite.mockSpendPolicyEvaluator.EXPECT().
				Authorize(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(reservation, nil).
				Times(1)
			suite.mockTypedDataSigner.EXPECT().
				WalletAddress(ctx).
				Return(common.Address{}, errors.New("node wallet not bootstrapped")).
				Times(1)
		}

		t.Run("Should return the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningX402PaymentBeforeEach(t)
			initSuite(suite)
			suite.mockSpendPolicyEvaluator.EXPECT().Release(gomock.Any(), reservation).Return(nil).Times(1)

			signed, err := suite.sut.Sign(ctx, eip3009SigningQuote())

			assert.Nil(t, signed)
			assert.EqualError(t, err, "node wallet not bootstrapped")
		})

		t.Run("Should give the spend back", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningX402PaymentBeforeEach(t)
			initSuite(suite)
			suite.mockSpendPolicyEvaluator.EXPECT().Release(gomock.Any(), reservation).Return(nil).Times(1)

			_, _ = suite.sut.Sign(ctx, eip3009SigningQuote())
		})
	})

	t.Run("Given the permit2 approval fails", func(t *testing.T) {
		t.Parallel()

		reservation := signingReservation()

		initSuite := func(suite *WhenSigningX402PaymentTestingSuite) {
			suite.mockSpendPolicyEvaluator.EXPECT().
				Authorize(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(reservation, nil).
				Times(1)
			suite.expectWalletSigning()
			suite.mockPermit2ApprovalEnsurer.EXPECT().
				EnsureApproval(ctx, gomock.Any(), gomock.Any(), gomock.Any()).
				Return(errors.New("insufficient gas")).
				Times(1)
		}

		t.Run("Should fail rather than return a signature that can never settle", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningX402PaymentBeforeEach(t)
			initSuite(suite)
			suite.mockSpendPolicyEvaluator.EXPECT().Release(gomock.Any(), reservation).Return(nil).Times(1)

			signed, err := suite.sut.Sign(ctx, &payment_interfaces.X402PaymentQuote{
				Network:             signingNetwork,
				Asset:               signingAsset,
				PayTo:               signingPayTo,
				Amount:              "10000",
				AssetTransferMethod: enums.AssetTransferMethodPermit2,
			})

			assert.Nil(t, signed)
			assert.ErrorContains(t, err, "permit2 approval")
		})

		t.Run("Should give the spend back", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningX402PaymentBeforeEach(t)
			initSuite(suite)
			suite.mockSpendPolicyEvaluator.EXPECT().Release(gomock.Any(), reservation).Return(nil).Times(1)

			_, _ = suite.sut.Sign(ctx, &payment_interfaces.X402PaymentQuote{
				Network:             signingNetwork,
				Asset:               signingAsset,
				PayTo:               signingPayTo,
				Amount:              "10000",
				AssetTransferMethod: enums.AssetTransferMethodPermit2,
			})
		})
	})

	t.Run("Given signing the digest fails", func(t *testing.T) {
		t.Parallel()

		reservation := signingReservation()

		initSuite := func(suite *WhenSigningX402PaymentTestingSuite) {
			suite.mockSpendPolicyEvaluator.EXPECT().
				Authorize(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(reservation, nil).
				Times(1)
			suite.mockTypedDataSigner.EXPECT().WalletAddress(ctx).Return(suite.walletAddress, nil).Times(1)
			suite.mockTypedDataSigner.EXPECT().
				SignDigest(ctx, gomock.Any()).
				Return(nil, errors.New("vault sealed")).
				Times(1)
		}

		t.Run("Should return the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningX402PaymentBeforeEach(t)
			initSuite(suite)
			suite.mockSpendPolicyEvaluator.EXPECT().Release(gomock.Any(), reservation).Return(nil).Times(1)

			signed, err := suite.sut.Sign(ctx, eip3009SigningQuote())

			assert.Nil(t, signed)
			assert.EqualError(t, err, "vault sealed")
		})

		t.Run("Should give the spend back", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningX402PaymentBeforeEach(t)
			initSuite(suite)
			suite.mockSpendPolicyEvaluator.EXPECT().Release(gomock.Any(), reservation).Return(nil).Times(1)

			_, _ = suite.sut.Sign(ctx, eip3009SigningQuote())
		})
	})

	t.Run("Given giving the spend back fails too", func(t *testing.T) {
		t.Parallel()

		reservation := signingReservation()

		initSuite := func(suite *WhenSigningX402PaymentTestingSuite) {
			suite.mockSpendPolicyEvaluator.EXPECT().
				Authorize(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(reservation, nil).
				Times(1)
			suite.mockTypedDataSigner.EXPECT().WalletAddress(ctx).Return(suite.walletAddress, nil).Times(1)
			suite.mockTypedDataSigner.EXPECT().
				SignDigest(ctx, gomock.Any()).
				Return(nil, errors.New("vault sealed")).
				Times(1)
			suite.mockSpendPolicyEvaluator.EXPECT().
				Release(gomock.Any(), reservation).
				Return(apperrors.NewInternalError("spend_policy_failed", nil)).
				Times(1)
		}

		// The signing failure is what the caller acts on; the failed refund is only logged.
		t.Run("Should return the signing error", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningX402PaymentBeforeEach(t)
			initSuite(suite)

			signed, err := suite.sut.Sign(ctx, eip3009SigningQuote())

			assert.Nil(t, signed)
			assert.EqualError(t, err, "vault sealed")
		})
	})

	t.Run("Given an eip3009 quote", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSigningX402PaymentTestingSuite) {
			suite.mockSpendPolicyEvaluator.EXPECT().
				Authorize(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(signingReservation(), nil).
				Times(1)
			suite.expectWalletSigning()
		}

		t.Run("Should build an authorization bound to the payee and a bounded validity window", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningX402PaymentBeforeEach(t)
			initSuite(suite)

			signed, err := suite.sut.Sign(ctx, &payment_interfaces.X402PaymentQuote{
				Network:             signingNetwork,
				Asset:               signingAsset,
				PayTo:               signingPayTo,
				Amount:              "10000",
				AssetTransferMethod: enums.AssetTransferMethodEip3009,
				Eip712Name:          "USDC",
				Eip712Version:       "2",
			})

			assert.NoError(t, err)
			payload := suite.decodePaymentPayload(t, signed.PaymentSignature)

			assert.Equal(t, x402.Version, payload.X402Version)
			assert.Equal(t, x402.SchemeExact, payload.Scheme)
			assert.Equal(t, signingNetwork, payload.Network)
			assert.Nil(t, payload.Payload.Permit2Authorization)

			authorization := payload.Payload.Eip3009Authorization
			assert.Equal(t, suite.walletAddress.Hex(), authorization.From)
			assert.Equal(t, signingPayTo, authorization.To)
			assert.Equal(t, "10000", authorization.Value)
			assert.Len(t, authorization.Nonce, 66)

			now := time.Now().Unix()
			validAfter, _ := new(big.Int).SetString(authorization.ValidAfter, 10)
			validBefore, _ := new(big.Int).SetString(authorization.ValidBefore, 10)
			assert.LessOrEqual(t, validAfter.Int64(), now)
			assert.Greater(t, validBefore.Int64(), now)
			assert.Equal(t, signed.ValidBefore, validBefore.Int64())
		})

	})

	t.Run("Given a quote whose seller allows a shorter signature life", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSigningX402PaymentTestingSuite) {
			suite.mockSpendPolicyEvaluator.EXPECT().
				Authorize(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(signingReservation(), nil).
				Times(1)
			suite.expectWalletSigning()
		}

		t.Run("Should expire the signature within the seller's limit", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningX402PaymentBeforeEach(t)
			initSuite(suite)

			before := time.Now().Unix()
			signed, err := suite.sut.Sign(ctx, &payment_interfaces.X402PaymentQuote{
				Network:             signingNetwork,
				Asset:               signingAsset,
				PayTo:               signingPayTo,
				Amount:              "10000",
				AssetTransferMethod: enums.AssetTransferMethodEip3009,
				Eip712Name:          "USDC",
				Eip712Version:       "2",
				MaxTimeoutSeconds:   30,
			})

			assert.NoError(t, err)
			assert.LessOrEqual(t, signed.ValidBefore, time.Now().Unix()+30)
			assert.GreaterOrEqual(t, signed.ValidBefore, before+30)
		})
	})

	t.Run("Given the same eip3009 quote is signed twice", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSigningX402PaymentTestingSuite) {
			suite.mockSpendPolicyEvaluator.EXPECT().
				Authorize(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(signingReservation(), nil).
				Times(2)
			suite.expectWalletSigning()
		}

		t.Run("Should draw a fresh nonce every time", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningX402PaymentBeforeEach(t)
			initSuite(suite)
			quote := &payment_interfaces.X402PaymentQuote{
				Network:             signingNetwork,
				Asset:               signingAsset,
				PayTo:               signingPayTo,
				Amount:              "10000",
				AssetTransferMethod: enums.AssetTransferMethodEip3009,
				Eip712Name:          "USDC",
				Eip712Version:       "2",
			}

			first, err := suite.sut.Sign(ctx, quote)
			assert.NoError(t, err)
			second, err := suite.sut.Sign(ctx, quote)
			assert.NoError(t, err)

			assert.NotEqual(t,
				suite.decodePaymentPayload(t, first.PaymentSignature).Payload.Eip3009Authorization.Nonce,
				suite.decodePaymentPayload(t, second.PaymentSignature).Payload.Eip3009Authorization.Nonce,
			)
		})
	})

	t.Run("Given a permit2 quote", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSigningX402PaymentTestingSuite) {
			suite.mockSpendPolicyEvaluator.EXPECT().
				Authorize(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(signingReservation(), nil).
				Times(1)
			suite.expectWalletSigning()
			suite.mockPermit2ApprovalEnsurer.EXPECT().
				EnsureApproval(ctx, signingNetwork, signingAsset, big.NewInt(10000)).
				Return(nil).
				Times(1)
		}

		t.Run("Should approve permit2 first and sign over its fixed spender", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningX402PaymentBeforeEach(t)
			initSuite(suite)

			signed, err := suite.sut.Sign(ctx, &payment_interfaces.X402PaymentQuote{
				Network:             signingNetwork,
				Asset:               signingAsset,
				PayTo:               signingPayTo,
				Amount:              "10000",
				AssetTransferMethod: enums.AssetTransferMethodPermit2,
			})

			assert.NoError(t, err)
			payload := suite.decodePaymentPayload(t, signed.PaymentSignature)

			assert.Nil(t, payload.Payload.Eip3009Authorization)
			authorization := payload.Payload.Permit2Authorization
			assert.Equal(t, x402.Permit2Spender, authorization.Spender)
			assert.Equal(t, signingAsset, authorization.Permitted.Token)
			assert.Equal(t, "10000", authorization.Permitted.Amount)
			assert.Equal(t, signingPayTo, authorization.Witness.To)
		})

	})

	t.Run("Given a zero-price permit2 quote", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSigningX402PaymentTestingSuite) {
			suite.mockSpendPolicyEvaluator.EXPECT().
				Authorize(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(signingReservation(), nil).
				Times(1)
			suite.expectWalletSigning()
		}

		t.Run("Should skip the approval", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningX402PaymentBeforeEach(t)
			initSuite(suite)

			signed, err := suite.sut.Sign(ctx, &payment_interfaces.X402PaymentQuote{
				Network:             signingNetwork,
				Asset:               signingAsset,
				PayTo:               signingPayTo,
				Amount:              "0",
				AssetTransferMethod: enums.AssetTransferMethodPermit2,
			})

			assert.NoError(t, err)
			assert.NotEmpty(t, signed.PaymentSignature)
		})

	})

	// A payment the signer produced must recover to the same wallet the verifier checks against.
	t.Run("Given an eip3009 payment signed by the agent wallet", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSigningX402PaymentTestingSuite) {
			suite.mockSpendPolicyEvaluator.EXPECT().
				Authorize(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(signingReservation(), nil).
				Times(1)
			suite.expectWalletSigning()
		}

		t.Run("Should recover as that same wallet", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningX402PaymentBeforeEach(t)
			initSuite(suite)

			signed, err := suite.sut.Sign(ctx, &payment_interfaces.X402PaymentQuote{
				Network:             signingNetwork,
				Asset:               signingAsset,
				PayTo:               signingPayTo,
				Amount:              "10000",
				AssetTransferMethod: enums.AssetTransferMethodEip3009,
				Eip712Name:          "USDC",
				Eip712Version:       "2",
			})
			assert.NoError(t, err)

			payer, err := services.NewX402PaymentVerifier().RecoverPayer(
				suite.decodePaymentPayload(t, signed.PaymentSignature),
				model_builders.NewPaymentReceiveRuleX402ConfigBuilder().
					WithNetwork(signingNetwork).
					WithAsset(signingAsset).
					WithPayTo(signingPayTo).
					WithAssetTransferMethod(enums.AssetTransferMethodEip3009).
					Build(),
			)

			assert.NoError(t, err)
			assert.Equal(t, suite.walletAddress, payer)
		})
	})

	t.Run("Given a permit2 payment signed by the agent wallet", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSigningX402PaymentTestingSuite) {
			suite.mockSpendPolicyEvaluator.EXPECT().
				Authorize(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(signingReservation(), nil).
				Times(1)
			suite.expectWalletSigning()
		}

		t.Run("Should recover as that same wallet", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningX402PaymentBeforeEach(t)
			initSuite(suite)

			// Zero-price keeps the on-chain approval out of a unit test.
			signed, err := suite.sut.Sign(ctx, &payment_interfaces.X402PaymentQuote{
				Network:             signingNetwork,
				Asset:               signingAsset,
				PayTo:               signingPayTo,
				Amount:              "0",
				AssetTransferMethod: enums.AssetTransferMethodPermit2,
			})
			assert.NoError(t, err)

			payer, err := services.NewX402PaymentVerifier().RecoverPayer(
				suite.decodePaymentPayload(t, signed.PaymentSignature),
				model_builders.NewPaymentReceiveRuleX402ConfigBuilder().
					WithNetwork(signingNetwork).
					WithAsset(signingAsset).
					WithPayTo(signingPayTo).
					WithAssetTransferMethod(enums.AssetTransferMethodPermit2).
					Build(),
			)

			assert.NoError(t, err)
			assert.Equal(t, suite.walletAddress, payer)
		})
	})
}

func eip3009SigningQuote() *payment_interfaces.X402PaymentQuote {
	return &payment_interfaces.X402PaymentQuote{
		Network:             signingNetwork,
		Asset:               signingAsset,
		PayTo:               signingPayTo,
		Amount:              "10000",
		AssetTransferMethod: enums.AssetTransferMethodEip3009,
		Eip712Name:          "USDC",
		Eip712Version:       "2",
	}
}
