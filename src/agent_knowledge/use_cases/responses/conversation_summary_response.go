package responses

import "time"

type ConversationSummary struct {
	ConversationId string    `json:"conversationId"`
	MessageCount   int64     `json:"messageCount"`
	FirstMessageAt time.Time `json:"firstMessageAt"`
	LastMessageAt  time.Time `json:"lastMessageAt"`
}

type ListConversationsResponse struct {
	Conversations []ConversationSummary `json:"conversations"`
}
