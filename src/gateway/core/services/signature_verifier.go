package services

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const ethSignatureLength = 65

type SignatureVerifier struct{}

func NewSignatureVerifier() *SignatureVerifier {
	return &SignatureVerifier{}
}

// Recover returns the address that produced an EIP-191 personal_sign signature
// over `message`.
func (v *SignatureVerifier) Recover(message string, signatureHex string) (common.Address, error) {
	sigBytes, err := hex.DecodeString(strings.TrimPrefix(signatureHex, "0x"))
	if err != nil {
		return common.Address{}, fmt.Errorf("signature is not valid hex: %w", err)
	}
	if len(sigBytes) != ethSignatureLength {
		return common.Address{}, fmt.Errorf("signature length must be %d bytes", ethSignatureLength)
	}
	if sigBytes[64] == 27 || sigBytes[64] == 28 {
		sigBytes[64] -= 27
	}
	if sigBytes[64] != 0 && sigBytes[64] != 1 {
		return common.Address{}, errors.New("invalid recovery id")
	}

	prefixed := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)
	hash := crypto.Keccak256Hash([]byte(prefixed))

	pubKey, err := crypto.SigToPub(hash.Bytes(), sigBytes)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to recover signer: %w", err)
	}
	return crypto.PubkeyToAddress(*pubKey), nil
}
