package requests

import (
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

const defaultImageKey = "default"

// SaveAgentRegistrationRequest is the draft form. Owner is injected by the
// gateway from the session; WalletAddress is never accepted (it's the node's).
type SaveAgentRegistrationRequest struct {
	CollectionAddress string            `json:"collectionAddress"`
	Owner             string            `json:"owner"`
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

	r.Owner = strings.TrimSpace(r.Owner)
	if !ethAddressPattern.MatchString(r.Owner) {
		return errors.NewValidationError("owner", "must be a 0x-prefixed 20-byte hex address")
	}

	r.Name = strings.TrimSpace(r.Name)
	if len(r.Name) == 0 {
		return errors.NewValidationError("name", "must not be empty")
	}

	var role enums.AgentRole
	if err := role.Set(r.Role); err != nil {
		return errors.NewValidationError("role", "must be one of: "+enums.GetAllAgentRoles().String())
	}
	r.Role = role.String()

	if _, ok := r.Images[defaultImageKey]; !ok {
		return errors.NewValidationError("images", "must include a \"default\" entry")
	}

	return nil
}
