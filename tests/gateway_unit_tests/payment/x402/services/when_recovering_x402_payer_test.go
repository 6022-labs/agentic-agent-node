package services_test

import (
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/x402/services"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/payment/x402/services/interfaces"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
)

type WhenRecoveringX402PayerTestingSuite struct {
	sut *services.X402PaymentVerifier
}

func WhenRecoveringX402PayerBeforeEach() *WhenRecoveringX402PayerTestingSuite {
	return &WhenRecoveringX402PayerTestingSuite{sut: services.NewX402PaymentVerifier()}
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

			assert.ErrorContains(t, err, "recovered signer does not match authorization from")
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

			assert.ErrorContains(t, err, "authorization recipient does not match payTo")
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

	t.Run("Given the authorization has expired", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return the expired sentinel", func(t *testing.T) {
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
	})

	t.Run("Given the authorization is not yet valid", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return the expired sentinel", func(t *testing.T) {
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

			assert.ErrorContains(t, err, "recovered signer does not match authorization from")
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

			assert.ErrorContains(t, err, "signature length must be 65 bytes")
		})
	})

	t.Run("Given an unsupported asset transfer method", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject the payment without defaulting to a method", func(t *testing.T) {
			t.Parallel()

			suite := WhenRecoveringX402PayerBeforeEach()
			key, _ := crypto.GenerateKey()
			signer := crypto.PubkeyToAddress(key.PublicKey)
			accept := x402TestAcceptOption()
			payload := signX402Payment(t, key, x402TestAuthorization(signer.Hex(), accept), accept)
			accept.AssetTransferMethod = "unknown"

			_, err := suite.sut.RecoverPayer(payload, accept)

			assert.ErrorContains(t, err, "unsupported asset transfer method")
		})
	})
}
