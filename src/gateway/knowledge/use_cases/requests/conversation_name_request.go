package requests

import (
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

// ConversationNameRequest is embedded by every request carrying a conversation label.
type ConversationNameRequest struct {
	Name string `json:"name"`
}

func (r *ConversationNameRequest) ValidateAndSanitize() error {
	r.Name = strings.Join(strings.Fields(r.Name), " ")
	if len(r.Name) == 0 {
		return errors.NewValidationError("name", "cannot be empty")
	}
	return nil
}
