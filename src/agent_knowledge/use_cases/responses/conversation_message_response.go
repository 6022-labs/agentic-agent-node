package responses

import (
	"time"

	"github.com/google/uuid"
)

type ConversationMessage struct {
	Id             uuid.UUID `json:"id"`
	ConversationId string    `json:"conversationId"`
	Role           string    `json:"role"`
	Content        string    `json:"content"`
	Timestamp      time.Time `json:"timestamp"`
}

type ListConversationMessagesResponse struct {
	Messages []ConversationMessage `json:"messages"`
}

type AppendConversationMessagesResponse struct {
	Inserted []ConversationMessage `json:"inserted"`
}
