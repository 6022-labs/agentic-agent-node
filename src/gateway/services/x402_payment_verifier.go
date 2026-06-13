package services

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/gateway/models/x402"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/services/interfaces"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

type X402PaymentVerifier struct{}

func NewX402PaymentVerifier() *X402PaymentVerifier {
	return &X402PaymentVerifier{}
}

func (v *X402PaymentVerifier) RecoverPayer(payload *x402.PaymentPayload, config *x402.X402Config) (common.Address, error) {
	auth := payload.Payload.Authorization

	if _, err := parseDecimal(auth.Value); err != nil {
		return common.Address{}, fmt.Errorf("invalid authorization value: %w", err)
	}
	validAfter, err := parseDecimal(auth.ValidAfter)
	if err != nil {
		return common.Address{}, fmt.Errorf("invalid authorization validAfter: %w", err)
	}
	validBefore, err := parseDecimal(auth.ValidBefore)
	if err != nil {
		return common.Address{}, fmt.Errorf("invalid authorization validBefore: %w", err)
	}
	if !common.IsHexAddress(auth.From) || !common.IsHexAddress(auth.To) {
		return common.Address{}, errors.New("authorization from/to must be hex addresses")
	}

	// The amount is not checked here: the facilitator prices paid calls and
	// free-access signatures carry value 0.
	if common.HexToAddress(auth.To) != common.HexToAddress(config.PayTo) {
		return common.Address{}, errors.New("authorization recipient does not match payTo")
	}
	now := big.NewInt(time.Now().Unix())
	if now.Cmp(validAfter) < 0 || now.Cmp(validBefore) > 0 {
		return common.Address{}, gateway_interfaces.ErrX402AuthorizationExpired
	}

	chainId, err := x402.ParseEip155ChainId(config.Network)
	if err != nil {
		return common.Address{}, err
	}

	digest, err := transferWithAuthorizationDigest(auth, config, chainId)
	if err != nil {
		return common.Address{}, err
	}

	sigBytes, err := hex.DecodeString(strings.TrimPrefix(payload.Payload.Signature, "0x"))
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

	pubKey, err := crypto.SigToPub(digest, sigBytes)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to recover signer: %w", err)
	}
	payer := crypto.PubkeyToAddress(*pubKey)
	if payer != common.HexToAddress(auth.From) {
		return common.Address{}, errors.New("recovered signer does not match authorization from")
	}
	return payer, nil
}

// transferWithAuthorizationDigest hashes the EIP-3009 typed data under the
// asset contract's EIP-712 domain.
func transferWithAuthorizationDigest(auth x402.ExactEvmAuthorization, config *x402.X402Config, chainId int64) ([]byte, error) {
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
			Name:              config.Eip712Name,
			Version:           config.Eip712Version,
			ChainId:           math.NewHexOrDecimal256(chainId),
			VerifyingContract: config.Asset,
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
	if err != nil {
		return nil, fmt.Errorf("failed to hash typed data: %w", err)
	}
	return digest, nil
}

func parseDecimal(s string) (*big.Int, error) {
	value, ok := new(big.Int).SetString(strings.TrimSpace(s), 10)
	if !ok {
		return nil, fmt.Errorf("%q is not a decimal integer", s)
	}
	return value, nil
}
