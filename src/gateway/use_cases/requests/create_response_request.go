package requests

import (
	"strconv"
	"strings"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/models/enums"
)

type InputMessageRequest struct {
	Id        string            `json:"id"`
	Author    string            `json:"author"`
	Role      enums.MessageRole `json:"role"`
	Content   string            `json:"content"`
	Timestamp time.Time         `json:"timestamp"`
}

type InputToolEndpointRequest struct {
	Kind       string `json:"kind"`
	Url        string `json:"url"`
	AuthHeader string `json:"authHeader,omitempty"`
}

type InputToolRequest struct {
	Name        string                    `json:"name"`
	Description string                    `json:"description"`
	Parameters  map[string]any            `json:"parameters"`
	Endpoint    *InputToolEndpointRequest `json:"endpoint,omitempty"`
}

type InputOptionsRequest struct {
	Temperature *float64 `json:"temperature,omitempty"`
	MaxTokens   *int     `json:"maxTokens,omitempty"`
	Stop        []string `json:"stop,omitempty"`
}

type CreateResponseRequest struct {
	ConversationId  string                     `json:"conversationId"`
	Platform        enums.ConversationPlatform `json:"platform"`
	Prompt          string                     `json:"prompt"`
	Input           []*InputMessageRequest     `json:"input"`
	ForceConclusion bool                       `json:"forceConclusion"`
	Tools           []*InputToolRequest        `json:"tools,omitempty"`
	Options         *InputOptionsRequest       `json:"options,omitempty"`
}

func (r *CreateResponseRequest) ValidateAndSanitize() error {
	if len(strings.TrimSpace(r.ConversationId)) == 0 {
		return errors.NewValidationError("conversationId", "cannot be empty")
	}

	if len(strings.TrimSpace(r.Prompt)) == 0 {
		return errors.NewValidationError("prompt", "cannot be empty")
	}

	if !enums.GetAllConversationPlatforms().Contains(r.Platform) {
		r.Platform = enums.ConversationPlatformNone
	}

	for index, message := range r.Input {
		if len(strings.TrimSpace(message.Id)) == 0 {
			return errors.NewValidationError("input.id", "cannot be empty, check input message at index "+strconv.Itoa(index))
		}
		if len(strings.TrimSpace(message.Content)) == 0 {
			return errors.NewValidationError("input.content", "cannot be empty, check input message at index "+strconv.Itoa(index))
		}
		if message.Timestamp.IsZero() {
			return errors.NewValidationError("input.timestamp", "cannot be empty, check input message at index "+strconv.Itoa(index))
		}
		if !enums.GetAllMessageRoles().Contains(message.Role) {
			return errors.NewValidationError("input.role", "invalid role, must be one of: "+enums.GetAllMessageRoles().String()+", check input message at index "+strconv.Itoa(index))
		}
	}

	for index, tool := range r.Tools {
		if len(strings.TrimSpace(tool.Name)) == 0 {
			return errors.NewValidationError("tools.name", "cannot be empty, check tool at index "+strconv.Itoa(index))
		}
		if tool.Endpoint != nil {
			if len(strings.TrimSpace(tool.Endpoint.Url)) == 0 {
				return errors.NewValidationError("tools.endpoint.url", "cannot be empty when endpoint is set, check tool at index "+strconv.Itoa(index))
			}
		}
	}

	return nil
}
