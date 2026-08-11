package services

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/common/helpers"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/x402"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/payment/x402/services/interfaces"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

const ethSignatureLength = 65

type X402PaymentVerifier struct{}

func NewX402PaymentVerifier() *X402PaymentVerifier {
	return &X402PaymentVerifier{}
}

func (v *X402PaymentVerifier) RecoverPayer(
	payload *x402.PaymentPayload,
	config *x402.X402Config,
) (common.Address, error) {
	switch config.AssetTransferMethod {
	case x402.AssetTransferMethodEip3009:
		return v.recoverEip3009(payload, config)
	case x402.AssetTransferMethodPermit2:
		return v.recoverPermit2(payload, config)
	default:
		return common.Address{}, fmt.Errorf("unsupported asset transfer method: %q", config.AssetTransferMethod)
	}
}

func (v *X402PaymentVerifier) recoverEip3009(
	payload *x402.PaymentPayload,
	config *x402.X402Config,
) (common.Address, error) {
	auth := payload.Payload.Authorization
	if auth == nil {
		return common.Address{}, errors.New("missing authorization")
	}

	if _, err := v.parseDecimal(auth.Value); err != nil {
		return common.Address{}, fmt.Errorf("invalid authorization value: %w", err)
	}
	validAfter, err := v.parseDecimal(auth.ValidAfter)
	if err != nil {
		return common.Address{}, fmt.Errorf("invalid authorization validAfter: %w", err)
	}
	validBefore, err := v.parseDecimal(auth.ValidBefore)
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

	parsedChainId, err := helpers.ParseCaip2(config.Network)
	if err != nil {
		return common.Address{}, err
	}
	chainId := int64(parsedChainId)

	digest, err := v.transferWithAuthorizationDigest(auth, config, chainId)
	if err != nil {
		return common.Address{}, err
	}

	payer, err := v.recoverSigner(digest, payload.Payload.Signature)
	if err != nil {
		return common.Address{}, err
	}
	if payer != common.HexToAddress(auth.From) {
		return common.Address{}, errors.New("recovered signer does not match authorization from")
	}
	return payer, nil
}

// recoverPermit2 validates a Permit2 PermitWitnessTransferFrom signature,
// hashed under Permit2's own fixed domain rather than the token's.
func (v *X402PaymentVerifier) recoverPermit2(
	payload *x402.PaymentPayload,
	config *x402.X402Config,
) (common.Address, error) {
	auth := payload.Payload.Permit2Authorization
	if auth == nil {
		return common.Address{}, errors.New("missing permit2 authorization")
	}

	if _, err := v.parseDecimal(auth.Permitted.Amount); err != nil {
		return common.Address{}, fmt.Errorf("invalid permitted amount: %w", err)
	}
	if _, err := v.parseDecimal(auth.Nonce); err != nil {
		return common.Address{}, fmt.Errorf("invalid permit2 nonce: %w", err)
	}
	deadline, err := v.parseDecimal(auth.Deadline)
	if err != nil {
		return common.Address{}, fmt.Errorf("invalid permit2 deadline: %w", err)
	}
	validAfter, err := v.parseDecimal(auth.Witness.ValidAfter)
	if err != nil {
		return common.Address{}, fmt.Errorf("invalid witness validAfter: %w", err)
	}
	if !common.IsHexAddress(auth.From) || !common.IsHexAddress(auth.Witness.To) ||
		!common.IsHexAddress(auth.Spender) || !common.IsHexAddress(auth.Permitted.Token) {
		return common.Address{}, errors.New("permit2 from/to/spender/token must be hex addresses")
	}

	// The amount is not checked here: the facilitator prices paid calls and
	// free-access signatures carry amount 0.
	if common.HexToAddress(auth.Witness.To) != common.HexToAddress(config.PayTo) {
		return common.Address{}, errors.New("witness recipient does not match payTo")
	}
	if common.HexToAddress(auth.Permitted.Token) != common.HexToAddress(config.Asset) {
		return common.Address{}, errors.New("permitted token does not match asset")
	}
	if common.HexToAddress(auth.Spender) != common.HexToAddress(x402.Permit2Spender) {
		return common.Address{}, errors.New("spender is not the x402 permit2 proxy")
	}
	now := big.NewInt(time.Now().Unix())
	if now.Cmp(validAfter) < 0 || now.Cmp(deadline) > 0 {
		return common.Address{}, gateway_interfaces.ErrX402AuthorizationExpired
	}

	parsedChainId, err := helpers.ParseCaip2(config.Network)
	if err != nil {
		return common.Address{}, err
	}
	chainId := int64(parsedChainId)

	digest, err := v.permitWitnessTransferFromDigest(auth, chainId)
	if err != nil {
		return common.Address{}, err
	}

	payer, err := v.recoverSigner(digest, payload.Payload.Signature)
	if err != nil {
		return common.Address{}, err
	}
	if payer != common.HexToAddress(auth.From) {
		return common.Address{}, errors.New("recovered signer does not match permit2 from")
	}
	return payer, nil
}

// recoverSigner decodes a 65-byte ECDSA signature (normalizing v 27/28→0/1)
// and returns the address that signed the EIP-712 digest.
func (v *X402PaymentVerifier) recoverSigner(digest []byte, signature string) (common.Address, error) {
	sigBytes, err := hex.DecodeString(strings.TrimPrefix(signature, "0x"))
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
	return crypto.PubkeyToAddress(*pubKey), nil
}

// transferWithAuthorizationDigest hashes the EIP-3009 typed data under the asset's EIP-712 domain.
func (v *X402PaymentVerifier) transferWithAuthorizationDigest(
	auth *x402.Eip3009Authorization,
	config *x402.X402Config,
	chainId int64,
) ([]byte, error) {
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

// permitWitnessTransferFromDigest hashes the Permit2 typed data under Permit2's
// fixed domain (NO version field) with the x402 {to, validAfter} witness bound in.
func (v *X402PaymentVerifier) permitWitnessTransferFromDigest(
	auth *x402.Permit2Authorization,
	chainId int64,
) ([]byte, error) {
	typedData := apitypes.TypedData{
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
			"permitted": apitypes.TypedDataMessage{
				"token":  auth.Permitted.Token,
				"amount": auth.Permitted.Amount,
			},
			"spender":  auth.Spender,
			"nonce":    auth.Nonce,
			"deadline": auth.Deadline,
			"witness": apitypes.TypedDataMessage{
				"to":         auth.Witness.To,
				"validAfter": auth.Witness.ValidAfter,
			},
		},
	}
	digest, _, err := apitypes.TypedDataAndHash(typedData)
	if err != nil {
		return nil, fmt.Errorf("failed to hash permit2 typed data: %w", err)
	}
	return digest, nil
}

func (v *X402PaymentVerifier) parseDecimal(s string) (*big.Int, error) {
	value, ok := new(big.Int).SetString(strings.TrimSpace(s), 10)
	if !ok {
		return nil, fmt.Errorf("%q is not a decimal integer", s)
	}
	return value, nil
}
