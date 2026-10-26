package x402_test

import (
	"encoding/hex"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/common/x402"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/stretchr/testify/assert"
)

const hashingChainId = int64(80002)

func testEip3009Authorization() *x402.Eip3009Authorization {
	return &x402.Eip3009Authorization{
		From:        "0x1111111111111111111111111111111111111111",
		To:          "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
		Value:       "10000",
		ValidAfter:  "1700000000",
		ValidBefore: "1700000300",
		Nonce:       "0x" + hex.EncodeToString(make([]byte, 32)),
	}
}

func testPermit2Authorization() *x402.Permit2Authorization {
	return &x402.Permit2Authorization{
		Permitted: x402.Permit2TokenPermissions{
			Token:  "0x41E94Eb019C0762f9Bfcf9Fb1E58725BfB0e7582",
			Amount: "10000",
		},
		From:     "0x1111111111111111111111111111111111111111",
		Spender:  x402.Permit2Spender,
		Nonce:    "42",
		Deadline: "1700000300",
		Witness: x402.Permit2Witness{
			To:         "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
			ValidAfter: "1700000000",
		},
	}
}

// Golden digests produced independently by ethers.js v6 (TypedDataEncoder.hash) over the same domain/types/message.
const (
	goldenEip3009Digest = "c08713f75413a1225021aaf18ddb94e18ef4a82b9208f345e773a7042cc74b82"
	goldenPermit2Digest = "cc75e64b59b0d9cdaacde47ec4804d1fae01189e6e495f6776b453f85133581a"
)

func TestWhenHashingX402TypedData(t *testing.T) {
	t.Parallel()

	t.Run("Given the reference EIP-3009 authorization", func(t *testing.T) {
		t.Parallel()

		t.Run("Should hash to the digest ethers.js produces", func(t *testing.T) {
			t.Parallel()

			digest, err := x402.TransferWithAuthorizationDigest(
				testEip3009Authorization(),
				"0x41E94Eb019C0762f9Bfcf9Fb1E58725BfB0e7582",
				"USDC",
				"2",
				hashingChainId,
			)

			assert.NoError(t, err)
			assert.Equal(t, goldenEip3009Digest, hex.EncodeToString(digest))
		})
	})

	t.Run("Given the reference Permit2 authorization", func(t *testing.T) {
		t.Parallel()

		t.Run("Should hash to the digest ethers.js produces", func(t *testing.T) {
			t.Parallel()

			digest, err := x402.PermitWitnessTransferFromDigest(testPermit2Authorization(), hashingChainId)

			assert.NoError(t, err)
			assert.Equal(t, goldenPermit2Digest, hex.EncodeToString(digest))
		})
	})

	t.Run("Given an EIP-3009 authorization", func(t *testing.T) {
		t.Parallel()

		t.Run("Should hash it under the asset's own EIP-712 domain", func(t *testing.T) {
			t.Parallel()

			asset := "0x41E94Eb019C0762f9Bfcf9Fb1E58725BfB0e7582"
			authorization := testEip3009Authorization()

			digest, err := x402.TransferWithAuthorizationDigest(authorization, asset, "USDC", "2", hashingChainId)
			assert.NoError(t, err)

			expected, _, err := apitypes.TypedDataAndHash(apitypes.TypedData{
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
					Name:              "USDC",
					Version:           "2",
					ChainId:           math.NewHexOrDecimal256(hashingChainId),
					VerifyingContract: asset,
				},
				Message: apitypes.TypedDataMessage{
					"from":        authorization.From,
					"to":          authorization.To,
					"value":       authorization.Value,
					"validAfter":  authorization.ValidAfter,
					"validBefore": authorization.ValidBefore,
					"nonce":       authorization.Nonce,
				},
			})
			assert.NoError(t, err)
			assert.Equal(t, expected, digest)
			assert.Len(t, digest, 32)
		})

		t.Run("Should change with the domain, so a wrong domain cannot verify", func(t *testing.T) {
			t.Parallel()

			asset := "0x41E94Eb019C0762f9Bfcf9Fb1E58725BfB0e7582"
			authorization := testEip3009Authorization()

			digest, err := x402.TransferWithAuthorizationDigest(authorization, asset, "USDC", "2", hashingChainId)
			assert.NoError(t, err)

			otherVersion, err := x402.TransferWithAuthorizationDigest(authorization, asset, "USDC", "1", hashingChainId)
			assert.NoError(t, err)
			otherChain, err := x402.TransferWithAuthorizationDigest(authorization, asset, "USDC", "2", 137)
			assert.NoError(t, err)

			assert.NotEqual(t, digest, otherVersion)
			assert.NotEqual(t, digest, otherChain)
		})

		t.Run("Should refuse a malformed authorization", func(t *testing.T) {
			t.Parallel()

			authorization := testEip3009Authorization()
			authorization.Value = "not a number"

			digest, err := x402.TransferWithAuthorizationDigest(authorization, "0x41E9", "USDC", "2", hashingChainId)

			assert.Nil(t, digest)
			assert.ErrorContains(t, err, "failed to hash typed data")
		})
	})

	t.Run("Given a Permit2 authorization", func(t *testing.T) {
		t.Parallel()

		t.Run("Should hash it under Permit2's fixed, versionless domain", func(t *testing.T) {
			t.Parallel()

			authorization := testPermit2Authorization()

			digest, err := x402.PermitWitnessTransferFromDigest(authorization, hashingChainId)
			assert.NoError(t, err)

			expected, _, err := apitypes.TypedDataAndHash(apitypes.TypedData{
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
					ChainId:           math.NewHexOrDecimal256(hashingChainId),
					VerifyingContract: x402.Permit2ContractAddress,
				},
				Message: apitypes.TypedDataMessage{
					"permitted": apitypes.TypedDataMessage{
						"token":  authorization.Permitted.Token,
						"amount": authorization.Permitted.Amount,
					},
					"spender":  authorization.Spender,
					"nonce":    authorization.Nonce,
					"deadline": authorization.Deadline,
					"witness": apitypes.TypedDataMessage{
						"to":         authorization.Witness.To,
						"validAfter": authorization.Witness.ValidAfter,
					},
				},
			})
			assert.NoError(t, err)
			assert.Equal(t, expected, digest)
		})

		t.Run("Should bind the witness recipient into the digest", func(t *testing.T) {
			t.Parallel()

			digest, err := x402.PermitWitnessTransferFromDigest(testPermit2Authorization(), hashingChainId)
			assert.NoError(t, err)

			rebound := testPermit2Authorization()
			rebound.Witness.To = "0x2222222222222222222222222222222222222222"
			reboundDigest, err := x402.PermitWitnessTransferFromDigest(rebound, hashingChainId)
			assert.NoError(t, err)

			assert.NotEqual(t, digest, reboundDigest)
		})
	})
}
