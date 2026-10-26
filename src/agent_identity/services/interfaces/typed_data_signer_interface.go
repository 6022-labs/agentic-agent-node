package interfaces

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

// TypedDataSignerInterface signs EIP-712 payloads with the node's EVM wallet key.
// Signatures come back with v as 27/28, the encoding on-chain verifiers expect.
type TypedDataSignerInterface interface {
	// SignDigest signs an already-hashed 32-byte EIP-712 digest, returning 65 bytes.
	SignDigest(ctx context.Context, digest []byte) ([]byte, error)
	// SignTypedData hashes then signs; SignDigest takes an already-hashed digest.
	SignTypedData(ctx context.Context, typedData *apitypes.TypedData) (string, error)
	// WalletAddress returns the address the signatures recover to.
	WalletAddress(ctx context.Context) (common.Address, error)
}
