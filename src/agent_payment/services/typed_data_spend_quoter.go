package services

import (
	"encoding/json"
	"math/big"
	"strings"

	"github.com/holiman/uint256"

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
		if err := q.requireDeclared(typedData, []string{"to"}, []string{"value"}); err != nil {
			return nil, nil, err
		}
		return q.quote(typedData, typedData.Domain.VerifyingContract, message["to"], message["value"],
			enums.AssetTransferMethodEip3009)
	case "Permit":
		if err := q.requireDeclared(typedData, []string{"spender"}); err != nil {
			return nil, nil, err
		}
		value, err := q.permitValue(typedData)
		if err != nil {
			return nil, nil, err
		}
		return q.quote(typedData, typedData.Domain.VerifyingContract, message["spender"], value, "")
	case "PermitTransferFrom", "PermitWitnessTransferFrom":
		if err := q.requirePermit2Domain(typedData); err != nil {
			return nil, nil, err
		}
		paths := [][]string{{"permitted", "token"}, {"permitted", "amount"}, {"spender"}}
		if q.isX402Proxy(message["spender"]) {
			paths = append(paths, []string{"witness", "to"})
		}
		if err := q.requireDeclared(typedData, paths...); err != nil {
			return nil, nil, err
		}
		permitted := q.object(message["permitted"])
		return q.quote(typedData, permitted["token"], q.permit2Payee(message), permitted["amount"],
			enums.AssetTransferMethodPermit2)
	case "PermitSingle":
		if err := q.requirePermit2Domain(typedData); err != nil {
			return nil, nil, err
		}
		if err := q.requireDeclared(
			typedData, []string{"details", "token"}, []string{"details", "amount"}, []string{"spender"},
		); err != nil {
			return nil, nil, err
		}
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

// requirePermit2Domain refuses a Permit2-shaped payload under another contract, which would charge the wrong token.
func (q *TypedDataSpendQuoter) requirePermit2Domain(typedData *apitypes.TypedData) error {
	if typedData.Domain.Name != x402.Permit2DomainName ||
		!strings.EqualFold(strings.TrimSpace(typedData.Domain.VerifyingContract), x402.Permit2ContractAddress) {
		return errors.NewValidationError(
			"typedData.domain",
			"a Permit2 payload must be signed for the Permit2 contract",
		)
	}
	return nil
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
	// Only declared domain fields are hashed, so an undeclared contract would not bind the signature to the token.
	if !q.declaresDomainOnce(typedData, "chainId", "uint256") ||
		!q.declaresDomainOnce(typedData, "verifyingContract", "address") {
		return nil, nil, errors.NewValidationError(
			"typedData.domain",
			"a token payload must declare chainId and verifyingContract once in EIP712Domain",
		)
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
	if q.isX402Proxy(message["spender"]) {
		return q.object(message["witness"])["to"]
	}
	return message["spender"]
}

func (q *TypedDataSpendQuoter) isX402Proxy(spender any) bool {
	address, _ := spender.(string)
	return strings.EqualFold(strings.TrimSpace(address), x402.Permit2Spender)
}

// requireDeclared refuses a quote read from a field the types don't declare, since only declared fields are signed.
func (q *TypedDataSpendQuoter) requireDeclared(typedData *apitypes.TypedData, paths ...[]string) error {
	for _, path := range paths {
		current := typedData.PrimaryType
		for _, name := range path {
			next := ""
			for _, declared := range typedData.Types[current] {
				if declared.Name == name {
					next = declared.Type
					break
				}
			}
			if next == "" {
				return errors.NewValidationError(
					"typedData",
					"the payload's "+strings.Join(path, ".")+" is not declared in its types",
				)
			}
			current = next
		}
	}
	return nil
}

// permitValue reads the amount the signed Permit type declares; a bool allowed is DAI's unlimited-or-revoke flag.
func (q *TypedDataSpendQuoter) permitValue(typedData *apitypes.TypedData) (any, error) {
	message := map[string]any(typedData.Message)
	declaresValue := strings.HasPrefix(q.declaredType(typedData, "value"), "uint")
	if q.declaredType(typedData, "allowed") == "bool" {
		if message["allowed"] == true {
			return math.MaxBig256.String(), nil
		}
		// A type declaring both still moves value, so the revoke reading must not hide it.
		if declaresValue {
			return message["value"], nil
		}
		return "0", nil
	}
	if declaresValue {
		return message["value"], nil
	}
	return nil, errors.NewValidationError("typedData", "the permit declares no amount the spend policy can read")
}

func (q *TypedDataSpendQuoter) declaredType(typedData *apitypes.TypedData, field string) string {
	for _, declared := range typedData.Types[typedData.PrimaryType] {
		if declared.Name == field {
			return declared.Type
		}
	}
	return ""
}

func (q *TypedDataSpendQuoter) declaresDomainOnce(typedData *apitypes.TypedData, field string, fieldType string) bool {
	count := 0
	for _, declared := range typedData.Types["EIP712Domain"] {
		if declared.Name != field {
			continue
		}
		if declared.Type != fieldType {
			return false
		}
		count++
	}
	return count == 1
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
	if parsed == nil || parsed.Sign() < 0 {
		return nil, false
	}
	if _, overflow := uint256.FromBig(parsed); overflow {
		return nil, false
	}
	return parsed, true
}
