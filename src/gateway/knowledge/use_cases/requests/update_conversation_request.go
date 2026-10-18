package requests

import (
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

// UpdateConversationRequest pairs the route id with the new label.
type UpdateConversationRequest struct {
	ConversationRequest
	Name string `json:"name"`
}

func (r *UpdateConversationRequest) ValidateAndSanitize() error {
	if err := r.ConversationRequest.ValidateAndSanitize(); err != nil {
		return err
	}

	r.Name = strings.Join(strings.Fields(r.Name), " ")
	if len(r.Name) == 0 {
		return errors.NewValidationError("name", "cannot be empty")
	}
	return nil
}
