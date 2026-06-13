package services_test

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/gateway/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway/models/x402"
	"github.com/6022-labs/agentic-agent-node/src/gateway/services"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/services/interfaces"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/stretchr/testify/assert"
)

type WhenRecoveringX402PayerTestingSuite struct {
	sut *services.X402PaymentVerifier
}

func WhenRecoveringX402PayerBeforeEach() *WhenRecoveringX402PayerTestingSuite {
	return &WhenRecoveringX402PayerTestingSuite{sut: services.NewX402PaymentVerifier()}
}

func x402TestAcceptOption() *models.ExternalPaymentRule {
	return &models.ExternalPaymentRule{
		Network:           "eip155:80002",
		Asset:             "0x41E94Eb019C0762f9Bfcf9Fb1E58725BfB0e7582",
		PayTo:             "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
		Price:             "10000",
		MaxTimeoutSeconds: 60,
		Eip712Name:        "USDC",
		Eip712Version:     "2",
	}
}

// signX402Payment signs the EIP-3009 typed data exactly like an x402 client.
func signX402Payment(t *testing.T, key *ecdsa.PrivateKey, auth x402.ExactEvmAuthorization, accept *models.ExternalPaymentRule) *x402.PaymentPayload {
	t.Helper()

	chainId, err := x402.ParseEip155ChainId(accept.Network)
	assert.NoError(t, err)

	typedData := apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": {
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"TransferWithAuthorization": {
				{Name: "from", Type: "address"},
				{Name: "to", Type: "address"},
				{Name: "value", Type: "uint256"},
				{Name: "validAfter", Type: "uint256"},
				{Name: "validBefore", Type: "uint256"},
				{Name: "nonce", Type: "bytes32"},
			},
		},
		PrimaryType: "TransferWithAuthorization",
		Domain: apitypes.TypedDataDomain{
			Name:              accept.Eip712Name,
			Version:           accept.Eip712Version,
			ChainId:           math.NewHexOrDecimal256(chainId),
			VerifyingContract: accept.Asset,
		},
		Message: apitypes.TypedDataMessage{
			"from":        auth.From,
			"to":          auth.To,
			"value":       auth.Value,
			"validAfter":  auth.ValidAfter,
			"validBefore": auth.ValidBefore,
			"nonce":       auth.Nonce,
		},
	}
	digest, _, err := apitypes.TypedDataAndHash(typedData)
	assert.NoError(t, err)

	signature, err := crypto.Sign(digest, key)
	assert.NoError(t, err)

	return &x402.PaymentPayload{
		X402Version: x402.Version,
		Scheme:      x402.SchemeExact,
		Network:     accept.Network,
		Payload: x402.ExactEvmPayload{
			Signature:     "0x" + hex.EncodeToString(signature),
			Authorization: auth,
		},
	}
}

func x402TestAuthorization(from string, accept *models.ExternalPaymentRule) x402.ExactEvmAuthorization {
	now := time.Now().Unix()
	return x402.ExactEvmAuthorization{
		From:        from,
		To:          accept.PayTo,
		Value:       accept.Price,
		ValidAfter:  fmt.Sprintf("%d", now-60),
		ValidBefore: fmt.Sprintf("%d", now+300),
		Nonce:       "0x" + hex.EncodeToString(crypto.Keccak256([]byte("test-nonce"))),
	}
}

func TestWhenRecoveringX402Payer(t *testing.T) {
	t.Parallel()

	t.Run("Given a valid signed authorization", func(t *testing.T) {
		t.Parallel()

		t.Run("Should recover the signer address", func(t *testing.T) {
			t.Parallel()

			suite := WhenRecoveringX402PayerBeforeEach()
			key, _ := crypto.GenerateKey()
			signer := crypto.PubkeyToAddress(key.PublicKey)
			accept := x402TestAcceptOption()
			payload := signX402Payment(t, key, x402TestAuthorization(signer.Hex(), accept), accept)

			payer, err := suite.sut.RecoverPayer(payload, accept)

			assert.NoError(t, err)
			assert.Equal(t, signer, payer)
		})

		t.Run("Should accept a 27/28 recovery id", func(t *testing.T) {
			t.Parallel()

			suite := WhenRecoveringX402PayerBeforeEach()
			key, _ := crypto.GenerateKey()
			signer := crypto.PubkeyToAddress(key.PublicKey)
			accept := x402TestAcceptOption()
			payload := signX402Payment(t, key, x402TestAuthorization(signer.Hex(), accept), accept)

			sigBytes, err := hex.DecodeString(payload.Payload.Signature[2:])
			assert.NoError(t, err)
			sigBytes[64] += 27
			payload.Payload.Signature = "0x" + hex.EncodeToString(sigBytes)

			payer, err := suite.sut.RecoverPayer(payload, accept)

			assert.NoError(t, err)
			assert.Equal(t, signer, payer)
		})
	})

	t.Run("Given the authorization was tampered with after signing", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject the payment", func(t *testing.T) {
			t.Parallel()

			suite := WhenRecoveringX402PayerBeforeEach()
			key, _ := crypto.GenerateKey()
			signer := crypto.PubkeyToAddress(key.PublicKey)
			accept := x402TestAcceptOption()
			payload := signX402Payment(t, key, x402TestAuthorization(signer.Hex(), accept), accept)
			payload.Payload.Authorization.Value = "999999999"

			_, err := suite.sut.RecoverPayer(payload, accept)

			assert.Error(t, err)
		})
	})

	t.Run("Given the recipient does not match payTo", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject the payment", func(t *testing.T) {
			t.Parallel()

			suite := WhenRecoveringX402PayerBeforeEach()
			key, _ := crypto.GenerateKey()
			signer := crypto.PubkeyToAddress(key.PublicKey)
			accept := x402TestAcceptOption()
			auth := x402TestAuthorization(signer.Hex(), accept)
			auth.To = "0x70997970C51812dc3A010C7d01b50e0d17dc79C8"
			payload := signX402Payment(t, key, auth, accept)

			_, err := suite.sut.RecoverPayer(payload, accept)

			assert.Error(t, err)
		})
	})

	t.Run("Given a zero-value free-auth authorization", func(t *testing.T) {
		t.Parallel()

		t.Run("Should recover the signer (the amount is the facilitator's concern)", func(t *testing.T) {
			t.Parallel()

			suite := WhenRecoveringX402PayerBeforeEach()
			key, _ := crypto.GenerateKey()
			signer := crypto.PubkeyToAddress(key.PublicKey)
			accept := x402TestAcceptOption()
			auth := x402TestAuthorization(signer.Hex(), accept)
			auth.Value = "0"
			payload := signX402Payment(t, key, auth, accept)

			payer, err := suite.sut.RecoverPayer(payload, accept)

			assert.NoError(t, err)
			assert.Equal(t, signer, payer)
		})
	})

	t.Run("Given the authorization is outside its validity window", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return the expired sentinel when validBefore has passed", func(t *testing.T) {
			t.Parallel()

			suite := WhenRecoveringX402PayerBeforeEach()
			key, _ := crypto.GenerateKey()
			signer := crypto.PubkeyToAddress(key.PublicKey)
			accept := x402TestAcceptOption()
			auth := x402TestAuthorization(signer.Hex(), accept)
			auth.ValidBefore = fmt.Sprintf("%d", time.Now().Unix()-10)
			payload := signX402Payment(t, key, auth, accept)

			_, err := suite.sut.RecoverPayer(payload, accept)

			assert.ErrorIs(t, err, gateway_interfaces.ErrX402AuthorizationExpired)
		})

		t.Run("Should return the expired sentinel when validAfter is in the future", func(t *testing.T) {
			t.Parallel()

			suite := WhenRecoveringX402PayerBeforeEach()
			key, _ := crypto.GenerateKey()
			signer := crypto.PubkeyToAddress(key.PublicKey)
			accept := x402TestAcceptOption()
			auth := x402TestAuthorization(signer.Hex(), accept)
			auth.ValidAfter = fmt.Sprintf("%d", time.Now().Unix()+300)
			payload := signX402Payment(t, key, auth, accept)

			_, err := suite.sut.RecoverPayer(payload, accept)

			assert.ErrorIs(t, err, gateway_interfaces.ErrX402AuthorizationExpired)
		})
	})

	t.Run("Given the from address is not the signer", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject the payment", func(t *testing.T) {
			t.Parallel()

			suite := WhenRecoveringX402PayerBeforeEach()
			key, _ := crypto.GenerateKey()
			accept := x402TestAcceptOption()
			auth := x402TestAuthorization("0x70997970C51812dc3A010C7d01b50e0d17dc79C8", accept)
			payload := signX402Payment(t, key, auth, accept)

			_, err := suite.sut.RecoverPayer(payload, accept)

			assert.Error(t, err)
		})
	})

	t.Run("Given the signature is garbage", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject the payment", func(t *testing.T) {
			t.Parallel()

			suite := WhenRecoveringX402PayerBeforeEach()
			key, _ := crypto.GenerateKey()
			signer := crypto.PubkeyToAddress(key.PublicKey)
			accept := x402TestAcceptOption()
			payload := signX402Payment(t, key, x402TestAuthorization(signer.Hex(), accept), accept)
			payload.Payload.Signature = "0xdeadbeef"

			_, err := suite.sut.RecoverPayer(payload, accept)

			assert.Error(t, err)
		})
	})
}
