package requests

import (
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

// UpdateConversationRequest pairs the route id with the new label.
type UpdateConversationRequest struct {
	ConversationId string `uri:"conversationId" json:"-"`
	Name           string `                     json:"name"`
}

func (r *UpdateConversationRequest) ValidateAndSanitize() error {
	r.ConversationId = strings.TrimSpace(r.ConversationId)
	if len(r.ConversationId) == 0 {
		return errors.NewValidationError("conversationId", "cannot be empty")
	}

	r.Name = strings.Join(strings.Fields(r.Name), " ")
	if len(r.Name) == 0 {
		return errors.NewValidationError("name", "cannot be empty")
	}
	return nil
}
