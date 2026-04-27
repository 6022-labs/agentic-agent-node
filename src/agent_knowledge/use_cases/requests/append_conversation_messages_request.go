package requests

import (
	"time"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

type AppendConversationMessagesRequest struct {
	Messages []ConversationMessageInput `json:"messages"`
}

// ConversationMessageInput carries the semantic content of a message.
// No id field — the runtime generates a uuid on insert and uses the SHA-512
// hash of the semantic fields as the dedupe key.
type ConversationMessageInput struct {
	Author    string    `json:"author"`
	Role      string    `json:"role"`
	Platform  string    `json:"platform,omitempty"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

func (r *AppendConversationMessagesRequest) ValidateAndSanitize() error {
	if len(r.Messages) == 0 {
		return errors.NewValidationError("messages", "must contain at least one message")
	}
	return nil
}
