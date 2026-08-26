package requests

import (
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

// CreateConversationRequest names a new memory session; the runtime stays the validation authority.
type CreateConversationRequest struct {
	Name string `json:"name"`
}

func (r *CreateConversationRequest) ValidateAndSanitize() error {
	r.Name = strings.Join(strings.Fields(r.Name), " ")
	if len(r.Name) == 0 {
		return errors.NewValidationError("name", "cannot be empty")
	}
	return nil
}
