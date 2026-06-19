package requests

import (
	"encoding/json"
	"mime/multipart"
	"regexp"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

const maxImageBytes = 5 * 1024 * 1024

var ethAddressPattern = regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`)

// SaveAgentRegistrationRequest is the multipart draft the connected wallet
// submits. Owner is injected by the gateway, not accepted here. Image is the
// single base image the runtime frames; Attributes arrives as a JSON object
// string in the "attributes" form field.
type SaveAgentRegistrationRequest struct {
	ChainId           string                `form:"chainId"`
	CollectionAddress string                `form:"collectionAddress"`
	Name              string                `form:"name"`
	Role              string                `form:"role"`
	CloneOf           *uint64               `form:"cloneOf"`
	AttributesRaw     string                `form:"attributes"`
	Image             *multipart.FileHeader `form:"image"`

	Attributes map[string]string `form:"-"`
}

func (r *SaveAgentRegistrationRequest) ValidateAndSanitize() error {
	r.ChainId = strings.TrimSpace(r.ChainId)
	if len(r.ChainId) == 0 {
		return errors.NewValidationError("chainId", "is required")
	}

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

	if r.Image == nil {
		return errors.NewValidationError("image", "is required")
	}
	if r.Image.Size > maxImageBytes {
		return errors.NewValidationError("image", "must be at most 5MB")
	}

	if strings.TrimSpace(r.AttributesRaw) != "" {
		if err := json.Unmarshal([]byte(r.AttributesRaw), &r.Attributes); err != nil {
			return errors.NewValidationError("attributes", "must be a JSON object of string values")
		}
	}

	return nil
}
