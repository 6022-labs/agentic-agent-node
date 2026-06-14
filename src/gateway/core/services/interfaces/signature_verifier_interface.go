package interfaces

import "github.com/ethereum/go-ethereum/common"

// SignatureVerifierInterface recovers the signer address from an EIP-191
// "personal_sign" hex signature over the given message. Returns a zero address
// + error on malformed input or recovery failure.
type SignatureVerifierInterface interface {
	Recover(message string, signatureHex string) (common.Address, error)
}
