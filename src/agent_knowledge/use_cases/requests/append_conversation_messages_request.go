package requests

import (
	"strconv"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

type AppendConversationMessagesRequest struct {
	Messages []ConversationMessageInput `json:"messages"`
}

// ConversationMessageInput carries the semantic content of a message.
// No id field — the runtime generates a uuid on insert and uses the SHA-512
// hash of the semantic fields as the dedupe key.
type ConversationMessageInput struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// appendableRoles excludes system (the built prompt owns that slot) and tool (orphan rows break pairing).
var appendableRoles = enums.ConversationMessageRoleList{
	enums.ConversationMessageRoleUser,
	enums.ConversationMessageRoleAssistant,
}

func (r *AppendConversationMessagesRequest) ValidateAndSanitize() error {
	if len(r.Messages) == 0 {
		return errors.NewValidationError("messages", "must contain at least one message")
	}
	for index, message := range r.Messages {
		if !appendableRoles.Contains(enums.ConversationMessageRole(message.Role)) {
			return errors.NewValidationError(
				"messages.role",
				"invalid role, must be one of: "+appendableRoles.String()+", check message at index "+strconv.Itoa(
					index,
				),
			)
		}
	}
	return nil
}
