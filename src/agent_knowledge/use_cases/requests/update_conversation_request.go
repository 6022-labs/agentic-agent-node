package requests

import (
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/google/uuid"
)

// UpdateConversationRequest pairs the route id with the new label.
type UpdateConversationRequest struct {
	ConversationId string `uri:"conversationId" json:"-"`
	Name           string `                     json:"name"`
}

// ValidateAndSanitize rejects uuid-shaped names: they would poison the caller-side scan that reads a session
// id out of a completion.
func (r *UpdateConversationRequest) ValidateAndSanitize() error {
	r.ConversationId = strings.TrimSpace(r.ConversationId)
	if len(r.ConversationId) == 0 {
		return errors.NewValidationError("conversationId", "cannot be empty")
	}

	r.Name = strings.Join(strings.Fields(r.Name), " ")
	if len(r.Name) == 0 {
		return errors.NewValidationError("name", "cannot be empty")
	}
	if len([]rune(r.Name)) > ConversationNameMaxLength {
		return errors.NewValidationError("name", "must be at most 128 characters")
	}
	if _, err := uuid.Parse(r.Name); err == nil {
		return errors.NewValidationError("name", "must not be a uuid")
	}
	return nil
}
