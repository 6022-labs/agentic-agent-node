package responses

import "time"

type ConversationResponse struct {
	ConversationId string    `json:"conversationId"`
	Name           *string   `json:"name,omitempty"`
	MessageCount   int64     `json:"messageCount"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type ListConversationsResponse struct {
	Page            uint                   `json:"page"`
	ElementsPerPage uint                   `json:"elementsPerPage"`
	TotalElements   uint                   `json:"totalElements"`
	Conversations   []ConversationResponse `json:"conversations"`
}

type ConversationMessage struct {
	Id             string    `json:"id"`
	ConversationId string    `json:"conversationId"`
	Role           string    `json:"role"`
	Content        *string   `json:"content,omitempty"`
	ToolCalls      *string   `json:"toolCalls,omitempty"`
	ToolCallId     *string   `json:"toolCallId,omitempty"`
	Timestamp      time.Time `json:"timestamp"`
}

type ListConversationMessagesResponse struct {
	Messages []ConversationMessage `json:"messages"`
}

type AppendConversationMessagesResponse struct {
	Inserted []ConversationMessage `json:"inserted"`
}
