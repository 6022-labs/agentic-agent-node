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

type ConversationMessage struct {
	Id             string    `json:"id"`
	ConversationId string    `json:"conversationId"`
	Role           string    `json:"role"`
	Content        string    `json:"content"`
	Timestamp      time.Time `json:"timestamp"`
}

type ListConversationMessagesResponse struct {
	Messages []ConversationMessage `json:"messages"`
}
