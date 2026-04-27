package responses

import (
	"time"

	"github.com/google/uuid"
)

type ConversationMessage struct {
	Id             uuid.UUID `json:"id"`
	ConversationId string    `json:"conversationId"`
	Author         string    `json:"author"`
	Role           string    `json:"role"`
	Platform       string    `json:"platform,omitempty"`
	Content        string    `json:"content"`
	Timestamp      time.Time `json:"timestamp"`
}

type ListConversationMessagesResponse struct {
	Messages []ConversationMessage `json:"messages"`
}

type AppendConversationMessagesResponse struct {
	Inserted []ConversationMessage `json:"inserted"`
}
