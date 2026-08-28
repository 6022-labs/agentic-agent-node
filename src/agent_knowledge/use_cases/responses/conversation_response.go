package responses

import (
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
)

type ConversationResponse struct {
	ConversationId string    `json:"conversationId"`
	Name           *string   `json:"name,omitempty"`
	MessageCount   int64     `json:"messageCount"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

func (r *ConversationResponse) FromModel(m models.Conversation) {
	r.ConversationId = m.Id
	r.Name = m.Name
	r.CreatedAt = m.CreatedAt
	r.UpdatedAt = m.UpdatedAt
}

type ListConversationsResponse struct {
	Page            uint                   `json:"page"`
	ElementsPerPage uint                   `json:"elementsPerPage"`
	TotalElements   uint                   `json:"totalElements"`
	Conversations   []ConversationResponse `json:"conversations"`
}
