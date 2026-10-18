package requests

import (
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/google/uuid"
)

// ConversationNameMaxLength matches the conversations.name column.
const ConversationNameMaxLength = 128

// ConversationNameRequest is embedded by every request carrying a conversation label.
type ConversationNameRequest struct {
	Name string `json:"name"`
}

// ValidateAndSanitize rejects uuid-shaped names: they would poison the caller-side session id scan.
func (r *ConversationNameRequest) ValidateAndSanitize() error {
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
