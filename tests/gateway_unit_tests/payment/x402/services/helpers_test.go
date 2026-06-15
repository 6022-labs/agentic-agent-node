package services_test

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/common/helpers"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/x402"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/stretchr/testify/assert"
)

// testPrice is the authorization value; price lives on the rule, not the config.
const testPrice = "10000"

func x402TestAcceptOption() *x402.X402Config {
	return &x402.X402Config{
		Network:             "eip155:80002",
		Asset:               "0x41E94Eb019C0762f9Bfcf9Fb1E58725BfB0e7582",
		PayTo:               "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
		AssetTransferMethod: x402.AssetTransferMethodEip3009,
		Eip712Name:          "USDC",
		Eip712Version:       "2",
	}
}

func x402TestAuthorization(from string, accept *x402.X402Config) x402.Eip3009Authorization {
	now := time.Now().Unix()
	return x402.Eip3009Authorization{
		From:        from,
		To:          accept.PayTo,
		Value:       testPrice,
		ValidAfter:  fmt.Sprintf("%d", now-60),
		ValidBefore: fmt.Sprintf("%d", now+300),
		Nonce:       "0x" + hex.EncodeToString(crypto.Keccak256([]byte("test-nonce"))),
	}
}

// signX402Payment signs the EIP-3009 typed data exactly like an x402 client.
func signX402Payment(t *testing.T, key *ecdsa.PrivateKey, auth x402.Eip3009Authorization, accept *x402.X402Config) *x402.PaymentPayload {
	t.Helper()

	parsedChainId, err := helpers.ParseCaip2(accept.Network)
	assert.NoError(t, err)
	chainId := int64(parsedChainId)

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
			Authorization: &auth,
		},
	}
}

// permit2TestNonce is a representative full-uint256 unordered nonce.
const permit2TestNonce = "33247007178036348590600198031289925668252061821958005840077069883511451257277"

func x402Permit2TestConfig() *x402.X402Config {
	return &x402.X402Config{
		Network:             "eip155:80002",
		Asset:               "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
		PayTo:               "0x209693Bc6afc0C5328bA36FaF03C514EF312287C",
		AssetTransferMethod: x402.AssetTransferMethodPermit2,
	}
}

func x402Permit2TestAuthorization(from string, config *x402.X402Config) *x402.Permit2Authorization {
	now := time.Now().Unix()
	return &x402.Permit2Authorization{
		Permitted: x402.Permit2TokenPermissions{Token: config.Asset, Amount: testPrice},
		From:      from,
		Spender:   x402.Permit2Spender,
		Nonce:     permit2TestNonce,
		Deadline:  fmt.Sprintf("%d", now+300),
		Witness:   x402.Permit2Witness{To: config.PayTo, ValidAfter: fmt.Sprintf("%d", now-60)},
	}
}

// permit2TypedData rebuilds the PermitWitnessTransferFrom typed data
// independently of the verifier — agreement on the digest proves both encode
// the spec the same way, not a shared bug.
func permit2TypedData(t *testing.T, auth *x402.Permit2Authorization, chainId int64) apitypes.TypedData {
	t.Helper()
	return apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": {
				{Name: "name", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"PermitWitnessTransferFrom": {
				{Name: "permitted", Type: "TokenPermissions"},
				{Name: "spender", Type: "address"},
				{Name: "nonce", Type: "uint256"},
				{Name: "deadline", Type: "uint256"},
				{Name: "witness", Type: "Witness"},
			},
			"TokenPermissions": {
				{Name: "token", Type: "address"},
				{Name: "amount", Type: "uint256"},
			},
			"Witness": {
				{Name: "to", Type: "address"},
				{Name: "validAfter", Type: "uint256"},
			},
		},
		PrimaryType: "PermitWitnessTransferFrom",
		Domain: apitypes.TypedDataDomain{
			Name:              x402.Permit2DomainName,
			ChainId:           math.NewHexOrDecimal256(chainId),
			VerifyingContract: x402.Permit2ContractAddress,
		},
		Message: apitypes.TypedDataMessage{
			"permitted": apitypes.TypedDataMessage{"token": auth.Permitted.Token, "amount": auth.Permitted.Amount},
			"spender":   auth.Spender,
			"nonce":     auth.Nonce,
			"deadline":  auth.Deadline,
			"witness":   apitypes.TypedDataMessage{"to": auth.Witness.To, "validAfter": auth.Witness.ValidAfter},
		},
	}
}

func signPermit2Payment(t *testing.T, key *ecdsa.PrivateKey, auth *x402.Permit2Authorization, config *x402.X402Config) *x402.PaymentPayload {
	t.Helper()

	parsedChainId, err := helpers.ParseCaip2(config.Network)
	assert.NoError(t, err)
	chainId := int64(parsedChainId)

	digest, _, err := apitypes.TypedDataAndHash(permit2TypedData(t, auth, chainId))
	assert.NoError(t, err)

	signature, err := crypto.Sign(digest, key)
	assert.NoError(t, err)

	return &x402.PaymentPayload{
		X402Version: x402.Version,
		Scheme:      x402.SchemeExact,
		Network:     config.Network,
		Payload: x402.ExactEvmPayload{
			Signature:            "0x" + hex.EncodeToString(signature),
			Permit2Authorization: auth,
		},
	}
}
