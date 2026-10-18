package requests

import (
	"encoding/json"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

type CreateResponseRequest struct {
	Model              string            `json:"model,omitempty"`
	Instructions       string            `json:"instructions,omitempty"`
	Input              ResponseInput     `json:"input"`
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
	Role    string
	Content string
}

type ResponseInput struct {
	Messages []InputMessage
}

func (r *ResponseInput) UnmarshalJSON(data []byte) error {
	var asString string
	if err := json.Unmarshal(data, &asString); err == nil {
		r.Messages = []InputMessage{{Role: "user", Content: asString}}
		return nil
	}

	var items []struct {
		Type    string       `json:"type"`
		Role    string       `json:"role"`
		Content inputContent `json:"content"`
	}
	if err := json.Unmarshal(data, &items); err != nil {
		return err
	}

	r.Messages = make([]InputMessage, 0, len(items))
	for _, item := range items {
		if item.Type != "" && item.Type != "message" {
			continue
		}
		role := item.Role
		if role == "" {
			role = "user"
		}
		r.Messages = append(r.Messages, InputMessage{Role: role, Content: item.Content.Text})
	}
	return nil
}

type inputContent struct {
	Text string
}

func (c *inputContent) UnmarshalJSON(data []byte) error {
	var asString string
	if err := json.Unmarshal(data, &asString); err == nil {
		c.Text = asString
		return nil
	}

	var parts []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(data, &parts); err != nil {
		return err
	}

	var builder strings.Builder
	for _, part := range parts {
		switch part.Type {
		case "", "text", "input_text", "output_text":
			builder.WriteString(part.Text)
		}
	}
	c.Text = builder.String()
	return nil
}

// ConversationKeyMaxLength matches the runtime's conversations.id column.
const ConversationKeyMaxLength = 128

type ConversationRef struct {
	Id string `json:"id"`
}

func (c *ConversationRef) UnmarshalJSON(data []byte) error {
	var asString string
	if err := json.Unmarshal(data, &asString); err == nil {
		c.Id = asString
		return nil
	}

	var object struct {
		Id string `json:"id"`
	}
	if err := json.Unmarshal(data, &object); err != nil {
		return err
	}
	c.Id = object.Id
	return nil
}

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
	if len(r.Input.Messages) == 0 {
		return errors.NewValidationError("input", "cannot be empty")
	}
	for index := range r.Input.Messages {
		if len(strings.TrimSpace(r.Input.Messages[index].Content)) == 0 {
			return errors.NewValidationError(
				"input",
				"message content cannot be empty, check input at index "+strconv.Itoa(index),
			)
		}
	}
	if utf8.RuneCountInString(r.ConversationKey()) > ConversationKeyMaxLength {
		return errors.NewValidationError(
			"conversation",
			"key must be at most "+strconv.Itoa(ConversationKeyMaxLength)+" characters",
		)
	}
	return nil
}
