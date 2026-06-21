package responses

import (
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	"github.com/google/uuid"
)

type ConversationMessage struct {
	Id             uuid.UUID `json:"id"`
	ConversationId string    `json:"conversationId"`
	Role           string    `json:"role"`
	Content        *string   `json:"content,omitempty"`
	ToolCalls      *string   `json:"toolCalls,omitempty"`
	ToolCallId     *string   `json:"toolCallId,omitempty"`
	Timestamp      time.Time `json:"timestamp"`
}

func (r *ConversationMessage) FromModel(m models.ConversationMessage) {
	r.Id = m.Id
	r.ConversationId = m.ConversationId
	r.Role = m.Role.String()
	r.Content = m.Content
	r.ToolCalls = m.ToolCalls
	r.ToolCallId = m.ToolCallId
	r.Timestamp = m.Timestamp
}

type ListConversationMessagesResponse struct {
	Messages []ConversationMessage `json:"messages"`
}

type AppendConversationMessagesResponse struct {
	Inserted []ConversationMessage `json:"inserted"`
}
