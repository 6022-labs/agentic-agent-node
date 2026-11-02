package requests

import (
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

// SignTypedDataRequest carries an EIP-712 payload to sign.
type SignTypedDataRequest struct {
	TypedData *apitypes.TypedData `json:"typedData"`
}

func (r *SignTypedDataRequest) ValidateAndSanitize() error {
	if r.TypedData == nil {
		return errors.NewValidationError("typedData", "is required")
	}
	if len(strings.TrimSpace(r.TypedData.PrimaryType)) == 0 {
		return errors.NewValidationError("typedData.primaryType", "is required")
	}
	if _, found := r.TypedData.Types[r.TypedData.PrimaryType]; !found {
		return errors.NewValidationError("typedData.types", "must declare the primaryType")
	}
	if _, found := r.TypedData.Types["EIP712Domain"]; !found {
		return errors.NewValidationError("typedData.types", "must declare EIP712Domain")
	}
	return nil
}
