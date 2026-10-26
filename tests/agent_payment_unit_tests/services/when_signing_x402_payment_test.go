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

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/services"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	payment_settings "github.com/6022-labs/agentic-agent-node/src/agent_payment/settings"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/x402"
	identity_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/services_mocks/interfaces_mocks"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

const (
	signingNetwork = "eip155:80002"
	signingAsset   = "0x41E94Eb019C0762f9Bfcf9Fb1E58725BfB0e7582"
	signingPayTo   = "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
)

type WhenSigningX402PaymentTestingSuite struct {
	sut *services.X402PaymentSigner

	mockTypedDataSigner        *identity_mocks.MockTypedDataSignerInterface
	mockPermit2ApprovalEnsurer *interfaces_mocks.MockPermit2ApprovalEnsurerInterface
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

	mockTypedDataSigner.EXPECT().WalletAddress(gomock.Any()).Return(walletAddress, nil).AnyTimes()
	// Sign the real digest so the gateway verifier can be pointed at the result.
	mockTypedDataSigner.EXPECT().
		SignDigest(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, digest []byte) ([]byte, error) {
			signature, err := crypto.Sign(digest, walletKey)
			if err != nil {
				return nil, err
			}
			signature[crypto.RecoveryIDOffset] += 27
			return signature, nil
		}).
		AnyTimes()

	return &WhenSigningX402PaymentTestingSuite{
		sut: services.NewX402PaymentSigner(
			mockTypedDataSigner,
			mockPermit2ApprovalEnsurer,
			newSpendSettings(),
		),
		mockTypedDataSigner:        mockTypedDataSigner,
		mockPermit2ApprovalEnsurer: mockPermit2ApprovalEnsurer,
		walletKey:                  walletKey,
		walletAddress:              walletAddress,
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

// decodePaymentPayload unwraps the base64 header the signer produced.
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

	t.Run("Given an eip3009 quote", func(t *testing.T) {
		t.Parallel()

		t.Run("Should build an authorization bound to the payee and a bounded validity window", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningX402PaymentBeforeEach(t)

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

			authorization := payload.Payload.Authorization
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

		t.Run("Should draw a fresh nonce every time", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningX402PaymentBeforeEach(t)
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
				suite.decodePaymentPayload(t, first.PaymentSignature).Payload.Authorization.Nonce,
				suite.decodePaymentPayload(t, second.PaymentSignature).Payload.Authorization.Nonce,
			)
		})
	})

	t.Run("Given a permit2 quote", func(t *testing.T) {
		t.Parallel()

		t.Run("Should approve permit2 first and sign over its fixed spender", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningX402PaymentBeforeEach(t)
			suite.mockPermit2ApprovalEnsurer.EXPECT().
				EnsureApproval(ctx, signingNetwork, signingAsset, big.NewInt(10000)).
				Return(nil)

			signed, err := suite.sut.Sign(ctx, &payment_interfaces.X402PaymentQuote{
				Network:             signingNetwork,
				Asset:               signingAsset,
				PayTo:               signingPayTo,
				Amount:              "10000",
				AssetTransferMethod: enums.AssetTransferMethodPermit2,
			})

			assert.NoError(t, err)
			payload := suite.decodePaymentPayload(t, signed.PaymentSignature)

			assert.Nil(t, payload.Payload.Authorization)
			authorization := payload.Payload.Permit2Authorization
			assert.Equal(t, x402.Permit2Spender, authorization.Spender)
			assert.Equal(t, signingAsset, authorization.Permitted.Token)
			assert.Equal(t, "10000", authorization.Permitted.Amount)
			assert.Equal(t, signingPayTo, authorization.Witness.To)
		})

		t.Run("Should skip the approval for a zero-price quote", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningX402PaymentBeforeEach(t)

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

		t.Run("Should fail rather than return a signature that can never settle", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningX402PaymentBeforeEach(t)
			suite.mockPermit2ApprovalEnsurer.EXPECT().
				EnsureApproval(ctx, gomock.Any(), gomock.Any(), gomock.Any()).
				Return(errors.New("insufficient gas"))

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
	})

	t.Run("Given an unsupported transfer method", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse to sign", func(t *testing.T) {
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
}
