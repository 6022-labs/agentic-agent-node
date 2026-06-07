package models

import "time"

type ExternalCreateResponseInputPayload struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

type ExternalCreateResponsePayload struct {
	ConversationId string                               `json:"conversationId"`
	Inputs         []ExternalCreateResponseInputPayload `json:"inputs"`
}

type ExternalCreateResponseMessagePayload struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

type ExternalCreateResponseResultPayload struct {
	Message *ExternalCreateResponseMessagePayload `json:"message,omitempty"`
}
