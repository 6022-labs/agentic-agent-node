package requests

import (
	"encoding/json"
	"mime/multipart"
	"regexp"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/helpers"
)

const maxImageBytes = 5 * 1024 * 1024

// ensLabelCharsPattern mirrors ENSUtils.isValidENSName's character set; length and hyphen-position rules are checked alongside.
var ensLabelCharsPattern = regexp.MustCompile(`^[a-z0-9-]+$`)

// SaveAgentRegistrationRequest is the multipart draft form. Owner is injected by
// the gateway from the session; Addresses are never accepted (it's the
// node's). Image is the single base image the node frames into the NFT set;
// Attributes arrives as a JSON object string in the "attributes" form field.
type SaveAgentRegistrationRequest struct {
	ChainId           string                `form:"chainId"`
	CollectionAddress string                `form:"collectionAddress"`
	Owner             string                `form:"owner"`
	Name              string                `form:"name"`
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
	if !helpers.IsEthAddress(r.CollectionAddress) {
		return errors.NewValidationError("collectionAddress", "must be a 0x-prefixed 20-byte hex address")
	}

	r.Owner = strings.TrimSpace(r.Owner)
	if !helpers.IsEthAddress(r.Owner) {
		return errors.NewValidationError("owner", "must be a 0x-prefixed 20-byte hex address")
	}

	if len(strings.TrimSpace(r.Name)) == 0 {
		return errors.NewValidationError("name", "must not be empty")
	}
	if len(r.Name) > 63 || r.Name[0] == '-' || r.Name[len(r.Name)-1] == '-' ||
		strings.Contains(r.Name, "--") || !ensLabelCharsPattern.MatchString(r.Name) {
		return errors.NewValidationError("name", "must be a valid ENS label: lowercase letters, digits and hyphens; no leading, trailing or repeated hyphens; 1-63 chars")
	}

	if r.Image != nil && r.Image.Size > maxImageBytes {
		return errors.NewValidationError("image", "must be at most 5MB")
	}

	r.Attributes = map[string]string{}
	if strings.TrimSpace(r.AttributesRaw) != "" {
		if err := json.Unmarshal([]byte(r.AttributesRaw), &r.Attributes); err != nil {
			return errors.NewValidationError("attributes", "must be a JSON object of string values")
		}
	}
	// The contract rejects mints missing a required attribute.
	if strings.TrimSpace(r.Attributes["role"]) == "" {
		return errors.NewValidationError("attributes", `must include a non-empty "role"`)
	}

	return nil
}
