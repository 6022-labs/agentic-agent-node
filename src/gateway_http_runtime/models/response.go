package models

import "time"

type CreateResponseInputPayload struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

type CreateResponsePayload struct {
	ConversationId string                       `json:"conversationId"`
	Inputs         []CreateResponseInputPayload `json:"inputs"`
}

type CreateResponseMessagePayload struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

type CreateResponseResultPayload struct {
	Message *CreateResponseMessagePayload `json:"message,omitempty"`
}
