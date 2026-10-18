package requests

import (
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

// ConversationIdMaxLength matches the conversations.id column.
const ConversationIdMaxLength = 128

// ConversationRequest is embedded by every request addressing one conversation by route id.
type ConversationRequest struct {
	ConversationId string `uri:"conversationId" json:"-"`
}

func (r *ConversationRequest) ValidateAndSanitize() error {
	r.ConversationId = strings.TrimSpace(r.ConversationId)
	if len(r.ConversationId) == 0 {
		return errors.NewValidationError("conversationId", "cannot be empty")
	}
	if utf8.RuneCountInString(r.ConversationId) > ConversationIdMaxLength {
		return errors.NewValidationError(
			"conversationId",
			"must be at most "+strconv.Itoa(ConversationIdMaxLength)+" characters",
		)
	}
	return nil
}
