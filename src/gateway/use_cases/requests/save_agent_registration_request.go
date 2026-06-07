package requests

import (
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

const defaultImageKey = "default"

// SaveAgentRegistrationRequest is the draft the connected wallet submits. Owner
// is injected by the gateway, not accepted here.
type SaveAgentRegistrationRequest struct {
	CollectionAddress string            `json:"collectionAddress"`
	Name              string            `json:"name"`
	Role              string            `json:"role"`
	CloneOf           *uint64           `json:"cloneOf"`
	Images            map[string]string `json:"images"`
	Attributes        map[string]string `json:"attributes"`
}

func (r *SaveAgentRegistrationRequest) ValidateAndSanitize() error {
	r.CollectionAddress = strings.TrimSpace(r.CollectionAddress)
	if !ethAddressPattern.MatchString(r.CollectionAddress) {
		return errors.NewValidationError("collectionAddress", "must be a 0x-prefixed 20-byte hex address")
	}

	r.Name = strings.TrimSpace(r.Name)
	if len(r.Name) == 0 {
		return errors.NewValidationError("name", "cannot be empty")
	}

	r.Role = strings.TrimSpace(r.Role)
	if len(r.Role) == 0 {
		return errors.NewValidationError("role", "cannot be empty")
	}

	if _, ok := r.Images[defaultImageKey]; !ok {
		return errors.NewValidationError("images", "must include a \"default\" entry")
	}

	return nil
}
