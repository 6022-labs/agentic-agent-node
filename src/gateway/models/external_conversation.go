package models

import "time"

// Mirrors the runtime's conversation list/messages wire shapes.

type ExternalConversationSummary struct {
	ConversationId string    `json:"conversationId"`
	MessageCount   int64     `json:"messageCount"`
	FirstMessageAt time.Time `json:"firstMessageAt"`
	LastMessageAt  time.Time `json:"lastMessageAt"`
}

type ExternalListConversationsPayload struct {
	Conversations []ExternalConversationSummary `json:"conversations"`
}

type ExternalConversationMessage struct {
	Id             string    `json:"id"`
	ConversationId string    `json:"conversationId"`
	Role           string    `json:"role"`
	Content        string    `json:"content"`
	Timestamp      time.Time `json:"timestamp"`
}

type ExternalListConversationMessagesPayload struct {
	Messages []ExternalConversationMessage `json:"messages"`
}
