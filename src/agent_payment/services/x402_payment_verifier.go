package services

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/common/helpers"
	"github.com/6022-labs/agentic-agent-node/src/common/x402"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const ethSignatureLength = 65

type X402PaymentVerifier struct{}

func NewX402PaymentVerifier() *X402PaymentVerifier {
	return &X402PaymentVerifier{}
}

func (v *X402PaymentVerifier) RecoverPayer(
	payload *x402.PaymentPayload,
	config *models.PaymentReceiveRuleX402Config,
) (common.Address, error) {
	switch config.AssetTransferMethod.String() {
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
	config *models.PaymentReceiveRuleX402Config,
) (common.Address, error) {
	auth := payload.Payload.Eip3009Authorization
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

	// The facilitator prices paid calls, and free-access signatures carry value 0.
	if common.HexToAddress(auth.To) != common.HexToAddress(config.PayTo) {
		return common.Address{}, errors.New("authorization recipient does not match payTo")
	}
	now := big.NewInt(time.Now().Unix())
	if now.Cmp(validAfter) < 0 || now.Cmp(validBefore) > 0 {
		return common.Address{}, payment_interfaces.ErrX402AuthorizationExpired
	}

	parsedChainId, err := helpers.ParseCaip2(config.Network)
	if err != nil {
		return common.Address{}, err
	}
	chainId := int64(parsedChainId)

	digest, err := x402.TransferWithAuthorizationDigest(
		auth,
		config.Asset,
		v.orEmpty(config.Eip712Name),
		v.orEmpty(config.Eip712Version),
		chainId,
	)
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

// recoverPermit2 validates a PermitWitnessTransferFrom signature under Permit2's fixed domain.
func (v *X402PaymentVerifier) recoverPermit2(
	payload *x402.PaymentPayload,
	config *models.PaymentReceiveRuleX402Config,
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

	// The facilitator prices paid calls, and free-access signatures carry amount 0.
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
		return common.Address{}, payment_interfaces.ErrX402AuthorizationExpired
	}

	parsedChainId, err := helpers.ParseCaip2(config.Network)
	if err != nil {
		return common.Address{}, err
	}
	chainId := int64(parsedChainId)

	digest, err := x402.PermitWitnessTransferFromDigest(auth, chainId)
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

// recoverSigner decodes a 65-byte signature (v 27/28→0/1) and returns the EIP-712 signer.
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

// orEmpty reads an optional domain field; permit2 rules leave both nil.
func (v *X402PaymentVerifier) orEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func (v *X402PaymentVerifier) parseDecimal(s string) (*big.Int, error) {
	value, ok := new(big.Int).SetString(strings.TrimSpace(s), 10)
	if !ok {
		return nil, fmt.Errorf("%q is not a decimal integer", s)
	}
	return value, nil
}
