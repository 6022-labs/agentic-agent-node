package services_test

import (
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/x402"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/x402/services"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/payment/x402/services/interfaces"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
)

type WhenRecoveringX402Permit2PayerTestingSuite struct {
	sut *services.X402PaymentVerifier
}

func WhenRecoveringX402Permit2PayerBeforeEach() *WhenRecoveringX402Permit2PayerTestingSuite {
	return &WhenRecoveringX402Permit2PayerTestingSuite{sut: services.NewX402PaymentVerifier()}
}

func TestWhenRecoveringX402Permit2Payer(t *testing.T) {
	t.Parallel()

	t.Run("Given the Permit2 type and domain encodings", func(t *testing.T) {
		t.Parallel()

		t.Run("Should match the x402 spec type strings", func(t *testing.T) {
			t.Parallel()

			typedData := permit2TypedData(
				t,
				x402Permit2TestAuthorization("0x0000000000000000000000000000000000000001", x402Permit2TestConfig()),
				80002,
			)

			// The witness-bound primary type, exactly as the x402 spec encodes it.
			expectedType := "PermitWitnessTransferFrom(TokenPermissions permitted,address spender,uint256 nonce," +
				"uint256 deadline,Witness witness)TokenPermissions(address token,uint256 amount)" +
				"Witness(address to,uint256 validAfter)"
			assert.Equal(
				t,
				crypto.Keccak256([]byte(expectedType)),
				[]byte(typedData.TypeHash("PermitWitnessTransferFrom")),
			)

			// Permit2's domain omits version — including it changes the hash.
			expectedDomain := "EIP712Domain(string name,uint256 chainId,address verifyingContract)"
			assert.Equal(t, crypto.Keccak256([]byte(expectedDomain)), []byte(typedData.TypeHash("EIP712Domain")))
		})
	})

	t.Run("Given a valid signed Permit2 authorization", func(t *testing.T) {
		t.Parallel()

		t.Run("Should recover the signer address", func(t *testing.T) {
			t.Parallel()

			suite := WhenRecoveringX402Permit2PayerBeforeEach()
			key, _ := crypto.GenerateKey()
			signer := crypto.PubkeyToAddress(key.PublicKey)
			config := x402Permit2TestConfig()
			payload := signPermit2Payment(t, key, x402Permit2TestAuthorization(signer.Hex(), config), config)

			payer, err := suite.sut.RecoverPayer(payload, config)

			assert.NoError(t, err)
			assert.Equal(t, signer, payer)
		})

		t.Run("Should accept a 27/28 recovery id", func(t *testing.T) {
			t.Parallel()

			suite := WhenRecoveringX402Permit2PayerBeforeEach()
			key, _ := crypto.GenerateKey()
			signer := crypto.PubkeyToAddress(key.PublicKey)
			config := x402Permit2TestConfig()
			payload := signPermit2Payment(t, key, x402Permit2TestAuthorization(signer.Hex(), config), config)

			sigBytes, err := hex.DecodeString(payload.Payload.Signature[2:])
			assert.NoError(t, err)
			sigBytes[64] += 27
			payload.Payload.Signature = "0x" + hex.EncodeToString(sigBytes)

			payer, err := suite.sut.RecoverPayer(payload, config)

			assert.NoError(t, err)
			assert.Equal(t, signer, payer)
		})
	})

	t.Run("Given the witness recipient does not match payTo", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject the payment", func(t *testing.T) {
			t.Parallel()

			suite := WhenRecoveringX402Permit2PayerBeforeEach()
			key, _ := crypto.GenerateKey()
			signer := crypto.PubkeyToAddress(key.PublicKey)
			config := x402Permit2TestConfig()
			auth := x402Permit2TestAuthorization(signer.Hex(), config)
			auth.Witness.To = "0x70997970C51812dc3A010C7d01b50e0d17dc79C8"
			payload := signPermit2Payment(t, key, auth, config)

			_, err := suite.sut.RecoverPayer(payload, config)

			assert.ErrorContains(t, err, "witness recipient does not match payTo")
		})
	})

	t.Run("Given the permitted token does not match the asset", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject the payment", func(t *testing.T) {
			t.Parallel()

			suite := WhenRecoveringX402Permit2PayerBeforeEach()
			key, _ := crypto.GenerateKey()
			signer := crypto.PubkeyToAddress(key.PublicKey)
			config := x402Permit2TestConfig()
			auth := x402Permit2TestAuthorization(signer.Hex(), config)
			auth.Permitted.Token = "0x70997970C51812dc3A010C7d01b50e0d17dc79C8"
			payload := signPermit2Payment(t, key, auth, config)

			_, err := suite.sut.RecoverPayer(payload, config)

			assert.ErrorContains(t, err, "permitted token does not match asset")
		})
	})

	t.Run("Given the spender is not the x402 permit2 proxy", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject the payment", func(t *testing.T) {
			t.Parallel()

			suite := WhenRecoveringX402Permit2PayerBeforeEach()
			key, _ := crypto.GenerateKey()
			signer := crypto.PubkeyToAddress(key.PublicKey)
			config := x402Permit2TestConfig()
			auth := x402Permit2TestAuthorization(signer.Hex(), config)
			auth.Spender = "0x70997970C51812dc3A010C7d01b50e0d17dc79C8"
			payload := signPermit2Payment(t, key, auth, config)

			_, err := suite.sut.RecoverPayer(payload, config)

			assert.ErrorContains(t, err, "spender is not the x402 permit2 proxy")
		})
	})

	t.Run("Given the authorization was tampered with after signing", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject the payment", func(t *testing.T) {
			t.Parallel()

			suite := WhenRecoveringX402Permit2PayerBeforeEach()
			key, _ := crypto.GenerateKey()
			signer := crypto.PubkeyToAddress(key.PublicKey)
			config := x402Permit2TestConfig()
			payload := signPermit2Payment(t, key, x402Permit2TestAuthorization(signer.Hex(), config), config)
			payload.Payload.Permit2Authorization.Permitted.Amount = "999999999"

			_, err := suite.sut.RecoverPayer(payload, config)

			assert.ErrorContains(t, err, "recovered signer does not match permit2 from")
		})
	})

	t.Run("Given a zero-amount free-auth authorization", func(t *testing.T) {
		t.Parallel()

		t.Run("Should recover the signer (the amount is the facilitator's concern)", func(t *testing.T) {
			t.Parallel()

			suite := WhenRecoveringX402Permit2PayerBeforeEach()
			key, _ := crypto.GenerateKey()
			signer := crypto.PubkeyToAddress(key.PublicKey)
			config := x402Permit2TestConfig()
			auth := x402Permit2TestAuthorization(signer.Hex(), config)
			auth.Permitted.Amount = "0"
			payload := signPermit2Payment(t, key, auth, config)

			payer, err := suite.sut.RecoverPayer(payload, config)

			assert.NoError(t, err)
			assert.Equal(t, signer, payer)
		})
	})

	t.Run("Given the authorization has expired", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return the expired sentinel", func(t *testing.T) {
			t.Parallel()

			suite := WhenRecoveringX402Permit2PayerBeforeEach()
			key, _ := crypto.GenerateKey()
			signer := crypto.PubkeyToAddress(key.PublicKey)
			config := x402Permit2TestConfig()
			auth := x402Permit2TestAuthorization(signer.Hex(), config)
			auth.Deadline = fmt.Sprintf("%d", time.Now().Unix()-10)
			payload := signPermit2Payment(t, key, auth, config)

			_, err := suite.sut.RecoverPayer(payload, config)

			assert.ErrorIs(t, err, gateway_interfaces.ErrX402AuthorizationExpired)
		})
	})

	t.Run("Given the authorization is not yet valid", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return the expired sentinel", func(t *testing.T) {
			t.Parallel()

			suite := WhenRecoveringX402Permit2PayerBeforeEach()
			key, _ := crypto.GenerateKey()
			signer := crypto.PubkeyToAddress(key.PublicKey)
			config := x402Permit2TestConfig()
			auth := x402Permit2TestAuthorization(signer.Hex(), config)
			auth.Witness.ValidAfter = fmt.Sprintf("%d", time.Now().Unix()+300)
			payload := signPermit2Payment(t, key, auth, config)

			_, err := suite.sut.RecoverPayer(payload, config)

			assert.ErrorIs(t, err, gateway_interfaces.ErrX402AuthorizationExpired)
		})
	})

	t.Run("Given the from address is not the signer", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject the payment", func(t *testing.T) {
			t.Parallel()

			suite := WhenRecoveringX402Permit2PayerBeforeEach()
			key, _ := crypto.GenerateKey()
			config := x402Permit2TestConfig()
			auth := x402Permit2TestAuthorization("0x70997970C51812dc3A010C7d01b50e0d17dc79C8", config)
			payload := signPermit2Payment(t, key, auth, config)

			_, err := suite.sut.RecoverPayer(payload, config)

			assert.ErrorContains(t, err, "recovered signer does not match permit2 from")
		})
	})

	t.Run("Given the permit2 authorization is missing", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject the payment", func(t *testing.T) {
			t.Parallel()

			suite := WhenRecoveringX402Permit2PayerBeforeEach()
			config := x402Permit2TestConfig()
			payload := &x402.PaymentPayload{
				X402Version: x402.Version,
				Scheme:      x402.SchemeExact,
				Network:     config.Network,
				Payload:     x402.ExactEvmPayload{Signature: "0xdeadbeef"},
			}

			_, err := suite.sut.RecoverPayer(payload, config)

			assert.ErrorContains(t, err, "missing permit2 authorization")
		})
	})
}
