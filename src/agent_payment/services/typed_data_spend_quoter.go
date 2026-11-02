package services

import (
	"encoding/json"
	"math/big"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/helpers"
	"github.com/6022-labs/agentic-agent-node/src/common/x402"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

// TypedDataSpendQuoter trusts the primary type's name, since the token contract checks its type hash.
type TypedDataSpendQuoter struct{}

func NewTypedDataSpendQuoter() *TypedDataSpendQuoter {
	return &TypedDataSpendQuoter{}
}

func (q *TypedDataSpendQuoter) Quote(typedData *apitypes.TypedData) (*interfaces.X402PaymentQuote, *big.Int, error) {
	message := map[string]any(typedData.Message)
	switch typedData.PrimaryType {
	case "TransferWithAuthorization", "ReceiveWithAuthorization":
		return q.quote(typedData, typedData.Domain.VerifyingContract, message["to"], message["value"],
			enums.AssetTransferMethodEip3009)
	case "Permit":
		return q.quote(typedData, typedData.Domain.VerifyingContract, message["spender"], q.permitValue(message), "")
	case "PermitTransferFrom", "PermitWitnessTransferFrom":
		permitted := q.object(message["permitted"])
		return q.quote(typedData, permitted["token"], q.permit2Payee(message), permitted["amount"],
			enums.AssetTransferMethodPermit2)
	case "PermitSingle":
		details := q.object(message["details"])
		return q.quote(typedData, details["token"], message["spender"], details["amount"],
			enums.AssetTransferMethodPermit2)
	case "PermitBatch", "PermitBatchTransferFrom", "PermitBatchWitnessTransferFrom":
		return nil, nil, errors.NewValidationError(
			"typedData.primaryType",
			"batch token permits cannot be checked against the spend policy; sign one token at a time",
		)
	default:
		return nil, nil, nil
	}
}

func (q *TypedDataSpendQuoter) quote(
	typedData *apitypes.TypedData,
	asset any,
	payTo any,
	value any,
	assetTransferMethod enums.AssetTransferMethod,
) (*interfaces.X402PaymentQuote, *big.Int, error) {
	if typedData.Domain.ChainId == nil {
		return nil, nil, errors.NewValidationError("typedData.domain.chainId", "is required for a token payload")
	}
	chainId := (*big.Int)(typedData.Domain.ChainId)
	if !chainId.IsUint64() || chainId.Sign() == 0 {
		return nil, nil, errors.NewValidationError("typedData.domain.chainId", "must be a positive chain id")
	}

	assetAddress, ok := asset.(string)
	if !ok || !common.IsHexAddress(assetAddress) {
		return nil, nil, errors.NewValidationError("typedData", "the token address is missing or malformed")
	}
	payToAddress, ok := payTo.(string)
	if !ok || !common.IsHexAddress(payToAddress) {
		return nil, nil, errors.NewValidationError("typedData", "the payee or spender address is missing or malformed")
	}
	amount, ok := q.amount(value)
	if !ok {
		return nil, nil, errors.NewValidationError("typedData", "the amount is missing or not a uint256")
	}

	return &interfaces.X402PaymentQuote{
		Network:             helpers.FormatCaip2(chainId.Uint64()),
		Asset:               assetAddress,
		PayTo:               payToAddress,
		Amount:              amount.String(),
		AssetTransferMethod: assetTransferMethod,
	}, amount, nil
}

// permit2Payee is the x402 witness recipient when the x402 proxy spends, else the spender, who picks the recipient.
func (q *TypedDataSpendQuoter) permit2Payee(message map[string]any) any {
	spender, _ := message["spender"].(string)
	if strings.EqualFold(strings.TrimSpace(spender), x402.Permit2Spender) {
		return q.object(message["witness"])["to"]
	}
	return message["spender"]
}

// permitValue reads a DAI-style permit's allowed flag as an unlimited or revoked allowance.
func (q *TypedDataSpendQuoter) permitValue(message map[string]any) any {
	allowed, found := message["allowed"]
	if !found {
		return message["value"]
	}
	if allowed == true {
		return math.MaxBig256.String()
	}
	return "0"
}

func (q *TypedDataSpendQuoter) object(value any) map[string]any {
	if typed, ok := value.(map[string]any); ok {
		return typed
	}
	return map[string]any{}
}

// amount accepts the integer forms the EIP-712 encoder accepts, bounded to uint256.
func (q *TypedDataSpendQuoter) amount(value any) (*big.Int, bool) {
	var parsed *big.Int
	switch typed := value.(type) {
	case string:
		parsed, _ = math.ParseBig256(strings.TrimSpace(typed))
	case json.Number:
		parsed, _ = math.ParseBig256(typed.String())
	case float64:
		// Beyond 2^53 a float64 no longer holds the integer the caller meant.
		if typed >= 0 && typed <= 1<<53 && typed == float64(int64(typed)) {
			parsed = big.NewInt(int64(typed))
		}
	case *big.Int:
		parsed = typed
	case *math.HexOrDecimal256:
		parsed = (*big.Int)(typed)
	}
	if parsed == nil {
		return nil, false
	}
	return helpers.ParseUint256(parsed.String())
}
