package requests

import (
	"encoding/json"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

const cardSignaturesMember = "signatures"

// SignAgentCardRequest keeps Card raw so the signed bytes are exactly what the caller serialized.
type SignAgentCardRequest struct {
	Card json.RawMessage `json:"card" swaggertype:"object"`
}

func (r *SignAgentCardRequest) ValidateAndSanitize() error {
	var members map[string]json.RawMessage
	if err := json.Unmarshal(r.Card, &members); err != nil || members == nil {
		return errors.NewValidationError("card", "must be a JSON object")
	}
	if _, present := members[cardSignaturesMember]; present {
		return errors.NewValidationError("card", "must not carry a signatures member")
	}
	return nil
}
