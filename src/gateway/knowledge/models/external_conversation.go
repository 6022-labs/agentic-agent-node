package models

import "time"

// Mirrors the runtime's conversation list/messages wire shapes.

type ExternalConversationPayload struct {
	ConversationId string    `json:"conversationId"`
	Name           *string   `json:"name,omitempty"`
	MessageCount   int64     `json:"messageCount"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type ExternalListConversationsPayload struct {
	Page            uint                          `json:"page"`
	ElementsPerPage uint                          `json:"elementsPerPage"`
	TotalElements   uint                          `json:"totalElements"`
	Conversations   []ExternalConversationPayload `json:"conversations"`
}

type ExternalCreateConversationPayload struct {
	Name string `json:"name"`
}

type ExternalUpdateConversationPayload struct {
	Name string `json:"name"`
}

type ExternalConversationMessage struct {
	Id             string    `json:"id"`
	ConversationId string    `json:"conversationId"`
	Role           string    `json:"role"`
	Content        *string   `json:"content,omitempty"`
	ToolCalls      *string   `json:"toolCalls,omitempty"`
	ToolCallId     *string   `json:"toolCallId,omitempty"`
	Timestamp      time.Time `json:"timestamp"`
}

type ExternalListConversationMessagesPayload struct {
	Messages []ExternalConversationMessage `json:"messages"`
}

type ExternalConversationMessageInput struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

type ExternalAppendConversationMessagesPayload struct {
	Messages []ExternalConversationMessageInput `json:"messages"`
}

type ExternalAppendConversationMessagesResultPayload struct {
	Inserted []ExternalConversationMessage `json:"inserted"`
}
