package requests

import (
	"strconv"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

// CreateResponseRequest is the /internal/responses body. The gateway normalizes
// the OpenAI Responses request before forwarding, so input arrives as a plain
// array of message items and conversation as a plain {id} — no union legs here.
type CreateResponseRequest struct {
	Model              string            `json:"model,omitempty"`
	Instructions       string            `json:"instructions,omitempty"`
	Input              []InputMessage    `json:"input"`
	Conversation       *ConversationRef  `json:"conversation,omitempty"`
	PreviousResponseId string            `json:"previous_response_id,omitempty"`
	Temperature        *float64          `json:"temperature,omitempty"`
	TopP               *float64          `json:"top_p,omitempty"`
	MaxOutputTokens    *int              `json:"max_output_tokens,omitempty"`
	Store              *bool             `json:"store,omitempty"`
	Stream             bool              `json:"stream,omitempty"`
	Metadata           map[string]string `json:"metadata,omitempty"`
}

type InputMessage struct {
	Type    string `json:"type,omitempty"`
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ConversationRef struct {
	Id string `json:"id"`
}

// ConversationKey is the history key — previous_response_id wins, then
// conversation. Empty means a fresh conversation.
func (r *CreateResponseRequest) ConversationKey() string {
	if key := strings.TrimSpace(r.PreviousResponseId); key != "" {
		return key
	}
	if r.Conversation != nil {
		return strings.TrimSpace(r.Conversation.Id)
	}
	return ""
}

func (r *CreateResponseRequest) ValidateAndSanitize() error {
	if len(r.Input) == 0 {
		return errors.NewValidationError("input", "cannot be empty")
	}
	for index := range r.Input {
		if len(strings.TrimSpace(r.Input[index].Content)) == 0 {
			return errors.NewValidationError("input", "message content cannot be empty, check input at index "+strconv.Itoa(index))
		}
	}
	return nil
}
