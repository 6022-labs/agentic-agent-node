package requests

import (
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/google/uuid"
)

// UpdateSelfToolRequest toggles one self tool; Enabled is a pointer so an absent field fails instead of disabling.
type UpdateSelfToolRequest struct {
	Id      uuid.UUID `uri:"id" json:"-"`
	Enabled *bool     `         json:"enabled"`
}

func (r *UpdateSelfToolRequest) ValidateAndSanitize() error {
	if r.Enabled == nil {
		return errors.NewValidationError("enabled", "enabled is required")
	}
	return nil
}
