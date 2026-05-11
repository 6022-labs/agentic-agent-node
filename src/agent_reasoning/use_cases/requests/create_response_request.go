package requests

import (
	"strconv"
	"strings"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

type CreateResponseRequestInput struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

type CreateResponseRequest struct {
	ConversationId string                       `json:"conversationId"`
	Inputs         []CreateResponseRequestInput `json:"inputs"`
}

func (r *CreateResponseRequest) ValidateAndSanitize() error {
	r.ConversationId = strings.TrimSpace(r.ConversationId)
	if len(r.ConversationId) == 0 {
		return errors.NewValidationError("conversationId", "cannot be empty")
	}
	if len(r.Inputs) == 0 {
		return errors.NewValidationError("inputs", "must contain at least one message")
	}
	for i, m := range r.Inputs {
		if len(strings.TrimSpace(m.Role)) == 0 {
			return errors.NewValidationError("inputs.role", "cannot be empty, check input message at index "+strconv.Itoa(i))
		}
		if len(strings.TrimSpace(m.Content)) == 0 {
			return errors.NewValidationError("inputs.content", "cannot be empty, check input message at index "+strconv.Itoa(i))
		}
		if m.Timestamp.IsZero() {
			return errors.NewValidationError("inputs.timestamp", "cannot be empty, check input message at index "+strconv.Itoa(i))
		}
	}
	return nil
}
