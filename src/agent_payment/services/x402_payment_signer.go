package services

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
	"time"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/settings"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/helpers"
	"github.com/6022-labs/agentic-agent-node/src/common/x402"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// nonceLength is the 32 bytes both an EIP-3009 nonce and a Permit2 uint256 nonce take.
const nonceLength = 32

// X402PaymentSigner signs an x402 payment, only ever for the quoted payee and amount.
type X402PaymentSigner struct {
	typedDataSigner        identity_interfaces.TypedDataSignerInterface
	permit2ApprovalEnsurer interfaces.Permit2ApprovalEnsurerInterface
	paymentSpendSettings   *settings.PaymentSpendSettings
}

func NewX402PaymentSigner(
	typedDataSigner identity_interfaces.TypedDataSignerInterface,
	permit2ApprovalEnsurer interfaces.Permit2ApprovalEnsurerInterface,
	paymentSpendSettings *settings.PaymentSpendSettings,
) *X402PaymentSigner {
	return &X402PaymentSigner{
		typedDataSigner:        typedDataSigner,
		permit2ApprovalEnsurer: permit2ApprovalEnsurer,
		paymentSpendSettings:   paymentSpendSettings,
	}
}

func (s *X402PaymentSigner) Sign(
	ctx context.Context,
	quote *interfaces.X402PaymentQuote,
) (*interfaces.X402SignedPayment, error) {
	chainId, err := helpers.ParseCaip2(quote.Network)
	if err != nil {
		return nil, errors.NewValidationError("network", err.Error())
	}

	amount, ok := new(big.Int).SetString(quote.Amount, 10)
	if !ok || amount.Sign() < 0 {
		return nil, errors.NewValidationError("amount", "must be a non-negative atomic amount")
	}
	payer, err := s.typedDataSigner.WalletAddress(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now().Unix()
	validAfter := now - s.paymentSpendSettings.ValidAfterSkew
	validBefore := now + s.paymentSpendSettings.ValiditySeconds

	nonce, err := s.randomNonce()
	if err != nil {
		return nil, err
	}

	var payload x402.ExactEvmPayload
	switch quote.AssetTransferMethod {
	case enums.AssetTransferMethodEip3009:
		payload, err = s.signEip3009(ctx, quote, payer.Hex(), amount, nonce, validAfter, validBefore, int64(chainId))
	case enums.AssetTransferMethodPermit2:
		payload, err = s.signPermit2(ctx, quote, payer.Hex(), amount, nonce, validAfter, validBefore, int64(chainId))
	default:
		return nil, errors.NewValidationError("assetTransferMethod", "must be one of eip3009, permit2")
	}
	if err != nil {
		return nil, err
	}

	header, err := x402.EncodeBase64Json(&x402.PaymentPayload{
		X402Version: x402.Version,
		Scheme:      x402.SchemeExact,
		Network:     quote.Network,
		Payload:     payload,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to encode payment payload: %w", err)
	}

	return &interfaces.X402SignedPayment{
		PaymentSignature: header,
		Payer:            payer.Hex(),
		Network:          quote.Network,
		Asset:            quote.Asset,
		PayTo:            quote.PayTo,
		Amount:           amount.String(),
		ValidBefore:      validBefore,
	}, nil
}

func (s *X402PaymentSigner) signEip3009(
	ctx context.Context,
	quote *interfaces.X402PaymentQuote,
	payer string,
	amount *big.Int,
	nonce []byte,
	validAfter int64,
	validBefore int64,
	chainId int64,
) (x402.ExactEvmPayload, error) {
	authorization := &x402.Eip3009Authorization{
		From:        payer,
		To:          quote.PayTo,
		Value:       amount.String(),
		ValidAfter:  strconv.FormatInt(validAfter, 10),
		ValidBefore: strconv.FormatInt(validBefore, 10),
		Nonce:       hexutil.Encode(nonce),
	}

	digest, err := x402.TransferWithAuthorizationDigest(
		authorization,
		quote.Asset,
		quote.Eip712Name,
		quote.Eip712Version,
		chainId,
	)
	if err != nil {
		return x402.ExactEvmPayload{}, err
	}

	signature, err := s.typedDataSigner.SignDigest(ctx, digest)
	if err != nil {
		return x402.ExactEvmPayload{}, err
	}

	return x402.ExactEvmPayload{
		Signature:     hexutil.Encode(signature),
		Authorization: authorization,
	}, nil
}

func (s *X402PaymentSigner) signPermit2(
	ctx context.Context,
	quote *interfaces.X402PaymentQuote,
	payer string,
	amount *big.Int,
	nonce []byte,
	validAfter int64,
	deadline int64,
	chainId int64,
) (x402.ExactEvmPayload, error) {
	// A signature the wallet has not approved Permit2 for is valid but unsettleable.
	if amount.Sign() > 0 {
		if err := s.permit2ApprovalEnsurer.EnsureApproval(ctx, quote.Network, quote.Asset, amount); err != nil {
			return x402.ExactEvmPayload{}, fmt.Errorf("permit2 approval: %w", err)
		}
	}

	authorization := &x402.Permit2Authorization{
		Permitted: x402.Permit2TokenPermissions{
			Token:  quote.Asset,
			Amount: amount.String(),
		},
		From:     payer,
		Spender:  x402.Permit2Spender,
		Nonce:    new(big.Int).SetBytes(nonce).String(),
		Deadline: strconv.FormatInt(deadline, 10),
		Witness: x402.Permit2Witness{
			To:         quote.PayTo,
			ValidAfter: strconv.FormatInt(validAfter, 10),
		},
	}

	digest, err := x402.PermitWitnessTransferFromDigest(authorization, chainId)
	if err != nil {
		return x402.ExactEvmPayload{}, err
	}

	signature, err := s.typedDataSigner.SignDigest(ctx, digest)
	if err != nil {
		return x402.ExactEvmPayload{}, err
	}

	return x402.ExactEvmPayload{
		Signature:            hexutil.Encode(signature),
		Permit2Authorization: authorization,
	}, nil
}

func (s *X402PaymentSigner) randomNonce() ([]byte, error) {
	nonce := make([]byte, nonceLength)
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}
	return nonce, nil
}
