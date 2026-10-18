package requests

import (
	"strconv"
	"strings"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/models/enums"
)

type AppendConversationMessageInput struct {
	Role      enums.MessageRole `json:"role"`
	Content   string            `json:"content"`
	Timestamp time.Time         `json:"timestamp"`
}

type AppendConversationMessagesRequest struct {
	ConversationRequest
	Messages []AppendConversationMessageInput `json:"messages"`
}

// appendableMessageRoles excludes system: out-of-band injection must not forge overlay rows.
var appendableMessageRoles = enums.MessageRoleList{
	enums.MessageRoleUser,
	enums.MessageRoleAssistant,
}

func (r *AppendConversationMessagesRequest) ValidateAndSanitize() error {
	if err := r.ConversationRequest.ValidateAndSanitize(); err != nil {
		return err
	}
	if len(r.Messages) == 0 {
		return errors.NewValidationError("messages", "must contain at least one message")
	}

	for index, message := range r.Messages {
		if len(strings.TrimSpace(message.Content)) == 0 {
			return errors.NewValidationError(
				"messages.content",
				"cannot be empty, check message at index "+strconv.Itoa(index),
			)
		}
		if message.Timestamp.IsZero() {
			return errors.NewValidationError(
				"messages.timestamp",
				"cannot be empty, check message at index "+strconv.Itoa(index),
			)
		}
		if !appendableMessageRoles.Contains(message.Role) {
			return errors.NewValidationError(
				"messages.role",
				"invalid role, must be one of: "+appendableMessageRoles.String()+", check message at index "+strconv.Itoa(
					index,
				),
			)
		}
	}

	return nil
}
